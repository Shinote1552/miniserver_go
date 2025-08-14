package delete_batch

import (
	"context"
	"encoding/json"
	"net/http"
	"urlshortener/internal/http/httputils"
)

type ServiceURLShortener interface {
	DeleteURLs(ctx context.Context, userID int64, shortURLs []string) error
}

func HandlerDeleteURLBatch(svc ServiceURLShortener) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := ctx.Value("user_id").(int64)
		if !ok || userID == 0 {
			httputils.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
			return
		}

		var shortURLs []string
		if err := json.NewDecoder(r.Body).Decode(&shortURLs); err != nil {
			httputils.WriteJSONError(w, http.StatusBadRequest, "invalid request format")
			return
		}

		if len(shortURLs) == 0 {
			httputils.WriteJSONError(w, http.StatusBadRequest, "empty URL list")
			return
		}

		// Вызываем сервис (удаление будет выполнено асинхронно)
		_ = svc.DeleteURLs(ctx, userID, shortURLs)

		w.WriteHeader(http.StatusAccepted)
	}
}
