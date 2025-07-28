package repository

import (
	"context"
	"urlshortener/internal/models"
)

// Storage - основной интерфейс хранилища URL и shotURL
type (
	Storage interface {
		// Основные CRUD операции
		CreateOrUpdate(ctx context.Context, shortURL, originalURL string) (*models.StorageURLModel, error)
		GetByShortURL(ctx context.Context, shortURL string) (*models.StorageURLModel, error)
		GetByOriginalURL(ctx context.Context, originalURL string) (*models.StorageURLModel, error)
		GetAll(ctx context.Context) ([]models.StorageURLModel, error)
		Delete(ctx context.Context, shortURL string) error

		// Пакетные операции
		BatchCreate(ctx context.Context, batchItems []models.APIBatchRequestItem) ([]models.APIBatchResponseItem, error)

		// Проверки существования
		Exists(ctx context.Context, originalURL string) (*models.StorageURLModel, error)
		ExistsBatch(ctx context.Context, originalURLs []string) ([]models.StorageURLModel, error)

		// Пагинация и листинг
		List(ctx context.Context, limit, offset int) ([]models.StorageURLModel, error)

		// Управление соединением
		Ping(ctx context.Context) error
		Close() error
	}
)
