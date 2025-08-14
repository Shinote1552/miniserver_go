package geturljsonbatch

import (
	"context"
	"net/http"
	"urlshortener/domain/models"
)

type ServiceURLShortener interface {
	GetUserLinks(ctx context.Context, userID int64) ([]models.ShortenedLink, error)
}

func HandlerGetURLJsonBatch(svc ServiceURLShortener, urlroot string) http.HandlerFunc {

}
