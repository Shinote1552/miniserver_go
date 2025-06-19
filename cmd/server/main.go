package main

import (
	"urlshortener/internal/config"
	"urlshortener/internal/server"
	"urlshortener/internal/storage/inmemory"
)

func main() {
	addr := "localhost:8080"
	baseURL := "http://" + addr

	mem := inmemory.NewInMemory()

	cfg := config.NewServerConfig(addr, baseURL)
	srv := server.NewServer(mem, cfg)

	srv.Start()
}
