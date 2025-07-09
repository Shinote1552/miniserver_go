package main

import (
	"os"
	"os/signal"
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

	if loadDir, err := filestore.Load(cfg.FileStoragePath, mem); err != nil {
		log.Warn().Err(err).Msg("Failed to load data from file" + loadDir)
	} else {
		log.Info().Msg("Data successfully loaded from: " + loadDir)
	}

	// Гарантированное сохранение при завершении
	defer func() {
		if saveDir, err := filestore.Save(cfg.FileStoragePath, mem); err != nil {
			log.Fatal().Err(err).Msg("Failed to save data in: " + saveDir)
		} else {
			log.Info().Msg("Data successfully saved in: " + saveDir)
		}
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

	// Простой обработчик прерывания
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		if err := srv.Start(); err != nil {
			log.Fatal().Err(err).Msg("Server failed")
		}
	}()

	// Ожидание сигнала прерывания
	<-stop
	log.Info().Msg("SIGINT (Ctrl+C) or SIGTERM shutdownning server...")
}
