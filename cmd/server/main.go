package main

import (
	"urlshortener/internal/server"
	"urlshortener/internal/storage/inmemory"
)

func main() {
	storage := inmemory.NewInMemory()
	srv := server.NewServer(storage, ":8080")

	srv.Start(":8080")
}
