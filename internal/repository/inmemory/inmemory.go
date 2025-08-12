package inmemory

import (
	"context"
	"sort"
	"time"
	"urlshortener/domain/models"
)

type InmemoryStorage struct {
	data       map[string]models.ShortenedLink // shortKey -> ShortenedLink
	users      map[int64]models.User           // userID -> User
	lastURLID  int64
	lastUserID int64
}

func NewStorage() *InmemoryStorage {
	return &InmemoryStorage{
		data:       make(map[string]models.ShortenedLink),
		users:      make(map[int64]models.User),
		lastURLID:  0,
		lastUserID: 0,
	}
}

// URLStorage methods

func (m *InmemoryStorage) ShortenedLinkCreate(ctx context.Context, url models.ShortenedLink) (models.ShortenedLink, error) {
	if err := ctx.Err(); err != nil {
		return models.ShortenedLink{}, models.ErrInvalidData
	}

	if url.ShortCode == "" || url.LongURL == "" {
		return models.ShortenedLink{}, models.ErrInvalidData
	}

	// Check for existing URL with same short code
	if existingURL, exists := m.data[url.ShortCode]; exists {
		if existingURL.LongURL == url.LongURL {
			return existingURL, nil
		}
		return models.ShortenedLink{}, models.ErrConflict
	}

	// Check for existing URL with same long URL
	for _, u := range m.data {
		if u.LongURL == url.LongURL {
			return u, models.ErrConflict
		}
	}

	m.lastURLID++
	url.ID = m.lastURLID
	if url.CreatedAt.IsZero() {
		url.CreatedAt = time.Now()
	}

	m.data[url.ShortCode] = url
	return url, nil
}

func (m *InmemoryStorage) ShortenedLinkGetByShortKey(ctx context.Context, shortKey string) (models.ShortenedLink, error) {
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

func (m *InmemoryStorage) ShortenedLinkGetByLongURL(ctx context.Context, originalURL string) (models.ShortenedLink, error) {
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

func (m *InmemoryStorage) ShortenedLinkBatchCreate(ctx context.Context, urls []models.ShortenedLink) ([]models.ShortenedLink, error) {
	if err := ctx.Err(); err != nil {
		return nil, models.ErrInvalidData
	}

	if len(urls) == 0 {
		return nil, models.ErrInvalidData
	}

	result := make([]models.ShortenedLink, 0, len(urls))
	for _, url := range urls {
		// Check for conflicts first
		conflict := false
		for _, existing := range m.data {
			if existing.LongURL == url.LongURL {
				result = append(result, existing)
				conflict = true
				break
			}
			if existing.ShortCode == url.ShortCode {
				conflict = true
				break
			}
		}

		if conflict {
			continue
		}

		m.lastURLID++
		url.ID = m.lastURLID
		if url.CreatedAt.IsZero() {
			url.CreatedAt = time.Now()
		}
		m.data[url.ShortCode] = url
		result = append(result, url)
	}

	return result, nil
}

func (m *InmemoryStorage) ShortenedLinkBatchExists(ctx context.Context, originalURLs []string) ([]models.ShortenedLink, error) {
	if err := ctx.Err(); err != nil {
		return nil, models.ErrInvalidData
	}

	if len(originalURLs) == 0 {
		return nil, models.ErrInvalidData
	}

	result := make([]models.ShortenedLink, 0, len(originalURLs))
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

// UserStorage methods

func (m *InmemoryStorage) UserCreate(ctx context.Context, user models.User) (models.User, error) {
	if err := ctx.Err(); err != nil {
		return models.User{}, models.ErrInvalidData
	}

	m.lastUserID++
	user.ID = m.lastUserID
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}

	m.users[user.ID] = user
	return user, nil
}

func (m *InmemoryStorage) UserGetByID(ctx context.Context, id int64) (models.User, error) {
	if err := ctx.Err(); err != nil {
		return models.User{}, models.ErrInvalidData
	}

	if id <= 0 {
		return models.User{}, models.ErrInvalidData
	}

	user, exists := m.users[id]
	if !exists {
		return models.User{}, models.ErrUnfound
	}
	return user, nil
}

func (m *InmemoryStorage) ShortenedLinkGetBatchByUser(ctx context.Context, userID int64) ([]models.ShortenedLink, error) {
	if err := ctx.Err(); err != nil {
		return nil, models.ErrInvalidData
	}

	if userID <= 0 {
		return nil, models.ErrInvalidData
	}

	var result []models.ShortenedLink
	for _, url := range m.data {
		if url.UserID == userID {
			result = append(result, url)
		}
	}

	// Sort by creation date
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})

	return result, nil
}

func (m *InmemoryStorage) List(ctx context.Context, limit, offset int) ([]models.ShortenedLink, error) {
	if err := ctx.Err(); err != nil {
		return nil, models.ErrInvalidData
	}

	if limit <= 0 || offset < 0 {
		return nil, models.ErrInvalidData
	}

	// Получаем все URL из хранилища
	allURLs := make([]models.ShortenedLink, 0, len(m.data))
	for _, url := range m.data {
		allURLs = append(allURLs, url)
	}

	// Сортируем по дате создания (от новых к старым)
	sort.Slice(allURLs, func(i, j int) bool {
		return allURLs[i].CreatedAt.After(allURLs[j].CreatedAt)
	})

	// Применяем limit и offset
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

// Common methods

func (m *InmemoryStorage) Close() error {
	m.data = make(map[string]models.ShortenedLink)
	m.users = make(map[int64]models.User)
	m.lastURLID = 0
	m.lastUserID = 0
	return nil
}
