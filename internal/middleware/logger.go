package middleware

import (
	"net/http"
	"time"
	"urlshortener/internal/deps"
)

// LoggingMiddleware реализует Middleware интерфейс для логирования
type LoggingMiddleware struct {
	log deps.Logger
}

// NewLoggingMiddleware создает новое middleware для логирования
func NewLoggingMiddleware(log deps.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{log: log}
}

// Handler реализует Middleware интерфейс
func (m *LoggingMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := NewResponseRecorder(w)

		m.log.Info().
			Str("method", r.Method).
			Str("uri", r.RequestURI).
			Msg("request started")

		next.ServeHTTP(recorder, r)

		m.log.Info().
			Str("method", r.Method).
			Str("uri", r.RequestURI).
			Int("status", recorder.StatusCode).
			Int("size", recorder.Size).
			Dur("duration", time.Since(start)).
			Msg("request completed")
	})
}
