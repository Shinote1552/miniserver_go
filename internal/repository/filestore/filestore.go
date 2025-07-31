package filestore

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"urlshortener/domain/models"

	"github.com/rs/zerolog"
)

var (
	ErrInvalidData = errors.New("invalid data")
	ErrUnfound     = errors.New("unfound data")
	ErrEmpty       = errors.New("storage is empty")
	ErrConflict    = errors.New("url already exists with different value")
)

var (
	ErrInvalidDir    = errors.New("invalid directory path")
	ErrEmptyFile     = errors.New("file is empty")
	ErrAbsPath       = errors.New("failed to get absolute path")
	ErrCreateDir     = errors.New("failed to create directory")
	ErrCreateFile    = errors.New("failed to create file")
	ErrOpenFile      = errors.New("failed to open file")
	ErrReadURL       = errors.New("failed to read URL from file")
	ErrSetURL        = errors.New("failed to set URL in storage")
	ErrMkdirAll      = errors.New("failed to create directory structure")
	ErrNewFileWriter = errors.New("failed to create file writer")
	ErrGetAllURLs    = errors.New("failed to get all URLs")
	ErrWriteURL      = errors.New("failed to write URL to file")
	ErrMarshalURL    = errors.New("failed to marshal URL")
	ErrWriteData     = errors.New("failed to write data")
	ErrWriteNewLine  = errors.New("failed to write new line")
	ErrFlushWriter   = errors.New("failed to flush writer")
	ErrCloseFile     = errors.New("failed to close file")
	ErrUnmarshalURL  = errors.New("failed to unmarshal URL")
)

// StorageInterface - ограниченный интерфейс для работы с filestore
type StorageInterface interface {
	CreateOrUpdate(ctx context.Context, url models.URL) (models.URL, error)
	GetByShortKey(ctx context.Context, shortKey string) (models.URL, error)
	List(ctx context.Context, limit, offset int) ([]models.URL, error)
}

// Load loads URLs from file into storage
func Load(ctx context.Context, log zerolog.Logger, filePath string, storage StorageInterface) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", logError(log, err, "context error")
	}

	if filePath == "" {
		return handleEmptyFilePath(log)
	}

	absPath, err := getAbsolutePath(filePath)
	if err != nil {
		return "", logAndWrapError(log, err, ErrAbsPath, "get absolute path")
	}

	if err := ensureFileExists(ctx, absPath, log); err != nil {
		return "", err
	}

	if isEmpty, err := checkFileEmpty(absPath, log); err != nil {
		return "", err
	} else if isEmpty {
		log.Info().Str("path", absPath).Msg("File is empty - starting with empty storage")
		return absPath, nil
	}

	loadedCount, err := loadURLsFromFile(ctx, absPath, storage, log)
	if err != nil {
		return "", err
	}

	return generateLoadResultMessage(loadedCount, absPath, log)
}

// Save saves URLs from storage to file
func Save(ctx context.Context, log *zerolog.Logger, filePath string, storage StorageInterface) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", logError(*log, err, "context error")
	}

	if filePath == "" {
		return "", logError(*log, ErrInvalidDir, "invalid directory path")
	}

	absPath, err := getAbsolutePath(filePath)
	if err != nil {
		return "", logAndWrapError(*log, err, ErrInvalidDir, "get absolute path")
	}

	dir := filepath.Dir(absPath)
	if err := createDirectoryStructure(ctx, dir, *log); err != nil {
		return dir, err
	}

	if err := writeURLsToFile(ctx, absPath, storage, *log); err != nil {
		return dir, err
	}

	log.Info().Str("dir", dir).Msg("Data successfully saved")
	return dir, nil
}

// Helper functions for error handling

func logError(log zerolog.Logger, err error, msg string) error {
	log.Error().Err(err).Msg(msg)
	return err
}

func logAndWrapError(log zerolog.Logger, err error, wrapErr error, context string) error {
	log.Error().Err(err).Str("context", context).Msg(wrapErr.Error())
	return fmt.Errorf("%w: %v", wrapErr, err)
}

// Helper functions for Load

func handleEmptyFilePath(log zerolog.Logger) (string, error) {
	msg := "No file path provided - using empty storage"
	log.Info().Msg(msg)
	return msg, nil
}

func getAbsolutePath(path string) (string, error) {
	return filepath.Abs(path)
}

func ensureFileExists(ctx context.Context, absPath string, log zerolog.Logger) error {
	if _, err := os.Stat(absPath); !os.IsNotExist(err) {
		return nil
	}

	if err := ctx.Err(); err != nil {
		return logError(log, err, "context error")
	}

	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return logAndWrapError(log, err, ErrCreateDir, "create directory")
	}

	return createEmptyFile(absPath, log)
}

func createEmptyFile(absPath string, log zerolog.Logger) error {
	file, err := os.OpenFile(absPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return logAndWrapError(log, err, ErrCreateFile, "create file")
	}
	return file.Close()
}

func checkFileEmpty(absPath string, log zerolog.Logger) (bool, error) {
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		return false, logAndWrapError(log, err, ErrOpenFile, "get file info")
	}
	return fileInfo.Size() == 0, nil
}

func loadURLsFromFile(ctx context.Context, absPath string, storage StorageInterface, log zerolog.Logger) (int, error) {
	reader, err := newFileReader(absPath)
	if err != nil {
		return 0, logAndWrapError(log, err, ErrOpenFile, "open file for reading")
	}
	defer reader.close()

	var loadedCount int
	for {
		if err := ctx.Err(); err != nil {
			return 0, logError(log, err, "context error")
		}

		url, err := reader.readURL()
		if err != nil {
			if err == io.EOF || errors.Is(err, ErrEmptyFile) {
				break
			}
			return 0, logAndWrapError(log, err, ErrReadURL, "read URL from file")
		}

		if err := storeURL(ctx, url, storage, log); err != nil {
			return 0, err
		}
		loadedCount++
	}
	return loadedCount, nil
}

func storeURL(ctx context.Context, url *models.URL, storage StorageInterface, log zerolog.Logger) error {
	_, err := storage.CreateOrUpdate(ctx, *url)
	if err == nil {
		return nil
	}

	if errors.Is(err, ErrConflict) {
		log.Info().Str("short_url", url.ShortKey).Msg("Skipping duplicate URL")
		return nil
	}

	return logAndWrapError(log, err, ErrSetURL, "set URL in storage")
}

func generateLoadResultMessage(loadedCount int, absPath string, log zerolog.Logger) (string, error) {
	if loadedCount > 0 {
		msg := fmt.Sprintf("Successfully loaded %d URLs from %s", loadedCount, absPath)
		log.Info().Int("count", loadedCount).Str("path", absPath).Msg(msg)
		return msg, nil
	}

	msg := fmt.Sprintf("No data loaded from %s (file exists but empty)", absPath)
	log.Info().Str("path", absPath).Msg(msg)
	return msg, nil
}

// Helper functions for Save

func createDirectoryStructure(ctx context.Context, dir string, log zerolog.Logger) error {
	if err := ctx.Err(); err != nil {
		return logError(log, err, "context error")
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return logAndWrapError(log, err, ErrMkdirAll, "create directory structure")
	}
	return nil
}

func writeURLsToFile(ctx context.Context, absPath string, storage StorageInterface, log zerolog.Logger) error {
	writer, err := newFileWriter(absPath)
	if err != nil {
		return logAndWrapError(log, err, ErrNewFileWriter, "create file writer")
	}
	defer writer.close()

	// Используем List с большим лимитом вместо GetAll
	urls, err := storage.List(ctx, 1000000, 0)
	if err != nil {
		return logAndWrapError(log, err, ErrGetAllURLs, "get all URLs from storage")
	}

	for _, url := range urls {
		if err := ctx.Err(); err != nil {
			return logError(log, err, "context error")
		}

		if err := writer.writeURL(&url); err != nil {
			return logAndWrapError(log, err, ErrWriteURL, "write URL to file")
		}
	}
	return nil
}

// File writer implementation (unchanged)
type fileWriter struct {
	file   *os.File
	writer *bufio.Writer
}

func newFileWriter(filePath string) (*fileWriter, error) {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOpenFile, err)
	}

	return &fileWriter{
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}

func (w *fileWriter) writeURL(url *models.URL) error {
	data, err := json.Marshal(url)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrMarshalURL, err)
	}

	if _, err := w.writer.Write(data); err != nil {
		return fmt.Errorf("%w: %v", ErrWriteData, err)
	}

	if err := w.writer.WriteByte('\n'); err != nil {
		return fmt.Errorf("%w: %v", ErrWriteNewLine, err)
	}

	return w.writer.Flush()
}

func (w *fileWriter) close() error {
	if err := w.writer.Flush(); err != nil {
		return fmt.Errorf("%w: %v", ErrFlushWriter, err)
	}
	if err := w.file.Close(); err != nil {
		return fmt.Errorf("%w: %v", ErrCloseFile, err)
	}
	return nil
}

// File reader implementation (unchanged)
type fileReader struct {
	file   *os.File
	reader *bufio.Reader
}

func newFileReader(filePath string) (*fileReader, error) {
	file, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOpenFile, err)
	}

	return &fileReader{
		file:   file,
		reader: bufio.NewReader(file),
	}, nil
}

func (r *fileReader) readURL() (*models.URL, error) {
	data, err := r.reader.ReadBytes('\n')
	if err != nil {
		if err == io.EOF && len(data) == 0 {
			return nil, ErrEmptyFile
		}
		return nil, fmt.Errorf("%w: %v", ErrReadURL, err)
	}

	var url models.URL
	if err := json.Unmarshal(data, &url); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnmarshalURL, err)
	}

	return &url, nil
}

func (r *fileReader) close() error {
	if err := r.file.Close(); err != nil {
		return fmt.Errorf("%w: %v", ErrCloseFile, err)
	}
	return nil
}
