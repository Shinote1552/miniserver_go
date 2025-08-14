package delete_batch

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"urlshortener/domain/models"
	"urlshortener/internal/http/httputils"
)

type ServiceURLShortener interface {
	DeleteURLs(ctx context.Context, userID int64, shortURLs []string) ([]string, []string, error)
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

		deleted, failed, err := svc.DeleteURLs(ctx, userID, shortURLs)

		switch {
		case errors.Is(err, models.ErrUnfound):
			httputils.WriteJSONResponse(w, http.StatusNotFound, map[string]interface{}{
				"error":  "none of the URLs were found or owned by user",
				"failed": failed,
			})
		case errors.Is(err, models.ErrPartialDeletion):
			httputils.WriteJSONResponse(w, http.StatusAccepted, map[string]interface{}{
				"message": "some URLs were not found or not owned by user",
				"deleted": deleted,
				"failed":  failed,
			})
		case err != nil:
			httputils.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
		default:
			httputils.WriteJSONResponse(w, http.StatusAccepted, map[string]interface{}{
				"message": "all URLs scheduled for deletion",
				"deleted": deleted,
			})
		}
	}
}
