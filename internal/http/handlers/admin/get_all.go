package admin

import (
	"context"
	"net/http"

	"urlshortener/internal/http/httputils"
	"urlshortener/internal/repository/dto"

	"github.com/rs/zerolog"
)

type Service interface {
	GetAllWithUsers(ctx context.Context) ([]dto.FullURLInfo, error)
}

func HandlerGetAll(svc Service, log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := log.With().Str("handler", "HandlerGetAll").Logger()

		log.Debug().Msg("fetching all URLs with user info")

		info, err := svc.GetAllWithUsers(ctx)
		if err != nil {
			log.Error().Err(err).Msg("failed to get all URLs")
			httputils.WriteJSONError(w, http.StatusInternalServerError, "failed to get all URLs")
			return
		}

		log.Debug().
			Int("count", len(info)).
			Msg("successfully retrieved URLs")

		httputils.WriteJSONResponse(w, http.StatusOK, info)
	}
}

// curl -v -X GET http://localhost:8080/admin/urls
