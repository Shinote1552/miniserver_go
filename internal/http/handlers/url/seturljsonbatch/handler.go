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

		var requestBatch []dto.ShortenedLinkBatchRequest
		if err := json.NewDecoder(r.Body).Decode(&requestBatch); err != nil {
			httputils.WriteJSONError(w, http.StatusBadRequest, "invalid request format")
			return
		}

		requestMap := make(map[string]dto.ShortenedLinkBatchRequest, len(requestBatch))
		urls := make([]models.ShortenedLink, 0, len(requestBatch))

		for _, req := range requestBatch {
			requestMap[req.OriginalURL] = req
			urls = append(urls, models.ShortenedLink{
				LongURL: req.OriginalURL,
			})
		}

		createdURLs, err := svc.BatchCreate(ctx, urls)
		if err != nil {
			if errors.Is(err, models.ErrConflict) {
				response := dto.ShortenedLinkBatchCreateResponseFromDomains(createdURLs, urlroot)

				// Сопоставлем по correlation_id // можно будет в функцию вывести
				for i := range response {
					if req, exists := requestMap[createdURLs[i].LongURL]; exists {
						response[i].CorrelationID = req.CorrelationID
					}
				}
				httputils.WriteJSONResponse(w, http.StatusConflict, response)
				return
			}
			httputils.WriteJSONError(w, http.StatusInternalServerError, "failed to process batch")
			return
		}

		response := dto.ShortenedLinkBatchCreateResponseFromDomains(createdURLs, urlroot)

		// Сопоставлем по correlation_id // можно будет в функцию вывести
		for i := range response {
			if req, exists := requestMap[createdURLs[i].LongURL]; exists {
				response[i].CorrelationID = req.CorrelationID
			}
		}

		httputils.WriteJSONResponse(w, http.StatusCreated, response)
	}
}
