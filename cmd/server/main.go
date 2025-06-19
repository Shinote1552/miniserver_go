package main

import (
	"urlshortener/internal/config"
	"urlshortener/internal/server"
	"urlshortener/internal/storage/inmemory"
)

func main() {
	// addr := "localhost:8080"
	// baseURL := "http://" + addr
	// cfg := config.NewServerConfig(addr, baseURL)

	mem := inmemory.NewInMemory()

	cfg := config.LoadConfig()

	srv := server.NewServer(mem, cfg.ServerAddr)

	srv.Start()
}
