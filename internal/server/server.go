package server

import (
	"errors"
	"net/http"
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
	GetURL(token string) (string, error)
	SetURL(url string) (string, error)
	PingDataBase() error
}

type Server struct {
	router *mux.Router
	log    *zerolog.Logger
	svc    URLServiceShortener
	cfg    *config.Config
}

func NewServer(cfg *config.Config, log *zerolog.Logger, svc URLServiceShortener) (*Server, error) {
	if cfg == nil {
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

func (s *Server) Start() error {
	s.log.Info().Str("address", s.cfg.ServerAddress).Msg("Starting server")
	return http.ListenAndServe(s.cfg.ServerAddress, s.router)
}
