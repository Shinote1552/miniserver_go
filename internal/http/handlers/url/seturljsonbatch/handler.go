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
	BatchCreate(ctx context.Context, urls []models.ShortenedLink) ([]models.ShortenedLink, error)
}

func HandlerSetURLJsonBatch(svc ServiceURLShortener, urlroot string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var requestBatch []dto.BatchShortenRequest
		if err := json.NewDecoder(r.Body).Decode(&requestBatch); err != nil {
			httputils.WriteJSONError(w, http.StatusBadRequest, "invalid request format")
			return
		}

		// сопостовляем по OriginalURL и сохроняем порядок corelation_id: corelation_id/url - corelation_id/key
		requestMap := make(map[string]dto.BatchShortenRequest, len(requestBatch))
		urls := make([]models.ShortenedLink, 0, len(requestBatch))

		for _, req := range requestBatch {
			requestMap[req.LongURL] = req
			urls = append(urls, models.ShortenedLink{
				LongURL: req.LongURL,
			})
		}

		createdURLs, err := svc.BatchCreate(ctx, urls)
		if err != nil {
			if errors.Is(err, models.ErrConflict) {
				// Формируем ответ при ErrConflict
				responseMap := make([]dto.BatchShortenResponse, 0, len(requestBatch))

				for _, url := range createdURLs {
					if req, exists := requestMap[url.LongURL]; exists {
						responseMap = append(responseMap, dto.BatchShortenResponse{
							CorrelationID: req.CorrelationID,
							ShortURL:      httputils.BuildShortURL(urlroot, url.ShortCode),
						})
					}
				}

				w.Header().Set("Content-Type", httputils.MIMEApplicationJSON)
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(responseMap)
				return
			}
			httputils.WriteJSONError(w, http.StatusInternalServerError, "failed to process batch")
			return
		}

		// Формируем ответ
		responseMap := make([]dto.BatchShortenResponse, 0, len(requestBatch))
		for _, url := range createdURLs {
			if req, exists := requestMap[url.LongURL]; exists {
				responseMap = append(responseMap, dto.BatchShortenResponse{
					CorrelationID: req.CorrelationID,
					ShortURL:      httputils.BuildShortURL(urlroot, url.ShortCode),
				})
			}
		}

		w.Header().Set("Content-Type", httputils.MIMEApplicationJSON)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(responseMap)
	}
}
