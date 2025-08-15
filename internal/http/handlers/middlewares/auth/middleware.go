package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"urlshortener/domain/models"
	"urlshortener/internal/http/httputils"
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

			// 1. Пытаемся получить валидного пользователя из куки
			cookie, cookieErr := r.Cookie("auth_token")
			if cookieErr == nil && cookie.Value != "" {
				authUser, validateErr := auth.ValidateAndGetUser(ctx, cookie.Value)
				if validateErr == nil {
					ctx = context.WithValue(ctx, "user_id", authUser.ID)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			// 2. Если куки нет или она невалидна - создаем нового пользователя
			var authUser models.User
			authUser, tokenString, tokenExpiry, registerErr := auth.Register(ctx, authUser)
			if registerErr != nil {
				http.Error(w, "Authentication failed", http.StatusInternalServerError)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     "auth_token",
				Value:    tokenString,
				Path:     "/",
				Expires:  tokenExpiry,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
				// Надо будет пробывать SameSiteStrictMode
			})

			httputils.WriteTextResponse(w, http.StatusCreated, fmt.Sprintf("Authentication token issued, authUserID: %d", authUser.ID))
		})
	}
}
