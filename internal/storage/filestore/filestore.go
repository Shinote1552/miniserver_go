package filestore

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"urlshortener/internal/models"
)

// StorageInterface определяет методы работы с хранилищем
type StorageInterface interface {
	Set(string, string) (*models.URL, error)
	Get(string) (*models.URL, error)
	GetAll() ([]models.URL, error)
}

// Load загружает данные из файла в хранилище и возвращает информационное сообщение при успехе
func Load(filePath string, storage StorageInterface) (string, error) {
	if filePath == "" {
		return "No file path provided - using empty storage", nil
	}

	// Получаем абсолютный путь для сообщения
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}

	reader, err := newFileReader(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Sprintf("Storage file %s not found - starting with empty storage", absPath), nil
		}
		return "", err
	}
	defer reader.close()

	var loadedCount int
	for {
		url, err := reader.readURL()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return "", err
		}

		if _, err := storage.Set(url.ShortURL, url.OriginalURL); err != nil {
			return "", err
		}
		loadedCount++
	}

	if loadedCount > 0 {
		return fmt.Sprintf("Successfully loaded %d URLs from %s", loadedCount, absPath), nil
	}
	return fmt.Sprintf("No data loaded from %s (file exists but empty)", absPath), nil
}

// Save сохраняет данные из хранилища в файл и возвращает путь к директории
func Save(filePath string, storage StorageInterface) (string, error) {
	if filePath == "" {
		return "", nil
	}

	// Получаем абсолютный путь к директории
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(absPath)

	// Создаем директорию, если ее нет
	if err := os.MkdirAll(dir, 0755); err != nil {
		return dir, err
	}

	writer, err := newFileWriter(filePath)
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
