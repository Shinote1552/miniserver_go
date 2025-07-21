package ping

import (
	"net/http"

	"github.com/rs/zerolog"
)

type Service interface {
	PingDataBase() error
}

func HandlerPing(svc Service, log *zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := svc.PingDataBase(); err != nil {
			log.Error().Err(err).Msg("Database ping failed")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Database unavailable"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}
