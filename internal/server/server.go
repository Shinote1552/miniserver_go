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

// все типы которые реализовывают ендпоинты
type Handlers struct {
	GetDefaultHandler http.Handler
	GetWithIdHandler  http.Handler
	PostTextHandler   http.Handler
	PostJSONHandler   http.Handler
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

	// добавляем middleware(deps.Middleware) к роутеру
	s.router.Use(mw.Handler)

	s.router.HandleFunc("/", handlers.GetDefaultHandler.ServeHTTP).Methods("GET")           // 400
	s.router.HandleFunc("/{id}", handlers.GetWithIdHandler.ServeHTTP).Methods("GET")        // 307
	s.router.HandleFunc("/", handlers.PostTextHandler.ServeHTTP).Methods("POST")            // 201
	s.router.HandleFunc("/api/shorten", handlers.PostJSONHandler.ServeHTTP).Methods("POST") // 201

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
