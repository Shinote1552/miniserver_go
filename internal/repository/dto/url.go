package dto

import (
	"time"
	"urlshortener/domain/models"
)

// DTO БД для работы с БД
type (
	ShortenedLinkDB struct {
		ID          int64     `db:"id"`
		OriginalURL string    `db:"original_url"`
		ShortCode   string    `db:"short_key"`
		UserID      int64     `db:"user_id"`
		IsDeleted   bool      `db:"is_deleted"`
		CreatedAt   time.Time `db:"created_at"`
	}
)

func ShortenedLinkDBToDomain(d ShortenedLinkDB) models.ShortenedLink {
	return models.ShortenedLink{
		ID:          d.ID,
		OriginalURL: d.OriginalURL,
		ShortCode:   d.ShortCode,
		UserID:      d.UserID,
		CreatedAt:   d.CreatedAt,
	}
}

func ShortenedLinkDBFromDomain(url models.ShortenedLink) ShortenedLinkDB {
	return ShortenedLinkDB{
		ID:          url.ID,
		OriginalURL: url.OriginalURL,
		ShortCode:   url.ShortCode,
		UserID:      url.UserID,
		CreatedAt:   url.CreatedAt,
	}
}
