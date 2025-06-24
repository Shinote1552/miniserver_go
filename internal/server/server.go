package server

import (
	"net/http"
	"urlshortener/internal/deps"

	"github.com/gorilla/mux"
)

type Server struct {
	addr   string
	router *mux.Router
	log    deps.Logger
}

func NewServer(addr string, mylog deps.Logger, middlware deps.Middleware, service deps.Handler) *Server {
	s :=
		&Server{
			addr:   addr,
			router: mux.NewRouter(),
			log:    mylog,
		}
	s.routerInit(service, middlware)
	return s
}

// muxRouter
func (s *Server) routerInit(h deps.Handler, mw deps.Middleware) {

	// Применяем middleware к основному роутеру
	s.router.Use(mw.Handler)

	s.router.HandleFunc("/api/shorten", h.SetURLwithJSON).Methods("POST") // 201

	s.router.HandleFunc("/{id}", h.GetURL).Methods("GET") // 307

	s.router.HandleFunc("/", h.SetURL).Methods("POST")    // 201
	s.router.HandleFunc("/", h.DefaultURL).Methods("GET") // 400
}

func (s *Server) Start() {
	fullURL := "http://" + s.addr
	s.log.Info().Str("address", fullURL).Msg("Starting server")

	err := http.ListenAndServe(s.addr, s.router)
	if err != nil {
		s.log.Error().Err(err).Msg("Server failed to start")
		panic(err)
	}
}
