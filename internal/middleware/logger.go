package middleware

import (
	"net/http"
	"time"

	"urlshortener/internal/httputils"

	"github.com/rs/zerolog"
)

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if r.statusCode == 0 {
		r.statusCode = http.StatusOK
	}
	size, err := r.ResponseWriter.Write(b)
	r.size += size
	return size, err
}

func MiddlewareLogging(log *zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Логируем информацию о запросе
			log.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("remote_addr", r.RemoteAddr).
				Str("user_agent", r.Header.Get(httputils.HeaderUserAgent)).
				Str("accept_encoding", r.Header.Get(httputils.HeaderAcceptEncoding)).
				Str("content_encoding", r.Header.Get(httputils.HeaderContentEncoding)).
				Msg("request started")

			recorder := &responseRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(recorder, r)

			// Логируем информацию о ответе
			logger := log.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", recorder.statusCode).
				Int("size", recorder.size).
				Dur("duration", time.Since(start)).
				Str("response_encoding", w.Header().Get(httputils.HeaderContentEncoding))

			if recorder.statusCode >= 400 {
				logger = logger.Str("error", http.StatusText(recorder.statusCode))
			}

			logger.Msg("request completed")
		})
	}
}
