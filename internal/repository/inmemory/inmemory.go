package inmemory

import (
	"context"
	"sort"
	"sync"
	"time"
	"urlshortener/internal/domain/models"
	"urlshortener/internal/repository/dto"
)

type InmemoryStorage struct {
	rwmu  sync.RWMutex
	data  map[string]dto.ShortenedLinkDB
	users map[int64]dto.UserDB

	// Дополнительные индексы(пока что почти такие же как и в Postgres)
	originalURLIndex map[string]string
	userURLsIndex    map[int64][]string
	urlsIsDeleted    map[string]bool

	createdAtIndex map[string][]string
	deletedAtIndex map[string][]string

	lastURLID  int64
	lastUserID int64
}

func NewStorage() *InmemoryStorage {
	return &InmemoryStorage{
		data:             make(map[string]dto.ShortenedLinkDB),
		users:            make(map[int64]dto.UserDB),
		originalURLIndex: make(map[string]string),
		createdAtIndex:   make(map[string][]string),
		deletedAtIndex:   make(map[string][]string),
		userURLsIndex:    make(map[int64][]string),
		urlsIsDeleted:    make(map[string]bool),
		lastURLID:        0,
		lastUserID:       0,
	}
}

func (m *InmemoryStorage) Close() error {
	m.rwmu.Lock()
	defer m.rwmu.Unlock()

	clear(m.data)
	clear(m.users)
	clear(m.originalURLIndex)
	clear(m.createdAtIndex)
	clear(m.userURLsIndex)

	m.lastURLID = 0
	m.lastUserID = 0

	return nil
}

// URLStorage methods
func (m *InmemoryStorage) ShortenedLinkCreate(ctx context.Context, url models.ShortenedLink) (models.ShortenedLink, error) {
	if err := ctx.Err(); err != nil {
		return models.ShortenedLink{}, models.ErrInvalidData
	}

	if url.ShortCode == "" || url.OriginalURL == "" || url.UserID <= 0 {
		return models.ShortenedLink{}, models.ErrInvalidData
	}

	m.rwmu.Lock()
	defer m.rwmu.Unlock()

	if existingURL, exists := m.data[url.ShortCode]; exists {
		if existingURL.OriginalURL == url.OriginalURL {
			return dto.ShortenedLinkDBToDomain(existingURL), nil
		}
		return models.ShortenedLink{}, models.ErrConflict
	}

	if shortCode, exists := m.originalURLIndex[url.OriginalURL]; exists {
		if existingURL, exists := m.data[shortCode]; exists {
			return dto.ShortenedLinkDBToDomain(existingURL), models.ErrConflict
		}
	}

	m.lastURLID++
	urlDB := dto.ShortenedLinkDBFromDomain(url)
	urlDB.ID = m.lastURLID
	if urlDB.CreatedAt.IsZero() {
		urlDB.CreatedAt = time.Now()
	}

	// Добавляем во все индексы
	m.data[urlDB.ShortCode] = urlDB
	m.originalURLIndex[urlDB.OriginalURL] = urlDB.ShortCode

	dateKey := urlDB.CreatedAt.Format("2006-01-02")
	m.createdAtIndex[dateKey] = append(m.createdAtIndex[dateKey], urlDB.ShortCode)
	m.userURLsIndex[urlDB.UserID] = append(m.userURLsIndex[urlDB.UserID], urlDB.ShortCode)

	return dto.ShortenedLinkDBToDomain(urlDB), nil
}
func (m *InmemoryStorage) ShortenedLinkGetByShortKey(ctx context.Context, shortKey string) (models.ShortenedLink, error) {
	if err := ctx.Err(); err != nil {
		return models.ShortenedLink{}, models.ErrInvalidData
	}

	if shortKey == "" {
		return models.ShortenedLink{}, models.ErrInvalidData
	}

	m.rwmu.RLock()
	defer m.rwmu.RUnlock()

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

	m.rwmu.RLock()
	defer m.rwmu.RUnlock()

	shortCode, exists := m.originalURLIndex[originalURL]
	if !exists {
		return models.ShortenedLink{}, models.ErrUnfound
	}

	url, exists := m.data[shortCode]
	if !exists {
		return models.ShortenedLink{}, models.ErrUnfound
	}
	return dto.ShortenedLinkDBToDomain(url), nil
}

func (m *InmemoryStorage) ShortenedLinkBatchCreate(ctx context.Context, urls []models.ShortenedLink) ([]models.ShortenedLink, error) {
	if err := ctx.Err(); err != nil {
		return nil, models.ErrInvalidData
	}

	if len(urls) == 0 {
		return nil, models.ErrInvalidData
	}

	m.rwmu.Lock()
	defer m.rwmu.Unlock()

	result := make([]models.ShortenedLink, 0, len(urls))
	for _, url := range urls {
		if url.ShortCode == "" || url.OriginalURL == "" || url.UserID <= 0 {
			return nil, models.ErrInvalidData
		}

		// Проверяем конфликты
		if existingURL, exists := m.data[url.ShortCode]; exists {
			if existingURL.OriginalURL != url.OriginalURL {
				return nil, models.ErrConflict
			}
			result = append(result, dto.ShortenedLinkDBToDomain(existingURL))
			continue
		}

		if shortCode, exists := m.originalURLIndex[url.OriginalURL]; exists {
			if existingURL, exists := m.data[shortCode]; exists {
				result = append(result, dto.ShortenedLinkDBToDomain(existingURL))
				continue
			}
		}

		m.lastURLID++
		urlDB := dto.ShortenedLinkDBFromDomain(url)
		urlDB.ID = m.lastURLID
		if urlDB.CreatedAt.IsZero() {
			urlDB.CreatedAt = time.Now()
		}

		// Добавляем во все индексы
		m.data[urlDB.ShortCode] = urlDB
		m.originalURLIndex[urlDB.OriginalURL] = urlDB.ShortCode

		dateKey := urlDB.CreatedAt.Format("2006-01-02")
		m.createdAtIndex[dateKey] = append(m.createdAtIndex[dateKey], urlDB.ShortCode)
		m.userURLsIndex[urlDB.UserID] = append(m.userURLsIndex[urlDB.UserID], urlDB.ShortCode)

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

	m.rwmu.RLock()
	defer m.rwmu.RUnlock()

	result := make([]models.ShortenedLink, 0, len(originalURLs))
	for _, originalURL := range originalURLs {
		if originalURL == "" {
			continue
		}

		if shortCode, exists := m.originalURLIndex[originalURL]; exists {
			if url, exists := m.data[shortCode]; exists {
				result = append(result, dto.ShortenedLinkDBToDomain(url))
			}
		}
	}

	return result, nil
}

func (m *InmemoryStorage) ShortenedLinkGetBatchByUser(ctx context.Context, userID int64) ([]models.ShortenedLink, error) {
	if err := ctx.Err(); err != nil {
		return nil, models.ErrInvalidData
	}

	if userID <= 0 {
		return nil, models.ErrInvalidData
	}

	m.rwmu.RLock()
	defer m.rwmu.RUnlock()

	shortKeys, exists := m.userURLsIndex[userID]
	if !exists {
		return []models.ShortenedLink{}, nil
	}

	result := make([]models.ShortenedLink, 0, len(shortKeys))
	for _, shortKey := range shortKeys {
		if url, exists := m.data[shortKey]; exists {
			result = append(result, dto.ShortenedLinkDBToDomain(url))
		}
	}

	// Сортируем по дате создания
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

	m.rwmu.RLock()
	defer m.rwmu.RUnlock()

	// Собираем все URL
	allURLs := make([]models.ShortenedLink, 0, len(m.data))
	for _, url := range m.data {
		allURLs = append(allURLs, dto.ShortenedLinkDBToDomain(url))
	}

	// Сортируем по дате создания (сначала новые)
	sort.Slice(allURLs, func(i, j int) bool {
		return allURLs[i].CreatedAt.After(allURLs[j].CreatedAt)
	})

	// Применяем limit и offset
	start := offset
	if start > len(allURLs) {
		return []models.ShortenedLink{}, nil
	}

	end := start + limit
	if end > len(allURLs) {
		end = len(allURLs)
	}

	return allURLs[start:end], nil
}

func (m *InmemoryStorage) Delete(ctx context.Context, shortKey string) error {
	if err := ctx.Err(); err != nil {
		return models.ErrInvalidData
	}

	if shortKey == "" {
		return models.ErrInvalidData
	}

	m.rwmu.Lock()
	defer m.rwmu.Unlock()

	url, exists := m.data[shortKey]
	if !exists {
		return models.ErrUnfound
	}

	// Удаляем из всех индексов
	delete(m.data, shortKey)
	delete(m.originalURLIndex, url.OriginalURL)

	dateKey := url.CreatedAt.Format("2006-01-02")
	if keys, exists := m.createdAtIndex[dateKey]; exists {
		for i, key := range keys {
			if key == shortKey {
				// Простое удаление элемента (меняем порядок, но это ок)
				m.createdAtIndex[dateKey] = append(keys[:i], keys[i+1:]...)
				break
			}
		}
		if len(m.createdAtIndex[dateKey]) == 0 {
			delete(m.createdAtIndex, dateKey)
		}
	}

	if keys, exists := m.userURLsIndex[url.UserID]; exists {
		for i, key := range keys {
			if key == shortKey {
				m.userURLsIndex[url.UserID] = append(keys[:i], keys[i+1:]...)
				break
			}
		}
		if len(m.userURLsIndex[url.UserID]) == 0 {
			delete(m.userURLsIndex, url.UserID)
		}
	}

	return nil
}

func (m *InmemoryStorage) Exists(ctx context.Context, originalURL string) (models.ShortenedLink, error) {
	if err := ctx.Err(); err != nil {
		return models.ShortenedLink{}, models.ErrInvalidData
	}

	if originalURL == "" {
		return models.ShortenedLink{}, models.ErrInvalidData
	}

	m.rwmu.RLock()
	defer m.rwmu.RUnlock()

	shortCode, exists := m.originalURLIndex[originalURL]
	if !exists {
		return models.ShortenedLink{}, models.ErrUnfound
	}

	url, exists := m.data[shortCode]
	if !exists {
		return models.ShortenedLink{}, models.ErrUnfound
	}

	return dto.ShortenedLinkDBToDomain(url), nil
}

// UserStorage methods
func (m *InmemoryStorage) UserCreate(ctx context.Context, user models.User) (models.User, error) {
	if err := ctx.Err(); err != nil {
		return models.User{}, models.ErrInvalidData
	}

	m.rwmu.Lock()
	defer m.rwmu.Unlock()

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

	m.rwmu.RLock()
	defer m.rwmu.RUnlock()

	user, exists := m.users[id]
	if !exists {
		return models.User{}, models.ErrUnfound
	}
	return dto.UserDBToDomain(user), nil
}

func (m *InmemoryStorage) Ping(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return models.ErrInvalidData
	}
	return nil
}

func (m *InmemoryStorage) WithinTx(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	return fn(ctx)
}

// Временный метод для отладки
func (m *InmemoryStorage) GetAll(ctx context.Context) ([]models.ShortenedLink, error) {
	if err := ctx.Err(); err != nil {
		return nil, models.ErrInvalidData
	}

	m.rwmu.RLock()
	defer m.rwmu.RUnlock()

	result := make([]models.ShortenedLink, 0, len(m.data))
	for _, url := range m.data {
		result = append(result, dto.ShortenedLinkDBToDomain(url))
	}

	return result, nil
}
