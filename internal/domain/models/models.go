package models

import (
	"errors"
	"time"
)

type (
	User struct {
		ID        int64
		CreatedAt time.Time
	}

	ShortenedLink struct {
		ID          int64  // Уникальный идентификатор
		OriginalURL string // Оригинальный URL в изначальном виде
		ShortCode   string // Короткий код (aBcD12) - сокращенный URL
		UserID      int64  // хэш сумма JWT подписи выданная пользователю
		DeletedFlag bool   // soft delete
		CreatedAt   time.Time
		DeletedAt   time.Time
	}
)

var (
	ErrInvalidData = errors.New("invalid input data")
	ErrUnfound     = errors.New("unfound data")
	ErrEmpty       = errors.New("storage is empty")
	ErrConflict    = errors.New("duplicate URL")
	ErrGone        = errors.New("url deleted")
)
