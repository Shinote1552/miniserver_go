package repository

import (
	"context"
	"urlshortener/internal/models"
)

// Storage - основной интерфейс хранилища URL и shotURL
type Storage interface {
	Set(ctx context.Context, shortURL, originalURL string) (*models.URL, error)
	Get(ctx context.Context, shortURL string) (*models.URL, error)
	GetAll(ctx context.Context) ([]models.URL, error)
	Exists(ctx context.Context, originalURL string) (bool, string, error)
	Ping(ctx context.Context) error
	Close() error
}
