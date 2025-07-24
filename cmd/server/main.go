package main

import (
	"context"
	"os"
	"os/signal"
	"urlshortener/internal/config"
	"urlshortener/internal/logger"
	"urlshortener/internal/server"
	"urlshortener/internal/service"
	"urlshortener/repository"
	"urlshortener/repository/filestore"
	"urlshortener/repository/inmemory"
	"urlshortener/repository/postgres"

	"github.com/rs/zerolog"
)

func main() {
	// ctxRoot := context.Background()
	// cfg := initConfig()
	// log := initLogger()

	// // storage := inmemory.NewMemoryStorage()
	// storage := initStorage(ctxRoot, cfg, log)

	// defer handleStorageDefer(ctxRoot, cfg, storage, log)

	// svc := initService(storage)
	// srv := initServer(cfg, log, svc)

	// stop := make(chan os.Signal, 1)
	// signal.Notify(stop, os.Interrupt)

	// go func() {
	// 	if err := srv.Start(ctxRoot); err != nil {
	// 		log.Fatal().Err(err).Msg("Server failed")
	// 	}
	// }()

	// <-stop
	// log.Info().Msg("SIGINT (Ctrl+C) or SIGTERM shutdownning server...")

	ctxRoot := context.Background()
	cfg := config.NewConfig()
	log := logger.NewLogger()

	storage := initStorage(ctxRoot, log, *cfg)
	defer closeStorage(log, storage)

	initData(ctxRoot, log, *cfg, storage)
	defer saveData(ctxRoot, log, *cfg, storage)

	service := service.NewServiceURLShortener(storage)
	srv, err := server.NewServer(log, *cfg, service)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create server")
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		if err := srv.Start(ctxRoot); err != nil {
			log.Fatal().Err(err).Msg("Server failed")
		}
	}()

	<-stop
	log.Info().Msg("SIGINT (Ctrl+C) or SIGTERM shutdownning server...")

}

func initStorage(ctx context.Context, log *zerolog.Logger, cfg config.Config) repository.Storage {
	if cfg.DatabaseDSN != "" {
		storage, err := postgres.NewStorage(ctx, cfg.DatabaseDSN)
		if err != nil {
			log.Fatal().
				Err(err).
				Msg("Failed to initialize PostgreSQL storage")
		}
		log.Info().
			Msg("Using PostgreSQL storage")
		return storage

	}
	log.Info().
		Msg("Using in-memory storage")
	storage := inmemory.NewStorage()
	return storage
}

func closeStorage(log *zerolog.Logger, storage repository.Storage) {
	if err := storage.Close(); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to close storage")
		return
	}
	log.Info().
		Msg("Storage closed successfully")
}

func initData(ctx context.Context, log *zerolog.Logger, cfg config.Config, storage repository.Storage) {
	if cfg.FileStoragePath == "" {
		log.Info().Msg("No file storage path specified, skip loading data")
		return
	}

	path, err := filestore.Load(ctx, *log, cfg.FileStoragePath, storage)
	if err != nil {
		log.Error().
			Err(err).
			Str("path", cfg.FileStoragePath).
			Msg("Failed to load data from file")
		return
	}

	log.Info().
		Str("path", path).
		Msg("Data loaded successfully from file")
}

func saveData(ctx context.Context, log *zerolog.Logger, cfg config.Config, storage repository.Storage) {
	if cfg.FileStoragePath == "" {
		log.Info().Msg("No file storage path specified, skip saving data")
		return
	}

	path, err := filestore.Save(ctx, log, cfg.FileStoragePath, storage)
	if err != nil {
		log.Error().
			Err(err).
			Str("path", path).
			Msg("Failed to save data to file")
		return
	}

	log.Info().
		Str("path", path).
		Msg("Data saved successfully to file")
}
