package middlewares

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

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
			recorder := &responseRecorder{ResponseWriter: w}

			// Логируем начало запроса
			log.Debug().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("ip", r.RemoteAddr).
				Msg("request started")

			// Перехватываем паники, чтобы залогировать их
			defer func() {
				if err := recover(); err != nil {
					log.Error().
						Str("panic", fmt.Sprintf("%v", err)).
						Str("stack", string(debug.Stack())).
						Msg("request panic")
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}

				// Логируем завершение запроса
				logEvent := log.Info().
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Int("status", recorder.statusCode).
					Dur("duration", time.Since(start)).
					Int("bytes", recorder.size)

				if recorder.statusCode >= 500 {
					logEvent = logEvent.Str("level", "error")
				} else if recorder.statusCode >= 400 {
					logEvent = logEvent.Str("level", "warn")
				}

				logEvent.Msg("request completed")
			}()

			next.ServeHTTP(recorder, r)
		})
	}
}

// func MiddlewareLogging(log *zerolog.Logger) func(http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			ctx := r.Context()
// 			start := time.Now()
// 			recorder := &responseRecorder{ResponseWriter: w}

// 			next.ServeHTTP(recorder, r.WithContext(ctx))

// 			logEvent := log.Info().
// 				Str("method", r.Method).
// 				Str("uri", r.RequestURI).
// 				Int("status", recorder.statusCode).
// 				Dur("latency", time.Since(start)).
// 				Str("ip", r.RemoteAddr)

// 			if recorder.statusCode >= 400 {
// 				logEvent = logEvent.
// 					Str("error", http.StatusText(recorder.statusCode)).
// 					Int("bytes_out", recorder.size)
// 			}

// 			msg := "request"
// 			if recorder.statusCode >= 500 {
// 				msg = "server error"
// 			} else if recorder.statusCode >= 400 {
// 				msg = "client error"
// 			}

// 			logEvent.Msg(msg)
// 		})
// 	}
// }
