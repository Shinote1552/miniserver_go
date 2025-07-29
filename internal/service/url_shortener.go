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
	existingURL, err := s.storage.GetByOriginalURL(ctx, originalURL)
	if err != nil && !errors.Is(err, models.ErrUnfound) {
		return "", err
	}

	if existingURL != nil {
		return existingURL.ShortURL, nil
	}

	token, err := s.generateUniqueToken(ctx)
	if err != nil {
		return "", err
	}

	createdURL, err := s.storage.CreateOrUpdate(ctx, token, originalURL)
	if err != nil {
		return "", err
	}

	return createdURL.ShortURL, nil
}

func (s *URLShortener) BatchCreate(ctx context.Context, batchItems []models.APIShortenRequestBatch) ([]models.APIShortenResponseBatch, error) {
	if len(batchItems) == 0 {
		return nil, fmt.Errorf("%w: empty batch", models.ErrInvalidData)
	}

	validatedRequestBatch := make([]string, 0, len(batchItems))
	for _, item := range batchItems {
		if item.CorrelationID == "" || item.OriginalURL == "" {
			return nil, fmt.Errorf("%w: correlation_id и original_url обязательны", models.ErrInvalidData)
		}
		// если хотябы одна keyval корректная то добавляем
		validatedRequestBatch = append(validatedRequestBatch, item.OriginalURL)
	}

	/*
		Разделение существущих URL от не существующих в БД
	*/

	// Проверяем существующие URL в хранилище
	existingURLs, err := s.storage.ExistsBatch(ctx, validatedRequestBatch)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing URLs: %w", err)
	}

	// Создаем мапу для быстрого поиска существующих URL
	existingMap := make(map[string]string)
	for _, url := range existingURLs {
		existingMap[url.OriginalURL] = url.ShortURL
	}

	// Разделяем URL на существующие и новые
	var itemsToCreate []models.APIShortenRequestBatch
	response := make([]models.APIShortenResponseBatch, 0, len(batchItems))

	for _, item := range batchItems {
		if shortURL, exists := existingMap[item.OriginalURL]; exists {
			// URL уже существует - добавляем в ответ
			response = append(response, models.APIShortenResponseBatch{
				CorrelationID: item.CorrelationID,
				ShortURL:      shortURL,
			})
		} else {
			// URL нужно создать
			itemsToCreate = append(itemsToCreate, item)
		}
	}

	// Если все URL уже существуют, возвращаем результат
	if len(itemsToCreate) == 0 {
		return response, nil
	}

	/*
		Обработка новых и корректных URL
	*/

	tokenToCorrelation := make(map[string]string)
	var storageBatch []models.APIShortenRequestBatch

	for _, item := range itemsToCreate {
		token, err := s.generateUniqueToken(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to generate token: %w", err)
		}
		tokenToCorrelation[token] = item.CorrelationID
		storageBatch = append(storageBatch, models.APIShortenRequestBatch{
			CorrelationID: token, // Используем токен как временный correlation_id
			OriginalURL:   item.OriginalURL,
		})
	}

	// Сохраняем новые URL одной транзакцией
	createdItems, err := s.storage.BatchCreate(ctx, storageBatch)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch in storage: %w", err)
	}

	// Сопоставляем созданные URL с исходными correlation_id
	for _, createdItem := range createdItems {
		response = append(response, models.APIShortenResponseBatch{
			CorrelationID: tokenToCorrelation[createdItem.CorrelationID],
			ShortURL:      createdItem.ShortURL,
		})
	}

	return response, nil

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
