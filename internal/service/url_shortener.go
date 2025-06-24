package service

import (
	"crypto/rand"
	"errors"
	"math/big"
	"urlshortener/internal/deps"
)

type URLShortenerService struct {
	storage deps.InMemoryStorage
}

func NewURLShortenerService(mem deps.InMemoryStorage) *URLShortenerService {
	return &URLShortenerService{
		storage: mem,
	}
}

func (s *URLShortenerService) GetURL(token string) (string, error) {
	url, err := s.storage.Get(token)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (s *URLShortenerService) SetURL(url string) (string, error) {
	allURL, err := s.storage.GetAll()

	if err == nil {
		for _, item := range allURL {
			if item == url {
				return "", errors.New("url is already exist")
			}
		}
	}

	token := s.tokenGenerator()

	if err := s.storage.Set(token, url); err != nil {
		return "", err
	}

	return token, nil
}

const (
	tokenGeneratorLength  = 8
	tokenGeneratorLetters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

func (s *URLShortenerService) tokenGenerator() string {
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
