package ping

import (
	"net/http"

	"github.com/rs/zerolog"
)

type DBService interface {
	Ping() error
}

func HandlerPing(dbService DBService, log *zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := dbService.Ping(); err != nil {
			log.Error().Err(err).Msg("Database ping failed")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
