package services

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"
	"urlshortener/domain/models"
	"urlshortener/internal/repository"
)

var (
	ErrInvalidData = errors.New("invalid data")
	ErrUnfound     = errors.New("unfound data")
	ErrEmpty       = errors.New("storage is empty")
	ErrConflict    = errors.New("url already exists with different value")
)

// URLShortener реализует бизнес-логику сервиса сокращения URL
type URLShortener struct {
	storage repository.Storage
	baseURL string
}

// NewServiceURLShortener создает новый экземпляр сервиса
func NewServiceURLShortener(storage repository.Storage, baseURL string) *URLShortener {
	return &URLShortener{
		storage: storage,
		baseURL: baseURL,
	}
}

// GetURL возвращает оригинальный URL по короткому ключу
func (s *URLShortener) GetURL(ctx context.Context, shortKey string) (models.URL, error) {
	if shortKey == "" {
		return models.URL{}, ErrInvalidData
	}

	url, err := s.storage.GetByShortKey(ctx, shortKey)
	if err != nil {
		if errors.Is(err, ErrUnfound) {
			return models.URL{}, fmt.Errorf("%w: URL not found", ErrUnfound)
		}
		return models.URL{}, fmt.Errorf("failed to get URL: %w", err)
	}
	return url, nil
}

// GetShortURL возвращает полный короткий URL
func (s *URLShortener) GetShortURL(shortKey string) string {
	return fmt.Sprintf("%s/%s", s.baseURL, shortKey)
}

// SetURL создает новую короткую ссылку или возвращает существующую
func (s *URLShortener) SetURL(ctx context.Context, originalURL string) (models.URL, error) {
	if originalURL == "" {
		return models.URL{}, ErrInvalidData
	}

	// Проверяем существование URL
	existing, err := s.storage.GetByOriginalURL(ctx, originalURL)
	if err == nil {
		return existing, ErrConflict
	}

	// Генерируем уникальный токен
	token, err := s.generateUniqueToken(ctx)
	if err != nil {
		return models.URL{}, fmt.Errorf("failed to generate token: %w", err)
	}

	// Создаем новую запись
	newURL := models.URL{
		OriginalURL: originalURL,
		ShortKey:    token,
		CreatedAt:   time.Now(),
	}

	createdURL, err := s.storage.CreateOrUpdate(ctx, newURL)
	if err != nil {
		if errors.Is(err, ErrConflict) {
			existing, err := s.storage.GetByOriginalURL(ctx, originalURL)
			if err != nil {
				return models.URL{}, fmt.Errorf("%w: %v", ErrConflict, err)
			}
			return existing, ErrConflict
		}
		return models.URL{}, fmt.Errorf("failed to create URL: %w", err)
	}

	return createdURL, nil
}

// BatchCreate создает несколько коротких ссылок за одну операцию
func (s *URLShortener) BatchCreate(ctx context.Context, urls []models.URL) ([]models.URL, error) {
	if len(urls) == 0 {
		return nil, ErrInvalidData
	}

	// Проверяем существующие URL
	originalURLs := make([]string, len(urls))
	for i, url := range urls {
		originalURLs[i] = url.OriginalURL
	}

	existingURLs, err := s.storage.ExistsBatch(ctx, originalURLs)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing URLs: %w", err)
	}

	existingMap := make(map[string]models.URL)
	for _, url := range existingURLs {
		existingMap[url.OriginalURL] = url
	}

	var (
		urlsToCreate []models.URL
		result       []models.URL
		allExist     = true
	)

	// Формируем результат для существующих URL
	for _, url := range urls {
		if existingURL, exists := existingMap[url.OriginalURL]; exists {
			result = append(result, existingURL)
		} else {
			// Генерируем короткий ключ для новых URL
			token, err := s.generateUniqueToken(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to generate token: %w", err)
			}
			url.ShortKey = token
			url.CreatedAt = time.Now()
			urlsToCreate = append(urlsToCreate, url)
			allExist = false
		}
	}

	// Если все URL уже существуют, возвращаем конфликт
	if allExist {
		return result, ErrConflict
	}

	// Создаем новые URL
	createdURLs, err := s.storage.BatchCreate(ctx, urlsToCreate)
	if err != nil {
		return nil, fmt.Errorf("failed to create URLs: %w", err)
	}

	return append(result, createdURLs...), nil
}

// PingDataBase проверяет соединение с хранилищем
func (s *URLShortener) PingDataBase(ctx context.Context) error {
	if err := s.storage.Ping(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	return nil
}

func (s *URLShortener) ListURLs(ctx context.Context, limit, offset int) ([]models.URL, error) {
	return s.storage.List(ctx, limit, offset)
}

const (
	maxAttempts  = 3
	tokenLength  = 8
	tokenLetters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

func (s *URLShortener) generateUniqueToken(ctx context.Context) (string, error) {
	for i := 0; i < maxAttempts; i++ {
		token := generateRandomToken()
		_, err := s.storage.GetByShortKey(ctx, token)

		if err != nil {
			if errors.Is(err, ErrUnfound) {
				return token, nil
			}
			return "", err
		}
	}

	return "", errors.New("failed to generate unique token after several attempts")
}

func generateRandomToken() string {
	b := make([]byte, tokenLength)
	letterCount := big.NewInt(int64(len(tokenLetters)))

	for i := range b {
		n, _ := rand.Int(rand.Reader, letterCount)
		b[i] = tokenLetters[n.Int64()]
	}
	return string(b)
}
