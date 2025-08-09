package repository

import (
	"context"
	"urlshortener/domain/models"
)

// Storage - основной интерфейс хранилища URL и shotURL
type (
	Storage interface {
		// Основные CRUD операции
		CreateOrUpdate(ctx context.Context, url models.ShortenedLink) (models.ShortenedLink, error)
		GetByShortKey(ctx context.Context, shortKey string) (models.ShortenedLink, error)
		GetByLongURL(ctx context.Context, originalURL string) (models.ShortenedLink, error)
		Delete(ctx context.Context, shortKey string) error

		// Пакетные операции
		BatchCreate(ctx context.Context, urls []models.ShortenedLink) ([]models.ShortenedLink, error)

		// BatchGetByUserID(ctx context.Context, UUID string) ([]models.ShortenedLink, error)

		ExistsBatch(ctx context.Context, originalURLs []string) ([]models.ShortenedLink, error)

		Exists(ctx context.Context, originalURL string) (models.ShortenedLink, error)

		// Пагинация/листинг
		List(ctx context.Context, limit, offset int) ([]models.ShortenedLink, error)

		// Управление соединением
		Ping(ctx context.Context) error
		Close() error

		// obly for experimental build
		GetAll(ctx context.Context) ([]models.ShortenedLink, error)
	}
)
