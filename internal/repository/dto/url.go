package dto

import (
	"time"
	"urlshortener/domain/models"
)

// DTO БД для работы с БД
type (
	URLDB struct {
		ID          int       `db:"id"`
		OriginalURL string    `db:"original_url"`
		ShortKey    string    `db:"short_key"`
		CreatedAt   time.Time `db:"created_at"`
	}
)

// ToDomain преобразует DTO БД в доменную модель
func (d *URLDB) ToDomain() *models.URL {
	return &models.URL{
		ID:          d.ID,
		OriginalURL: d.OriginalURL,
		ShortKey:    d.ShortKey,
		CreatedAt:   d.CreatedAt,
	}
}

// FromDomain преобразует доменную модель в DTO БД
func FromDomain(url models.URL) *URLDB {
	return &URLDB{
		ID:          url.ID,
		OriginalURL: url.OriginalURL,
		ShortKey:    url.ShortKey,
		CreatedAt:   url.CreatedAt,
	}
}



