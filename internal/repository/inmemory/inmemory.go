package inmemory

import (
	"context"
	"sort"
	"time"
	"urlshortener/domain/models"
	"urlshortener/internal/repository/dto"
)

type InmemoryStorage struct {
	data       map[string]dto.ShortenedLinkDB
	users      map[int64]dto.UserDB
	lastURLID  int64
	lastUserID int64
}

func NewStorage() *InmemoryStorage {
	return &InmemoryStorage{
		data:       make(map[string]dto.ShortenedLinkDB),
		users:      make(map[int64]dto.UserDB),
		lastURLID:  0,
		lastUserID: 0,
	}
}

// URLStorage methods

func (m *InmemoryStorage) ShortenedLinkCreate(ctx context.Context, url models.ShortenedLink) (models.ShortenedLink, error) {
	if err := ctx.Err(); err != nil {
		return models.ShortenedLink{}, models.ErrInvalidData
	}

	if url.ShortCode == "" || url.OriginalURL == "" {
		return models.ShortenedLink{}, models.ErrInvalidData
	}

	// Check for existing URL with same short code
	if existingURL, exists := m.data[url.ShortCode]; exists {
		if existingURL.OriginalURL == url.OriginalURL {
			return dto.ShortenedLinkDBToDomain(existingURL), nil
		}
		return models.ShortenedLink{}, models.ErrConflict
	}

	// Check for existing URL with same long URL
	for _, u := range m.data {
		if u.OriginalURL == url.OriginalURL {
			return dto.ShortenedLinkDBToDomain(u), models.ErrConflict
		}
	}

	m.lastURLID++
	urlDB := dto.ShortenedLinkDBFromDomain(url)
	urlDB.ID = m.lastURLID
	if urlDB.CreatedAt.IsZero() {
		urlDB.CreatedAt = time.Now()
	}

	m.data[urlDB.ShortCode] = urlDB
	return dto.ShortenedLinkDBToDomain(urlDB), nil
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
	return dto.ShortenedLinkDBToDomain(url), nil
}

func (m *InmemoryStorage) ShortenedLinkGetByOriginalURL(ctx context.Context, originalURL string) (models.ShortenedLink, error) {
	if err := ctx.Err(); err != nil {
		return models.ShortenedLink{}, models.ErrInvalidData
	}

	if originalURL == "" {
		return models.ShortenedLink{}, models.ErrInvalidData
	}

	for _, url := range m.data {
		if url.OriginalURL == originalURL {
			return dto.ShortenedLinkDBToDomain(url), nil
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
		var existingDB dto.ShortenedLinkDB

		for _, existing := range m.data {
			if existing.OriginalURL == url.OriginalURL {
				existingDB = existing
				conflict = true
				break
			}
			if existing.ShortCode == url.ShortCode {
				conflict = true
				break
			}
		}

		if conflict {
			if existingDB.ID != 0 {
				result = append(result, dto.ShortenedLinkDBToDomain(existingDB))
			}
			continue
		}

		m.lastURLID++
		urlDB := dto.ShortenedLinkDBFromDomain(url)
		urlDB.ID = m.lastURLID
		if urlDB.CreatedAt.IsZero() {
			urlDB.CreatedAt = time.Now()
		}
		m.data[urlDB.ShortCode] = urlDB
		result = append(result, dto.ShortenedLinkDBToDomain(urlDB))
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
			if url.OriginalURL == originalURL {
				result = append(result, dto.ShortenedLinkDBToDomain(url))
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
	userDB := dto.UserDBFromDomain(user)
	userDB.ID = m.lastUserID
	if userDB.CreatedAt.IsZero() {
		userDB.CreatedAt = time.Now()
	}

	m.users[userDB.ID] = userDB
	return dto.UserDBToDomain(userDB), nil
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
	return dto.UserDBToDomain(user), nil
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
			result = append(result, dto.ShortenedLinkDBToDomain(url))
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

	// Get all URLs from storage
	allURLs := make([]models.ShortenedLink, 0, len(m.data))
	for _, url := range m.data {
		allURLs = append(allURLs, dto.ShortenedLinkDBToDomain(url))
	}

	// Sort by creation date (newest first)
	sort.Slice(allURLs, func(i, j int) bool {
		return allURLs[i].CreatedAt.After(allURLs[j].CreatedAt)
	})

	// Apply limit and offset
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
	m.data = make(map[string]dto.ShortenedLinkDB)
	m.users = make(map[int64]dto.UserDB)
	m.lastURLID = 0
	m.lastUserID = 0
	return nil
}

func (m *InmemoryStorage) Delete(ctx context.Context, shortKey string) error {
	if err := ctx.Err(); err != nil {
		return models.ErrInvalidData
	}

	if shortKey == "" {
		return models.ErrInvalidData
	}

	if _, exists := m.data[shortKey]; !exists {
		return models.ErrUnfound
	}

	delete(m.data, shortKey)
	return nil
}

func (m *InmemoryStorage) GetAll(ctx context.Context) ([]models.ShortenedLink, error) {
	if err := ctx.Err(); err != nil {
		return nil, models.ErrInvalidData
	}

	result := make([]models.ShortenedLink, 0, len(m.data))
	for _, url := range m.data {
		result = append(result, dto.ShortenedLinkDBToDomain(url))
	}

	return result, nil
}

func (m *InmemoryStorage) Exists(ctx context.Context, originalURL string) (models.ShortenedLink, error) {
	if err := ctx.Err(); err != nil {
		return models.ShortenedLink{}, models.ErrInvalidData
	}

	if originalURL == "" {
		return models.ShortenedLink{}, models.ErrInvalidData
	}

	for _, url := range m.data {
		if url.OriginalURL == originalURL {
			return dto.ShortenedLinkDBToDomain(url), nil
		}
	}

	return models.ShortenedLink{}, nil
}
