package models

import (
	"errors"
	"time"
)

type (
	URL struct {
		ID          int
		OriginalURL string
		ShortKey    string
		CreatedAt   time.Time
	}
)

var (
	ErrInvalidData = errors.New("invalid data")
	ErrUnfound     = errors.New("unfound data")
	ErrEmpty       = errors.New("storage is empty")
	ErrConflict    = errors.New("url already exists with different value")
)
