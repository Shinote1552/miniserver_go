package seturljsonbatch

import (
	"context"
	"encoding/json"
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

		if r.Method != http.MethodPost {
			writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		var reqBatch []models.APIShortenRequestBatch
		if err := json.NewDecoder(r.Body).
			Decode(&reqBatch); err != nil {
			writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
			return
		}

		if len(reqBatch) == 0 {
			writeJSONError(w, http.StatusBadRequest, "empty batch request")
			return
		}

		// ХЗ насколько правильно массив данных валидировать
		for _, item := range reqBatch {
			if item.CorrelationID == "" || item.OriginalURL == "" {
				writeJSONError(w, http.StatusBadRequest, "correlation_id and original_url are required for all items")
				return
			}
		}

		resBatch, err := svc.BatchCreate(ctx, reqBatch)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to process batch: %v", err))
			return
		}

		for i := range resBatch {
			resBatch[i].ShortURL = buildShortURL(urlroot, resBatch[i].ShortURL)
		}

		w.Header().Set("Content-Type", httputils.MIMEApplicationJSON)
		w.WriteHeader(http.StatusCreated)

		if err := json.NewEncoder(w).Encode(resBatch); err != nil {
			writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to encode response: %v", err))
			return
		}
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
