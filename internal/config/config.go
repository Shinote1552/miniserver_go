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
)

const (
	defaultServerAddress   = "localhost:8080"
	defaultBaseURL         = "http://localhost:8080"
	defaultFileStoragePath = "tmp/short-url-db.json"
)

type Config struct {
	ServerAddress   string
	BaseURL         string
	FileStoragePath string
}

func NewConfig() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.ServerAddress, "a", defaultServerAddress, "Server address")
	flag.StringVar(&cfg.BaseURL, "b", defaultBaseURL, "Base URL")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "File storage path (default: "+defaultFileStoragePath+")")
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
}

func (c *Config) resolveFilePath() string {
	// Если путь не задан явно, используем относительный путь по умолчанию
	if c.FileStoragePath == "" {
		c.FileStoragePath = defaultFileStoragePath
	}

	// Если путь уже абсолютный (начинается с /), возвращаем как есть
	if filepath.IsAbs(c.FileStoragePath) {
		return c.FileStoragePath
	}

	// Для относительных путей - делаем абсолютным относительно директории запуска
	absPath, err := filepath.Abs(c.FileStoragePath)
	if err != nil {
		return c.FileStoragePath // возвращаем как есть в случае ошибки
	}
	return absPath
}

func (c *Config) normalizeServerAddress() {
	if strings.HasPrefix(c.ServerAddress, ":") {
		c.ServerAddress = "localhost" + c.ServerAddress
	}
}
