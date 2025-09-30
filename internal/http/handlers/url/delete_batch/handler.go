package delete_batch

import (
	"context"
	"encoding/json"
	"net/http"
	"urlshortener/internal/http/httputils"
)

type ServiceURLShortener interface {
	BatchDelete(ctx context.Context, userID int64, shortCode []string) error
}

func HandlerDeleteURLBatch(svc ServiceURLShortener) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value("user_id").(int64)
		if !ok || userID == 0 {
			httputils.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
			return
		}

		var shortCode []string
		if err := json.NewDecoder(r.Body).Decode(&shortCode); err != nil {
			httputils.WriteJSONError(w, http.StatusBadRequest, "invalid request format")
		}

		if len(shortCode) == 0 {
			httputils.WriteJSONError(w, http.StatusBadRequest, "Empty URL list")
			return
		}

		if err := svc.BatchDelete(r.Context(), userID, shortCode); err != nil {
			httputils.WriteJSONError(w, http.StatusInternalServerError, "deletion failed")
		}

		w.WriteHeader(http.StatusAccepted)
	}
}
