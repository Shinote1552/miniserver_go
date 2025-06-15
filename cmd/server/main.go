package main

import (
	"urlshortener/internal/server"
	"urlshortener/internal/storage/inmemory"
)

func main() {
	baseURL := ":8080"

	storage := inmemory.NewInMemory()
	srv := server.NewServer(storage, baseURL)

	srv.Start(baseURL)
}
