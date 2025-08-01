package inmemory

import (
	"context"
	"errors"
	"sort"
	"time"
	"urlshortener/domain/models"
)

const initLastID = 0

type InmemoryStorage struct {
	data   map[string]models.ShortenedLink
	lastID int
}

func NewStorage() *InmemoryStorage {
	return &InmemoryStorage{
		data:   make(map[string]models.ShortenedLink),
		lastID: 0,
	}
}

func (m *InmemoryStorage) CreateOrUpdate(ctx context.Context, url models.ShortenedLink) (models.ShortenedLink, error) {
	if err := ctx.Err(); err != nil {
		return models.ShortenedLink{}, models.ErrInvalidData
	}

	if url.ShortCode == "" || url.LongURL == "" {
		return models.ShortenedLink{}, models.ErrInvalidData
	}

	if existingURL, exists := m.data[url.ShortCode]; exists {
		if existingURL.LongURL == url.LongURL {
			return existingURL, nil
		}
		return models.ShortenedLink{}, models.ErrConflict
	}

	m.lastID++
	url.ID = m.lastID
	if url.CreatedAt.IsZero() {
		url.CreatedAt = time.Now()
	}

	m.data[url.ShortCode] = url
	return url, nil
}

func (m *InmemoryStorage) GetByShortKey(ctx context.Context, shortKey string) (models.ShortenedLink, error) {
	if err := ctx.Err(); err != nil {
		return models.ShortenedLink{}, models.ErrInvalidData
	}

	if shortKey == "" {
		return models.ShortenedLink{}, models.ErrInvalidData
	}

	url, exists := m.data[shortKey]
	if !exists {
		return models.ShortenedLink{}, models.ErrUnfound
	}
	return url, nil
}

func (m *InmemoryStorage) GetByLongURL(ctx context.Context, originalURL string) (models.ShortenedLink, error) {
	if err := ctx.Err(); err != nil {
		return models.ShortenedLink{}, models.ErrInvalidData
	}

	if originalURL == "" {
		return models.ShortenedLink{}, models.ErrInvalidData
	}

	for _, url := range m.data {
		if url.LongURL == originalURL {
			return url, nil
		}
	}
	return models.ShortenedLink{}, models.ErrUnfound
}

func (m *InmemoryStorage) GetAll(ctx context.Context) ([]models.ShortenedLink, error) {
	if err := ctx.Err(); err != nil {
		return nil, models.ErrInvalidData
	}

	urls := make([]models.ShortenedLink, 0, len(m.data))
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
		return models.ErrInvalidData
	}

	if shortKey == "" {
		return models.ErrInvalidData
	}

	delete(m.data, shortKey)
	return nil
}

func (m *InmemoryStorage) BatchCreate(ctx context.Context, urls []models.ShortenedLink) ([]models.ShortenedLink, error) {
	if err := ctx.Err(); err != nil {
		return nil, models.ErrInvalidData
	}

	if len(urls) == 0 {
		return nil, models.ErrInvalidData
	}

	result := make([]models.ShortenedLink, 0, len(urls))
	for _, url := range urls {
		createdURL, err := m.CreateOrUpdate(ctx, url)
		if err != nil && !errors.Is(err, models.ErrConflict) {
			return nil, err
		}
		result = append(result, createdURL)
	}

	return result, nil
}

func (m *InmemoryStorage) Exists(ctx context.Context, originalURL string) (models.ShortenedLink, error) {
	if err := ctx.Err(); err != nil {
		return models.ShortenedLink{}, models.ErrInvalidData
	}

	for _, url := range m.data {
		if url.LongURL == originalURL {
			return url, nil
		}
	}
	return models.ShortenedLink{}, nil
}

func (m *InmemoryStorage) ExistsBatch(ctx context.Context, originalURLs []string) ([]models.ShortenedLink, error) {
	if err := ctx.Err(); err != nil {
		return nil, models.ErrInvalidData
	}

	if len(originalURLs) == 0 {
		return nil, models.ErrInvalidData
	}

	var result []models.ShortenedLink
	for _, originalURL := range originalURLs {
		for _, url := range m.data {
			if url.LongURL == originalURL {
				result = append(result, url)
				break
			}
		}
	}
	return result, nil
}

func (m *InmemoryStorage) List(ctx context.Context, limit, offset int) ([]models.ShortenedLink, error) {
	if err := ctx.Err(); err != nil {
		return nil, models.ErrInvalidData
	}

	if limit <= 0 || offset < 0 {
		return nil, models.ErrInvalidData
	}

	allURLs := make([]models.ShortenedLink, 0, len(m.data))
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
		return models.ErrInvalidData
	}
	return nil
}

func (m *InmemoryStorage) Close() error {
	m.data = make(map[string]models.ShortenedLink)
	m.lastID = 0
	return nil
}
