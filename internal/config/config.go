package config

import (
	"flag"
	"strings"
)

type ServerConfig struct {
	ServerAddr string // flag -a
	BaseURL    string // flag -b
}

func NewServerConfig(serverAddr, baseURL string) ServerConfig {
	return ServerConfig{
		ServerAddr: serverAddr,
		BaseURL:    baseURL,
	}
}

func LoadServerConfigCLI() ServerConfig {
	var cfg ServerConfig

	flag.StringVar(&cfg.ServerAddr, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "Base URL for shortened links")
	flag.Parse()

	if strings.HasPrefix(cfg.ServerAddr, ":") {
		cfg.ServerAddr = "localhost" + cfg.ServerAddr
	}

	return cfg
}
