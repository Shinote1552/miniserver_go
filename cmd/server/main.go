package main

import (
	"os"
	"os/signal"
	"urlshortener/internal/config"
	"urlshortener/internal/logger"
	"urlshortener/internal/server"
	"urlshortener/internal/service"
	"urlshortener/internal/storage/filestore"
	"urlshortener/internal/storage/postgres"

	"github.com/rs/zerolog"
)

func main() {
	cfg := initConfig()
	log := initLogger()
	storage := initStorage(cfg, log)
	defer handleStorageDefer(cfg, storage, log)

	svc := initService(storage)
	srv := initServer(cfg, log, svc)

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

func initConfig() *config.Config {
	return config.NewConfig()
}

func initLogger() zerolog.Logger {
	return *logger.GetLogger()
}

func initStorage(cfg *config.Config, log zerolog.Logger) *postgres.PostgresStorage {
	storage, err := postgres.NewPostgresStorage(cfg.DatabaseDSN)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database storage")
	}

	if loadDir, err := filestore.Load(cfg.FileStoragePath, storage); err != nil {
		log.Warn().Err(err).Msg("Failed to load data from file" + loadDir)
	} else {
		log.Info().Msg("Data successfully loaded from: " + loadDir)
	}

	return storage
}

func handleStorageDefer(cfg *config.Config, storage *postgres.PostgresStorage, log zerolog.Logger) {
	if saveDir, err := filestore.Save(cfg.FileStoragePath, storage); err != nil {
		log.Warn().Err(err).Msg("Failed to save data in: " + saveDir)
	} else {
		log.Info().Msg("Data successfully saved in: " + saveDir)
	}
}

func initService(storage *postgres.PostgresStorage) *service.ServiceURLShortener {
	return service.NewServiceURLShortener(storage)
}

func initServer(cfg *config.Config, log zerolog.Logger, svc *service.ServiceURLShortener) *server.Server {
	srv, err := server.NewServer(cfg, &log, svc)
	if err != nil {
		log.Fatal().Err(err)
	}
	return srv
}
