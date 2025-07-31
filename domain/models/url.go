package models

import "time"

type (
	URL struct {
		ID          int
		OriginalURL string
		ShortKey    string
		CreatedAt   time.Time
	}
)
