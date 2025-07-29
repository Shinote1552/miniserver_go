package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"urlshortener/internal/models"

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
			short_url VARCHAR(10) UNIQUE NOT NULL,
			original_url TEXT NOT NULL,
			UNIQUE (original_url)
		)`)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	return nil
}

func (p *PostgresStorage) CreateOrUpdate(ctx context.Context, shortURL, originalURL string) (*models.StorageURLModel, error) {
	if shortURL == "" || originalURL == "" {
		return nil, models.ErrInvalidData
	}

	var url models.StorageURLModel
	err := p.db.QueryRowContext(ctx, `
        INSERT INTO urls (short_url, original_url)
        VALUES ($1, $2)
        ON CONFLICT (original_url) DO NOTHING
        RETURNING id, short_url, original_url`,
		shortURL, originalURL,
	).Scan(&url.ID, &url.ShortURL, &url.OriginalURL)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			existing, err := p.GetByOriginalURL(ctx, originalURL)
			if err != nil {
				return nil, fmt.Errorf("%w: %v", models.ErrConflict, err)
			}
			return existing, models.ErrConflict
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &url, nil
}

func (p *PostgresStorage) GetByShortURL(ctx context.Context, shortURL string) (*models.StorageURLModel, error) {
	if shortURL == "" {
		return nil, fmt.Errorf("%w: shortURL must not be empty", models.ErrInvalidData)
	}

	var url models.StorageURLModel
	err := p.db.QueryRowContext(ctx,
		"SELECT id, short_url, original_url FROM urls WHERE short_url = $1",
		shortURL,
	).Scan(&url.ID, &url.ShortURL, &url.OriginalURL)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: shortURL not found", models.ErrUnfound)
		}
		return nil, fmt.Errorf("failed to get URL: %w", err)
	}

	return &url, nil
}

func (p *PostgresStorage) GetByOriginalURL(ctx context.Context, originalURL string) (*models.StorageURLModel, error) {
	if originalURL == "" {
		return nil, fmt.Errorf("%w: originalURL must not be empty", models.ErrInvalidData)
	}

	var url models.StorageURLModel
	err := p.db.QueryRowContext(ctx,
		"SELECT id, short_url, original_url FROM urls WHERE original_url = $1",
		originalURL,
	).Scan(&url.ID, &url.ShortURL, &url.OriginalURL)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: originalURL not found", models.ErrUnfound)
		}
		return nil, fmt.Errorf("failed to get URL: %w", err)
	}

	return &url, nil
}

func (p *PostgresStorage) BatchCreate(
	ctx context.Context,
	batchItems []models.APIShortenRequestBatch,
) ([]models.APIShortenResponseBatch, error) {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO urls (short_url, original_url)
		VALUES ($1, $2)
		ON CONFLICT (original_url) DO NOTHING
		RETURNING short_url, original_url`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	var result []models.APIShortenResponseBatch
	for _, item := range batchItems {
		var shortURL, originalURL string
		err := stmt.QueryRowContext(ctx, item.CorrelationID, item.OriginalURL).Scan(
			&shortURL, &originalURL,
		)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// URL уже существует, получаем существующую запись
				existing, err := p.GetByOriginalURL(ctx, item.OriginalURL)
				if err != nil {
					return nil, fmt.Errorf("failed to get existing URL: %w", err)
				}
				result = append(result, models.APIShortenResponseBatch{
					CorrelationID: item.CorrelationID,
					ShortURL:      existing.ShortURL,
				})
				continue
			}
			return nil, fmt.Errorf("failed to insert URL: %w", err)
		}

		result = append(result, models.APIShortenResponseBatch{
			CorrelationID: item.CorrelationID,
			ShortURL:      shortURL,
		})
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}

func (p *PostgresStorage) Delete(ctx context.Context, shortURL string) error {
	if shortURL == "" {
		return fmt.Errorf("%w: shortURL must not be empty", models.ErrInvalidData)
	}

	result, err := p.db.ExecContext(ctx,
		"DELETE FROM urls WHERE short_url = $1",
		shortURL,
	)
	if err != nil {
		return fmt.Errorf("failed to delete URL: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%w: URL not found", models.ErrUnfound)
	}

	return nil
}

func (p *PostgresStorage) List(ctx context.Context, limit, offset int) ([]models.StorageURLModel, error) {
	if limit <= 0 || offset < 0 {
		return nil, fmt.Errorf("%w: invalid pagination params", models.ErrInvalidData)
	}

	rows, err := p.db.QueryContext(ctx,
		"SELECT id, short_url, original_url FROM urls LIMIT $1 OFFSET $2",
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query URLs: %w", err)
	}
	defer rows.Close()

	var urls []models.StorageURLModel
	for rows.Next() {
		var url models.StorageURLModel
		if err := rows.Scan(&url.ID, &url.ShortURL, &url.OriginalURL); err != nil {
			return nil, fmt.Errorf("failed to scan URL: %w", err)
		}
		urls = append(urls, url)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return urls, nil
}

func (p *PostgresStorage) Exists(ctx context.Context, originalURL string) (*models.StorageURLModel, error) {
	var url models.StorageURLModel
	err := p.db.QueryRowContext(ctx,
		"SELECT id, short_url, original_url FROM urls WHERE original_url = $1",
		originalURL,
	).Scan(&url.ID, &url.ShortURL, &url.OriginalURL)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to check URL existence: %w", err)
	}

	return &url, nil
}

func (p *PostgresStorage) ExistsBatch(ctx context.Context, originalURLs []string) ([]models.StorageURLModel, error) {
	if len(originalURLs) == 0 {
		return nil, fmt.Errorf("%w: empty URL list", models.ErrInvalidData)
	}

	rows, err := p.db.QueryContext(ctx,
		"SELECT id, short_url, original_url FROM urls WHERE original_url = ANY($1)",
		originalURLs,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query existing URLs: %w", err)
	}
	defer rows.Close()

	var result []models.StorageURLModel
	for rows.Next() {
		var url models.StorageURLModel
		if err := rows.Scan(&url.ID, &url.ShortURL, &url.OriginalURL); err != nil {
			return nil, fmt.Errorf("failed to scan URL row: %w", err)
		}
		result = append(result, url)
	}

	return result, nil
}

func (p *PostgresStorage) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, storagePingTimeout)
	defer cancel()

	if err := p.db.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	return nil
}

func (p *PostgresStorage) Close() error {
	if err := p.db.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}
	return nil
}

func (p *PostgresStorage) GetAll(ctx context.Context) ([]models.StorageURLModel, error) {
	rows, err := p.db.QueryContext(ctx,
		"SELECT id, short_url, original_url FROM urls",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query URLs: %w", err)
	}
	defer rows.Close()

	var urls []models.StorageURLModel
	for rows.Next() {
		var url models.StorageURLModel
		if err := rows.Scan(&url.ID, &url.ShortURL, &url.OriginalURL); err != nil {
			return nil, fmt.Errorf("failed to scan URL: %w", err)
		}
		urls = append(urls, url)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return urls, nil
}
