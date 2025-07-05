package filebased

import (
	"encoding/json"
	"os"
	"urlshortener/internal/models"
)

type StorageInterface interface {
	Set(string, string) (*models.URL, error)
	Get(string) (*models.URL, error)
	GetAll() ([]models.URL, error)
}

type FileStorage struct {
	memory   StorageInterface
	file     *os.File
	encoder  *json.Encoder
	filePath string
}

func NewFileStorage(memory StorageInterface, filePath string) (*FileStorage, error) {
	fs := &FileStorage{
		memory:   memory,
		filePath: filePath,
	}

	if filePath != "" {
		if err := fs.initFile(); err != nil {
			return nil, err
		}
	}

	return fs, nil
}

func (fs *FileStorage) initFile() error {
	if _, err := os.Stat(fs.filePath); err == nil {
		file, err := os.Open(fs.filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		decoder := json.NewDecoder(file)
		for decoder.More() {
			var url models.URL
			if err := decoder.Decode(&url); err != nil {
				return err
			}
			fs.memory.Set(url.ShortURL, url.OriginalURL)
		}
	}

	file, err := os.OpenFile(fs.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	fs.file = file
	fs.encoder = json.NewEncoder(file)
	return nil
}

func (fs *FileStorage) Set(shortURL, originalURL string) (*models.URL, error) {
	url, err := fs.memory.Set(shortURL, originalURL)
	if err != nil {
		return nil, err
	}

	if err := fs.encoder.Encode(url); err != nil {
		return nil, err
	}

	return url, nil
}

func (fs *FileStorage) Get(shortURL string) (*models.URL, error) {
	return fs.memory.Get(shortURL)
}

func (fs *FileStorage) GetAll() ([]models.URL, error) {
	return fs.memory.GetAll()
}

func (fs *FileStorage) Close() error {
	if fs.file != nil {
		return fs.file.Close()
	}
	return nil
}
