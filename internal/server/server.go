package server

import (
	"context"
	"errors"
	"net/http"
	"time"
	"urlshortener/internal/config"
	"urlshortener/internal/handlers/getdefault"
	"urlshortener/internal/handlers/geturl"
	"urlshortener/internal/handlers/ping"
	"urlshortener/internal/handlers/seturljson"
	"urlshortener/internal/handlers/seturltext"
	"urlshortener/internal/middleware"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

//go:generate mockgen -destination=mocks/url_shortener_mock.go -package=mocks urlshortener/internal/deps ServiceURLShortener
type URLServiceShortener interface {
	GetURL(context.Context, string) (string, error)
	SetURL(context.Context, string) (string, error)
	PingDataBase(context.Context) error
}

type Server struct {
	httpServer *http.Server
	router     *mux.Router
	log        *zerolog.Logger
	svc        URLServiceShortener
	cfg        config.Config
}

func NewServer(log *zerolog.Logger, cfg config.Config, svc URLServiceShortener) (*Server, error) {

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
			router: mux.NewRouter(),
			cfg:    cfg,
			log:    log,
			svc:    svc,
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

	s.router.Use(middleware.MiddlewareLogging(s.log))
	s.router.Use(middleware.MiddlewareCompressing())

	s.router.HandleFunc("/ping", ping.HandlerPing(s.svc)).Methods("GET")
	s.router.HandleFunc("/{id}", geturl.HandlerGetURLWithID(s.svc)).Methods("GET") // 307
	s.router.HandleFunc("/", getdefault.HandlerGetDefault()).Methods("GET")        // 400

	s.router.HandleFunc("/api/shorten", seturljson.HandlerSetURLJSON(s.svc, s.cfg.ServerAddress)).Methods("POST") // 201
	s.router.HandleFunc("/", seturltext.HandlerSetURLText(s.svc, s.cfg.ServerAddress)).Methods("POST")            // 201

}

func (s *Server) Start(ctx context.Context) error {
	s.log.Info().Str("address", s.cfg.ServerAddress).Msg("Starting server")
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil

}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
