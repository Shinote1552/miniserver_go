package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"
	"urlshortener/domain/models"

	"github.com/golang-jwt/jwt/v4"
)

type UserStorage interface {
	UserCreate(ctx context.Context, user models.User) (models.User, error)
	UserGetByID(ctx context.Context, id int64) (models.User, error)
	ShortenedLinkGetBatchByUser(ctx context.Context, id int64) ([]models.ShortenedLink, error)
}

type Authentication struct {
	storage   UserStorage
	secretKey []byte
	accessExp time.Duration
}

func NewAuthentication(userStorage UserStorage, secretKey string, accessExp time.Duration) (*Authentication, error) {
	key, err := base64.StdEncoding.DecodeString(secretKey)
	if err != nil || len(key) < 32 {
		return nil, fmt.Errorf("invalid JWT secret key: must be at least 32 bytes when decoded")
	}

	return &Authentication{
		storage:   userStorage,
		secretKey: key,
		accessExp: accessExp,
	}, nil
}

func (a *Authentication) Register(ctx context.Context, user models.User) (models.User, string, error) {
	// надо ли добавлять валидацию сюда?
	// if user.ID != 0{
	// 	return  user,
	// }

	user.CreatedAt = time.Now().UTC()

	createdUser, err := a.storage.UserCreate(ctx, user)
	if err != nil {
		return user, "", fmt.Errorf("failed to create user: %w", err)
	}

	jwtToken, err := a.jwtGenerate(createdUser.ID)
	if err != nil {
		// надо ли удлаить юзера из за ошибки генерации токена?
		return createdUser, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return createdUser, jwtToken, nil

}

func (a *Authentication) ValidateAndGetUser(ctx context.Context, jwtToken string) (models.User, error) {
	userID, err := a.getUserId(jwtToken)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to validate token: %w", err)
	}

	user, err := a.storage.UserGetByID(ctx, userID)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (a *Authentication) GetUserLinks(ctx context.Context, jwtToken string) ([]models.ShortenedLink, error) {
	userID, err := a.getUserId(jwtToken)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	userLinks, err := a.storage.ShortenedLinkGetBatchByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get links: %w", err)
	}

	return userLinks, nil

}

type Claims struct {
	jwt.RegisteredClaims
	UserID int64
}

func (a *Authentication) jwtGenerate(userID int64) (string, error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.accessExp)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: userID,
	}
	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	jwtToken, err := newToken.SignedString([]byte(a.secretKey))
	if err != nil {
		return "", err
	}

	return jwtToken, nil
}

// Одновременно здесь происходит валидация токена
func (a *Authentication) getUserId(tokenString string) (int64, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(a.secretKey), nil
		})
	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, fmt.Errorf("Token is not valid")
	}

	fmt.Println("Token os valid")
	return claims.UserID, nil
}
