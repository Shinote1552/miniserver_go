package config

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	envServerAddress    = "SERVER_ADDRESS"
	envBaseURL          = "BASE_URL"
	envFileStoragePath  = "FILE_STORAGE_PATH"
	envDatabaseDSN      = "DATABASE_DSN"
	envJWTSecretKey     = "JWT_SECRET_KEY"
	envJWTAccessExpire  = "JWT_ACCESS_EXPIRE"
	envJWTRefreshExpire = "JWT_REFRESH_EXPIRE"
)

const (
	defaultServerAddress       = "localhost:8080"
	defaultBaseURL             = "http://localhost:8080"
	defaultFilestoreStorageDir = "tmp"
	defaultStorageFile         = "short-url-db.json"
	defaultDatabaseDSN         = "postgres://postgres:admin@localhost:5432/gpx_test?sslmode=disable"
	defaultJWTSecretKey        = "YuHiAYxgw4WDdhxduFavo1/202YPUSwbn9AbO0R4dhs="
	defaultJWTAccessExpire     = 15 * time.Minute
	defaultJWTRefreshExpire    = 24 * time.Hour * 7
)

type Config struct {
	ServerAddress    string
	BaseURL          string
	FileStoragePath  string
	DatabaseDSN      string
	JWTSecretKey     string // Минимум 32 байта для HS256
	JWTAccessExpire  time.Duration
	JWTRefreshExpire time.Duration
}

func NewConfig() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.ServerAddress, "a", defaultServerAddress, "Server address")
	flag.StringVar(&cfg.BaseURL, "b", defaultBaseURL, "Base URL")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "File storage path")
	flag.StringVar(&cfg.DatabaseDSN, "d", defaultDatabaseDSN, "Database DSN")
	flag.DurationVar(&cfg.JWTAccessExpire, "jwt-access-expire", defaultJWTAccessExpire, "JWT access token expiration")
	flag.DurationVar(&cfg.JWTRefreshExpire, "jwt-refresh-expire", defaultJWTRefreshExpire, "JWT refresh token expiration")
	flag.Parse()

	cfg.applyEnvOverrides()
	cfg.validateJWTSecret()
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
	if envKey := os.Getenv(envJWTSecretKey); envKey != "" {
		c.JWTSecretKey = envKey
	}
	if envExp := os.Getenv(envJWTAccessExpire); envExp != "" {
		if d, err := time.ParseDuration(envExp); err == nil {
			c.JWTAccessExpire = d
		}
	}
	if envExp := os.Getenv(envJWTRefreshExpire); envExp != "" {
		if d, err := time.ParseDuration(envExp); err == nil {
			c.JWTRefreshExpire = d
		}
	}
}

func (c *Config) validateJWTSecret() {
	if c.JWTSecretKey == "" {
		// Генерируем случайный ключ при запуске (только для разработки!)
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			panic("failed to generate JWT secret key")
		}
		c.JWTSecretKey = base64.StdEncoding.EncodeToString(key)
		fmt.Println("WARNING: Using auto-generated JWT secret key. For production, set JWT_SECRET_KEY environment variable.")
	}

	// Проверяем длину ключа
	keyBytes, err := base64.StdEncoding.DecodeString(c.JWTSecretKey)
	if err != nil || len(keyBytes) < 32 {
		panic("JWT secret key must be at least 32 bytes long (base64 encoded)")
	}
}

func (c *Config) resolveFilePath() string {
	if c.FileStoragePath == "" {
		wd, err := os.Getwd()
		if err != nil {
			return filepath.Join(defaultFilestoreStorageDir, defaultStorageFile)
		}
		return filepath.Join(wd, defaultFilestoreStorageDir, defaultStorageFile)
	}

	if filepath.IsAbs(c.FileStoragePath) {
		return c.FileStoragePath
	}

	absPath, err := filepath.Abs(c.FileStoragePath)
	if err != nil {
		return c.FileStoragePath
	}
	return absPath
}

func (c *Config) normalizeServerAddress() {
	if strings.HasPrefix(c.ServerAddress, ":") {
		c.ServerAddress = "localhost" + c.ServerAddress
	}
}
