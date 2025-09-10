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
	originalURLIndex map[string]dto.ShortenedLinkDB
	createdAtIndex   map[time.Time][]string
	userURLsIndex    map[int64][]string

	lastURLID  int64
	lastUserID int64
}

func NewStorage() *InmemoryStorage {
	return &InmemoryStorage{
		data:             make(map[string]dto.ShortenedLinkDB),
		users:            make(map[int64]dto.UserDB),
		originalURLIndex: make(map[string]dto.ShortenedLinkDB),
		createdAtIndex:   make(map[time.Time][]string),
		userURLsIndex:    make(map[int64][]string),
		lastURLID:        0,
		lastUserID:       0,
	}
}

func (m *InmemoryStorage) Close() error {
	m.rwmu.Lock()
	defer m.rwmu.Unlock()

	// Очищаем все основные структуры данных
	m.data = make(map[string]dto.ShortenedLinkDB)
	m.users = make(map[int64]dto.UserDB)

	// Очищаем все индексы
	m.originalURLIndex = make(map[string]dto.ShortenedLinkDB)
	m.createdAtIndex = make(map[time.Time][]string)
	m.userURLsIndex = make(map[int64][]string)

	// Сбрасываем счетчики
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

	// Check for existing URL with same short code (индекс short_key)
	if existingURL, exists := m.data[url.ShortCode]; exists {
		if existingURL.OriginalURL == url.OriginalURL {
			return dto.ShortenedLinkDBToDomain(existingURL), nil
		}
		return models.ShortenedLink{}, models.ErrConflict
	}

	// Check for existing URL with same long URL (индекс original_url)
	if existingURL, exists := m.originalURLIndex[url.OriginalURL]; exists {
		return dto.ShortenedLinkDBToDomain(existingURL), models.ErrConflict
	}

	m.lastURLID++
	urlDB := dto.ShortenedLinkDBFromDomain(url)
	urlDB.ID = m.lastURLID
	if urlDB.CreatedAt.IsZero() {
		urlDB.CreatedAt = time.Now()
	}

	// Добавляем во все индексы
	m.data[urlDB.ShortCode] = urlDB
	m.originalURLIndex[urlDB.OriginalURL] = urlDB
	m.createdAtIndex[urlDB.CreatedAt] = append(m.createdAtIndex[urlDB.CreatedAt], urlDB.ShortCode)
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

	// Используем индекс по short_key
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

	// Используем индекс по original_url
	url, exists := m.originalURLIndex[originalURL]
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
		if url.UserID <= 0 {
			return nil, models.ErrInvalidData
		}

		// Проверяем конфликты через индексы
		if existingURL, exists := m.data[url.ShortCode]; exists {
			if existingURL.OriginalURL == url.OriginalURL {
				result = append(result, dto.ShortenedLinkDBToDomain(existingURL))
			}
			continue
		}

		if existingURL, exists := m.originalURLIndex[url.OriginalURL]; exists {
			result = append(result, dto.ShortenedLinkDBToDomain(existingURL))
			continue
		}

		m.lastURLID++
		urlDB := dto.ShortenedLinkDBFromDomain(url)
		urlDB.ID = m.lastURLID
		if urlDB.CreatedAt.IsZero() {
			urlDB.CreatedAt = time.Now()
		}

		// Добавляем во все индексы
		m.data[urlDB.ShortCode] = urlDB
		m.originalURLIndex[urlDB.OriginalURL] = urlDB
		m.createdAtIndex[urlDB.CreatedAt] = append(m.createdAtIndex[urlDB.CreatedAt], urlDB.ShortCode)
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
		// Используем индекс по original_url
		if url, exists := m.originalURLIndex[originalURL]; exists {
			result = append(result, dto.ShortenedLinkDBToDomain(url))
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

	// Используем индекс по user_id
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

	// by creation date
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

	// Собираем все URL через индекс created_at
	allURLs := make([]models.ShortenedLink, 0, len(m.data))
	for _, shortKeys := range m.createdAtIndex {
		for _, shortKey := range shortKeys {
			if url, exists := m.data[shortKey]; exists {
				allURLs = append(allURLs, dto.ShortenedLinkDBToDomain(url))
			}
		}
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

	// Удаляем из created_at индекса
	if keys, exists := m.createdAtIndex[url.CreatedAt]; exists {
		for i, key := range keys {
			if key == shortKey {
				m.createdAtIndex[url.CreatedAt] = append(keys[:i], keys[i+1:]...)
				break
			}
		}
		if len(m.createdAtIndex[url.CreatedAt]) == 0 {
			delete(m.createdAtIndex, url.CreatedAt)
		}
	}

	// Удаляем из user_urls индекса
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

	// Используем индекс по original_url
	if url, exists := m.originalURLIndex[originalURL]; exists {
		return dto.ShortenedLinkDBToDomain(url), nil
	}

	return models.ShortenedLink{}, nil
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
	// Пока что это заглушка
	// Для inmemory просто вызываем функцию без транзакций
	// так как все операции и так атомарны благодаря мьютексу
	return fn(ctx)
}

// Временный метод!
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
