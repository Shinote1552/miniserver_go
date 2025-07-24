// internal/storage/inmemory/memory.go
package inmemory

import (
	"context"
	"urlshortener/internal/models"
)

const initLastID = 0

type InmemoryStorage struct {
	data   map[string]models.URL
	lastID int
}

func NewStorage() *InmemoryStorage {
	return &InmemoryStorage{
		data:   make(map[string]models.URL),
		lastID: initLastID,
	}
}

func (m *InmemoryStorage) Set(ctx context.Context, shortURL, originalURL string) (*models.URL, error) {
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
	url := models.URL{
		ID:          m.lastID,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}

	m.data[shortURL] = url
	return &url, nil
}

func (m *InmemoryStorage) Get(ctx context.Context, shortURL string) (*models.URL, error) {
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

func (m *InmemoryStorage) GetAll(ctx context.Context) ([]models.URL, error) {
	if err := ctx.Err(); err != nil {
		return nil, models.ErrInvalidData
	}

	if len(m.data) == 0 {
		return nil, models.ErrEmpty
	}

	urls := make([]models.URL, 0, len(m.data))
	for _, url := range m.data {
		urls = append(urls, url)
	}

	return urls, nil
}

func (m *InmemoryStorage) Ping(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return models.ErrInvalidData
	}
	return nil
}

func (m *InmemoryStorage) Exists(ctx context.Context, originalURL string) (bool, string, error) {
	for _, url := range m.data {
		if url.OriginalURL == originalURL {
			return true, url.ShortURL, nil
		}
	}
	return false, "", nil
}

// Cleaning data and reset the counter
func (m *InmemoryStorage) Close() error {
	m.data = make(map[string]models.URL)
	m.lastID = initLastID
	return nil
}
