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
			id SERIAL PRIMARY KEY,
			short_key VARCHAR(10) UNIQUE NOT NULL,
			original_url TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			UNIQUE (original_url)
		)`)
	return err
}

func (p *PostgresStorage) CreateOrUpdate(ctx context.Context, url models.ShortenedLink) (models.ShortenedLink, error) {
	if url.ShortCode == "" || url.LongURL == "" {
		return models.ShortenedLink{}, models.ErrInvalidData
	}

	dbURL := dto.FromDomain(url)
	var result dto.URLDB
	err := p.db.QueryRowContext(ctx, `
        INSERT INTO urls (short_key, original_url, created_at)
        VALUES ($1, $2, $3)
        ON CONFLICT (original_url) DO NOTHING
        RETURNING id, short_key, original_url, created_at`,
		dbURL.ShortCode, dbURL.LongURL, dbURL.CreatedAt,
	).Scan(&result.ID, &result.ShortCode, &result.LongURL, &result.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			existing, err := p.GetByLongURL(ctx, url.LongURL)
			if err != nil {
				return models.ShortenedLink{}, fmt.Errorf("%w: %v", models.ErrConflict, err)
			}
			return existing, models.ErrConflict
		}
		return models.ShortenedLink{}, fmt.Errorf("database error: %w", err)
	}

	return *result.ToDomain(), nil
}

func (p *PostgresStorage) GetByShortKey(ctx context.Context, shortKey string) (models.ShortenedLink, error) {
	if shortKey == "" {
		return models.ShortenedLink{}, models.ErrInvalidData
	}

	var result dto.URLDB
	err := p.db.QueryRowContext(ctx,
		"SELECT id, short_key, original_url, created_at FROM urls WHERE short_key = $1",
		shortKey,
	).Scan(&result.ID, &result.ShortCode, &result.LongURL, &result.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.ShortenedLink{}, models.ErrUnfound
		}
		return models.ShortenedLink{}, fmt.Errorf("failed to get URL: %w", err)
	}

	return *result.ToDomain(), nil
}

func (p *PostgresStorage) GetByLongURL(ctx context.Context, originalURL string) (models.ShortenedLink, error) {
	if originalURL == "" {
		return models.ShortenedLink{}, models.ErrInvalidData
	}

	var result dto.URLDB
	err := p.db.QueryRowContext(ctx,
		"SELECT id, short_key, original_url, created_at FROM urls WHERE original_url = $1",
		originalURL,
	).Scan(&result.ID, &result.ShortCode, &result.LongURL, &result.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.ShortenedLink{}, models.ErrUnfound
		}
		return models.ShortenedLink{}, fmt.Errorf("failed to get URL: %w", err)
	}

	return *result.ToDomain(), nil
}

func (p *PostgresStorage) BatchCreate(ctx context.Context, urls []models.ShortenedLink) ([]models.ShortenedLink, error) {
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
		var dbURL dto.URLDB
		err := stmt.QueryRowContext(ctx, url.ShortCode, url.LongURL, url.CreatedAt).Scan(
			&dbURL.ID, &dbURL.ShortCode, &dbURL.LongURL, &dbURL.CreatedAt,
		)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				existing, err := p.GetByLongURL(ctx, url.LongURL)
				if err != nil {
					return nil, fmt.Errorf("failed to get existing URL: %w", err)
				}
				result = append(result, existing)
				continue
			}
			return nil, fmt.Errorf("failed to insert URL: %w", err)
		}

		result = append(result, *dbURL.ToDomain())
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
		var dbURL dto.URLDB
		if err := rows.Scan(&dbURL.ID, &dbURL.ShortCode, &dbURL.LongURL, &dbURL.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan URL: %w", err)
		}
		urls = append(urls, *dbURL.ToDomain())
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return urls, nil
}

func (p *PostgresStorage) Exists(ctx context.Context, originalURL string) (models.ShortenedLink, error) {
	var dbURL dto.URLDB
	err := p.db.QueryRowContext(ctx,
		"SELECT id, short_key, original_url, created_at FROM urls WHERE original_url = $1",
		originalURL,
	).Scan(&dbURL.ID, &dbURL.ShortCode, &dbURL.LongURL, &dbURL.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.ShortenedLink{}, nil
		}
		return models.ShortenedLink{}, fmt.Errorf("failed to check URL existence: %w", err)
	}

	return *dbURL.ToDomain(), nil
}

func (p *PostgresStorage) ExistsBatch(ctx context.Context, originalURLs []string) ([]models.ShortenedLink, error) {
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
		var dbURL dto.URLDB
		if err := rows.Scan(&dbURL.ID, &dbURL.ShortCode, &dbURL.LongURL, &dbURL.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan URL row: %w", err)
		}
		result = append(result, *dbURL.ToDomain())
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
		var dbURL dto.URLDB
		if err := rows.Scan(&dbURL.ID, &dbURL.ShortCode, &dbURL.LongURL, &dbURL.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan URL: %w", err)
		}
		urls = append(urls, *dbURL.ToDomain())
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return urls, nil
}
