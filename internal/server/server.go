package server

import (
	"net/http"
	"urlshortener/internal/deps"

	"github.com/gorilla/mux"
)

type Server struct {
	router *mux.Router
	addr   string
	log    deps.Logger
}

func NewServer(addr string, mylog deps.Logger, service deps.Handler) *Server {
	s :=
		&Server{
			router: mux.NewRouter(),
			addr:   addr,
			log:    mylog,
		}
	s.routerInit(service)
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
func (s *Server) routerInit(h deps.Handler) {
	s.router.HandleFunc("/{id}", h.GetURL).Methods("GET") // 307
	s.router.HandleFunc("/", h.SetURL).Methods("POST")    // 201
	s.router.HandleFunc("/", h.DefaultURL).Methods("GET") // 400

}

func (s *Server) Start() {
	fullURL := "http://" + s.addr
	s.log.Info().Str("address", fullURL).Msg("Starting server")

	// fmt.Println("Server started on:", fullURL)
	err := http.ListenAndServe(s.addr, s.router)
	if err != nil {
		panic(err)
	}
}
