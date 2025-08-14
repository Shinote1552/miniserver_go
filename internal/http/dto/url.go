// Файл: internal/http/dto/shortened_link.go
package dto

import (
	"urlshortener/domain/models"
)

// Request types
type (
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

	// Для POST /api/shorten/batch
	ShortenedLinkBatchCreateResponse struct {
		CorrelationID string `json:"correlation_id"`
		ShortURL      string `json:"short_url"`
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

func ShortenedLinkSingleRequestToDomain(r ShortenedLinkSingleRequest) models.ShortenedLink {
	return models.ShortenedLink{
		LongURL: r.URL,
	}
}

func ShortenedLinkSingleResponseFromDomain(url models.ShortenedLink, baseURL string) ShortenedLinkSingleResponse {
	return ShortenedLinkSingleResponse{
		Result: baseURL + "/" + url.ShortCode,
	}
}

func ShortenedLinkBatchRequestsToDomains(reqs []ShortenedLinkBatchRequest) []models.ShortenedLink {
	urls := make([]models.ShortenedLink, len(reqs))
	for i, r := range reqs {
		urls[i] = models.ShortenedLink{
			LongURL: r.OriginalURL,
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
			OriginalURL: url.LongURL,
		}
	}
	return responses
}
