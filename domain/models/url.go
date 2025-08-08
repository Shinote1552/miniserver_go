package models

import (
	"errors"
	"time"
)

type (
	ShortenedLink struct {
		ID        int    // Уникальный идентификатор в БД
		LongURL   string // Оригинальный URL в изначальном виде
		ShortCode string // Короткий код (aBcD12) - сокращенный URL
		UserToken string // хэш сумма JWT подписи выданная пользователю
		CreatedAt time.Time
	}
)

var (
	ErrInvalidData = errors.New("invalid input data")
	ErrUnfound     = errors.New("unfound data")
	ErrEmpty       = errors.New("storage is empty")
	ErrConflict    = errors.New("duplicate URL")
)
