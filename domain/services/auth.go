package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"
	"urlshortener/domain/models"

	"github.com/golang-jwt/jwt/v4"
)

//go:generate mockgen -source=auth.go -destination=../../mocks/mock_auth.go -package=mocks
type UserStorage interface {
	UserCreate(ctx context.Context, user models.User) (models.User, error)
	UserGetByID(ctx context.Context, id int64) (models.User, error)
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

func (a *Authentication) Register(ctx context.Context, user models.User) (models.User, string, time.Time, error) {
	user.CreatedAt = time.Now().UTC()

	createdUser, err := a.storage.UserCreate(ctx, user)
	if err != nil {
		return user, "", time.Time{}, fmt.Errorf("failed to create user: %w", err)
	}

	jwtToken, tokenExpiry, err := a.jwtGenerate(createdUser.ID)
	if err != nil {
		return createdUser, "", time.Time{}, fmt.Errorf("failed to generate token: %w", err)
	}

	return createdUser, jwtToken, tokenExpiry, nil
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

type Claims struct {
	jwt.RegisteredClaims
	UserID int64 `json:"UserID"`
}

func (a *Authentication) jwtGenerate(userID int64) (string, time.Time, error) {
	expiryTime := time.Now().Add(a.accessExp)
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiryTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: userID,
	}

	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtToken, err := newToken.SignedString([]byte(a.secretKey))
	if err != nil {
		return "", time.Time{}, err
	}

	return jwtToken, expiryTime, nil
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
		return 0, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return 0, fmt.Errorf("token is invalid")
	}

	fmt.Printf("Decoded claims: %+v\n", claims) // <- добавить для отладки
	return claims.UserID, nil
}
