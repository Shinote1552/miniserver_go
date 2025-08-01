package seturltext

import (
	"context"
	"errors"
	"io"
	"net/http"
	"urlshortener/domain/models"
	"urlshortener/internal/http/httputils"
)

type ServiceURLShortener interface {
	SetURL(ctx context.Context, longUrl string) (models.ShortenedLink, error)
}

func HandlerSetURLText(svc ServiceURLShortener, urlroot string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			httputils.WriteTextError(w, http.StatusBadRequest, httputils.ErrInvalidData.Error())
			return
		}
		defer r.Body.Close()

		url := string(body)
		if url == "" {
			httputils.WriteTextError(w, http.StatusBadRequest, httputils.ErrInvalidData.Error())
			return
		}

		urlModel, err := svc.SetURL(ctx, url)
		if err != nil {
			if errors.Is(err, httputils.ErrConflict) {
				w.Header().Set("Content-Type", httputils.MIMETextPlain)
				w.WriteHeader(http.StatusConflict)
				w.Write([]byte(httputils.BuildShortURL(urlroot, urlModel.ShortCode)))
				return
			}
			httputils.WriteTextError(w, http.StatusBadRequest, err.Error())
			return
		}

		w.Header().Set("Content-Type", httputils.MIMETextPlain)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(httputils.BuildShortURL(urlroot, urlModel.ShortCode)))
	}
}
