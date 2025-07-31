package repository

import (
	"context"
	"urlshortener/domain/models"
)

// Storage - основной интерфейс хранилища URL и shotURL
type (
	Storage interface {
		// Основные CRUD операции
		CreateOrUpdate(ctx context.Context, url models.URL) (models.URL, error)
		GetByShortKey(ctx context.Context, shortKey string) (models.URL, error)
		GetByOriginalURL(ctx context.Context, originalURL string) (models.URL, error)
		GetAll(ctx context.Context) ([]models.URL, error)
		Delete(ctx context.Context, shortKey string) error

		// Пакетные операции
		BatchCreate(ctx context.Context, urls []models.URL) ([]models.URL, error)

		// Проверки существования
		Exists(ctx context.Context, originalURL string) (models.URL, error)
		ExistsBatch(ctx context.Context, originalURLs []string) ([]models.URL, error)

		// Пагинация/листинг
		List(ctx context.Context, limit, offset int) ([]models.URL, error)

		// Управление соединением
		Ping(ctx context.Context) error
		Close() error
	}
)
