package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"urlshortener/internal/models"
	"urlshortener/repository"
)

type URLShortener struct {
	storage repository.Storage
}

func NewServiceURLShortener(storage repository.Storage) *URLShortener {
	return &URLShortener{storage: storage}
}

func (s *URLShortener) GetURL(ctx context.Context, shortURL string) (string, error) {
	urlModel, err := s.storage.GetByShortURL(ctx, shortURL)
	if err != nil {
		if errors.Is(err, models.ErrUnfound) {
			return "", models.ErrUnfound
		}
		return "", err
	}
	return urlModel.OriginalURL, nil
}
func (s *URLShortener) SetURL(ctx context.Context, originalURL string) (string, error) {
	// Сначала проверяем существование URL
	existing, err := s.storage.GetByOriginalURL(ctx, originalURL)
	if err == nil && existing != nil {
		return existing.ShortURL, models.ErrConflict
	}

	token, err := s.generateUniqueToken(ctx)
	if err != nil {
		return "", err
	}

	_, err = s.storage.CreateOrUpdate(ctx, token, originalURL)
	if err != nil {
		if errors.Is(err, models.ErrConflict) {
			existing, err := s.storage.GetByOriginalURL(ctx, originalURL)
			if err != nil {
				return "", fmt.Errorf("%w: %v", models.ErrConflict, err)
			}
			return existing.ShortURL, models.ErrConflict
		}
		return "", err
	}

	return token, nil
}

func (s *URLShortener) BatchCreate(ctx context.Context, batchItems []models.APIShortenRequestBatch) ([]models.APIShortenResponseBatch, error) {
	if len(batchItems) == 0 {
		return nil, fmt.Errorf("%w: empty batch", models.ErrInvalidData)
	}

	// Валидация элементов
	for _, item := range batchItems {
		if item.CorrelationID == "" || item.OriginalURL == "" {
			return nil, fmt.Errorf("%w: correlation_id и original_url обязательны", models.ErrInvalidData)
		}
	}

	// Проверяем существующие URL
	existingURLs, err := s.storage.ExistsBatch(ctx, getOriginalURLs(batchItems))
	if err != nil {
		return nil, fmt.Errorf("failed to check existing URLs: %w", err)
	}

	existingMap := make(map[string]string)
	for _, url := range existingURLs {
		existingMap[url.OriginalURL] = url.ShortURL
	}

	var (
		itemsToCreate []models.APIShortenRequestBatch
		response      []models.APIShortenResponseBatch
		allExist      = true
	)

	// Формируем ответ для существующих URL
	for _, item := range batchItems {
		if shortURL, exists := existingMap[item.OriginalURL]; exists {
			response = append(response, models.APIShortenResponseBatch{
				CorrelationID: item.CorrelationID,
				ShortURL:      shortURL,
			})
		} else {
			itemsToCreate = append(itemsToCreate, item)
			allExist = false
		}
	}

	// Если все URL уже существуют, возвращаем конфликт
	if allExist {
		return response, models.ErrConflict
	}

	// Создаем новые URL
	newItems, err := s.createNewBatchItems(ctx, itemsToCreate)
	if err != nil {
		return nil, err
	}

	return append(response, newItems...), nil
}

func getOriginalURLs(items []models.APIShortenRequestBatch) []string {
	urls := make([]string, len(items))
	for i, item := range items {
		urls[i] = item.OriginalURL
	}
	return urls
}

func (s *URLShortener) createNewBatchItems(ctx context.Context, items []models.APIShortenRequestBatch) ([]models.APIShortenResponseBatch, error) {
	tokenToCorrelation := make(map[string]string)
	var storageBatch []models.APIShortenRequestBatch

	for _, item := range items {
		token, err := s.generateUniqueToken(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to generate token: %w", err)
		}
		tokenToCorrelation[token] = item.CorrelationID
		storageBatch = append(storageBatch, models.APIShortenRequestBatch{
			CorrelationID: token,
			OriginalURL:   item.OriginalURL,
		})
	}

	createdItems, err := s.storage.BatchCreate(ctx, storageBatch)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch in storage: %w", err)
	}

	result := make([]models.APIShortenResponseBatch, len(createdItems))
	for i, item := range createdItems {
		result[i] = models.APIShortenResponseBatch{
			CorrelationID: tokenToCorrelation[item.CorrelationID],
			ShortURL:      item.ShortURL,
		}
	}

	return result, nil
}

func (s *URLShortener) PingDataBase(ctx context.Context) error {
	return s.storage.Ping(ctx)
}

func (s *URLShortener) ListURLs(ctx context.Context, limit, offset int) ([]models.StorageURLModel, error) {
	return s.storage.List(ctx, limit, offset)
}

const (
	tokenLength  = 8
	tokenLetters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

func (s *URLShortener) generateUniqueToken(ctx context.Context) (string, error) {
	const maxAttempts = 3
	for i := 0; i < maxAttempts; i++ {
		token := generateRandomToken()
		_, err := s.storage.GetByShortURL(ctx, token)
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
