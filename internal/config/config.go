package config

import (
	"flag"
	"os"
	"strings"
)

const (
	EnvListenPort = "SERVER_ADDRESS"
	EnvBaseURL    = "BASE_URL"
)

const (
	defaultListenPort string = "localhost:8080"
	defaultBaseURL    string = "http://localhost:8080"
)

type ServerConfig struct {
	ListenPort string
	BaseURL    string
}

func LoadConfig() ServerConfig {
	var cfg ServerConfig

	// Читаем переменные окружения
	envAddr := os.Getenv(EnvListenPort)
	envBaseURL := os.Getenv(EnvBaseURL)

	// Парсим флаги командной строки
	flag.StringVar(&cfg.ListenPort, "a", defaultListenPort, "HTTP server address")
	flag.StringVar(&cfg.BaseURL, "b", defaultBaseURL, "Base URL for shortened links")
	flag.Parse()

	// Применяем приоритет: env vars > flags > defaults
	if envAddr != "" {
		cfg.ListenPort = envAddr
	}
	if envBaseURL != "" {
		cfg.BaseURL = envBaseURL
	}

	// Автоматически добавляем localhost если указан только порт
	if strings.HasPrefix(cfg.ListenPort, ":") {
		cfg.ListenPort = "localhost" + cfg.ListenPort
	}

	return cfg
}
