package config

import (
	"flag"
	"os"
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
	defaultFileStoragePath = "/tmp/short-url-db.json"
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
	flag.StringVar(&cfg.FileStoragePath, "f", defaultFileStoragePath, "File storage path")
	flag.Parse()

	if envAddr := os.Getenv(envServerAddress); envAddr != "" {
		cfg.ServerAddress = envAddr
	}
	if envURL := os.Getenv(envBaseURL); envURL != "" {
		cfg.BaseURL = envURL
	}
	if envPath := os.Getenv(envFileStoragePath); envPath != "" {
		cfg.FileStoragePath = envPath
	}

	if strings.HasPrefix(cfg.ServerAddress, ":") {
		cfg.ServerAddress = "localhost" + cfg.ServerAddress
	}

	return cfg
}
