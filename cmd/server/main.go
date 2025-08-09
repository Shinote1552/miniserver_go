package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"urlshortener/domain/services"
	"urlshortener/internal/config"
	"urlshortener/internal/http/server"
	"urlshortener/internal/logger"
	"urlshortener/internal/repository/filestore"
	"urlshortener/internal/repository/inmemory"
	"urlshortener/internal/repository/postgres"

	"github.com/rs/zerolog"
)

func main() {
	ctxRoot := context.Background()
	cfg := config.NewConfig()
	log := logger.NewLogger()

	storage := initStorage(ctxRoot, log, *cfg)
	defer closeStorage(log, storage)

	initData(ctxRoot, log, *cfg, storage)
	defer saveData(ctxRoot, log, *cfg, storage)
	service := services.NewServiceURLShortener(storage, cfg.BaseURL)

	srv, err := server.NewServer(log, *cfg, service)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to create server")
	}

	runServer(srv, log)

}

func runServer(srv *server.Server, log *zerolog.Logger) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := srv.Start(context.Background()); err != nil && err != http.ErrServerClosed {
			log.Fatal().
				Err(err).
				Msg("Server failed to start")
		}
	}()

	// Shutdown, TODO gracefull shutdown
	sig := <-stop
	log.Info().
		Str("signal", sig.String()).
		Msg("Received signal, shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server shutdown error")
	} else {
		log.Info().Msg("Server shutdown completed successfully")
	}
}

func initPostgres(ctx context.Context, log *zerolog.Logger, cfg config.Config) *postgres.PostgresStorage {
	if cfg.DatabaseDSN != "" {
		storage, err := postgres.NewStorage(ctx, cfg.DatabaseDSN)
		if err != nil {
			log.Error().
				Err(err).
				Msg("Failed to initialize PostgreSQL storage")

			log.Info().
				Msg("Falling back to in-memory storage")
			storage := inmemory.NewStorage()
			return storage
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

func initInMemory() *inmemory.InmemoryStorage {

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
