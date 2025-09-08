package filestore

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"urlshortener/internal/domain/models"

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

// FileStore представляет потокобезопасный менеджер файлового хранилища
type FileStore struct {
	mu       sync.Mutex
	filePath string
	log      zerolog.Logger
}

// NewFileStore создает новый экземпляр FileStore
func NewFileStore(log zerolog.Logger, filePath string) *FileStore {
	return &FileStore{
		filePath: filePath,
		log:      log,
	}
}

// StorageInterface - ограниченный интерфейс для работы с filestore
type StorageInterface interface {
	ShortenedLinkCreate(ctx context.Context, url models.ShortenedLink) (models.ShortenedLink, error)
	ShortenedLinkGetByShortKey(ctx context.Context, shortKey string) (models.ShortenedLink, error)
	List(ctx context.Context, limit, offset int) ([]models.ShortenedLink, error)
}

// Load loads URLs from file into storage and returns whether file was empty
func (fs *FileStore) Load(ctx context.Context, storage StorageInterface) (string, bool, error) {

	if err := ctx.Err(); err != nil {
		return "", false, fs.logError(err, "context error")
	}

	if fs.filePath == "" {
		msg := "No file path provided - using empty storage"
		fs.log.Info().Msg(msg)
		return msg, true, nil
	}

	absPath, err := filepath.Abs(fs.filePath)
	if err != nil {
		return "", false, fs.logAndWrapError(err, ErrAbsPath, "get absolute path")
	}

	fs.mu.Lock()
	defer fs.mu.Unlock()

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		dir := filepath.Dir(absPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", false, fs.logAndWrapError(err, ErrCreateDir, "create directory")
		}

		file, err := os.Create(absPath)
		if err != nil {
			return "", false, fs.logAndWrapError(err, ErrCreateFile, "create file")
		}
		file.Close()

		msg := fmt.Sprintf("File %s created as empty", absPath)
		fs.log.Info().Msg(msg)
		return msg, true, nil
	}

	isEmpty, err := fs.isFileEmpty(absPath)
	if err != nil {
		return "", false, fs.logAndWrapError(err, ErrOpenFile, "check file size")
	}

	if isEmpty {
		msg := fmt.Sprintf("File %s is empty", absPath)
		fs.log.Info().Msg(msg)
		return msg, true, nil
	}

	loadedCount, err := fs.loadDataFromFile(ctx, absPath, storage)
	if err != nil {
		return "", false, err
	}

	msg := fmt.Sprintf("Successfully loaded %d URLs from %s", loadedCount, absPath)
	fs.log.Info().Int("count", loadedCount).Str("path", absPath).Msg(msg)
	return msg, false, nil
}

// Save saves URLs from storage to file
func (fs *FileStore) Save(ctx context.Context, storage StorageInterface) (string, error) {

	if err := ctx.Err(); err != nil {
		return "", fs.logError(err, "context error")
	}

	if fs.filePath == "" {
		return "", fs.logError(ErrInvalidDir, "invalid directory path")
	}

	absPath, err := filepath.Abs(fs.filePath)
	if err != nil {
		return "", fs.logAndWrapError(err, ErrInvalidDir, "get absolute path")
	}

	fs.mu.Lock()
	defer fs.mu.Unlock()

	dir := filepath.Dir(absPath)
	if err := fs.createDirectoryIfNotExists(ctx, dir); err != nil {
		return dir, err
	}

	if err := fs.writeDataToFile(ctx, absPath, storage); err != nil {
		return dir, err
	}

	msg := fmt.Sprintf("Data successfully saved to %s", absPath)
	fs.log.Info().Str("path", absPath).Msg(msg)
	return msg, nil
}

// Helper functions for error handling
func (fs *FileStore) logError(err error, msg string) error {
	fs.log.Error().Err(err).Msg(msg)
	return err
}

func (fs *FileStore) logAndWrapError(err error, wrapErr error, context string) error {
	fs.log.Error().Err(err).Str("context", context).Msg(wrapErr.Error())
	return fmt.Errorf("%w: %v", wrapErr, err)
}

// Helper functions for file operations
func (fs *FileStore) isFileEmpty(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.Size() == 0, nil
}

func (fs *FileStore) createDirectoryIfNotExists(ctx context.Context, dir string) error {
	if err := ctx.Err(); err != nil {
		return fs.logError(err, "context error")
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fs.logAndWrapError(err, ErrMkdirAll, "create directory structure")
	}
	return nil
}

// Core loading and saving functions
func (fs *FileStore) loadDataFromFile(ctx context.Context, filePath string, storage StorageInterface) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, fs.logAndWrapError(err, ErrOpenFile, "open file")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	loadedCount := 0

	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return 0, fs.logError(err, "context error")
		}

		data := scanner.Bytes()
		if len(data) == 0 {
			continue
		}

		var url models.ShortenedLink
		if err := json.Unmarshal(data, &url); err != nil {
			fs.log.Warn().Err(err).Msg("Failed to unmarshal URL, skipping line")
			continue
		}

		if err := fs.storeURL(ctx, &url, storage); err != nil {
			return 0, err
		}
		loadedCount++
	}

	if err := scanner.Err(); err != nil {
		return 0, fs.logAndWrapError(err, ErrReadURL, "read file")
	}

	return loadedCount, nil
}

func (fs *FileStore) storeURL(ctx context.Context, url *models.ShortenedLink, storage StorageInterface) error {
	_, err := storage.ShortenedLinkCreate(ctx, *url)
	if err == nil {
		return nil
	}

	if errors.Is(err, ErrConflict) {
		fs.log.Info().Str("short_url", url.ShortCode).Msg("Skipping duplicate URL")
		return nil
	}

	return fs.logAndWrapError(err, ErrSetURL, "set URL in storage")
}

func (fs *FileStore) writeDataToFile(ctx context.Context, filePath string, storage StorageInterface) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fs.logAndWrapError(err, ErrCreateFile, "create file")
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// Используем пагинацию для больших объемов данных
	limit := 1000
	offset := 0
	totalWritten := 0

	for {
		if err := ctx.Err(); err != nil {
			return fs.logError(err, "context error")
		}

		urls, err := storage.List(ctx, limit, offset)
		if err != nil {
			return fs.logAndWrapError(err, ErrGetAllURLs, "get URLs from storage")
		}

		if len(urls) == 0 {
			break
		}

		for _, url := range urls {
			data, err := json.Marshal(url)
			if err != nil {
				return fs.logAndWrapError(err, ErrMarshalURL, "marshal URL")
			}

			if _, err := writer.Write(data); err != nil {
				return fs.logAndWrapError(err, ErrWriteData, "write data")
			}

			if err := writer.WriteByte('\n'); err != nil {
				return fs.logAndWrapError(err, ErrWriteNewLine, "write newline")
			}
			totalWritten++
		}

		// Если получили меньше записей, чем лимит, значит это последняя страница
		if len(urls) < limit {
			break
		}

		offset += limit
	}

	fs.log.Info().Int("count", totalWritten).Msg("Total URLs written to file")
	return writer.Flush()
}

// GetFilePath возвращает путь к файлу (только для чтения)
func (fs *FileStore) GetFilePath() string {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	return fs.filePath
}

// SetFilePath устанавливает новый путь к файлу (потокобезопасно)
func (fs *FileStore) SetFilePath(newPath string) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.filePath = newPath
}
