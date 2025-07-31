package dto

import (
	"urlshortener/domain/models"
)

// Request
type (
	SingleShortenRequest struct {
		URL string `json:"url"`
	}

	BatchShortenRequest struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"`
	}
)

// Response
type (
	SingleShortenResponse struct {
		Result string `json:"result"`
	}

	BatchShortenResponse struct {
		CorrelationID string `json:"correlation_id"`
		ShortURL      string `json:"short_url"`
	}

	ErrorResponse struct {
		Error string `json:"error"`
	}
)

// Request → Domain
func (r *SingleShortenRequest) ToDomain() models.URL {
	return models.URL{
		OriginalURL: r.URL,
	}
}

func BatchRequestsToDomains(reqs []BatchShortenRequest) []models.URL {
	urls := make([]models.URL, len(reqs))
	for i, r := range reqs {
		urls[i] = models.URL{
			OriginalURL: r.OriginalURL,
		}
	}
	return urls
}

// Domain → Response
func DomainToSingleResponse(url models.URL, baseURL string) SingleShortenResponse {
	return SingleShortenResponse{
		Result: baseURL + "/" + url.ShortKey,
	}
}

func DomainsToBatchResponse(urls []models.URL, baseURL string) []BatchShortenResponse {
	responses := make([]BatchShortenResponse, len(urls))
	for i, url := range urls {
		responses[i] = BatchShortenResponse{
			CorrelationID: "",
			ShortURL:      baseURL + "/" + url.ShortKey,
		}
	}
	return responses
}
