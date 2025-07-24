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

func (s *URLShortener) GetURL(ctx context.Context, token string) (string, error) {
	url, err := s.storage.Get(ctx, token)
	if err != nil {
		if errors.Is(err, models.ErrUnfound) {
			return "", models.ErrUnfound
		}
		return "", err
	}
	return url.OriginalURL, nil
}

func (s *URLShortener) SetURL(ctx context.Context, originalURL string) (string, error) {
	exists, shortURL, err := s.storage.Exists(ctx, originalURL)
	if err != nil {
		return "", err
	}
	if exists {
		return shortURL, nil
	}

	// Генерируем уникальный токен
	token, err := s.generateUniqueToken(ctx)
	if err != nil {
		return "", err
	}

	// Сохраняем новую запись
	_, err = s.storage.Set(ctx, token, originalURL)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *URLShortener) PingDataBase(ctx context.Context) error {
	return s.storage.Ping(ctx)
}

const (
	tokenLength  = 8
	tokenLetters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

func (s *URLShortener) generateUniqueToken(ctx context.Context) (string, error) {
	const maxAttempts = 3
	for i := 0; i < maxAttempts; i++ {
		token := generateRandomToken()
		// Проверяем, не существует ли уже такого токена
		_, err := s.storage.Get(ctx, token)
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
