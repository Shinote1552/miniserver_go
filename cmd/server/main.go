package main

import (
	"urlshortener/internal/config"
	"urlshortener/internal/handlers/defaulthandler"
	"urlshortener/internal/handlers/geturl"
	"urlshortener/internal/handlers/seturljson"
	"urlshortener/internal/handlers/seturltext"
	"urlshortener/internal/logger"
	"urlshortener/internal/middleware"
	"urlshortener/internal/server"
	"urlshortener/internal/service"
	"urlshortener/internal/storage/inmemory"
)

func main() {
	cfg := config.LoadConfig()
	// Config может содержать:
	// - ListenPort (порт сервера, например "localhost:8080")
	// - BaseURL (базовый URL для коротких ссылок, например "http://localhost:8080")

	/*

		// можно будет еще добавить:

		// - LogLevel (уровень логирования)
		// - StorageType (тип хранилища: "inmemory", "postgres", etc.)
		// - и другие параметры

	*/

	// Создание хэндлеров

	storage := inmemory.NewInMemoryStorage()

	service := service.NewURLShortenerService(storage)

	mylog := logger.GetLogger()
	loggingMiddleware := middleware.NewLoggingMiddleware(mylog)

	handlers := server.Handlers{
		GetHandler:      geturl.New(service),
		PostTextHandler: seturltext.New(service, cfg.BaseURL),
		PostJSONHandler: seturljson.New(service, cfg.BaseURL),
		DefaultHandler:  defaulthandler.New(),
	}

	srv := server.NewServer(cfg.ListenPort, mylog, loggingMiddleware, handlers)

	srv.Start()
}
