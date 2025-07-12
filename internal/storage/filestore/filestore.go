package filestore

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"urlshortener/internal/models"
)

var (
	ErrInvalidDir = errors.New("using bad path to the save data")
	ErrEmpty      = errors.New("file is empty")
)

// StorageInterface определяет методы работы с хранилищем БД который есть у Сервера
type StorageInterface interface {
	Set(string, string) (*models.URL, error)
	Get(string) (*models.URL, error)
	GetAll() ([]models.URL, error)
}

func Load(filePath string, storage StorageInterface) (string, error) {
	if filePath == "" {
		return "No file path provided - using empty storage", nil
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Проверяем существование файла
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		// Файла нет - создаем пустой файл и возвращаем сообщение
		dir := filepath.Dir(absPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		file, err := os.OpenFile(absPath, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return "", fmt.Errorf("failed to create file %s: %w", absPath, err)
		}
		file.Close()
		return fmt.Sprintf("Storage file %s created - starting with empty storage", absPath), nil
	}

	reader, err := newFileReader(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", absPath, err)
	}
	defer reader.close()

	var loadedCount int
	for {
		url, err := reader.readURL()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("failed to read URL from file: %w", err)
		}

		if _, err := storage.Set(url.ShortURL, url.OriginalURL); err != nil {
			return "", fmt.Errorf("failed to set URL in storage: %w", err)
		}
		loadedCount++
	}

	if loadedCount > 0 {
		return fmt.Sprintf("Successfully loaded %d URLs from %s", loadedCount, absPath), nil
	}
	return fmt.Sprintf("No data loaded from %s (file exists but empty)", absPath), nil
}

func Save(filePath string, storage StorageInterface) (string, error) {
	if filePath == "" {
		return "", ErrInvalidDir
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", ErrInvalidDir
	}

	dir := filepath.Dir(absPath)

	// Создаем директорию, если ее нет (кросс-платформенный способ)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return dir, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	writer, err := newFileWriter(absPath)
	if err != nil {
		return dir, err
	}
	defer writer.close()

	urls, err := storage.GetAll()
	if err != nil {
		return dir, err
	}

	for _, url := range urls {
		if err := writer.writeURL(&url); err != nil {
			return dir, err
		}
	}

	return dir, nil
}

// fileWriter реализует запись данных в файл *Producer*
type fileWriter struct {
	file   *os.File
	writer *bufio.Writer
}

// newFileWriter создает новый fileWriter
func newFileWriter(filePath string) (*fileWriter, error) {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}

	return &fileWriter{
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}

// writeURL записывает URL в файл
func (w *fileWriter) writeURL(url *models.URL) error {
	data, err := json.Marshal(url)
	if err != nil {
		return err
	}

	if _, err := w.writer.Write(data); err != nil {
		return err
	}

	if err := w.writer.WriteByte('\n'); err != nil {
		return err
	}

	return w.writer.Flush()
}

// close закрывает файл
func (w *fileWriter) close() error {
	if err := w.writer.Flush(); err != nil {
		return err
	}
	return w.file.Close()
}

// fileReader реализует чтение данных из файла *Consumer*
type fileReader struct {
	file   *os.File
	reader *bufio.Reader
}

// newFileReader создает новый fileReader
func newFileReader(filePath string) (*fileReader, error) {
	file, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &fileReader{
		file:   file,
		reader: bufio.NewReader(file),
	}, nil
}

// readURL читает URL из файла
func (r *fileReader) readURL() (*models.URL, error) {
	data, err := r.reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	var url models.URL
	if err := json.Unmarshal(data, &url); err != nil {
		return nil, err
	}

	return &url, nil
}

// close закрывает файл
func (r *fileReader) close() error {
	return r.file.Close()
}
