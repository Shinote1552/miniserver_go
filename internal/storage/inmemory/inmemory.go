package inmemory

import (
	"crypto/rand"
	"errors"
	"math/big"
)

const (
	tokenGeneratorLength   = 8                                                                // fixme: можешь ближе к методу расположить
	lettersToGenerateToken = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz" // fixme: сохрани иерархию
)

var (
	SetErr           = errors.New("nil url") // нейминг странный
	GenerateTokenErr = errors.New("failed to generate new token")
	GetErr           = errors.New("not found") // fixme: посмотри различие неймингу между 4xx и 5xx
	// GetAllErr = errors.New("empty memory ib DB")
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

type InMemory struct {
	mem     map[string]urlshortener
	lastKey string
}

func NewInMemory() *InMemory { // // fixme: storage ?
	return &InMemory{
		mem:     make(map[string]urlshortener),
		lastKey: "",
	}
}

func (i *InMemory) generateToken() string {
	token := make([]byte, 0, tokenGeneratorLength)
	initRandRange := big.NewInt(int64(len(lettersToGenerateToken)))
	for i := 0; i < tokenGeneratorLength; i++ {
		num, err := rand.Int(rand.Reader, initRandRange)
		if err != nil {
			return ""
		}

		index := num.Int64()
		symbol := lettersToGenerateToken[index]
		token = append(token, symbol)
	}

	return string(token)
}

func (i *InMemory) Set(url string) (string, error) { // fixme: Сущеостей будет больше, Назоави более явно: SetURL, к примеру

	if url == "" {
		return "", SetErr
	}

	if i.mem[i.lastKey].url == url {
		return i.mem[i.lastKey].token, nil
	}

	for _, object := range i.mem {
		if object.url == url {
			return object.token, nil
		}
	}

	var id uint64
	object, ok := i.mem[i.lastKey]
	if ok {
		id += object.id
	} else {
		id = initLastID
	}

	token := i.generateToken()
	if token == "" {
		return "", GenerateTokenErr
	}
	id++
	i.lastKey = token
	us := newurlshortener(url, token, id)
	i.mem[token] = *us
	return token, nil
}

func (i *InMemory) Get(token string) (string, error) { // fixme: GetToken
	us, exists := i.mem[token]
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
