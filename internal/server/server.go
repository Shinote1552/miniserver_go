package server

import (
	"fmt"
	"net/http"
	"urlshortener/internal/handlers"
	"urlshortener/internal/service"
	"urlshortener/internal/storage"

	"github.com/gorilla/mux"
)

type Server struct {
	storage *storage.InMemoryStorage
	router  *mux.Router
}

func NewServer(mem storage.InMemoryStorage, baseURL string) *Server {
	s :=
		&Server{
			storage: &mem,
			router:  mux.NewRouter(),
		}
	// Возможно стоит отсюда вынести и передовать в эти обьекты в NewServer
	urlService := service.NewURLshortener(*s.storage, baseURL)
	urlHandler := handlers.NewHandlerURL(&urlService, baseURL)
	s.routerInit(*urlHandler)
	return s
}

// serveMux
//
//	func (s *Server) routerInit(h handlers.HandlderURL) {
//		s.router.HandleFunc("GET /{id}", h.GetURL) // 307
//		s.router.HandleFunc("POST /", h.SetURL)    // 201
//		s.router.HandleFunc("/", h.DefaultURL)     // 400
//	}

// muxRouter
func (s *Server) routerInit(h handlers.HandlderURL) {
	s.router.HandleFunc("/{id}", h.GetURL).Methods("GET")
	s.router.HandleFunc("/", h.SetURL).Methods("POST")    // 201
	s.router.HandleFunc("/", h.DefaultURL).Methods("GET") // 400

}

func (s *Server) Start(addr string) {
	fullURL := "http://" + addr
	fmt.Println("Server started on:", fullURL)
	err := http.ListenAndServe(addr, s.router)
	if err != nil {
		panic(err)
	}
}
