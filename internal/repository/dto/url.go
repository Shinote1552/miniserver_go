package dto

import (
	"time"
	"urlshortener/domain/models"
)

// DTO БД для работы с БД
type (
	URLDB struct {
		ID        int       `db:"id"`
		LongURL   string    `db:"original_url"`
		ShortCode string    `db:"short_key"`
		UserToken string    `db:"user_token_hash"`
		CreatedAt time.Time `db:"created_at"`
	}
)

// ToDomain преобразует DTO БД в доменную модель
func (d *URLDB) ToDomain() *models.ShortenedLink {
	return &models.ShortenedLink{
		ID:        d.ID,
		LongURL:   d.LongURL,
		ShortCode: d.ShortCode,
		UserToken: d.UserToken,
		CreatedAt: d.CreatedAt,
	}
}

// FromDomain преобразует доменную модель в DTO БД
func FromDomain(url models.ShortenedLink) *URLDB {
	return &URLDB{
		ID:        url.ID,
		LongURL:   url.LongURL,
		ShortCode: url.ShortCode,
		UserToken: url.UserToken,
		CreatedAt: url.CreatedAt,
	}
}
