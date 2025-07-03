package middleware

import (
	"net/http"
	"time"
	"urlshortener/internal/httputils"

	"github.com/rs/zerolog"
)

// MiddlewareLogging создает middleware для логирования HTTP запросов
func MiddlewareLogging(log zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			recorder := httputils.NewResponseRecorder(w)

			log.Info().
				Str("method", r.Method).
				Str("uri", r.RequestURI).
				Msg("request started")

			next.ServeHTTP(recorder, r)

			log.Info().
				Str("method", r.Method).
				Str("uri", r.RequestURI).
				Int("status", recorder.StatusCode).
				Int("size", recorder.Size).
				Dur("duration", time.Since(start)).
				Msg("request completed")
		})
	}
}
