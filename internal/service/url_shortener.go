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
