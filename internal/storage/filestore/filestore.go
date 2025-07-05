package filestore

import (
	"bufio"
	"encoding/json"
	"os"
	"urlshortener/internal/models"
)

// StorageInterface определяет методы работы с хранилищем
type StorageInterface interface {
	Set(string, string) (*models.URL, error)
	Get(string) (*models.URL, error)
	GetAll() ([]models.URL, error)
}

// Load загружает данные из файла в хранилище
func Load(filePath string, storage StorageInterface) error {
	if filePath == "" {
		return nil
	}

	reader, err := newFileReader(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer reader.close()

	for {
		url, err := reader.readURL()
		if err != nil {
			// EOF означает корректное завершение файла
			if err.Error() == "EOF" {
				break
			}
			return err
		}

		if _, err := storage.Set(url.ShortURL, url.OriginalURL); err != nil {
			return err
		}
	}

	return nil
}

// Save сохраняет данные из хранилища в файл
func Save(filePath string, storage StorageInterface) error {
	if filePath == "" {
		return nil
	}

	writer, err := newFileWriter(filePath)
	if err != nil {
		return err
	}
	defer writer.close()

	urls, err := storage.GetAll()
	if err != nil {
		return err
	}

	for _, url := range urls {
		if err := writer.writeURL(&url); err != nil {
			return err
		}
	}

	return nil
}

// fileWriter реализует запись данных в файл *Producer*
type fileWriter struct {
	file   *os.File
	writer *bufio.Writer
}

// newFileWriter создает новый fileWriter
func newFileWriter(filePath string) (*fileWriter, error) {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
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
