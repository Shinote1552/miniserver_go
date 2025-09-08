package create_json

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"urlshortener/internal/domain/models"
	"urlshortener/internal/http/dto"
	"urlshortener/internal/http/httputils"
)

type ServiceURLShortener interface {
	SetURL(ctx context.Context, model models.ShortenedLink) (models.ShortenedLink, error)
}

func HandlerSetURLJson(svc ServiceURLShortener, urlroot string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if r.Method != http.MethodPost {
			httputils.WriteJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		userID, ok := ctx.Value("user_id").(int64)
		if !ok || userID == 0 {
			httputils.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
			return
		}

		var req dto.ShortenedLinkSingleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httputils.WriteJSONError(w, http.StatusBadRequest, httputils.ErrInvalidData.Error())
			return
		}

		if req.URL == "" {
			httputils.WriteJSONError(w, http.StatusBadRequest, httputils.ErrInvalidData.Error())
			return
		}

		model := dto.ShortenedLinkSingleRequestToDomain(req, userID)
		urlModel, err := svc.SetURL(ctx, model)

		if err != nil {
			if errors.Is(err, httputils.ErrConflict) {
				resp := dto.ShortenedLinkSingleResponseFromDomain(urlModel, urlroot)
				httputils.WriteJSONResponse(w, http.StatusConflict, resp)
				return
			}
			httputils.WriteJSONError(w, http.StatusBadRequest, err.Error())
			return
		}

		resp := dto.ShortenedLinkSingleResponseFromDomain(urlModel, urlroot)
		httputils.WriteJSONResponse(w, http.StatusCreated, resp)
	}
}
