package main

import (
	"urlshortener/internal/config"
	"urlshortener/internal/logger"
	"urlshortener/internal/server"
	"urlshortener/internal/service"
	"urlshortener/internal/storage/inmemory"
)

func main() {
	// Config может содержать:
	// - ListenPort (порт сервера, например "localhost:8080")
	// - BaseURL (базовый URL для коротких ссылок, например "http://localhost:8080")

	/*
		// можно будет еще добавить:

		// - LogLevel (уровень логирования)
		// - StorageType (тип хранилища: "inmemory", "postgres", etc.)
		// - и другие параметры

	*/

	cfg := config.LoadConfig()

	storage := inmemory.NewInMemoryStorage()
	svc := service.NewURLShortenerService(storage)
	logger := logger.GetLogger()

	srv, err := server.NewServer(
		cfg,
		logger,
		svc,
	)

	if err != nil {
		logger.Fatal().Err(err)
	}

	srv.Start()
}
