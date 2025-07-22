package geturl

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"urlshortener/internal/httputils"
)

type ServiceURLShortener interface {
	GetURL(ctx context.Context, token string) (string, error)
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

		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}
