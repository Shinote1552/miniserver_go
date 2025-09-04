package filestore

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	ErrInvalidDir   = errors.New("invalid directory path")
	ErrEmptyFile    = errors.New("file is empty")
	ErrAbsPath      = errors.New("failed to get absolute path")
	ErrCreateDir    = errors.New("failed to create directory")
	ErrCreateFile   = errors.New("failed to create file")
	ErrOpenFile     = errors.New("failed to open file")
	ErrReadURL      = errors.New("failed to read URL from file")
	ErrSetURL       = errors.New("failed to set URL in storage")
	ErrMkdirAll     = errors.New("failed to create directory structure")
	ErrGetAllURLs   = errors.New("failed to get all URLs")
	ErrWriteURL     = errors.New("failed to write URL to file")
	ErrMarshalURL   = errors.New("failed to marshal URL")
	ErrWriteData    = errors.New("failed to write data")
	ErrWriteNewLine = errors.New("failed to write new line")
)

// StorageInterface - ограниченный интерфейс для работы с filestore
type StorageInterface interface {
	ShortenedLinkCreate(ctx context.Context, url models.ShortenedLink) (models.ShortenedLink, error)
	ShortenedLinkGetByShortKey(ctx context.Context, shortKey string) (models.ShortenedLink, error)
	List(ctx context.Context, limit, offset int) ([]models.ShortenedLink, error)
}

// Load loads URLs from file into storage and returns whether file was empty
func Load(ctx context.Context, log zerolog.Logger, filePath string, storage StorageInterface) (string, bool, error) {
	if err := ctx.Err(); err != nil {
		return "", false, logError(log, err, "context error")
	}

	if filePath == "" {
		msg := "No file path provided - using empty storage"
		log.Info().Msg(msg)
		return msg, true, nil // Файла нет = считаем его пустым
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", false, logAndWrapError(log, err, ErrAbsPath, "get absolute path")
	}

	// Проверяем существование файла
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		// Файл не существует - создаем директорию и пустой файл
		dir := filepath.Dir(absPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", false, logAndWrapError(log, err, ErrCreateDir, "create directory")
		}

		file, err := os.Create(absPath)
		if err != nil {
			return "", false, logAndWrapError(log, err, ErrCreateFile, "create file")
		}
		file.Close()

		msg := fmt.Sprintf("File %s created as empty", absPath)
		log.Info().Msg(msg)
		return msg, true, nil
	}

	// Проверяем, пуст ли файл
	isEmpty, err := isFileEmpty(absPath)
	if err != nil {
		return "", false, logAndWrapError(log, err, ErrOpenFile, "check file size")
	}

	if isEmpty {
		msg := fmt.Sprintf("File %s is empty", absPath)
		log.Info().Msg(msg)
		return msg, true, nil
	}

	// Загружаем данные из файла
	loadedCount, err := loadURLsFromFile(ctx, absPath, storage, log)
	if err != nil {
		return "", false, err
	}

	msg := fmt.Sprintf("Successfully loaded %d URLs from %s", loadedCount, absPath)
	log.Info().Int("count", loadedCount).Str("path", absPath).Msg(msg)
	return msg, false, nil // Файл не был пустым
}

// Save saves URLs from storage to file
func Save(ctx context.Context, log *zerolog.Logger, filePath string, storage StorageInterface) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", logError(*log, err, "context error")
	}

	if filePath == "" {
		return "", logError(*log, ErrInvalidDir, "invalid directory path")
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", logAndWrapError(*log, err, ErrInvalidDir, "get absolute path")
	}

	dir := filepath.Dir(absPath)
	if err := createDirectoryIfNotExists(ctx, dir, *log); err != nil {
		return dir, err
	}

	if err := writeURLsToFile(ctx, absPath, storage, *log); err != nil {
		return dir, err
	}

	msg := fmt.Sprintf("Data successfully saved to %s", absPath)
	log.Info().Str("path", absPath).Msg(msg)
	return msg, nil
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

// Helper functions for file operations
func isFileEmpty(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.Size() == 0, nil
}

func createDirectoryIfNotExists(ctx context.Context, dir string, log zerolog.Logger) error {
	if err := ctx.Err(); err != nil {
		return logError(log, err, "context error")
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return logAndWrapError(log, err, ErrMkdirAll, "create directory structure")
	}
	return nil
}

// Core loading and saving functions
func loadURLsFromFile(ctx context.Context, filePath string, storage StorageInterface, log zerolog.Logger) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, logAndWrapError(log, err, ErrOpenFile, "open file")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	loadedCount := 0

	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return 0, logError(log, err, "context error")
		}

		data := scanner.Bytes()
		if len(data) == 0 {
			continue
		}

		var url models.ShortenedLink
		if err := json.Unmarshal(data, &url); err != nil {
			log.Warn().Err(err).Msg("Failed to unmarshal URL, skipping line")
			continue
		}

		if err := storeURL(ctx, &url, storage, log); err != nil {
			return 0, err
		}
		loadedCount++
	}

	if err := scanner.Err(); err != nil {
		return 0, logAndWrapError(log, err, ErrReadURL, "read file")
	}

	return loadedCount, nil
}

func storeURL(ctx context.Context, url *models.ShortenedLink, storage StorageInterface, log zerolog.Logger) error {
	_, err := storage.ShortenedLinkCreate(ctx, *url)
	if err == nil {
		return nil
	}

	if errors.Is(err, ErrConflict) {
		log.Info().Str("short_url", url.ShortCode).Msg("Skipping duplicate URL")
		return nil
	}

	return logAndWrapError(log, err, ErrSetURL, "set URL in storage")
}

func writeURLsToFile(ctx context.Context, filePath string, storage StorageInterface, log zerolog.Logger) error {
	file, err := os.Create(filePath)
	if err != nil {
		return logAndWrapError(log, err, ErrCreateFile, "create file")
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	urls, err := storage.List(ctx, 1000000, 0)
	if err != nil {
		return logAndWrapError(log, err, ErrGetAllURLs, "get URLs from storage")
	}

	for _, url := range urls {
		if err := ctx.Err(); err != nil {
			return logError(log, err, "context error")
		}

		data, err := json.Marshal(url)
		if err != nil {
			return logAndWrapError(log, err, ErrMarshalURL, "marshal URL")
		}

		if _, err := writer.Write(data); err != nil {
			return logAndWrapError(log, err, ErrWriteData, "write data")
		}

		if err := writer.WriteByte('\n'); err != nil {
			return logAndWrapError(log, err, ErrWriteNewLine, "write newline")
		}
	}

	return writer.Flush()
}
