package config

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
)

const (
	envServerAddress   = "SERVER_ADDRESS"
	envBaseURL         = "BASE_URL"
	envFileStoragePath = "FILE_STORAGE_PATH"
	envDatabaseDSN     = "DATABASE_DSN"
)

const (
	defaultServerAddress       = "localhost:8080"
	defaultBaseURL             = "http://localhost:8080"
	defaultFilestoreStorageDir = "tmp"
	defaultStorageFile         = "short-url-db.json"
	defaultDatabaseDSN         = "postgres://postgres:admin@localhost:5432/gpx_test?sslmode=disable"
)

type Config struct {
	ServerAddress   string
	BaseURL         string
	FileStoragePath string
	DatabaseDSN     string
}

func NewConfig() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.ServerAddress, "a", defaultServerAddress, "Server address")
	flag.StringVar(&cfg.BaseURL, "b", defaultBaseURL, "Base URL")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "File storage path (default: "+filepath.Join(defaultFilestoreStorageDir, defaultStorageFile)+")")
	flag.StringVar(&cfg.DatabaseDSN, "d", defaultDatabaseDSN, "Database DSN")
	flag.Parse()

	cfg.applyEnvOverrides()
	cfg.FileStoragePath = cfg.resolveFilePath()
	cfg.normalizeServerAddress()

	return cfg
}

func (c *Config) applyEnvOverrides() {
	if envAddr := os.Getenv(envServerAddress); envAddr != "" {
		c.ServerAddress = envAddr
	}
	if envURL := os.Getenv(envBaseURL); envURL != "" {
		c.BaseURL = envURL
	}
	if envPath := os.Getenv(envFileStoragePath); envPath != "" {
		c.FileStoragePath = envPath
	}
	if envDSN := os.Getenv(envDatabaseDSN); envDSN != "" {
		c.DatabaseDSN = envDSN
	}
}

func (c *Config) resolveFilePath() string {
	// Если путь не задан явно, формируем путь по умолчанию
	if c.FileStoragePath == "" {
		wd, err := os.Getwd()
		if err != nil {
			// В случае ошибки возвращаем относительный путь
			return filepath.Join(defaultFilestoreStorageDir, defaultStorageFile)
		}
		return filepath.Join(wd, defaultFilestoreStorageDir, defaultStorageFile)
	}

	// Если путь уже абсолютный, возвращаем как есть
	if filepath.IsAbs(c.FileStoragePath) {
		return c.FileStoragePath
	}

	// Для относительных путей делаем абсолютным относительно рабочей директории
	absPath, err := filepath.Abs(c.FileStoragePath)
	if err != nil {
		// В случае ошибки возвращаем как есть
		return c.FileStoragePath
	}
	return absPath
}

func (c *Config) normalizeServerAddress() {
	if strings.HasPrefix(c.ServerAddress, ":") {
		c.ServerAddress = "localhost" + c.ServerAddress
	}
}
