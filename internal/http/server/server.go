package server

import (
	"context"
	"errors"
	"net/http"
	"time"
	"urlshortener/domain/models"
	"urlshortener/internal/config"
	"urlshortener/internal/http/handlers/middlewares"
	"urlshortener/internal/http/handlers/system/getping"
	"urlshortener/internal/http/handlers/url/getdefault"
	"urlshortener/internal/http/handlers/url/geturltext"
	"urlshortener/internal/http/handlers/url/seturljson"
	"urlshortener/internal/http/handlers/url/seturljsonbatch"
	"urlshortener/internal/http/handlers/url/seturltext"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

//go:generate mockgen -destination=mocks/url_shortener_mock.go -package=mocks urlshortener/internal/deps ServiceURLShortener
type URLShortener interface {
	GetURL(context.Context, string) (models.ShortenedLink, error)
	SetURL(context.Context, string) (models.ShortenedLink, error)
	BatchCreate(ctx context.Context, urls []models.ShortenedLink) ([]models.ShortenedLink, error)
	PingDataBase(context.Context) error
}

type Server struct {
	httpServer *http.Server
	router     *mux.Router
	log        *zerolog.Logger
	svc        URLShortener
	cfg        config.Config
}

func NewServer(log *zerolog.Logger, cfg config.Config, svc URLShortener) (*Server, error) {

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

	s.router.Use(middlewares.MiddlewareLogging(s.log))
	s.router.Use(middlewares.MiddlewareCompressing())

	s.router.HandleFunc("/ping", getping.HandlerPing(s.svc)).Methods("GET")
	s.router.HandleFunc("/{id}", geturltext.HandlerGetURLWithID(s.svc)).Methods("GET") // 307
	s.router.HandleFunc("/", getdefault.HandlerGetDefault()).Methods("GET")            // 400

	s.router.HandleFunc("/api/shorten/batch", seturljsonbatch.HandlerSetURLJsonBatch(s.svc, s.cfg.ServerAddress)).Methods("POST") // 201
	s.router.HandleFunc("/api/shorten", seturljson.HandlerSetURLJson(s.svc, s.cfg.ServerAddress)).Methods("POST")                 // 201
	s.router.HandleFunc("/", seturltext.HandlerSetURLText(s.svc, s.cfg.ServerAddress)).Methods("POST")                            // 201

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
