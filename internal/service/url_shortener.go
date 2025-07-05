package service

import (
	"crypto/rand"
	"math/big"
	"urlshortener/internal/models"
	"urlshortener/internal/storage/filestore"

	"github.com/rs/zerolog/log"
)

type StorageInterface interface {
	Set(string, string) (*models.URL, error)
	Get(string) (*models.URL, error)
	GetAll() ([]models.URL, error)
}

type ServiceURLShortener struct {
	storage StorageInterface
}

func NewServiceURLShortener(mem StorageInterface) *ServiceURLShortener {
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

func (s *ServiceURLShortener) SetURL(originalURL string) (string, error) {
	allURLs, err := s.storage.GetAll()
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
	if err := filestore.Save("/tmp/short-url-db.json", s.storage); err != nil {
		log.Error().Err(err).Msg("Failed to save after update")
	} else {
		log.Info().Msg("Data updated and save in /tmp/short-url-db.json")

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
