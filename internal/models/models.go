package models

import "errors"

var (
	ErrInvalidData = errors.New("invalid data")
	ErrUnfound     = errors.New("unfound data")
	ErrEmpty       = errors.New("storage is empty")
	ErrConflict    = errors.New("url already exists with different value")
)

// Storage модели (для работы с хранилищем)
type (
	// StorageURLModel - основная модель хранения URL
	StorageURLModel struct {
		ID          int    `json:"id"`
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}
)

// API модели (для HTTP-взаимодействия)
type (
	// APIShortenRequest - запрос на сокращение URL
	APIShortenRequest struct {
		URL string `json:"url"`
	}

	// APIShortenResponse - ответ с сокращённым URL
	APIShortenResponse struct {
		Result string `json:"result"`
	}

	// APIErrorResponse - модель ошибки API
	APIErrorResponse struct {
		Error string `json:"error"`
	}

	// APIBatchRequestItem - элемент пакетного запроса
	APIShortenRequestBatch struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"`
	}

	// APIBatchResponseItem - элемент пакетного ответа
	APIShortenResponseBatch struct {
		CorrelationID string `json:"correlation_id"`
		ShortURL      string `json:"short_url"`
	}
)

// // Service модели (для бизнес-логики, если потребуется)
// type (
// 	// ServiceURL - модель для работы сервиса
// 	ServiceURL struct {
// 		Short    string
// 		Original string
// 	}
// )
