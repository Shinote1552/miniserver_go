package logger

import (
	"net/http"
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

			// Логируем начало запроса только в debug режиме
			if log.GetLevel() <= zerolog.DebugLevel {
				log.Debug().
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Str("ip", r.RemoteAddr).
					Msg("request started")
			}

			next.ServeHTTP(recorder, r)

			duration := time.Since(start)

			// Определяем тип сообщения по статусу
			var msg string
			switch {
			case recorder.statusCode >= 500:
				msg = "server error"
			case recorder.statusCode >= 400:
				msg = "client error"
			default:
				msg = "request completed"
			}

			// Базовое логирование для всех запросов
			logEntry := log.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", recorder.statusCode).
				Dur("duration_ms", duration/time.Millisecond).
				Int("bytes", recorder.size).
				Str("ip", r.RemoteAddr) // IP адрес клиента

			// Добавляем предупреждение для медленных запросов
			if duration > 100*time.Millisecond {
				logEntry = logEntry.Str("slow", "true")
			}

			// Логируем ошибки клиента как warning
			if recorder.statusCode >= 400 && recorder.statusCode < 500 {
				logEntry = logEntry.Str("error_type", "client_error")
			}

			// Логируем ошибки сервера как error
			if recorder.statusCode >= 500 {
				logEntry = logEntry.Str("error_type", "server_error")
			}

			logEntry.Msg(msg)
		})
	}
}
