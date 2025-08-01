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

		// сопостовляем по OriginalURL и сохроняем порядок: corelation_id/url - corelation_id/key
		reqMap := make(map[string]dto.BatchShortenRequest, len(reqBatch))
		urls := make([]models.URL, 0, len(reqBatch))

		for _, req := range reqBatch {
			reqMap[req.OriginalURL] = req
			urls = append(urls, models.URL{
				OriginalURL: req.OriginalURL,
			})
		}

		createdURLs, err := svc.BatchCreate(ctx, urls)
		if err != nil {
			if errors.Is(err, models.ErrConflict) {
				response := make([]dto.BatchShortenResponse, 0, len(reqBatch))

				for _, url := range createdURLs {
					if req, exists := reqMap[url.OriginalURL]; exists {
						response = append(response, dto.BatchShortenResponse{
							CorrelationID: req.CorrelationID,
							ShortURL:      httputils.BuildShortURL(urlroot, url.ShortKey),
						})
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

		// Формируем ответ
		response := make([]dto.BatchShortenResponse, 0, len(reqBatch))
		for _, url := range createdURLs {
			if req, exists := reqMap[url.OriginalURL]; exists {
				response = append(response, dto.BatchShortenResponse{
					CorrelationID: req.CorrelationID,
					ShortURL:      httputils.BuildShortURL(urlroot, url.ShortKey),
				})
			}
		}

		w.Header().Set("Content-Type", httputils.MIMEApplicationJSON)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}
