package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"
	"urlshortener/domain/models"
)

type UserStorage interface {
	Create(ctx context.Context, user *models.User) (models.User, error)
	GetByID(ctx context.Context, id int64) (models.User, error)
	GetByUUID(ctx context.Context, uuid string) (models.User, error)
	GetUserLinks(ctx context.Context, userID int64) ([]models.ShortenedLink, error)
}

type Authentication struct {
	storage    UserStorage
	secretKey  []byte
	accessExp  time.Duration
	refreshExp time.Duration
}

func NewAuthService(userStorage UserStorage, secretKey string, accessExp, refreshExp time.Duration) (Authentication, error) {
	key, err := base64.StdEncoding.DecodeString(secretKey)
	if err != nil || len(key) < 32 {
		return Authentication{}, fmt.Errorf("invalid JWT secret key: must be at least 32 bytes when decoded")
	}

	return Authentication{
		storage:    userStorage,
		secretKey:  key,
		accessExp:  accessExp,
		refreshExp: refreshExp,
	}, nil
}

/*
JWT lib: "github.com/golang-jwt/jwt/v4"

*/
