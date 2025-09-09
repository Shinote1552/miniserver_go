package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"urlshortener/internal/domain/models"
	"urlshortener/internal/repository/dto"
	"urlshortener/internal/repository/postgres/txmanager"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	storageMaxOpenConnections     = 5
	storageMaxIdleConnections     = 2
	storageConnectionsMaxIdleTime = 2 * time.Minute
	storageConnectionsLifetime    = 30 * time.Minute
	storagePingTimeout            = 5 * time.Second
)

type PostgresStorage struct {
	db  *sql.DB
	txm txmanager.TxManager
}

func NewStorage(ctx context.Context, dsn string) (*PostgresStorage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	initConnectionPools(db)

	ctxPing, cancel := context.WithTimeout(ctx, storagePingTimeout)
	defer cancel()

	if err := db.PingContext(ctxPing); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresStorage{
		db:  db,
		txm: txmanager.NewSQLTxManager(db),
	}, nil
}

func initConnectionPools(db *sql.DB) {
	db.SetMaxOpenConns(storageMaxOpenConnections)
	db.SetMaxIdleConns(storageMaxIdleConnections)
	db.SetConnMaxIdleTime(storageConnectionsMaxIdleTime)
	db.SetConnMaxLifetime(storageConnectionsLifetime)
}

func (p *PostgresStorage) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return p.txm.WithinTx(ctx, fn)
}

func (p *PostgresStorage) UserCreate(ctx context.Context, user models.User) (models.User, error) {
	userDB := dto.UserDBFromDomain(user)

	tx, err := txmanager.GetTx(ctx)
	if err != nil {
		if !errors.Is(err, txmanager.ErrNoTransaction) {
			return models.User{}, fmt.Errorf("failed to get transaction: %w", err)
		}
		err = p.db.QueryRowContext(ctx,
			`INSERT INTO users(created_at) VALUES ($1) RETURNING id`,
			userDB.CreatedAt).Scan(&userDB.ID)
	} else {
		err = tx.QueryRowContext(ctx,
			`INSERT INTO users(created_at) VALUES ($1) RETURNING id`,
			userDB.CreatedAt).Scan(&userDB.ID)
	}

	if err != nil {
		return models.User{}, fmt.Errorf("failed to create user: %w", err)
	}

	return dto.UserDBToDomain(userDB), nil
}

func (p *PostgresStorage) UserGetByID(ctx context.Context, id int64) (models.User, error) {
	if err := ctx.Err(); err != nil {
		return models.User{}, models.ErrInvalidData
	}

	if id <= 0 {
		return models.User{}, models.ErrInvalidData
	}

	var userDB dto.UserDB

	tx, err := txmanager.GetTx(ctx)
	if err != nil {
		err = p.db.QueryRowContext(ctx,
			`SELECT id, created_at
			FROM users
			WHERE id = $1`, id).Scan(&userDB.ID, &userDB.CreatedAt)
	} else {
		err = tx.QueryRowContext(ctx,
			`SELECT id, created_at
			FROM users
			WHERE id = $1`, id).Scan(&userDB.ID, &userDB.CreatedAt)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, models.ErrUnfound
		}
		return models.User{}, fmt.Errorf("failed to get user: %w", err)
	}

	resultUser := dto.UserDBToDomain(userDB)
	return resultUser, nil
}

func (p *PostgresStorage) ShortenedLinkGetBatchByUser(ctx context.Context, id int64) ([]models.ShortenedLink, error) {
	if id <= 0 {
		return nil, models.ErrInvalidData
	}

	var rows *sql.Rows
	var err error

	tx, txErr := txmanager.GetTx(ctx)
	if txErr != nil {
		rows, err = p.db.QueryContext(ctx,
			`SELECT id, short_key, original_url, user_id, created_at 
			FROM urls 
			WHERE user_id = $1 
			ORDER BY created_at`, id)
	} else {
		rows, err = tx.QueryContext(ctx,
			`SELECT id, short_key, original_url, user_id, created_at 
			FROM urls 
			WHERE user_id = $1 
			ORDER BY created_at`, id)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query user links: %w", err)
	}
	defer rows.Close()

	var shortLinks []models.ShortenedLink

	for rows.Next() {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("operation canceled: %w", err)
		}

		var linkDB dto.ShortenedLinkDB
		if err := rows.Scan(
			&linkDB.ID,
			&linkDB.ShortCode,
			&linkDB.OriginalURL,
			&linkDB.UserID,
			&linkDB.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan link: %w", err)
		}
		shortLinks = append(shortLinks, dto.ShortenedLinkDBToDomain(linkDB))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return shortLinks, nil
}

func (p *PostgresStorage) ShortenedLinkCreate(ctx context.Context, url models.ShortenedLink) (models.ShortenedLink, error) {
	var result models.ShortenedLink
	err := p.WithinTx(ctx, func(txCtx context.Context) error {
		tx, err := txmanager.GetTx(txCtx)
		if err != nil {
			return fmt.Errorf("failed to get transaction: %w", err)
		}

		dbURL := dto.ShortenedLinkDBFromDomain(url)
		var dbResult dto.ShortenedLinkDB

		err = tx.QueryRowContext(txCtx, `
            INSERT INTO urls (short_key, original_url, user_id, created_at)
            VALUES ($1, $2, $3, $4)
            ON CONFLICT (original_url) DO NOTHING
            RETURNING id, short_key, original_url, user_id, created_at`,
			dbURL.ShortCode, dbURL.OriginalURL, dbURL.UserID, dbURL.CreatedAt,
		).Scan(&dbResult.ID, &dbResult.ShortCode, &dbResult.OriginalURL, &dbResult.UserID, &dbResult.CreatedAt)

		if errors.Is(err, sql.ErrNoRows) {
			var existing dto.ShortenedLinkDB
			err := tx.QueryRowContext(txCtx,
				"SELECT id, short_key, original_url, user_id, created_at FROM urls WHERE original_url = $1",
				url.OriginalURL,
			).Scan(&existing.ID, &existing.ShortCode, &existing.OriginalURL, &existing.UserID, &existing.CreatedAt)

			if err != nil {
				return fmt.Errorf("failed to get existing URL: %w", err)
			}
			result = dto.ShortenedLinkDBToDomain(existing)
			return models.ErrConflict
		}

		if err != nil {
			return fmt.Errorf("database error: %w", err)
		}

		result = dto.ShortenedLinkDBToDomain(dbResult)
		return nil
	})

	if err != nil && !errors.Is(err, models.ErrConflict) {
		return models.ShortenedLink{}, err
	}

	return result, err
}

func (p *PostgresStorage) ShortenedLinkGetByShortKey(ctx context.Context, shortKey string) (models.ShortenedLink, error) {
	if shortKey == "" {
		return models.ShortenedLink{}, models.ErrInvalidData
	}

	var result dto.ShortenedLinkDB

	tx, err := txmanager.GetTx(ctx)
	if err != nil {
		err = p.db.QueryRowContext(ctx,
			"SELECT id, short_key, original_url, user_id, created_at FROM urls WHERE short_key = $1",
			shortKey,
		).Scan(&result.ID, &result.ShortCode, &result.OriginalURL, &result.UserID, &result.CreatedAt)
	} else {
		err = tx.QueryRowContext(ctx,
			"SELECT id, short_key, original_url, user_id, created_at FROM urls WHERE short_key = $1",
			shortKey,
		).Scan(&result.ID, &result.ShortCode, &result.OriginalURL, &result.UserID, &result.CreatedAt)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.ShortenedLink{}, models.ErrUnfound
		}
		return models.ShortenedLink{}, fmt.Errorf("failed to get URL: %w", err)
	}

	return dto.ShortenedLinkDBToDomain(result), nil
}

func (p *PostgresStorage) ShortenedLinkGetByOriginalURL(ctx context.Context, originalURL string) (models.ShortenedLink, error) {
	if originalURL == "" {
		return models.ShortenedLink{}, models.ErrInvalidData
	}

	var result dto.ShortenedLinkDB

	tx, err := txmanager.GetTx(ctx)
	if err != nil {
		err = p.db.QueryRowContext(ctx,
			"SELECT id, short_key, original_url, user_id, created_at FROM urls WHERE original_url = $1",
			originalURL,
		).Scan(&result.ID, &result.ShortCode, &result.OriginalURL, &result.UserID, &result.CreatedAt)
	} else {
		err = tx.QueryRowContext(ctx,
			"SELECT id, short_key, original_url, user_id, created_at FROM urls WHERE original_url = $1",
			originalURL,
		).Scan(&result.ID, &result.ShortCode, &result.OriginalURL, &result.UserID, &result.CreatedAt)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.ShortenedLink{}, models.ErrUnfound
		}
		return models.ShortenedLink{}, fmt.Errorf("failed to get URL: %w", err)
	}

	return dto.ShortenedLinkDBToDomain(result), nil
}

func (p *PostgresStorage) ShortenedLinkBatchCreate(ctx context.Context, urls []models.ShortenedLink) ([]models.ShortenedLink, error) {
	var result []models.ShortenedLink

	if _, err := txmanager.GetTx(ctx); err != nil {
		err := p.WithinTx(ctx, func(txCtx context.Context) error {
			var innerErr error
			result, innerErr = p.createBatchInTx(txCtx, urls)
			return innerErr
		})
		return result, err
	}

	return p.createBatchInTx(ctx, urls)
}

func (p *PostgresStorage) createBatchInTx(ctx context.Context, urls []models.ShortenedLink) ([]models.ShortenedLink, error) {
	tx, err := txmanager.GetTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("no transaction in context")
	}

	var result []models.ShortenedLink
	for _, url := range urls {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("operation canceled: %w", err)
		}

		var dbURL dto.ShortenedLinkDB
		err := tx.QueryRowContext(ctx, `
			INSERT INTO urls (short_key, original_url, user_id, created_at)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (original_url) DO NOTHING
			RETURNING id, short_key, original_url, user_id, created_at`,
			url.ShortCode, url.OriginalURL, url.UserID, url.CreatedAt,
		).Scan(&dbURL.ID, &dbURL.ShortCode, &dbURL.OriginalURL, &dbURL.UserID, &dbURL.CreatedAt)

		if errors.Is(err, sql.ErrNoRows) {
			existing, err := p.getByOriginalURLInTx(ctx, url.OriginalURL)
			if err != nil {
				return nil, err
			}
			result = append(result, existing)
			continue
		}

		if err != nil {
			return nil, fmt.Errorf("failed to insert URL: %w", err)
		}

		result = append(result, dto.ShortenedLinkDBToDomain(dbURL))
	}

	return result, nil
}

func (p *PostgresStorage) getByOriginalURLInTx(ctx context.Context, originalURL string) (models.ShortenedLink, error) {
	tx, err := txmanager.GetTx(ctx)
	if err != nil {
		return models.ShortenedLink{}, fmt.Errorf("no transaction in context")
	}

	var result dto.ShortenedLinkDB
	err = tx.QueryRowContext(ctx,
		"SELECT id, short_key, original_url, user_id, created_at FROM urls WHERE original_url = $1",
		originalURL,
	).Scan(&result.ID, &result.ShortCode, &result.OriginalURL, &result.UserID, &result.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.ShortenedLink{}, models.ErrUnfound
		}
		return models.ShortenedLink{}, fmt.Errorf("failed to get URL: %w", err)
	}

	return dto.ShortenedLinkDBToDomain(result), nil
}

func (p *PostgresStorage) Delete(ctx context.Context, shortKey string) error {
	if shortKey == "" {
		return models.ErrInvalidData
	}

	tx, err := txmanager.GetTx(ctx)
	if err != nil {
		result, err := p.db.ExecContext(ctx,
			"DELETE FROM urls WHERE short_key = $1",
			shortKey,
		)
		if err != nil {
			return fmt.Errorf("failed to delete URL: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return models.ErrUnfound
		}
		return nil
	}

	result, err := tx.ExecContext(ctx,
		"DELETE FROM urls WHERE short_key = $1",
		shortKey,
	)
	if err != nil {
		return fmt.Errorf("failed to delete URL: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.ErrUnfound
	}

	return nil
}

func (p *PostgresStorage) List(ctx context.Context, limit, offset int) ([]models.ShortenedLink, error) {
	if limit <= 0 || offset < 0 {
		return nil, models.ErrInvalidData
	}

	var rows *sql.Rows
	var err error

	tx, txErr := txmanager.GetTx(ctx)
	if txErr != nil {
		rows, err = p.db.QueryContext(ctx,
			"SELECT id, short_key, original_url, user_id, created_at FROM urls ORDER BY created_at DESC LIMIT $1 OFFSET $2",
			limit, offset,
		)
	} else {
		rows, err = tx.QueryContext(ctx,
			"SELECT id, short_key, original_url, user_id, created_at FROM urls ORDER BY created_at DESC LIMIT $1 OFFSET $2",
			limit, offset,
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query URLs: %w", err)
	}
	defer rows.Close()

	var urls []models.ShortenedLink
	for rows.Next() {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("operation canceled: %w", err)
		}

		var dbURL dto.ShortenedLinkDB
		if err := rows.Scan(&dbURL.ID, &dbURL.ShortCode, &dbURL.OriginalURL, &dbURL.UserID, &dbURL.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan URL: %w", err)
		}
		urls = append(urls, dto.ShortenedLinkDBToDomain(dbURL))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return urls, nil
}

func (p *PostgresStorage) Exists(ctx context.Context, originalURL string) (models.ShortenedLink, error) {
	var dbURL dto.ShortenedLinkDB

	tx, err := txmanager.GetTx(ctx)
	if err != nil {
		err = p.db.QueryRowContext(ctx,
			"SELECT id, short_key, original_url, user_id, created_at FROM urls WHERE original_url = $1",
			originalURL,
		).Scan(&dbURL.ID, &dbURL.ShortCode, &dbURL.OriginalURL, &dbURL.UserID, &dbURL.CreatedAt)
	} else {
		err = tx.QueryRowContext(ctx,
			"SELECT id, short_key, original_url, user_id, created_at FROM urls WHERE original_url = $1",
			originalURL,
		).Scan(&dbURL.ID, &dbURL.ShortCode, &dbURL.OriginalURL, &dbURL.UserID, &dbURL.CreatedAt)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.ShortenedLink{}, nil
		}
		return models.ShortenedLink{}, fmt.Errorf("failed to check URL existence: %w", err)
	}

	return dto.ShortenedLinkDBToDomain(dbURL), nil
}

func (p *PostgresStorage) ShortenedLinkBatchExists(ctx context.Context, originalURLs []string) ([]models.ShortenedLink, error) {
	if len(originalURLs) == 0 {
		return nil, models.ErrInvalidData
	}

	var rows *sql.Rows
	var err error

	tx, txErr := txmanager.GetTx(ctx)
	if txErr != nil {
		rows, err = p.db.QueryContext(ctx,
			"SELECT id, short_key, original_url, user_id, created_at FROM urls WHERE original_url = ANY($1)",
			originalURLs,
		)
	} else {
		rows, err = tx.QueryContext(ctx,
			"SELECT id, short_key, original_url, user_id, created_at FROM urls WHERE original_url = ANY($1)",
			originalURLs,
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query existing URLs: %w", err)
	}
	defer rows.Close()

	var result []models.ShortenedLink
	for rows.Next() {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("operation canceled: %w", err)
		}

		var dbURL dto.ShortenedLinkDB
		if err := rows.Scan(&dbURL.ID, &dbURL.ShortCode, &dbURL.OriginalURL, &dbURL.UserID, &dbURL.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan URL row: %w", err)
		}
		result = append(result, dto.ShortenedLinkDBToDomain(dbURL))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return result, nil
}

func (p *PostgresStorage) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, storagePingTimeout)
	defer cancel()

	return p.db.PingContext(ctx)
}

func (p *PostgresStorage) Close() error {
	return p.db.Close()
}
