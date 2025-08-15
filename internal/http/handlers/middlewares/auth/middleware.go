package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"urlshortener/domain/models"
)

//go:generate mockgen
type Authentication interface {
	Register(ctx context.Context, user models.User) (models.User, string, time.Time, error)
	ValidateAndGetUser(ctx context.Context, jwtToken string) (models.User, error)
}

func MiddlewareAuth(auth Authentication) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			fmt.Println("Incoming request to:", r.URL.Path) // Логируем путь

			// 1. Проверка куки
			cookie, cookieErr := r.Cookie("auth_token")
			if cookieErr != nil {
				fmt.Println("No auth_token cookie found:", cookieErr)
			} else {
				fmt.Printf("Found auth token: %s\n", cookie.Value)

				// Добавляем проверку пустого значения
				if cookie.Value == "" {
					fmt.Println("Empty auth_token value")
				} else {
					authUser, validateErr := auth.ValidateAndGetUser(ctx, cookie.Value)
					if validateErr != nil {
						fmt.Printf("Token validation failed: %v\n", validateErr)
					} else {
						fmt.Printf("Successfully authenticated user ID: %d\n", authUser.ID)
						ctx = context.WithValue(ctx, "user_id", authUser.ID)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
				}
			}

			// 2. Создание нового пользователя (только если не удалось аутентифицироваться)
			fmt.Println("Creating new user...")
			var authUser models.User
			authUser, tokenString, tokenExpiry, registerErr := auth.Register(ctx, authUser)
			if registerErr != nil {
				fmt.Printf("User registration failed: %v\n", registerErr)
				http.Error(w, "Authentication failed", http.StatusInternalServerError)
				return
			}

			fmt.Printf("Created new user with ID: %d\n", authUser.ID)
			http.SetCookie(w, &http.Cookie{
				Name:     "auth_token",
				Value:    tokenString,
				Path:     "/",
				Expires:  tokenExpiry.UTC(),
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
			})

			ctx = context.WithValue(ctx, "user_id", authUser.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
