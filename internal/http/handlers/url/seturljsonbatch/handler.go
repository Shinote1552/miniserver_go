package seturljsonbatch

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"urlshortener/domain/models"
	"urlshortener/internal/http/httputils"
)

type ServiceURLShortener interface {
	BatchCreate(ctx context.Context, urls []models.URL) ([]models.URL, error)
}

func HandlerSetURLJsonBatch(svc ServiceURLShortener, urlroot string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var reqBatch []models.URL
		if err := json.NewDecoder(r.Body).Decode(&reqBatch); err != nil {
			httputils.WriteJSONError(w, http.StatusBadRequest, "invalid request format")
			return
		}

		resBatch, err := svc.BatchCreate(ctx, reqBatch)
		if err != nil {
			if errors.Is(err, httputils.ErrConflict) {
				for i := range resBatch {
					resBatch[i].ShortKey = httputils.BuildShortURL(urlroot, resBatch[i].ShortKey)
				}
				w.Header().Set("Content-Type", httputils.MIMEApplicationJSON)
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(resBatch)
				return
			}
			httputils.WriteJSONError(w, http.StatusInternalServerError, "failed to process batch")
			return
		}

		for i := range resBatch {
			resBatch[i].ShortKey = httputils.BuildShortURL(urlroot, resBatch[i].ShortKey)
		}

		w.Header().Set("Content-Type", httputils.MIMEApplicationJSON)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resBatch)
	}
}
