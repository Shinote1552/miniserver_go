package service

import (
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

func (s *URLshortener) GetURL(token string) (string, error) {

	url, err := s.storage.Get(token)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (s *URLshortener) SetURL(url string) (string, error) {
	token, err := s.storage.Set(url)
	if err != nil {
		return "", err
	}

	return token, nil
}
