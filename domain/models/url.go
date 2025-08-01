package models

import (
	"errors"
	"time"
)

type (
	ShortenedLink struct {
		ID        int
		LongURL   string
		ShortCode string
		CreatedAt time.Time
	}
)

var (
	ErrInvalidData = errors.New("invalid input data")
	ErrUnfound     = errors.New("unfound data")
	ErrEmpty       = errors.New("storage is empty")
	ErrConflict    = errors.New("duplicate URL")
)
