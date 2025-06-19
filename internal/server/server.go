package server

import (
	"fmt"
	"net/http"
	"urlshortener/internal/config"
	"urlshortener/internal/handlers"
	"urlshortener/internal/service"
	"urlshortener/internal/storage"

	"github.com/gorilla/mux"
)

type Server struct {
	storage *storage.InMemoryStorage
	router  *mux.Router
	config  *config.ServerConfig
}

func NewServer(mem storage.InMemoryStorage, cfg config.ServerConfig) *Server {
	s :=
		&Server{
			storage: &mem,
			router:  mux.NewRouter(),
			config:  &cfg,
		}
	// Возможно стоит отсюда вынести и передовать в эти обьекты в NewServer
	urlService := service.NewURLshortener(*s.storage)
	urlHandler := handlers.NewHandlerURL(&urlService, cfg.ServerAddr)
	s.routerInit(*urlHandler)
	return s
}

// serveMux
//
//	func (s *Server) routerInit(h handlers.HandlderURL) {
//		s.router.HandleFunc("GET /{id}", h.GetURL) // 307
//		s.router.HandleFunc("POST /", h.SetURL)    // 201
//		s.router.HandleFunc("/", h.DefaultURL)      // 400
//	}

// muxRouter
func (s *Server) routerInit(h handlers.HandlderURL) {
	s.router.HandleFunc("/{id}", h.GetURL).Methods("GET") // 307
	s.router.HandleFunc("/", h.SetURL).Methods("POST")    // 201
	s.router.HandleFunc("/", h.DefaultURL).Methods("GET") // 400

}

func (s *Server) Start() {
	fullURL := s.config.BaseURL
	fmt.Println("Server started on:", fullURL)
	err := http.ListenAndServe(s.config.ServerAddr, s.router)
	if err != nil {
		panic(err)
	}
}
