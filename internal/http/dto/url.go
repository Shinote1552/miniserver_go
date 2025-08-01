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
		LongURL       string `json:"original_url"`
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
func (r *SingleShortenRequest) ToDomain() models.ShortenedLink {
	return models.ShortenedLink{
		LongURL: r.URL,
	}
}

func BatchRequestsToDomains(reqs []BatchShortenRequest) []models.ShortenedLink {
	urls := make([]models.ShortenedLink, len(reqs))
	for i, r := range reqs {
		urls[i] = models.ShortenedLink{
			LongURL: r.LongURL,
		}
	}
	return urls
}

// Domain → Response
func DomainToSingleResponse(url models.ShortenedLink, baseURL string) SingleShortenResponse {
	return SingleShortenResponse{
		Result: baseURL + "/" + url.ShortCode,
	}
}

func DomainsToBatchResponse(urls []models.ShortenedLink, baseURL string) []BatchShortenResponse {
	responses := make([]BatchShortenResponse, len(urls))
	for i, url := range urls {
		responses[i] = BatchShortenResponse{
			CorrelationID: "",
			ShortURL:      baseURL + "/" + url.ShortCode,
		}
	}
	return responses
}
