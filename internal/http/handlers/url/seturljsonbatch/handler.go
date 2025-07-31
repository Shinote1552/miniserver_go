package seturljsonbatch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"urlshortener/internal/httputils"
	"urlshortener/internal/models"
)

type ServiceURLShortener interface {
	BatchCreate(ctx context.Context, batchItems []models.APIShortenRequestBatch) ([]models.APIShortenResponseBatch, error)
}

func HandlerSetURLJsonBatch(svc ServiceURLShortener, urlroot string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var reqBatch []models.APIShortenRequestBatch
		if err := json.NewDecoder(r.Body).Decode(&reqBatch); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid request format")
			return
		}

		resBatch, err := svc.BatchCreate(ctx, reqBatch)
		if err != nil {
			if errors.Is(err, models.ErrConflict) {
				for i := range resBatch {
					resBatch[i].ShortURL = buildShortURL(urlroot, resBatch[i].ShortURL)
				}
				w.Header().Set("Content-Type", httputils.MIMEApplicationJSON)
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(resBatch)
				return
			}
			writeJSONError(w, http.StatusInternalServerError, "failed to process batch")
			return
		}

		for i := range resBatch {
			resBatch[i].ShortURL = buildShortURL(urlroot, resBatch[i].ShortURL)
		}

		w.Header().Set("Content-Type", httputils.MIMEApplicationJSON)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resBatch)
	}
}
func buildShortURL(urlroot, id string) string {
	return fmt.Sprintf("http://%s/%s", urlroot, id)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", httputils.MIMEApplicationJSON)
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(models.APIErrorResponse{Error: message})
}
