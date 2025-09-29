package dto

import (
	"time"
	"urlshortener/internal/domain/models"
)

// DTO БД для работы с БД
type (
	ShortenedLinkDB struct {
		ID          int64     `db:"id"`
		OriginalURL string    `db:"original_url"`
		ShortCode   string    `db:"short_key"`
		UserID      int64     `db:"user_id"`
		DeletedFlag bool      `db:"is_deleted"`
		CreatedAt   time.Time `db:"created_at"`
		DeletedAt   time.Time `db:"deleted_at"`
	}
)

func ShortenedLinkDBToDomain(domain ShortenedLinkDB) models.ShortenedLink {
	return models.ShortenedLink{
		ID:          domain.ID,
		OriginalURL: domain.OriginalURL,
		ShortCode:   domain.ShortCode,
		UserID:      domain.UserID,
		DeletedFlag: domain.DeletedFlag,
		CreatedAt:   domain.CreatedAt,
		DeletedAt:   domain.DeletedAt,
	}
}

func ShortenedLinkDBFromDomain(db models.ShortenedLink) ShortenedLinkDB {
	return ShortenedLinkDB{
		ID:          db.ID,
		OriginalURL: db.OriginalURL,
		ShortCode:   db.ShortCode,
		UserID:      db.UserID,
		DeletedFlag: db.DeletedFlag,
		CreatedAt:   db.CreatedAt,
		DeletedAt:   db.DeletedAt,
	}
}
