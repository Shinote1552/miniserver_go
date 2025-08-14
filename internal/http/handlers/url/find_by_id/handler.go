package find_by_id

import (
	"context"
	"errors"
	"net/http"
	"urlshortener/domain/models"
	"urlshortener/internal/http/httputils"

	"github.com/gorilla/mux"
)

type ServiceURLShortener interface {
	GetURL(ctx context.Context, shortKey string) (models.ShortenedLink, error)
}

func HandlerGetURLWithID(svc ServiceURLShortener) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		vars := mux.Vars(r)
		shortKey := vars["id"]

		url, err := svc.GetURL(ctx, shortKey)
		if err != nil {
			if errors.Is(err, models.ErrUnfound) {
				httputils.WriteJSONError(w, http.StatusNotFound, "URL не найден")
				return
			}
			httputils.WriteJSONError(w, http.StatusInternalServerError, "ошибка получения URL")
			return
		}

		if url.IsDeleted {
			w.WriteHeader(http.StatusGone) // 410 для удаленных URL
			return
		}

		http.Redirect(w, r, url.OriginalURL, http.StatusTemporaryRedirect)
	}
}
