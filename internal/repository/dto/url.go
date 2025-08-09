package dto

import (
	"time"
	"urlshortener/domain/models"
)

// DTO БД для работы с БД
type (
	ShortenedLinkDB struct {
		ID        int64     `db:"id"`
		LongURL   string    `db:"original_url"`
		ShortCode string    `db:"short_key"`
		UserID    int64     `db:"user_id"`
		CreatedAt time.Time `db:"created_at"`
	}
)

func ShortenedLinkDBToDomain(d ShortenedLinkDB) models.ShortenedLink {
	return models.ShortenedLink{
		ID:        d.ID,
		LongURL:   d.LongURL,
		ShortCode: d.ShortCode,
		UserID:    d.UserID,
		CreatedAt: d.CreatedAt,
	}
}

func ShortenedLinkDBFromDomain(url models.ShortenedLink) ShortenedLinkDB {
	return ShortenedLinkDB{
		ID:        url.ID,
		LongURL:   url.LongURL,
		ShortCode: url.ShortCode,
		UserID:    url.UserID,
		CreatedAt: url.CreatedAt,
	}
}
