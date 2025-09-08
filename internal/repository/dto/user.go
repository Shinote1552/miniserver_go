package dto

import (
	"time"
	"urlshortener/internal/domain/models"
)

type (
	UserDB struct {
		ID        int64     `db:"id"`
		CreatedAt time.Time `db:"created_at"`
	}
)

func UserDBToDomain(u UserDB) models.User {
	return models.User{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
	}
}

func UserDBFromDomain(u models.User) UserDB {
	return UserDB{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
	}
}
