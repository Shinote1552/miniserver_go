package seturltext

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"urlshortener/internal/httputils"
	"urlshortener/internal/models"
)

type ServiceURLShortener interface {
	SetURL(ctx context.Context, url string) (string, error)
}

func HandlerSetURLText(svc ServiceURLShortener, urlroot string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeTextPlainError(w, http.StatusBadRequest, models.ErrInvalidData.Error())
			return
		}
		defer r.Body.Close()

		url := string(body)
		if url == "" {
			writeTextPlainError(w, http.StatusBadRequest, models.ErrInvalidData.Error())
			return
		}

		id, err := svc.SetURL(ctx, url)
		if err != nil {
			if errors.Is(err, models.ErrConflict) {
				w.Header().Set("Content-Type", httputils.MIMETextPlain)
				w.WriteHeader(http.StatusConflict)
				w.Write([]byte(buildShortURL(urlroot, id)))
				return
			}
			writeTextPlainError(w, http.StatusBadRequest, err.Error())
			return
		}

		w.Header().Set("Content-Type", httputils.MIMETextPlain)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(buildShortURL(urlroot, id)))
	}
}

func buildShortURL(urlroot, id string) string {
	return fmt.Sprintf("http://%s/%s", urlroot, id)
}

func writeTextPlainError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", httputils.MIMETextPlain)
	w.WriteHeader(status)
	w.Write([]byte(message))
}
