package server

import (
	"net/http"
	"urlshortener/internal/handlers"
	"urlshortener/internal/service"
	"urlshortener/internal/storage"
)

type Server struct {
	storage *storage.InMemoryStorage
	router  *http.ServeMux
}

func NewServer(mem storage.InMemoryStorage, baseURL string) *Server {
	s :=
		&Server{
			storage: &mem,
			router:  http.NewServeMux(),
		}

	urlService := service.NewURLshortener(*s.storage, baseURL)
	urlHandler := handlers.NewHandlderURL(&urlService)
	routerInit(*urlHandler)
	return s
}

func routerInit(handlers.HandlderURL) {
	/*
		GET
		middleware
		POST
	*/

}

func (*Server) Start(addr string) {

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		panic(err)
	}
}
