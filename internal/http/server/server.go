package server

import (
	"context"
	"errors"
	"net/http"
	"time"
	"urlshortener/internal/config"
	"urlshortener/internal/http/handlers/middlewares/authorization"
	"urlshortener/internal/http/handlers/middlewares/compressor"
	"urlshortener/internal/http/handlers/middlewares/logger"
	"urlshortener/internal/http/handlers/system/ping"
	"urlshortener/internal/http/handlers/url/create_json"
	"urlshortener/internal/http/handlers/url/create_json_batch"
	"urlshortener/internal/http/handlers/url/create_text"
	"urlshortener/internal/http/handlers/url/delete_batch"
	"urlshortener/internal/http/handlers/url/find_by_id"
	"urlshortener/internal/http/handlers/url/get_default"
	"urlshortener/internal/http/handlers/url/list_user_urls"
	"urlshortener/internal/services/auth"
	"urlshortener/internal/services/url_shortener"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

type Server struct {
	httpServer  *http.Server
	router      *mux.Router
	log         *zerolog.Logger
	authService *auth.Authentication
	urlService  *url_shortener.URLShortener
	cfg         config.Config
}

func NewServer(log *zerolog.Logger, cfg config.Config, svc *url_shortener.URLShortener, auth *auth.Authentication) (*Server, error) {
	/*
		хз по идее конфиг создается через фабрику где уже есть валидация и
		стандартные значения, сюда по идее нереально подать пустую cfg
	*/

	if cfg.ServerAddress == "" {
		return nil, errors.New("server config cannot be nil")
	}
	if log == nil {
		return nil, errors.New("logger cannot be nil")
	}
	if svc == nil {
		return nil, errors.New("service cannot be nil")
	}

	s :=
		&Server{
			router:      mux.NewRouter(),
			cfg:         cfg,
			log:         log,
			authService: auth,
			urlService:  svc,
		}

	s.httpServer = &http.Server{
		Addr:              cfg.ServerAddress,
		Handler:           s.router,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	s.setupRoutes()
	return s, nil
}

func (s *Server) setupRoutes() {

	s.router.Use(logger.MiddlewareLogging(s.log))
	s.router.Use(compressor.MiddlewareCompressing())

	// Public routes (no auth required)
	s.router.HandleFunc("/ping", ping.HandlerPing(s.urlService)).Methods("GET")
	s.router.HandleFunc("/{id}", find_by_id.HandlerGetURLWithID(s.urlService)).Methods("GET") // 307
	s.router.HandleFunc("/", get_default.HandlerGetDefault()).Methods("GET")                  // 400

	authRouter := s.router.PathPrefix("/").Subrouter()
	authRouter.Use(authorization.MiddlewareAuth(s.authService))

	// Protected routes (with auth)
	authRouter.HandleFunc("/api/shorten/batch", create_json_batch.HandlerSetURLJsonBatch(s.urlService, s.cfg.ServerAddress)).Methods("POST") // 201
	authRouter.HandleFunc("/api/shorten", create_json.HandlerSetURLJson(s.urlService, s.cfg.ServerAddress)).Methods("POST")                  // 201
	authRouter.HandleFunc("/api/user/urls", list_user_urls.HandlerGetURLJsonBatch(s.urlService, s.cfg.ServerAddress)).Methods("GET")
	authRouter.HandleFunc("/api/user/urls", delete_batch.(s.urlService, s.cfg.ServerAddress)).Methods("DELETE")
	authRouter.HandleFunc("/", create_text.HandlerSetURLText(s.urlService, s.cfg.ServerAddress)).Methods("POST") // 201
}

func (s *Server) Start(ctx context.Context) error {
	s.log.
		Info().
		Str("address", s.cfg.ServerAddress).
		Msg("Starting server")
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil

}

func (s *Server) Shutdown(ctx context.Context) error {
	s.log.
		Info().
		Msg("Shutting down server")
	return s.httpServer.Shutdown(ctx)
}
