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

	cfg := config.NewConfig()
	log := logger.GetLogger()
	storage := inmemory.NewMemoryStorage()

	svc := service.NewServiceURLShortener(storage)

	srv, err := server.NewServer(
		cfg,
		log,
		svc,
	)

	if err != nil {
		log.Fatal().Err(err)
	}

	if err := srv.Start(); err != nil {
		log.Fatal().Err(err).Msg("Server failed")
	}
}
