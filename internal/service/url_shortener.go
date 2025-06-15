package service

import "urlshortener/internal/storage"

type URLshortener struct {
	storage storage.InMemoryStorage
	baseURL string
}

func NewURLshortener(mem storage.InMemoryStorage, url string) URLshortener {
	return URLshortener{
		storage: mem,
		baseURL: url,
	}
}

func (s *URLshortener) GetURL(key uint64) (string, error) {
	url, err := s.storage.Get(key)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (s *URLshortener) SetURL(url string) (uint64, error) {
	key, err := s.storage.Set(url)
	if err != nil {
		return 0, err
	}
	return key, nil
}
