package server

import (
	"errors"
	"net/http"
	"urlshortener/internal/config"
	"urlshortener/internal/handlers/getdefault"
	"urlshortener/internal/handlers/geturl"
	"urlshortener/internal/handlers/seturljson"
	"urlshortener/internal/handlers/seturltext"
	"urlshortener/internal/middleware"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

type Server struct {
	cfg    *config.ServerConfig
	router *mux.Router
	log    *zerolog.Logger
	svc    ServiceURLShortener
}

//go:generate mockgen -destination=mocks/url_shortener_mock.go -package=mocks urlshortener/internal/deps ServiceURLShortener
type ServiceURLShortener interface {
	GetURL(token string) (string, error)
	SetURL(url string) (string, error)
}

func NewServer(cfg *config.ServerConfig, logger *zerolog.Logger, svc ServiceURLShortener) (*Server, error) {
	if cfg == nil {
		return nil, errors.New("server config cannot be nil")
	}
	if logger == nil {
		return nil, errors.New("logger cannot be nil")
	}
	if svc == nil {
		return nil, errors.New("service cannot be nil")
	}

	s :=
		&Server{
			cfg:    cfg,
			router: mux.NewRouter(),
			log:    logger,
			svc:    svc,
		}

	s.registerRoutes()
	return s, nil
}

func (s *Server) registerRoutes() {

	s.router.Use(middleware.MiddlewareLogging(s.log))
	s.router.Use(middleware.MiddlewareCompressing())

	s.router.HandleFunc("/", getdefault.HandlerGetDefault()).Methods("GET")                                    // 400
	s.router.HandleFunc("/{id}", geturl.HandlerGetURLWithID(s.svc)).Methods("GET")                             // 307
	s.router.HandleFunc("/api/shorten", seturljson.HandlerSetURLJSON(s.svc, s.cfg.ListenPort)).Methods("POST") // 201
	s.router.HandleFunc("/", seturltext.HandlerSetURLText(s.svc, s.cfg.ListenPort)).Methods("POST")            // 201

}

func (s *Server) Start() {
	fullURL := "http://" + s.cfg.ListenPort
	s.log.Info().Str("address", fullURL).Msg("Starting server")

	err := http.ListenAndServe(s.cfg.ListenPort, s.router)
	if err != nil {
		s.log.Error().Err(err).Msg("Server failed to start")
		return
	}
}
