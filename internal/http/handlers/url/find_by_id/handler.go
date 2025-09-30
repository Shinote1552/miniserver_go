package find_by_id

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"urlshortener/internal/domain/models"
	"urlshortener/internal/http/httputils"
)

type ServiceURLShortener interface {
	GetURL(ctx context.Context, shortKey string) (models.ShortenedLink, error)
}

func HandlerGetURLWithID(svc ServiceURLShortener) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := strings.TrimPrefix(r.URL.Path, "/")

		url, err := svc.GetURL(ctx, id)

		if err != nil {
			if errors.Is(err, models.ErrGone) {
				httputils.WriteTextError(w, http.StatusGone, fmt.Sprintf("GetURL Error(): %v", err))
				return
			}
			httputils.WriteTextError(w, http.StatusBadRequest, fmt.Sprintf("GetURL Error(): %v", err))
			return
		}
		httputils.WriteRedirect(w, url.OriginalURL, false)
	}
}
