package dto

import (
	"time"
	"urlshortener/internal/domain/models"
)

// Request types
type (

	// ShortenedLinkTextRequest представляет DTO для текстового запроса на сокращение URL
	ShortenedLinkTextRequest struct {
		URL string // тут text/plain
	}

	ShortenedLinkSingleRequest struct {
		URL string `json:"url"`
	}

	ShortenedLinkBatchRequest struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"`
	}
)

// Response types
type (

	// ShortenedLinkTextResponse представляет DTO для текстового ответа с сокращенным URL
	ShortenedLinkTextResponse struct {
		ShortURL string
	}

	// Для POST /api/shorten/batch
	ShortenedLinkBatchCreateResponse struct {
		CorrelationID string `json:"correlation_id"`
		ShortURL      string `json:"short_url"`
	}

	ShortenedLinkBatchCreateRequest struct {
		OriginalURL string `json:"original_url"`
		UserID      uint64 `json:"user_id"`
	}

	// Для GET /api/user/urls
	ShortenedLinkBatchGetResponse struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}

	ShortenedLinkSingleResponse struct {
		Result string `json:"result"`
	}

	ShortenedLinkErrorResponse struct {
		Error string `json:"error"`
	}
)

// ShortenedLinkTextRequestToDomain преобразует текстовый запрос в доменную модель
func ShortenedLinkTextRequestToDomain(r ShortenedLinkTextRequest, userID int64) models.ShortenedLink {
	return models.ShortenedLink{
		OriginalURL: r.URL,
		UserID:      userID,
		CreatedAt:   time.Now().UTC(),
	}
}

// ShortenedLinkTextResponseFromDomain создает текстовый ответ из доменной модели
func ShortenedLinkTextResponseFromDomain(model models.ShortenedLink, baseURL string) ShortenedLinkTextResponse {
	return ShortenedLinkTextResponse{
		ShortURL: baseURL + "/" + model.ShortCode,
	}
}

func ShortenedLinkSingleRequestToDomain(r ShortenedLinkSingleRequest, userID int64) models.ShortenedLink {
	return models.ShortenedLink{
		OriginalURL: r.URL,
		UserID:      userID,
		CreatedAt:   time.Now().UTC(),
	}
}

func ShortenedLinkSingleResponseFromDomain(model models.ShortenedLink, baseURL string) ShortenedLinkSingleResponse {
	return ShortenedLinkSingleResponse{
		Result: baseURL + "/" + model.ShortCode,
	}
}

// ShortenedLinkBatchRequest -> domain
func ShortenedLinkBatchRequestToDomain(reqs []ShortenedLinkBatchRequest, userID int64) []models.ShortenedLink {
	urls := make([]models.ShortenedLink, len(reqs))
	for i, r := range reqs {
		urls[i] = models.ShortenedLink{
			OriginalURL: r.OriginalURL,
			UserID:      userID,
			CreatedAt:   time.Now().UTC(),
		}
	}
	return urls
}

// Для POST /api/shorten/batch
func ShortenedLinkBatchCreateResponseFromDomains(urls []models.ShortenedLink, baseURL string) []ShortenedLinkBatchCreateResponse {
	responses := make([]ShortenedLinkBatchCreateResponse, len(urls))
	for i, url := range urls {
		responses[i] = ShortenedLinkBatchCreateResponse{
			CorrelationID: "", // Заполняется в обработчике
			ShortURL:      baseURL + "/" + url.ShortCode,
		}
	}
	return responses
}

// Для GET /api/user/urls
func ShortenedLinkBatchGetResponseFromDomains(urls []models.ShortenedLink, baseURL string) []ShortenedLinkBatchGetResponse {
	responses := make([]ShortenedLinkBatchGetResponse, len(urls))
	for i, url := range urls {
		responses[i] = ShortenedLinkBatchGetResponse{
			ShortURL:    baseURL + "/" + url.ShortCode,
			OriginalURL: url.OriginalURL,
		}
	}
	return responses
}
