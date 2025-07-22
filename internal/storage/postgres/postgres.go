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

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(ctx context.Context, dsn string) (*PostgresStorage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Ping с контекстом
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Создание таблиц с контекстом
	if err := createTables(ctx, db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &PostgresStorage{db: db}, nil
}

func createTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS urls (
			id SERIAL PRIMARY KEY,
			short_url VARCHAR(10) UNIQUE NOT NULL,
			original_url TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		);
	`)
	return err
}

func (p *PostgresStorage) Set(ctx context.Context, shortURL, originalURL string) (*models.URL, error) {
	if shortURL == "" || originalURL == "" {
		return nil, models.ErrInvalidData
	}

	existingURL, err := p.Get(ctx, shortURL)
	if err == nil && existingURL != nil {
		if existingURL.OriginalURL == originalURL {
			return existingURL, nil
		}
		return nil, models.ErrConflict
	}

	var id int
	err = p.db.QueryRowContext(ctx,
		"INSERT INTO urls (short_url, original_url) VALUES ($1, $2) RETURNING id",
		shortURL, originalURL,
	).Scan(&id)

	if err != nil {
		return nil, fmt.Errorf("failed to insert url: %w", err)
	}

	return &models.URL{
		ID:          id,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}, nil
}

func (p *PostgresStorage) Get(ctx context.Context, shortURL string) (*models.URL, error) {
	var url models.URL
	err := p.db.QueryRowContext(ctx,
		"SELECT id, short_url, original_url FROM urls WHERE short_url = $1",
		shortURL,
	).Scan(&url.ID, &url.ShortURL, &url.OriginalURL)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrUnfound
		}
		return nil, fmt.Errorf("failed to get url: %w", err)
	}

	return &url, nil
}

func (p *PostgresStorage) GetAll(ctx context.Context) ([]models.URL, error) {
	rows, err := p.db.QueryContext(ctx, "SELECT id, short_url, original_url FROM urls")
	if err != nil {
		return nil, fmt.Errorf("failed to query urls: %w", err)
	}
	defer rows.Close()

	var urls []models.URL
	for rows.Next() {
		var url models.URL
		if err := rows.Scan(&url.ID, &url.ShortURL, &url.OriginalURL); err != nil {
			return nil, fmt.Errorf("failed to scan url: %w", err)
		}
		urls = append(urls, url)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	if len(urls) == 0 {
		return nil, models.ErrEmpty
	}

	return urls, nil
}

func (p *PostgresStorage) PingDataBase(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return p.db.PingContext(ctx)
}
