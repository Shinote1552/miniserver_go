// Самостоятельный пакет?

package config // 1. Сначала пакет

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type ServerConfig struct { // 2. Название
	ServerAddr string // TODO: Addr
	BaseURL    string
}

const (
	defaultAddr    string = "localhost:8080"
	defaultBaseURL string = "http://localhost:8080"
)

func LoadConfig() ServerConfig {
	var cfg ServerConfig

	// Читаем переменные окружения
	envAddr := os.Getenv("SERVER_ADDRESS") // fixme: в константы
	envBaseURL := os.Getenv("BASE_URL")    // fixme: в костаны

	fmt.Println("SERVER_ADDRESS=", envAddr)
	fmt.Println("BASE_URL=", envBaseURL)

	// Парсим флаги командной строки
	flag.StringVar(&cfg.ServerAddr, "a", defaultAddr, "HTTP server address")
	flag.StringVar(&cfg.BaseURL, "b", defaultBaseURL, "Base URL for shortened links")
	flag.Parse()

	// Применяем приоритет: env vars > flags > defaults
	if envAddr != "" {
		cfg.ServerAddr = envAddr
	}
	if envBaseURL != "" {
		cfg.BaseURL = envBaseURL
	}
	
	// Автоматически добавляем localhost если указан только порт
	if strings.HasPrefix(cfg.ServerAddr, ":") {
		cfg.ServerAddr = "localhost" + cfg.ServerAddr
	}

	return cfg
}
