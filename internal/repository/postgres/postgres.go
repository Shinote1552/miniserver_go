package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"urlshortener/internal/domain/models"
	"urlshortener/internal/repository/dto"

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
	txm TxManager
	que Querier
}

type TxManager interface {
	// Если opts == nil, применяется по умолчанию уровень ReadCommitted
	WithTx(ctx context.Context, opts *sql.TxOptions, fn func(ctx context.Context) error) error
	GetQuerier(ctx context.Context) (Querier, error)
}

// Querier предоставляет единый интерфейс для SQL-запросов,
// автоматически работающий как с транзакциями, так и с обычными соединениями.
// Избавляет от ручных проверок контекста транзакций.
type Querier interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
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
		txm: NewSQLTxManager(db),
	}, nil
}

func initConnectionPools(db *sql.DB) {
	db.SetMaxOpenConns(storageMaxOpenConnections)
	db.SetMaxIdleConns(storageMaxIdleConnections)
	db.SetConnMaxIdleTime(storageConnectionsMaxIdleTime)
	db.SetConnMaxLifetime(storageConnectionsLifetime)
}

// Выполняет функцию в транзакции, если opts == nil, применяется по умолчанию уровень ReadCommitted
func (p *PostgresStorage) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return p.txm.WithTx(ctx, nil, fn)
}

// Возвращает унифицированный интерфейс для выполнения SQL-запросов.
// Автоматически определяет контекст транзакции и возвращает соответствующий Querier:
// - TxQuerier если выполняется транзакция
// - SQLQuerier если транзакции нет
func (p *PostgresStorage) GetQuerier(ctx context.Context) (Querier, error) {
	return p.txm.GetQuerier(ctx)
}

func (p *PostgresStorage) UserCreate(ctx context.Context, user models.User) (models.User, error) {
	userDB := dto.UserDBFromDomain(user)

	querier, err := p.GetQuerier(ctx)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to get querier: %w", err)
	}

	err = querier.QueryRowContext(ctx,
		`INSERT INTO users(created_at) VALUES ($1) RETURNING id`,
		userDB.CreatedAt).Scan(&userDB.ID)

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

	querier, err := p.GetQuerier(ctx)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to get querier: %w", err)
	}

	err = querier.QueryRowContext(ctx,
		`SELECT id, created_at FROM users WHERE id = $1`, id).
		Scan(&userDB.ID, &userDB.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, models.ErrUnfound
		}
		return models.User{}, fmt.Errorf("failed to get user: %w", err)
	}

	return dto.UserDBToDomain(userDB), nil
}

func (p *PostgresStorage) ShortenedLinkGetBatchByUser(ctx context.Context, id int64) ([]models.ShortenedLink, error) {
	if id <= 0 {
		return nil, models.ErrInvalidData
	}

	querier, err := p.GetQuerier(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get querier: %w", err)
	}

	rows, err := querier.QueryContext(ctx,
		`SELECT id, short_key, original_url, user_id, created_at 
         FROM urls 
         WHERE user_id = $1 
         ORDER BY created_at`, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query user links: %w", err)
	}
	defer rows.Close()

	return p.scanShortLinks(ctx, rows)
}

func (p *PostgresStorage) ShortenedLinkCreate(ctx context.Context, url models.ShortenedLink) (models.ShortenedLink, error) {
	querier, err := p.GetQuerier(ctx)
	if err != nil {
		return models.ShortenedLink{}, fmt.Errorf("failed to get querier: %w", err)
	}

	dbURL := dto.ShortenedLinkDBFromDomain(url)
	var dbResult dto.ShortenedLinkDB

	// Пытаемся вставить запись, при конфликте НИЧЕГО не делаем
	err = querier.QueryRowContext(ctx, `
        INSERT INTO urls (short_key, original_url, user_id, created_at)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (original_url) DO NOTHING
        RETURNING id, short_key, original_url, user_id, created_at`,
		dbURL.ShortCode, dbURL.OriginalURL, dbURL.UserID, dbURL.CreatedAt,
	).Scan(&dbResult.ID, &dbResult.ShortCode, &dbResult.OriginalURL, &dbResult.UserID, &dbResult.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		// Конфликт - запись уже существует, возвращаем существующую
		var existing dto.ShortenedLinkDB
		err := querier.QueryRowContext(ctx,
			"SELECT id, short_key, original_url, user_id, created_at FROM urls WHERE original_url = $1",
			url.OriginalURL,
		).Scan(&existing.ID, &existing.ShortCode, &existing.OriginalURL, &existing.UserID, &existing.CreatedAt)

		if err != nil {
			return models.ShortenedLink{}, fmt.Errorf("failed to get existing URL: %w", err)
		}
		return dto.ShortenedLinkDBToDomain(existing), models.ErrConflict
	}

	if err != nil {
		return models.ShortenedLink{}, fmt.Errorf("database error: %w", err)
	}

	return dto.ShortenedLinkDBToDomain(dbResult), nil
}

func (p *PostgresStorage) ShortenedLinkGetByShortKey(ctx context.Context, shortKey string) (models.ShortenedLink, error) {
	if shortKey == "" {
		return models.ShortenedLink{}, models.ErrInvalidData
	}

	var result dto.ShortenedLinkDB

	querier, err := p.GetQuerier(ctx)
	if err != nil {
		return models.ShortenedLink{}, fmt.Errorf("failed to get querier: %w", err)
	}

	err = querier.QueryRowContext(ctx,
		"SELECT id, short_key, original_url, user_id, created_at FROM urls WHERE short_key = $1",
		shortKey,
	).Scan(&result.ID, &result.ShortCode, &result.OriginalURL, &result.UserID, &result.CreatedAt)

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

	querier, err := p.GetQuerier(ctx)
	if err != nil {
		return models.ShortenedLink{}, fmt.Errorf("failed to get querier: %w", err)
	}

	err = querier.QueryRowContext(ctx,
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

func (p *PostgresStorage) ShortenedLinkBatchCreate(ctx context.Context, urls []models.ShortenedLink) ([]models.ShortenedLink, error) {
	var result []models.ShortenedLink

	err := p.txm.WithTx(ctx, nil, func(txCtx context.Context) error {
		querier, err := p.GetQuerier(txCtx)
		if err != nil {
			return fmt.Errorf("failed to get querier: %w", err)
		}

		for _, url := range urls {
			if err := txCtx.Err(); err != nil {
				return fmt.Errorf("operation canceled: %w", err)
			}

			var dbURL dto.ShortenedLinkDB
			err := querier.QueryRowContext(txCtx, `
				INSERT INTO urls (short_key, original_url, user_id, created_at)
				VALUES ($1, $2, $3, $4)
				ON CONFLICT (original_url) DO NOTHING
				RETURNING id, short_key, original_url, user_id, created_at`,
				url.ShortCode, url.OriginalURL, url.UserID, url.CreatedAt,
			).Scan(&dbURL.ID, &dbURL.ShortCode, &dbURL.OriginalURL, &dbURL.UserID, &dbURL.CreatedAt)

			if errors.Is(err, sql.ErrNoRows) {
				var existing dto.ShortenedLinkDB
				err := querier.QueryRowContext(txCtx,
					"SELECT id, short_key, original_url, user_id, created_at FROM urls WHERE original_url = $1",
					url.OriginalURL,
				).Scan(&existing.ID, &existing.ShortCode, &existing.OriginalURL, &existing.UserID, &existing.CreatedAt)

				if err != nil {
					return fmt.Errorf("failed to get existing URL: %w", err)
				}
				result = append(result, dto.ShortenedLinkDBToDomain(existing))
				continue
			}

			if err != nil {
				return fmt.Errorf("failed to insert URL: %w", err)
			}

			result = append(result, dto.ShortenedLinkDBToDomain(dbURL))
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// List остается без изменений
func (p *PostgresStorage) List(ctx context.Context, limit, offset int) ([]models.ShortenedLink, error) {
	if limit <= 0 || offset < 0 {
		return nil, models.ErrInvalidData
	}

	querier, err := p.GetQuerier(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get querier: %w", err)
	}

	rows, err := querier.QueryContext(ctx,
		"SELECT id, short_key, original_url, user_id, created_at FROM urls ORDER BY created_at DESC LIMIT $1 OFFSET $2",
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query URLs: %w", err)
	}
	defer rows.Close()

	return p.scanShortLinks(ctx, rows)
}

func (p *PostgresStorage) Exists(ctx context.Context, originalURL string) (models.ShortenedLink, error) {
	var dbURL dto.ShortenedLinkDB

	querier, err := p.GetQuerier(ctx)
	if err != nil {
		return models.ShortenedLink{}, fmt.Errorf("failed to get querier: %w", err)
	}

	err = querier.QueryRowContext(ctx,
		"SELECT id, short_key, original_url, user_id, created_at FROM urls WHERE original_url = $1",
		originalURL,
	).Scan(&dbURL.ID, &dbURL.ShortCode, &dbURL.OriginalURL, &dbURL.UserID, &dbURL.CreatedAt)

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

	querier, err := p.GetQuerier(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get querier: %w", err)
	}

	rows, err := querier.QueryContext(ctx,
		"SELECT id, short_key, original_url, user_id, created_at FROM urls WHERE original_url = ANY($1)",
		originalURLs,
	)
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

// scanShortLinks - утилитарный метод для сканирования результатов запроса URL
func (p *PostgresStorage) scanShortLinks(ctx context.Context, rows *sql.Rows) ([]models.ShortenedLink, error) {
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

func (p *PostgresStorage) ShortenedLinkBatchDelete(ctx context.Context, userID int64, shortCode []string) error {
	if userID <= 0 || len(shortCode) == 0 {
		return models.ErrInvalidData
	}
	querier, err := p.GetQuerier(ctx)

	if err != nil {
		return fmt.Errorf("failed to get querier: %w", err)
	}

	// в случае проблем со слайсом ANY($3::text[])

	_, err = querier.ExecContext(ctx, `
	UPDATE urls
	SET is_deleted = true, deleted_at = $1
	WHERE user_id = $2 AND short_key = ANY($3) AND is_deleted = false`,
		time.Now().UTC(), userID, shortCode)

	if err != nil {
		return fmt.Errorf("failed to batch delete URLs: %w", err)
	}

	// rowsAffected, err := result.RowsAffected()
	// if err != nil {
	// 	return fmt.Errorf("failed to get rows affected: %w", err)
	// }

	// p.log.Debug().Int64("rows_affected", rowsAffected).Msg("URLs soft deleted")

	return nil
}
