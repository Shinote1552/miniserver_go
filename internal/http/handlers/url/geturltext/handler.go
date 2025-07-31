package geturltext

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"urlshortener/domain/models"
	"urlshortener/internal/http/httputils"
)

type ServiceURLShortener interface {
	GetURL(ctx context.Context, shortKey string) (models.URL, error)
}

func writeTextPlainError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", httputils.MIMETextPlain)
	w.WriteHeader(status)
	w.Write([]byte(message))
}

func HandlerGetURLWithID(svc ServiceURLShortener) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := strings.TrimPrefix(r.URL.Path, "/")

		url, err := svc.GetURL(ctx, id)
		if err != nil {
			writeTextPlainError(w, http.StatusBadRequest, fmt.Sprintf("GetURL Error(): %v", err))
			return
		}

		w.Header().Set("Location", url.OriginalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}
