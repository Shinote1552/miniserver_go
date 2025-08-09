package seturljson

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"urlshortener/domain/models"
	"urlshortener/internal/http/dto"
	"urlshortener/internal/http/httputils"
)

type ServiceURLShortener interface {
	SetURL(ctx context.Context, originalURL string) (models.ShortenedLink, error)
}

func HandlerSetURLJson(svc ServiceURLShortener, urlroot string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if r.Method != http.MethodPost {
			httputils.WriteJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
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

		urlModel, err := svc.SetURL(ctx, req.URL)
		if err != nil {
			if errors.Is(err, httputils.ErrConflict) {
				res := dto.ShortenedLinkSingleResponse{
					Result: httputils.BuildShortURL(urlroot, urlModel.ShortCode),
				}
				w.Header().Set("Content-Type", httputils.MIMEApplicationJSON)
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(res)
				return
			}
			httputils.WriteJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}

		res := dto.ShortenedLinkSingleResponse{Result: httputils.BuildShortURL(urlroot, urlModel.ShortCode)}

		w.Header().Set("Content-Type", httputils.MIMEApplicationJSON)
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(res); err != nil {
			httputils.WriteJSONError(w, http.StatusInternalServerError, "failed to encode response")
			return
		}
	}
}
