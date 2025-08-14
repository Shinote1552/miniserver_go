package create_json_batch

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

		// Получаем ID пользователя из контекста
		userID, ok := ctx.Value("user_id").(int64)
		if !ok || userID == 0 {
			httputils.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
			return
		}

		var requestBatch []dto.ShortenedLinkBatchRequest
		if err := json.NewDecoder(r.Body).Decode(&requestBatch); err != nil {
			httputils.WriteJSONError(w, http.StatusBadRequest, "invalid request format")
			return
		}

		// Создаем map для быстрого поиска correlation_id по URL
		urlToCorrelation := make(map[string]string, len(requestBatch))
		for _, req := range requestBatch {
			urlToCorrelation[req.OriginalURL] = req.CorrelationID
		}

		modelsBatch := dto.ShortenedLinkBatchRequestToDomain(requestBatch, userID)

		createdURLs, err := svc.BatchCreate(ctx, modelsBatch)
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, models.ErrConflict) {
				status = http.StatusConflict
			}

			response := dto.ShortenedLinkBatchCreateResponseFromDomains(createdURLs, urlroot)
			// Добавляем correlation_id внутри обработчика
			for i := range response {
				response[i].CorrelationID = urlToCorrelation[createdURLs[i].OriginalURL]
			}
			httputils.WriteJSONResponse(w, status, response)
			return
		}

		response := dto.ShortenedLinkBatchCreateResponseFromDomains(createdURLs, urlroot)
		// Добавляем correlation_id внутри обработчика
		for i := range response {
			response[i].CorrelationID = urlToCorrelation[createdURLs[i].OriginalURL]
		}

		httputils.WriteJSONResponse(w, http.StatusCreated, response)
	}
}
