package middlewares

import (
	"context"
	"net/http"
	"urlshortener/domain/models"
)

//go:generate mockgen
type Authentication interface {
	Register(ctx context.Context, user models.User) (models.User, string, error)
	ValidateAndGetUser(ctx context.Context, jwtToken string) (models.User, error)
	GetUserLinks(ctx context.Context, jwtToken string) ([]models.ShortenedLink, error)
}

func MiddlewareAuth(auth Authentication) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			var user models.User
			var tokenString string
			var err error

			cookie, err := r.Cookie("auth_token")
			if err == nil {
				tokenString = cookie.Value
				user, err = auth.ValidateAndGetUser(ctx, tokenString)
			}

			if err != nil {
				user, tokenString, err = auth.Register(ctx)
			}

			/*
				увидел вот такуб идею еще:
				Добавляем пользователя в контекст
				ctx := context.WithValue(r.Context(), "user", user)
			*/
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
