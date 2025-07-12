package models

import "errors"

var (
	ErrInvalidData = errors.New("invalid data")
	ErrUnfound     = errors.New("unfound data")
	ErrEmpty       = errors.New("storage is empty")
)

// URL - основная модель для хранения
type URL struct {
	ID          int    `json:"id"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// ShortenRequest - модель запроса
type ShortenRequest struct {
	URL string `json:"url"`
}

// ShortenResponse - модель ответа
type ShortenResponse struct {
	Result string `json:"result"`
}

// ErrorResponse - модель ошибки
type ErrorResponse struct {
	Error string `json:"error"`
}
