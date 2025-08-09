package services

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"
	"urlshortener/domain/models"
)

// Storage - основной интерфейс хранилища URL и shotURL
type Storage interface {
	// Основные CRUD операции
	CreateOrUpdate(ctx context.Context, url models.ShortenedLink) (models.ShortenedLink, error)
	GetByShortKey(ctx context.Context, shortKey string) (models.ShortenedLink, error)
	GetByLongURL(ctx context.Context, originalURL string) (models.ShortenedLink, error)
	Delete(ctx context.Context, shortKey string) error

	// Пакетные операции
	BatchCreate(ctx context.Context, urls []models.ShortenedLink) ([]models.ShortenedLink, error)

	// BatchGetByUserID(ctx context.Context, UUID string) ([]models.ShortenedLink, error)

	ExistsBatch(ctx context.Context, originalURLs []string) ([]models.ShortenedLink, error)

	Exists(ctx context.Context, originalURL string) (models.ShortenedLink, error)

	// Пагинация/листинг
	List(ctx context.Context, limit, offset int) ([]models.ShortenedLink, error)

	// Управление соединением
	Ping(ctx context.Context) error
	Close() error

	// only for experimental build
	GetAll(ctx context.Context) ([]models.ShortenedLink, error)
}

// URLShortener реализует бизнес-логику сервиса сокращения URL
type URLShortener struct {
	storage Storage
	baseURL string
}

// NewServiceURLShortener создает новый экземпляр сервиса
func NewServiceURLShortener(storage Storage, baseURL string) *URLShortener {
	return &URLShortener{
		storage: storage,
		baseURL: baseURL,
	}
}

// GetURL возвращает оригинальный URL по короткому ключу
func (s *URLShortener) GetURL(ctx context.Context, shortKey string) (models.ShortenedLink, error) {
	if shortKey == "" {
		return models.ShortenedLink{}, models.ErrInvalidData
	}

	url, err := s.storage.GetByShortKey(ctx, shortKey)
	if err != nil {
		if errors.Is(err, models.ErrUnfound) {
			return models.ShortenedLink{}, fmt.Errorf("%w: URL not found", models.ErrUnfound)
		}
		return models.ShortenedLink{}, fmt.Errorf("failed to get URL: %w", err)
	}
	return url, nil
}

// GetShortURL возвращает полный короткий URL
func (s *URLShortener) GetShortURL(shortKey string) string {
	return fmt.Sprintf("%s/%s", s.baseURL, shortKey)
}

// SetURL создает новую короткую ссылку или возвращает существующую
func (s *URLShortener) SetURL(ctx context.Context, longUrl string) (models.ShortenedLink, error) {
	if longUrl == "" {
		return models.ShortenedLink{}, models.ErrInvalidData
	}

	// Проверяем существование URL
	existing, err := s.storage.GetByLongURL(ctx, longUrl)
	if err == nil {
		return existing, models.ErrConflict
	}

	// Генерируем уникальный токен
	token, err := s.generateUniqueToken(ctx)
	if err != nil {
		return models.ShortenedLink{}, fmt.Errorf("failed to generate token: %w", err)
	}

	// Создаем новую запись
	newURL := models.ShortenedLink{
		LongURL:   longUrl,
		ShortCode: token,
		CreatedAt: time.Now(),
	}

	createdURL, err := s.storage.CreateOrUpdate(ctx, newURL)
	if err != nil {
		if errors.Is(err, models.ErrConflict) {
			existing, err := s.storage.GetByLongURL(ctx, longUrl)
			if err != nil {
				return models.ShortenedLink{}, fmt.Errorf("%w: %v", models.ErrConflict, err)
			}
			return existing, models.ErrConflict
		}
		return models.ShortenedLink{}, fmt.Errorf("failed to create URL: %w", err)
	}

	return createdURL, nil
}

// BatchCreate создает несколько коротких ссылок за одну операцию
func (s *URLShortener) BatchCreate(ctx context.Context, urls []models.ShortenedLink) ([]models.ShortenedLink, error) {
	if len(urls) == 0 {
		return nil, models.ErrInvalidData
	}

	// Проверяем существующие URL
	longUrls := make([]string, len(urls))
	for i, url := range urls {
		longUrls[i] = url.LongURL
	}

	existingURLs, err := s.storage.ExistsBatch(ctx, longUrls)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing URLs: %w", err)
	}

	existingMap := make(map[string]models.ShortenedLink)
	for _, url := range existingURLs {
		existingMap[url.LongURL] = url
	}

	var (
		urlsToCreate []models.ShortenedLink
		result       []models.ShortenedLink
		allExist     = true
	)

	// Формируем результат для существующих URL
	for _, url := range urls {
		if existingURL, exists := existingMap[url.LongURL]; exists {
			result = append(result, existingURL)
		} else {
			// Генерируем короткий ключ для новых URL
			token, err := s.generateUniqueToken(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to generate token: %w", err)
			}
			url.ShortCode = token
			url.CreatedAt = time.Now()
			urlsToCreate = append(urlsToCreate, url)
			allExist = false
		}
	}

	// Если все URL уже существуют, возвращаем конфликт
	if allExist {
		return result, models.ErrConflict
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

func (s *URLShortener) ListURLs(ctx context.Context, limit, offset int) ([]models.ShortenedLink, error) {
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
			if errors.Is(err, models.ErrUnfound) {
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
