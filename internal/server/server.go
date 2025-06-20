package server

import (
	"fmt"
	"log"
	"net/http"
	"urlshortener/internal/handlers"
	"urlshortener/internal/service"
	"urlshortener/internal/storage"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

type Server struct {
	storage *storage.InMemoryStorage
	router  *mux.Router
	addr    string
}

func NewServer(mem storage.InMemoryStorage, addr string) *Server {
	s :=
		&Server{
			storage: &mem,
			router:  mux.NewRouter(),
			addr:    addr,
		}
	// Возможно стоит отсюда вынести и передовать в эти обьекты в NewServer
	urlService := service.NewURLshortener(*s.storage)
	urlHandler := handlers.NewHandlerURL(&urlService, addr)
	s.routerInit(*urlHandler)
	s.logerInit()
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

func (s *Server) logerInit() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Print("hello world")
}

func (s *Server) Start() {
	fullURL := "http://" + s.addr
	fmt.Println("Server started on:", fullURL)
	err := http.ListenAndServe(s.addr, s.router)
	if err != nil {
		panic(err)
	}
}
