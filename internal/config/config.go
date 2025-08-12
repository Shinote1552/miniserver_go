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
	defaultFileStoragePath     = "short-url-db.json"
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

	// Initialize with defaults
	*cfg = Config{
		ServerAddress:    defaultServerAddress,
		BaseURL:          defaultBaseURL,
		FileStoragePath:  defaultFileStoragePath,
		DatabaseDSN:      defaultDatabaseDSN,
		JWTAccessExpire:  defaultJWTAccessExpire,
		JWTRefreshExpire: defaultJWTRefreshExpire,
	}

	// Parse flags
	flag.StringVar(&cfg.ServerAddress, "server-address", cfg.ServerAddress, "Server address")
	flag.StringVar(&cfg.BaseURL, "base-url", cfg.BaseURL, "Base URL")
	flag.StringVar(&cfg.FileStoragePath, "file-storage-path", cfg.FileStoragePath, "File storage path")
	flag.StringVar(&cfg.DatabaseDSN, "database-dsn", cfg.DatabaseDSN, "Database DSN")
	flag.DurationVar(&cfg.JWTAccessExpire, "jwt-access-expire", cfg.JWTAccessExpire, "JWT access token expiration")
	flag.DurationVar(&cfg.JWTRefreshExpire, "jwt-refresh-expire", cfg.JWTRefreshExpire, "JWT refresh token expiration")
	flag.Parse()

	// Apply environment variables
	cfg.applyEnv("SERVER_ADDRESS", &cfg.ServerAddress)
	cfg.applyEnv("BASE_URL", &cfg.BaseURL)
	cfg.applyEnv("FILE_STORAGE_PATH", &cfg.FileStoragePath)
	cfg.applyEnv("DATABASE_DSN", &cfg.DatabaseDSN)
	cfg.applyEnv("JWT_SECRET_KEY", &cfg.JWTSecretKey)
	cfg.applyEnvDuration("JWT_ACCESS_EXPIRE", &cfg.JWTAccessExpire)
	cfg.applyEnvDuration("JWT_REFRESH_EXPIRE", &cfg.JWTRefreshExpire)

	// Final setup
	cfg.validateJWTSecret()
	cfg.FileStoragePath = cfg.resolveFilePath()
	cfg.normalizeServerAddress()

	return cfg
}

func (c *Config) applyEnv(key string, target *string) {
	if val, ok := os.LookupEnv(key); ok {
		*target = val
	}
}

func (c *Config) applyEnvDuration(key string, target *time.Duration) {
	if val, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(val); err == nil {
			*target = d
		}
	}
}

func (c *Config) validateJWTSecret() {
	if c.JWTSecretKey == "" {
		// Generate random key for development
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			panic("failed to generate JWT secret key")
		}
		c.JWTSecretKey = base64.StdEncoding.EncodeToString(key)
		fmt.Println("WARNING: Using auto-generated JWT secret key. For production, set JWT_SECRET_KEY environment variable.")
	}

	// Validate key length
	if _, err := base64.StdEncoding.DecodeString(c.JWTSecretKey); err != nil || len(c.JWTSecretKey) < 32 {
		panic("JWT secret key must be at least 32 bytes long (base64 encoded)")
	}
}

func (c *Config) resolveFilePath() string {
	if filepath.IsAbs(c.FileStoragePath) {
		return c.FileStoragePath
	}

	absPath, err := filepath.Abs(c.FileStoragePath)
	if err != nil {
		return filepath.Clean(c.FileStoragePath)
	}
	return absPath
}

func (c *Config) normalizeServerAddress() {
	if strings.HasPrefix(c.ServerAddress, ":") {
		c.ServerAddress = "localhost" + c.ServerAddress
	}
}
