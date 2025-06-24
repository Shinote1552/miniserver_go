package server

import (
	"net/http"
	"urlshortener/internal/deps"

	"github.com/gorilla/mux"
)

type Server struct {
	addr     string
	router   *mux.Router
	log      deps.Logger
	handlers Handlers
}

type Handlers struct {
	GetHandler      http.Handler
	PostTextHandler http.Handler
	PostJSONHandler http.Handler
	DefaultHandler  http.Handler
}

func NewServer(addr string, mylog deps.Logger, middlware deps.Middleware, handlers Handlers) *Server {
	s :=
		&Server{
			addr:     addr,
			router:   mux.NewRouter(),
			log:      mylog,
			handlers: handlers,
		}
	s.routerInit(middlware, handlers)
	return s
}

// muxRouter
func (s *Server) routerInit(mw deps.Middleware, handlers Handlers) {

	// Применяем middleware к основному роутеру
	s.router.Use(mw.Handler)

	s.router.HandleFunc("/api/shorten", handlers.PostJSONHandler.ServeHTTP).Methods("POST") // 201
	s.router.HandleFunc("/{id}", handlers.GetHandler.ServeHTTP).Methods("GET")              // 307
	s.router.HandleFunc("/", handlers.PostTextHandler.ServeHTTP).Methods("POST")            // 201
	s.router.HandleFunc("/", handlers.DefaultHandler.ServeHTTP).Methods("GET")              // 400

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

/*

фвыфвыфвыыфывфвцвцй

*/
