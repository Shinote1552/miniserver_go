package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"urlshortener/domain/models"
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

const (
	pgErrCodeUniqueViolation = "23505"
)

type PostgresStorage struct {
	db *sql.DB
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

	if err := createTable(ctx, db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &PostgresStorage{db: db}, nil
}

func initConnectionPools(db *sql.DB) {
	db.SetMaxOpenConns(storageMaxOpenConnections)
	db.SetMaxIdleConns(storageMaxIdleConnections)
	db.SetConnMaxIdleTime(storageConnectionsMaxIdleTime)
	db.SetConnMaxLifetime(storageConnectionsLifetime)
}

func createTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS urls (
            id BIGSERIAL PRIMARY KEY,
            short_key VARCHAR(10) UNIQUE NOT NULL,
            original_url TEXT NOT NULL,
            user_id BIGINT NOT NULL,
            created_at TIMESTAMP NOT NULL,
            is_deleted BOOLEAN DEFAULT FALSE,  // Добавляем колонку
            UNIQUE (original_url)
        )`)
	if err != nil {
		return fmt.Errorf("failed to create urls table: %w", err)
	}
	return nil
}

// internal/repository/postgres/storage.go
func (p *PostgresStorage) DeleteURLsBatch(ctx context.Context, userID int64, shortURLs []string) error {
	if len(shortURLs) == 0 {
		return nil
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		UPDATE urls 
		SET is_deleted = TRUE 
		WHERE short_key = $1 AND user_id = $2`)
	if err != nil {
		return fmt.Errorf("ошибка подготовки запроса: %w", err)
	}
	defer stmt.Close()

	for _, shortKey := range shortURLs {
		if _, err := stmt.ExecContext(ctx, shortKey, userID); err != nil {
			return fmt.Errorf("ошибка удаления URL%s:: %w", shortKey, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("ошибка коммита транзакции: %w", err)
	}

	return nil
}

func (p *PostgresStorage) UserCreate(ctx context.Context, user models.User) (models.User, error) {

	userDB := dto.UserDBFromDomain(user)

	err := p.db.QueryRowContext(ctx,
		`INSERT INTO users(created_at)
		VALUES ($1)
		RETURNING id`, userDB.CreatedAt).Scan(&userDB.ID)

	if err != nil {
		return models.User{}, fmt.Errorf("failed to create user: %w", err)
	}

	resultUser := dto.UserDBToDomain(userDB)
	return resultUser, nil
}

func (p *PostgresStorage) UserGetByID(ctx context.Context, id int64) (models.User, error) {
	if id <= 0 {
		return models.User{}, models.ErrInvalidData
	}
	var userDB dto.UserDB

	err := p.db.QueryRowContext(ctx,
		`SELECT id, created_at
		FROM users
		WHERE id = $1`, id).Scan(&userDB)

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

	user := dto.UserDBFromDomain(models.User{
		ID:        id,
		CreatedAt: time.Now().UTC(),
	})

	rows, err := p.db.QueryContext(ctx,
		`SELECT id, short_key, original_url, user_id, created_at 
		FROM urls 
		WHERE user_id = $1`, user.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to query user links: %w", err)
	}

	defer rows.Close()
	/*
	   надо ли range перекопировать dto->domen еще раз чтобы внутри цикла
	   были чистые рачсты на dto или как у меня сразу же append в слайс доменной модели?
	*/
	var shortLinks []models.ShortenedLink

	for rows.Next() {
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
	if url.ShortCode == "" || url.OriginalURL == "" {
		return models.ShortenedLink{}, models.ErrInvalidData
	}
	dbURL := dto.ShortenedLinkDBFromDomain(url)
	var result dto.ShortenedLinkDB
	err := p.db.QueryRowContext(ctx, `
        INSERT INTO urls (short_key, original_url, user_id, created_at)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (original_url) DO NOTHING
        RETURNING id, short_key, original_url, user_id, created_at`,
		dbURL.ShortCode, dbURL.OriginalURL, dbURL.CreatedAt,
	).Scan(&result.ID, &result.ShortCode, &result.OriginalURL, &result.UserID, &result.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			existing, err := p.ShortenedLinkGetByOriginalURL(ctx, url.OriginalURL)
			if err != nil {
				return models.ShortenedLink{}, fmt.Errorf("%w: %v", models.ErrConflict, err)
			}
			return existing, models.ErrConflict
		}
		return models.ShortenedLink{}, fmt.Errorf("database error: %w", err)
	}

	return dto.ShortenedLinkDBToDomain(result), nil
}

func (p *PostgresStorage) ShortenedLinkGetByShortKey(ctx context.Context, shortKey string) (models.ShortenedLink, error) {
	if shortKey == "" {
		return models.ShortenedLink{}, models.ErrInvalidData
	}

	var result dto.ShortenedLinkDB
	err := p.db.QueryRowContext(ctx,
		`SELECT id, short_key, original_url, user_id, created_at, is_deleted 
		 FROM urls WHERE short_key = $1`,
		shortKey,
	).Scan(&result.ID, &result.ShortCode, &result.OriginalURL, &result.UserID, &result.CreatedAt, &result.IsDeleted)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.ShortenedLink{}, models.ErrUnfound
		}
		return models.ShortenedLink{}, fmt.Errorf("ошибка получения URL: %w", err)
	}

	return dto.ShortenedLinkDBToDomain(result), nil
}

func (p *PostgresStorage) ShortenedLinkGetByOriginalURL(ctx context.Context, originalURL string) (models.ShortenedLink, error) {
	if originalURL == "" {
		return models.ShortenedLink{}, models.ErrInvalidData
	}

	var result dto.ShortenedLinkDB
	err := p.db.QueryRowContext(ctx,
		"SELECT id, short_key, original_url, created_at FROM urls WHERE original_url = $1",
		originalURL,
	).Scan(&result.ID, &result.ShortCode, &result.OriginalURL, &result.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.ShortenedLink{}, models.ErrUnfound
		}
		return models.ShortenedLink{}, fmt.Errorf("failed to get URL: %w", err)
	}

	return dto.ShortenedLinkDBToDomain(result), nil
}

func (p *PostgresStorage) ShortenedLinkBatchCreate(ctx context.Context, urls []models.ShortenedLink) ([]models.ShortenedLink, error) {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO urls (short_key, original_url, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (original_url) DO NOTHING
		RETURNING id, short_key, original_url, created_at`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	var result []models.ShortenedLink
	for _, url := range urls {
		var dbURL dto.ShortenedLinkDB
		err := stmt.QueryRowContext(ctx, url.ShortCode, url.OriginalURL, url.CreatedAt).Scan(
			&dbURL.ID, &dbURL.ShortCode, &dbURL.OriginalURL, &dbURL.CreatedAt,
		)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				existing, err := p.ShortenedLinkGetByOriginalURL(ctx, url.OriginalURL)
				if err != nil {
					return nil, fmt.Errorf("failed to get existing URL: %w", err)
				}
				result = append(result, existing)
				continue
			}
			return nil, fmt.Errorf("failed to insert URL: %w", err)
		}

		result = append(result, dto.ShortenedLinkDBToDomain(dbURL))
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}

func (p *PostgresStorage) Delete(ctx context.Context, shortKey string) error {
	if shortKey == "" {
		return models.ErrInvalidData
	}

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

func (p *PostgresStorage) List(ctx context.Context, limit, offset int) ([]models.ShortenedLink, error) {
	if limit <= 0 || offset < 0 {
		return nil, models.ErrInvalidData
	}

	rows, err := p.db.QueryContext(ctx,
		"SELECT id, short_key, original_url, created_at FROM urls LIMIT $1 OFFSET $2",
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query URLs: %w", err)
	}
	defer rows.Close()

	var urls []models.ShortenedLink
	for rows.Next() {
		var dbURL dto.ShortenedLinkDB
		if err := rows.Scan(&dbURL.ID, &dbURL.ShortCode, &dbURL.OriginalURL, &dbURL.CreatedAt); err != nil {
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
	err := p.db.QueryRowContext(ctx,
		"SELECT id, short_key, original_url, created_at FROM urls WHERE original_url = $1",
		originalURL,
	).Scan(&dbURL.ID, &dbURL.ShortCode, &dbURL.OriginalURL, &dbURL.CreatedAt)

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

	rows, err := p.db.QueryContext(ctx,
		"SELECT id, short_key, original_url, created_at FROM urls WHERE original_url = ANY($1)",
		originalURLs,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query existing URLs: %w", err)
	}
	defer rows.Close()

	var result []models.ShortenedLink
	for rows.Next() {
		var dbURL dto.ShortenedLinkDB
		if err := rows.Scan(&dbURL.ID, &dbURL.ShortCode, &dbURL.OriginalURL, &dbURL.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan URL row: %w", err)
		}
		result = append(result, dto.ShortenedLinkDBToDomain(dbURL))
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

func (p *PostgresStorage) GetAll(ctx context.Context) ([]models.ShortenedLink, error) {
	rows, err := p.db.QueryContext(ctx,
		"SELECT id, short_key, original_url, created_at FROM urls")
	if err != nil {
		return nil, fmt.Errorf("failed to query URLs: %w", err)
	}
	defer rows.Close()

	var urls []models.ShortenedLink
	for rows.Next() {
		var dbURL dto.ShortenedLinkDB
		if err := rows.Scan(&dbURL.ID, &dbURL.ShortCode, &dbURL.OriginalURL, &dbURL.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan URL: %w", err)
		}
		urls = append(urls, dto.ShortenedLinkDBToDomain(dbURL))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return urls, nil
}
