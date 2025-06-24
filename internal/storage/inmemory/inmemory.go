package inmemory

import (
	"errors"
)

var (
	SetErr           = errors.New("key or value is not valid")
	GetErr           = errors.New("value is not exist")
	GenerateTokenErr = errors.New("failed to generate new token")
)

const (
	initLastID uint64 = 0
)

type urlshortener struct {
	url   string
	token string
	id    uint64
}

func newurlshortener(url, token string, id uint64) *urlshortener {
	return &urlshortener{
		url:   url,
		token: token,
		id:    id,
	}
}

type InMemoryStorage struct {
	mem     map[string]urlshortener
	lastKey string
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		mem:     make(map[string]urlshortener),
		lastKey: "",
	}
}

func (i *InMemoryStorage) Set(key string, value string) error {

	if value == "" || key == "" {
		return SetErr
	}

	var id uint64
	object, ok := i.mem[i.lastKey]
	if ok {
		id += object.id
	} else {
		id = initLastID
	}

	id++
	i.lastKey = key
	us := newurlshortener(value, key, id)
	i.mem[key] = *us
	return nil
}

func (i *InMemoryStorage) Get(token string) (string, error) {
	us, exists := i.mem[token]
	if !exists {
		return "", GetErr
	}

	url := us.url
	return url, nil
}

func (i *InMemoryStorage) GetAll() ([]string, error) {
	if len(i.mem) == 0 {
		return nil, errors.New("storage si empty")
	}

	urls := make([]string, 0, len(i.mem))
	for _, v := range i.mem {
		if v.url == "" {
			continue
		}
		urls = append(urls, v.url)
	}

	if len(urls) == 0 {
		return nil, errors.New("all stored URLs are empty")
	}

	return urls, nil
}
