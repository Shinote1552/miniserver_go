package inmemory

import (
	"errors"
)

var (
	SetErr = errors.New("nil url")
	GetErr = errors.New("not found")
	// GetAllErr = errors.New("empty memory ib DB")
)

const (
	initLastID uint64 = 0
)

type urlshortener struct {
	url string
	id  uint64
}

func newurlshortener(url string, id uint64) *urlshortener {
	return &urlshortener{
		url: url,
		id:  id,
	}
}

type InMemory struct {
	mem    map[uint64]urlshortener
	lastID uint64
}

func NewInMemory() *InMemory {
	return &InMemory{
		mem:    make(map[uint64]urlshortener),
		lastID: initLastID,
	}
}

func (i *InMemory) Set(url string) (uint64, error) {

	if url == "" {
		return 0, SetErr
	}

	i.lastID++
	key := i.lastID

	us := newurlshortener(url, key)
	i.mem[key] = *us

	return key, nil
}

func (i *InMemory) Get(id uint64) (string, error) {
	key := id
	us, exists := i.mem[key]
	if !exists {
		return "", GetErr
	}

	url := us.url
	return url, nil
}

// func (i *InMemory) GetAll() ([]string, error) {
// 	if len(i.mem) == 0 {
// 		return nil, GetAllErr
// 	}
// 	allByName := make([]string, 0, len(i.mem))

// 	for _, db := range i.mem {
// 		allByName = append(allByName, db)
// 	}

// 	return allByName, nil
// }
