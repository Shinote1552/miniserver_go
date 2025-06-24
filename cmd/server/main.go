package main

import (
	"urlshortener/internal/config"
	"urlshortener/internal/handlers"
	"urlshortener/internal/logger"
	"urlshortener/internal/middleware"
	"urlshortener/internal/server"
	"urlshortener/internal/service"
	"urlshortener/internal/storage/inmemory"
)

func main() {
	// addr := "localhost:8080"
	// baseURL := "http://" + addr
	// cfg := config.NewServerConfig(addr, baseURL)
	cfg := config.LoadConfig() // Для чего?
	// AI: описать что это и какие есть юзкейсы

	mem := inmemory.NewInMemory()
	urlService := service.NewURLshortener(mem)
	urlHandler := handlers.NewHandlerURL(&urlService, cfg.ServerAddr)

	mylog := logger.GetLogger()
	loggingMiddleware := middleware.NewLoggingMiddleware(mylog)

	srv := server.NewServer(cfg.ServerAddr, mylog, loggingMiddleware, urlHandler)

	srv.Start()
}
