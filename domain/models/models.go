package models

import (
	"errors"
	"time"
)

type (
	User struct {
		ID        int64 // Уникальный идентификатор
		CreatedAt time.Time
	}

	ShortenedLink struct {
		ID          int64  // Уникальный идентификатор
		OriginalURL string // Оригинальный URL в изначальном виде
		ShortCode   string // Короткий код (aBcD12) - сокращенный URL
		UserID      int64  // хэш сумма JWT подписи выданная пользователю
		IsDeleted   bool   //TODO хз нужен он вообще?
		CreatedAt   time.Time
	}
)

var (
	ErrInvalidData     = errors.New("invalid input data")
	ErrUnfound         = errors.New("url not found")
	ErrEmpty           = errors.New("storage is empty")
	ErrConflict        = errors.New("duplicate URL")
	ErrDeleted         = errors.New("url is already deleted")
	ErrNotOwner        = errors.New("user is not owner of url")
	ErrPartialDeletion = errors.New("some urls were not deleted")
)
