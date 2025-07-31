package inmemory

import (
	"context"
	"errors"
	"sort"
	"time"
	"urlshortener/domain/models"
)

var (
	ErrInvalidData = errors.New("invalid data")
	ErrUnfound     = errors.New("unfound data")
	ErrEmpty       = errors.New("storage is empty")
	ErrConflict    = errors.New("url already exists with different value")
)

const initLastID = 0

type InmemoryStorage struct {
	data   map[string]models.URL
	lastID int
}

func NewStorage() *InmemoryStorage {
	return &InmemoryStorage{
		data:   make(map[string]models.URL),
		lastID: 0,
	}
}

func (m *InmemoryStorage) CreateOrUpdate(ctx context.Context, url models.URL) (models.URL, error) {
	if err := ctx.Err(); err != nil {
		return models.URL{}, ErrInvalidData
	}

	if url.ShortKey == "" || url.OriginalURL == "" {
		return models.URL{}, ErrInvalidData
	}

	if existingURL, exists := m.data[url.ShortKey]; exists {
		if existingURL.OriginalURL == url.OriginalURL {
			return existingURL, nil
		}
		return models.URL{}, ErrConflict
	}

	m.lastID++
	url.ID = m.lastID
	if url.CreatedAt.IsZero() {
		url.CreatedAt = time.Now()
	}

	m.data[url.ShortKey] = url
	return url, nil
}

func (m *InmemoryStorage) GetByShortKey(ctx context.Context, shortKey string) (models.URL, error) {
	if err := ctx.Err(); err != nil {
		return models.URL{}, ErrInvalidData
	}

	if shortKey == "" {
		return models.URL{}, ErrInvalidData
	}

	url, exists := m.data[shortKey]
	if !exists {
		return models.URL{}, ErrUnfound
	}
	return url, nil
}

func (m *InmemoryStorage) GetByOriginalURL(ctx context.Context, originalURL string) (models.URL, error) {
	if err := ctx.Err(); err != nil {
		return models.URL{}, ErrInvalidData
	}

	if originalURL == "" {
		return models.URL{}, ErrInvalidData
	}

	for _, url := range m.data {
		if url.OriginalURL == originalURL {
			return url, nil
		}
	}
	return models.URL{}, ErrUnfound
}

func (m *InmemoryStorage) GetAll(ctx context.Context) ([]models.URL, error) {
	if err := ctx.Err(); err != nil {
		return nil, ErrInvalidData
	}

	urls := make([]models.URL, 0, len(m.data))
	for _, url := range m.data {
		urls = append(urls, url)
	}

	sort.Slice(urls, func(i, j int) bool {
		return urls[i].ID < urls[j].ID
	})

	return urls, nil
}

func (m *InmemoryStorage) Delete(ctx context.Context, shortKey string) error {
	if err := ctx.Err(); err != nil {
		return ErrInvalidData
	}

	if shortKey == "" {
		return ErrInvalidData
	}

	delete(m.data, shortKey)
	return nil
}

func (m *InmemoryStorage) BatchCreate(ctx context.Context, urls []models.URL) ([]models.URL, error) {
	if err := ctx.Err(); err != nil {
		return nil, ErrInvalidData
	}

	if len(urls) == 0 {
		return nil, ErrInvalidData
	}

	result := make([]models.URL, 0, len(urls))
	for _, url := range urls {
		createdURL, err := m.CreateOrUpdate(ctx, url)
		if err != nil && !errors.Is(err, ErrConflict) {
			return nil, err
		}
		result = append(result, createdURL)
	}

	return result, nil
}

func (m *InmemoryStorage) Exists(ctx context.Context, originalURL string) (models.URL, error) {
	if err := ctx.Err(); err != nil {
		return models.URL{}, ErrInvalidData
	}

	for _, url := range m.data {
		if url.OriginalURL == originalURL {
			return url, nil
		}
	}
	return models.URL{}, nil
}

func (m *InmemoryStorage) ExistsBatch(ctx context.Context, originalURLs []string) ([]models.URL, error) {
	if err := ctx.Err(); err != nil {
		return nil, ErrInvalidData
	}

	if len(originalURLs) == 0 {
		return nil, ErrInvalidData
	}

	var result []models.URL
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

func (m *InmemoryStorage) List(ctx context.Context, limit, offset int) ([]models.URL, error) {
	if err := ctx.Err(); err != nil {
		return nil, ErrInvalidData
	}

	if limit <= 0 || offset < 0 {
		return nil, ErrInvalidData
	}

	allURLs := make([]models.URL, 0, len(m.data))
	for _, url := range m.data {
		allURLs = append(allURLs, url)
	}

	sort.Slice(allURLs, func(i, j int) bool {
		return allURLs[i].ID < allURLs[j].ID
	})

	start := offset
	if start > len(allURLs) {
		start = len(allURLs)
	}

	end := start + limit
	if end > len(allURLs) {
		end = len(allURLs)
	}

	return allURLs[start:end], nil
}

func (m *InmemoryStorage) Ping(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return ErrInvalidData
	}
	return nil
}

func (m *InmemoryStorage) Close() error {
	m.data = make(map[string]models.URL)
	m.lastID = 0
	return nil
}
