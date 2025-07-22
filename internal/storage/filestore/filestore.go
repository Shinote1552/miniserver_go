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
	"urlshortener/internal/models"

	"github.com/rs/zerolog"
)

// Error variables
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

// StorageInterface defines storage methods
type StorageInterface interface {
	Set(context.Context, string, string) (*models.URL, error)
	Get(context.Context, string) (*models.URL, error)
	GetAll(context.Context) ([]models.URL, error)
}

func Load(ctx context.Context, filePath string, storage StorageInterface, log zerolog.Logger) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	if filePath == "" {
		log.Info().Msg("No file path provided - using empty storage")
		return "No file path provided - using empty storage", nil
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get absolute path")
		return "", fmt.Errorf("%w: %v", ErrAbsPath, err)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		dir := filepath.Dir(absPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Error().Err(err).Str("dir", dir).Msg("Failed to create directory")
			return "", fmt.Errorf("%w: %v", ErrCreateDir, err)
		}

		file, err := os.OpenFile(absPath, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Error().Err(err).Str("path", absPath).Msg("Failed to create file")
			return "", fmt.Errorf("%w: %v", ErrCreateFile, err)
		}
		file.Close()

		msg := fmt.Sprintf("Storage file %s created - starting with empty storage", absPath)
		log.Info().Str("path", absPath).Msg(msg)
		return msg, nil
	}

	// Check if file is empty
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		log.Error().Err(err).Str("path", absPath).Msg("Failed to get file info")
		return "", fmt.Errorf("%w: %v", ErrOpenFile, err)
	}

	if fileInfo.Size() == 0 {
		msg := fmt.Sprintf("File %s is empty - starting with empty storage", absPath)
		log.Info().Str("path", absPath).Msg(msg)
		return msg, nil
	}

	reader, err := newFileReader(absPath)
	if err != nil {
		log.Error().Err(err).Str("path", absPath).Msg("Failed to open file")
		return "", fmt.Errorf("%w: %v", ErrOpenFile, err)
	}
	defer reader.close()

	var loadedCount int
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		url, err := reader.readURL()
		if err != nil {
			if err == io.EOF || errors.Is(err, ErrEmptyFile) {
				break
			}
			log.Error().Err(err).Msg("Failed to read URL from file")
			return "", fmt.Errorf("%w: %v", ErrReadURL, err)
		}

		if _, err := storage.Set(ctx, url.ShortURL, url.OriginalURL); err != nil {
			if errors.Is(err, models.ErrConflict) {
				log.Info().Str("short_url", url.ShortURL).Msg("Skipping duplicate URL")
				continue
			}
			log.Error().Err(err).Msg("Failed to set URL in storage")
			return "", fmt.Errorf("%w: %v", ErrSetURL, err)
		}
		loadedCount++
	}

	if loadedCount > 0 {
		msg := fmt.Sprintf("Successfully loaded %d URLs from %s", loadedCount, absPath)
		log.Info().Int("count", loadedCount).Str("path", absPath).Msg(msg)
		return msg, nil
	}

	msg := fmt.Sprintf("No data loaded from %s (file exists but empty)", absPath)
	log.Info().Str("path", absPath).Msg(msg)
	return msg, nil
}

func Save(ctx context.Context, filePath string, storage StorageInterface, log zerolog.Logger) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	if filePath == "" {
		log.Error().Msg("Invalid directory path")
		return "", ErrInvalidDir
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		log.Error().Err(err).Msg("Invalid directory path")
		return "", ErrInvalidDir
	}

	dir := filepath.Dir(absPath)
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Error().Err(err).Str("dir", dir).Msg("Failed to create directory structure")
			return dir, fmt.Errorf("%w: %v", ErrMkdirAll, err)
		}
	}

	writer, err := newFileWriter(absPath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create file writer")
		return dir, fmt.Errorf("%w: %v", ErrNewFileWriter, err)
	}
	defer writer.close()

	urls, err := storage.GetAll(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get all URLs")
		return dir, fmt.Errorf("%w: %v", ErrGetAllURLs, err)
	}

	for _, url := range urls {
		select {
		case <-ctx.Done():
			return dir, ctx.Err()
		default:
			if err := writer.writeURL(&url); err != nil {
				log.Error().Err(err).Msg("Failed to write URL to file")
				return dir, fmt.Errorf("%w: %v", ErrWriteURL, err)
			}
		}
	}

	log.Info().Str("dir", dir).Msg("Data successfully saved")
	return dir, nil
}

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
