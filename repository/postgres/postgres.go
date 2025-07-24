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

func (p *PostgresStorage) Set(ctx context.Context, shortURL, originalURL string) (*models.URL, error) {
	if shortURL == "" || originalURL == "" {
		return nil, fmt.Errorf("%w: shortURL and originalURL must not be empty", models.ErrInvalidData)
	}

	existingURL, err := p.Get(ctx, shortURL)
	switch {
	case err == nil && existingURL != nil:
		if existingURL.OriginalURL == originalURL {
			return existingURL, nil
		}
		return nil, fmt.Errorf("%w: shortURL '%s' already exists with different originalURL", models.ErrConflict, shortURL)
	case errors.Is(err, models.ErrUnfound):
	case err != nil:
		return nil, fmt.Errorf("failed to check existing URL: %w", err)
	}

	var id int
	row := p.db.QueryRowContext(ctx,
		"INSERT INTO urls (short_url, original_url) VALUES ($1, $2) RETURNING id",
		shortURL, originalURL,
	)

	if err := row.Scan(&id); err != nil {
		return nil, fmt.Errorf("failed to insert URL: %w", err)
	}

	return &models.URL{
		ID:          id,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}, nil
}

func (p *PostgresStorage) Get(ctx context.Context, shortURL string) (*models.URL, error) {
	if shortURL == "" {
		return nil, fmt.Errorf("%w: shortURL must not be empty", models.ErrInvalidData)
	}

	var url models.URL
	row := p.db.QueryRowContext(ctx,
		"SELECT id, short_url, original_url FROM urls WHERE short_url = $1",
		shortURL,
	)

	if err := row.Scan(&url.ID, &url.ShortURL, &url.OriginalURL); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: shortURL '%s' not found", models.ErrUnfound, shortURL)
		}
		return nil, fmt.Errorf("failed to scan URL row: %w", err)
	}

	return &url, nil
}

func (p *PostgresStorage) GetAll(ctx context.Context) ([]models.URL, error) {
	rows, err := p.db.QueryContext(ctx, "SELECT id, short_url, original_url FROM urls")
	if err != nil {
		return nil, fmt.Errorf("failed to query URLs: %w", err)
	}
	defer rows.Close()

	var urls []models.URL
	for rows.Next() {
		var url models.URL
		if err := rows.Scan(&url.ID, &url.ShortURL, &url.OriginalURL); err != nil {
			return nil, fmt.Errorf("failed to scan URL row: %w", err)
		}
		urls = append(urls, url)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	if len(urls) == 0 {
		return nil, fmt.Errorf("%w: no URLs found", models.ErrEmpty)
	}

	return urls, nil
}

func (p *PostgresStorage) Ping(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := p.db.PingContext(pingCtx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	return nil
}

func createTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS urls (
			id SERIAL PRIMARY KEY,
			short_url VARCHAR(10) UNIQUE NOT NULL,
			original_url TEXT NOT NULL
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create table 'urls': %w", err)
	}
	return nil
}

func (p *PostgresStorage) Exists(ctx context.Context, originalURL string) (bool, string, error) {
	var shortURL string
	err := p.db.QueryRowContext(ctx,
		"SELECT short_url FROM urls WHERE original_url = $1",
		originalURL,
	).Scan(&shortURL)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, "", nil
		}
		return false, "", fmt.Errorf("failed to check URL existence: %w", err)
	}
	return true, shortURL, nil
}

func (p *PostgresStorage) Close() error {
	if err := p.db.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}
	return nil
}
