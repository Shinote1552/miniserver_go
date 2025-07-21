package service

import (
	"context"
	"crypto/rand"
	"math/big"
	"urlshortener/internal/models"
)

type Storage interface {
	Set(string, string) (*models.URL, error)
	Get(string) (*models.URL, error)
	GetAll(ctx context.Context) ([]models.URL, error)
	PingDataBase(context.Context) error
}

type ServiceURLShortener struct {
	storage Storage
}

func NewServiceURLShortener(mem Storage) *ServiceURLShortener {
	return &ServiceURLShortener{
		storage: mem,
	}
}

func (s *ServiceURLShortener) GetURL(token string) (string, error) {
	url, err := s.storage.Get(token)
	if err != nil {
		return "", err
	}
	return url.OriginalURL, nil
}

func (s *ServiceURLShortener) SetURL(ctx context.Context, originalURL string) (string, error) {
	allURLs, err := s.storage.GetAll(ctx)
	if err != nil && err != models.ErrEmpty {
		return "", err
	}

	for _, url := range allURLs {
		if url.OriginalURL == originalURL {
			return url.ShortURL, nil
		}
	}

	token := s.tokenGenerator()

	_, err = s.storage.Set(token, originalURL)
	if err != nil {
		return "", err
	}

	return token, nil
}

const (
	tokenGeneratorLength  = 8
	tokenGeneratorLetters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

func (s *ServiceURLShortener) tokenGenerator() string {
	token := make([]byte, 0, tokenGeneratorLength)
	initRandRange := big.NewInt(int64(len(tokenGeneratorLetters)))
	for i := 0; i < tokenGeneratorLength; i++ {
		num, err := rand.Int(rand.Reader, initRandRange)
		if err != nil {
			return ""
		}

		index := num.Int64()
		symbol := tokenGeneratorLetters[index]
		token = append(token, symbol)
	}

	return string(token)
}

func (s *ServiceURLShortener) PingDataBase(ctx context.Context) error {
	return s.storage.PingDataBase(ctx)
}
