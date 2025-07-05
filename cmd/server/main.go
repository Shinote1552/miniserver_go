package main

import (
	"fmt"
	"urlshortener/internal/config"
	"urlshortener/internal/logger"
	"urlshortener/internal/server"
	"urlshortener/internal/service"
	"urlshortener/internal/storage/filestore"
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
	mem := inmemory.NewMemoryStorage()

	if err := filestore.Load(cfg.FileStoragePath, mem); err != nil {
		log.Fatal().Err(err).Msg("Failed to load data from file")
	}

	fmt.Println("\n\nDEBUG", cfg.FileStoragePath)

	defer func() {

		fmt.Println("\n\nDEBUG: defer is done\n")

	}()

	// Гарантированное сохранение при завершении
	defer func() {
		if err := filestore.Save(cfg.FileStoragePath, mem); err != nil {
			log.Error().Err(err).Msg("Failed to save data")
		} else {
			log.Info().Msg("Data successfully saved")
		}
	}()

	defer func() {

		fmt.Println("\n\nDEFER IS WORKING?\n\n")
	}()
	svc := service.NewServiceURLShortener(mem)

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
