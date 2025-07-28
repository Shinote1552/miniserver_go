package service

import (
	"context"
	"crypto/rand"
	"errors"
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

func (s *URLShortener) BatchCreate(ctx context.Context, batchItems []models.APIBatchRequestItem) ([]models.APIBatchResponseItem, error) {
	return s.storage.BatchCreate(ctx, batchItems)
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
