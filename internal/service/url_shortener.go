package service

import (
	"strconv"
	"urlshortener/internal/storage"
)

type URLshortener struct {
	storage storage.InMemoryStorage
	BaseURL string
}

func NewURLshortener(mem storage.InMemoryStorage, url string) URLshortener {
	return URLshortener{
		storage: mem,
		BaseURL: url,
	}
}

func (s *URLshortener) GetURL(id string) (string, error) {
	key, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return "", err
	}

	url, err := s.storage.Get(key)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (s *URLshortener) SetURL(url string) (string, error) {
	key, err := s.storage.Set(url)
	if err != nil {
		return "", err
	}

	id := strconv.FormatUint(key, 10)
	return id, nil
}
