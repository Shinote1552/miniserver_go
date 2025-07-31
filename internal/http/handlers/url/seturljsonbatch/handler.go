package seturljsonbatch

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
	BatchCreate(ctx context.Context, urls []models.URL) ([]models.URL, error)
}

func HandlerSetURLJsonBatch(svc ServiceURLShortener, urlroot string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var reqBatch []dto.BatchShortenRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBatch); err != nil {
			httputils.WriteJSONError(w, http.StatusBadRequest, "invalid request format")
			return
		}

		// Convert DTO to domain models
		urls := make([]models.URL, len(reqBatch))
		for i, req := range reqBatch {
			urls[i] = models.URL{
				OriginalURL: req.OriginalURL,
				// Здесь можно сохранить correlation_id если нужно
			}
		}

		resBatch, err := svc.BatchCreate(ctx, urls)
		if err != nil {
			if errors.Is(err, httputils.ErrConflict) {
				response := make([]dto.BatchShortenResponse, len(resBatch))
				for i, url := range resBatch {
					response[i] = dto.BatchShortenResponse{
						CorrelationID: reqBatch[i].CorrelationID, // Сохраняем оригинальный ID
						ShortURL:      httputils.BuildShortURL(urlroot, url.ShortKey),
					}
				}
				w.Header().Set("Content-Type", httputils.MIMEApplicationJSON)
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(response)
				return
			}
			httputils.WriteJSONError(w, http.StatusInternalServerError, "failed to process batch")
			return
		}

		response := make([]dto.BatchShortenResponse, len(resBatch))
		for i, url := range resBatch {
			response[i] = dto.BatchShortenResponse{
				CorrelationID: reqBatch[i].CorrelationID,
				ShortURL:      httputils.BuildShortURL(urlroot, url.ShortKey),
			}
		}

		w.Header().Set("Content-Type", httputils.MIMEApplicationJSON)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}
