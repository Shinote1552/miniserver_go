package ping

import (
	"net/http"

	"github.com/rs/zerolog"
)

type ServiceURLShortener interface {
	PingDataBase() error
}

func HandlerPing(dbService ServiceURLShortener, log *zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := dbService.PingDataBase(); err != nil {
			log.Error().Err(err).Msg("Database ping failed")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
