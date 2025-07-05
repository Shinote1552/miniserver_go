package inmemory

import "urlshortener/internal/models"

const initLastID = 0

type MemoryStorage struct {
	data   map[string]models.URL
	lastID int
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data:   make(map[string]models.URL),
		lastID: initLastID,
	}
}

func (m *MemoryStorage) Set(shortURL, originalURL string) (*models.URL, error) {
	if shortURL == "" || originalURL == "" {
		return nil, models.ErrInvalidData
	}

	m.lastID++
	url := models.URL{
		ID:          m.lastID,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}

	m.data[shortURL] = url
	return &url, nil
}

func (m *MemoryStorage) Get(shortURL string) (*models.URL, error) {
	url, exists := m.data[shortURL]
	if !exists {
		return nil, models.ErrNotFound
	}
	return &url, nil
}

func (m *MemoryStorage) GetAll() ([]models.URL, error) {
	if len(m.data) == 0 {
		return nil, models.ErrEmpty
	}

	urls := make([]models.URL, 0, len(m.data))
	for _, url := range m.data {
		urls = append(urls, url)
	}

	return urls, nil
}

func (m *MemoryStorage) Close() error {
	return nil
}
