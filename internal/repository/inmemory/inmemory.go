package inmemory

import (
	"context"
	"sort"
	"urlshortener/internal/models"
)

const initLastID = 0

type InmemoryStorage struct {
	data   map[string]models.StorageURLModel
	lastID int
}

func NewStorage() *InmemoryStorage {
	return &InmemoryStorage{
		data:   make(map[string]models.StorageURLModel),
		lastID: initLastID,
	}
}

func (m *InmemoryStorage) CreateOrUpdate(ctx context.Context, shortURL, originalURL string) (*models.StorageURLModel, error) {
	if err := ctx.Err(); err != nil {
		return nil, models.ErrInvalidData
	}

	if shortURL == "" || originalURL == "" {
		return nil, models.ErrInvalidData
	}

	if existingURL, exists := m.data[shortURL]; exists {
		if existingURL.OriginalURL == originalURL {
			return &existingURL, nil
		}
		return nil, models.ErrConflict
	}

	m.lastID++
	url := models.StorageURLModel{
		ID:          m.lastID,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}

	m.data[shortURL] = url
	return &url, nil
}

func (m *InmemoryStorage) GetByShortURL(ctx context.Context, shortURL string) (*models.StorageURLModel, error) {
	if err := ctx.Err(); err != nil {
		return nil, models.ErrInvalidData
	}

	if shortURL == "" {
		return nil, models.ErrInvalidData
	}

	url, exists := m.data[shortURL]
	if !exists {
		return nil, models.ErrUnfound
	}
	return &url, nil
}

func (m *InmemoryStorage) GetByOriginalURL(ctx context.Context, originalURL string) (*models.StorageURLModel, error) {
	if err := ctx.Err(); err != nil {
		return nil, models.ErrInvalidData
	}

	for _, url := range m.data {
		if url.OriginalURL == originalURL {
			return &url, nil
		}
	}
	return nil, models.ErrUnfound
}

func (m *InmemoryStorage) GetAll(ctx context.Context) ([]models.StorageURLModel, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	urls := make([]models.StorageURLModel, 0, len(m.data))
	for _, url := range m.data {
		urls = append(urls, url)
	}

	// TODO Опционально: сортировка по ID, возможно лучше убрать
	sort.Slice(urls, func(i, j int) bool {
		return urls[i].ID < urls[j].ID
	})

	return urls, nil
}

func (m *InmemoryStorage) Delete(ctx context.Context, shortURL string) error {
	if err := ctx.Err(); err != nil {
		return models.ErrInvalidData
	}

	delete(m.data, shortURL)
	return nil
}

func (m *InmemoryStorage) BatchCreate(ctx context.Context, batchItems []models.APIShortenRequestBatch) ([]models.APIShortenResponseBatch, error) {
	if err := ctx.Err(); err != nil {
		return nil, models.ErrInvalidData
	}

	results := make([]models.APIShortenResponseBatch, 0, len(batchItems))
	for _, item := range batchItems {
		shortURL := item.CorrelationID
		_, err := m.CreateOrUpdate(ctx, shortURL, item.OriginalURL)
		if err != nil && err != models.ErrConflict {
			return nil, err
		}

		results = append(results, models.APIShortenResponseBatch{
			CorrelationID: item.CorrelationID,
			ShortURL:      shortURL,
		})
	}

	return results, nil
}

func (m *InmemoryStorage) Exists(ctx context.Context, originalURL string) (*models.StorageURLModel, error) {
	for _, url := range m.data {
		if url.OriginalURL == originalURL {
			return &url, nil
		}
	}
	return nil, nil
}

func (m *InmemoryStorage) ExistsBatch(ctx context.Context, originalURLs []string) ([]models.StorageURLModel, error) {
	if err := ctx.Err(); err != nil {
		return nil, models.ErrInvalidData
	}

	var result []models.StorageURLModel
	for _, originalURL := range originalURLs {
		for _, url := range m.data {
			if url.OriginalURL == originalURL {
				result = append(result, url)
				break
			}
		}
	}
	return result, nil
}

func (m *InmemoryStorage) List(ctx context.Context, limit, offset int) ([]models.StorageURLModel, error) {
	if err := ctx.Err(); err != nil {
		return nil, models.ErrInvalidData
	}

	if len(m.data) == 0 {
		return nil, models.ErrEmpty
	}

	urls := make([]models.StorageURLModel, 0, len(m.data))
	for _, url := range m.data {
		urls = append(urls, url)
	}

	sort.Slice(urls, func(i, j int) bool {
		return urls[i].ID < urls[j].ID
	})

	start := offset
	if start > len(urls) {
		start = len(urls)
	}

	end := start + limit
	if end > len(urls) {
		end = len(urls)
	}

	return urls[start:end], nil
}

func (m *InmemoryStorage) Ping(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return models.ErrInvalidData
	}
	return nil
}

func (m *InmemoryStorage) Close() error {
	m.data = make(map[string]models.StorageURLModel)
	m.lastID = initLastID
	return nil
}
