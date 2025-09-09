package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"urlshortener/internal/config"
	"urlshortener/internal/http/server"
	"urlshortener/internal/logger"
	"urlshortener/internal/repository/filestore"
	"urlshortener/internal/repository/inmemory"
	"urlshortener/internal/repository/postgres"
	"urlshortener/internal/services/auth"
	"urlshortener/internal/services/url_shortener"

	"github.com/rs/zerolog"
)

func main() {
	ctxRoot := context.Background()
	cfg := config.NewConfig()
	log := logger.NewLogger()
	fileStore := filestore.NewFileStore(*log, cfg.FileStoragePath)

	var urlService *url_shortener.URLShortener
	var authService *auth.Authentication

	if cfg.DatabaseDSN != "" {
		storage, err := initPostgres(ctxRoot, log, cfg.DatabaseDSN)
		if err != nil {
			log.
				Error().
				Err(err).
				Msg("Failed to initialize PostgreSQL storage")

		} else {
			defer closePostgresStorage(log, storage)
			defer savePostgresData(ctxRoot, log, storage, fileStore)
			initPostgresData(ctxRoot, log, storage, fileStore)

			var errAuth error
			urlService = url_shortener.NewServiceURLShortener(storage, cfg.BaseURL)
			authService, errAuth = auth.NewAuthentication(storage, cfg.JWTSecretKey, cfg.JWTAccessExpire)
			if errAuth != nil {
				log.
					Error().
					Err(errAuth).
					Msg("Failed to initialize authentication with PostgreSQL")

				urlService = nil
				authService = nil
			}
		}
	}
	if urlService == nil || authService == nil {
		log.
			Info().
			Msg("Using in-memory storage as fallback")

		storage := initInMemory(log)
		defer closeInMemoryStorage(log, storage)
		defer saveInMemoryData(ctxRoot, log, storage, fileStore)
		initInMemoryData(ctxRoot, log, storage, fileStore)

		var errAuth error
		urlService = url_shortener.NewServiceURLShortener(storage, cfg.BaseURL)
		authService, errAuth = auth.NewAuthentication(storage, cfg.JWTSecretKey, cfg.JWTAccessExpire)
		if errAuth != nil {
			log.Error().Err(errAuth).Msg("Failed to initialize authentication with in-memory storage")
			urlService = nil
			authService = nil
		}
	}

	if urlService == nil || authService == nil {
		log.
			Fatal().
			Msg("URL shortener service initialization failed")
		return
	}

	srv, err := server.
		NewServer(log,
			*cfg,
			urlService,
			authService,
		)
	if err != nil {
		log.
			Fatal().
			Err(err).
			Msg("Failed to create server")
	} else {
		runServer(srv, log)
	}

}

/*
Пример:

	func run(

		ctx    context.Context,
		args   []string,
		getenv func(string) string,
		stdin  io.Reader,
		stdout, stderr io.Writer,

	) error
*/
func runServer(srv *server.Server, log *zerolog.Logger) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := srv.Start(context.Background()); err != nil && err != http.ErrServerClosed {
			log.
				Fatal().
				Err(err).
				Msg("Server failed to start")
		}
	}()

	// Shutdown, TODO gracefull shutdown
	sig := <-stop
	log.
		Info().
		Str("signal", sig.String()).
		Msg("Received signal, shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.
			Error().
			Err(err).
			Msg("Server shutdown error")
	} else {
		log.
			Info().
			Msg("Server shutdown completed successfully")
	}
}

func initPostgres(ctx context.Context, log *zerolog.Logger, dsn string) (*postgres.PostgresStorage, error) {
	storage, err := postgres.NewStorage(ctx, dsn)
	if err != nil {
		log.
			Error().
			Err(err).
			Msg("Failed to initialize PostgreSQL storage")
		return nil, err
	}

	log.
		Info().
		Msg("Using PostgreSQL storage")
	return storage, nil
}

func initInMemory(log *zerolog.Logger) *inmemory.InmemoryStorage {
	log.
		Info().
		Msg("Using in-memory storage")
	return inmemory.NewStorage()
}

func closePostgresStorage(log *zerolog.Logger, storage *postgres.PostgresStorage) {
	if err := storage.Close(); err != nil {
		log.
			Error().
			Err(err).
			Msg("Failed to close PostgreSQL storage")
		return
	}
	log.
		Info().
		Msg("PostgreSQL storage closed successfully")
}

func closeInMemoryStorage(log *zerolog.Logger, storage *inmemory.InmemoryStorage) {
	if err := storage.Close(); err != nil {
		log.
			Error().
			Err(err).
			Msg("Failed to close in-memory storage")
		return
	}
	log.
		Info().
		Msg("In-memory storage closed successfully")
}

// Для PostgreSQL хранилища
func initPostgresData(ctx context.Context, log *zerolog.Logger, storage *postgres.PostgresStorage, fileStore *filestore.FileStore) {
	if fileStore.GetFilePath() == "" {
		log.
			Info().
			Msg("No file storage path specified, skip loading data to PostgreSQL")
		return
	}

	path, isEmpty, err := fileStore.Load(ctx, storage)
	if err != nil {
		log.
			Error().
			Err(err).
			Str("path", fileStore.GetFilePath()).
			Msg("Failed to load data from file to PostgreSQL")
		return
	}

	if !isEmpty {
		log.
			Info().
			Str("path", path).
			Msg("Data loaded successfully from file to PostgreSQL")
	}
}

func savePostgresData(ctx context.Context, log *zerolog.Logger, storage *postgres.PostgresStorage, fileStore *filestore.FileStore) {
	if fileStore.GetFilePath() == "" {
		log.
			Info().
			Msg("No file storage path specified, skip saving PostgreSQL data")
		return
	}

	path, err := fileStore.Save(ctx, storage)
	if err != nil {
		log.
			Error().
			Err(err).
			Str("path", path).
			Msg("Failed to save PostgreSQL data to file")
		return
	}

	log.
		Info().
		Str("path", path).
		Msg("PostgreSQL data saved successfully to file")
}

func initInMemoryData(ctx context.Context, log *zerolog.Logger, storage *inmemory.InmemoryStorage, fileStore *filestore.FileStore) {
	if fileStore.GetFilePath() == "" {
		log.
			Info().
			Msg("No file storage path specified, skip loading data to in-memory storage")
		return
	}

	path, isEmpty, err := fileStore.Load(ctx, storage)
	if err != nil {
		log.
			Error().
			Err(err).
			Str("path", fileStore.GetFilePath()).
			Msg("Failed to load data from file to in-memory storage")
		return
	}

	// Выводим сообщение только если файл не был пустым
	if !isEmpty {
		log.
			Info().
			Str("path", path).
			Msg("Data loaded successfully from file to in-memory storage")
	}
}

func saveInMemoryData(ctx context.Context, log *zerolog.Logger, storage *inmemory.InmemoryStorage, fileStore *filestore.FileStore) {
	if fileStore.GetFilePath() == "" {

		log.
			Info().
			Msg("No file storage path specified, skip saving in-memory data")
		return
	}

	path, err := fileStore.Save(ctx, storage)
	if err != nil {
		log.
			Error().
			Err(err).
			Str("path", path).
			Msg("Failed to save in-memory data to file")
		return
	}

	log.
		Info().
		Str("path", path).
		Msg("In-memory data saved successfully to file")
}
