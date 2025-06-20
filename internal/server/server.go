package server

import (
	"net/http"
	"time"
	"urlshortener/internal/deps"
	"urlshortener/internal/middleware"

	"github.com/gorilla/mux"
)

type Server struct {
	addr   string
	router *mux.Router
	log    deps.Logger
}

func NewServer(addr string, mylog deps.Logger, service deps.Handler) *Server {
	s :=
		&Server{
			addr:   addr,
			router: mux.NewRouter(),
			log:    mylog,
		}
	s.routerInit(service)
	return s
}

// muxRouter
func (s *Server) routerInit(h deps.Handler) {

	// Создаем middleware
	loggingMiddleware := middleware.NewLoggingMiddleware(s.log)

	// Применяем middleware к основному роутеру
	s.router.Use(loggingMiddleware.Handler)

	s.router.HandleFunc("/{id}", h.GetURL).Methods("GET") // 307
	s.router.HandleFunc("/", h.SetURL).Methods("POST")    // 201
	s.router.HandleFunc("/", h.DefaultURL).Methods("GET") // 400

}

func (s *Server) WithLogging(next http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		// функция Now() возвращает текущее время
		start := time.Now()
		// Логируем информацию о запросе
		s.log.Info().
			Str("method", r.Method).
			Str("uri", r.RequestURI).
			Msg("request started")

		next.ServeHTTP(w, r) // обслуживание оригинального запроса
		duration := time.Since(start)

		// Логируем информацию о ответе
		s.log.Info().
			Str("method", r.Method).
			Str("url", r.RequestURI).
			// Int("status", w.statusCode).
			// Int("size", w.size).
			Dur("duration", duration).
			Msg("request completed")
	}
	return http.HandlerFunc(logFn)
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
