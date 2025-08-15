package list_user_urls

import (
	"context"
	"errors"
	"net/http"
	"urlshortener/domain/models"
	"urlshortener/internal/http/dto"
	"urlshortener/internal/http/httputils"
)

type ServiceURLShortener interface {
	GetUserLinks(ctx context.Context, userID int64) ([]models.ShortenedLink, error)
}

func HandlerGetURLJsonBatch(svc ServiceURLShortener, urlroot string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := ctx.Value("user_id").(int64)
		if !ok || userID == 0 {
			httputils.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
			return
		}

		shortLinks, err := svc.GetUserLinks(ctx, userID)

		if err != nil || len(shortLinks) == 0 {
			if errors.Is(err, models.ErrUnfound) || errors.Is(err, models.ErrEmpty) {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			httputils.WriteJSONError(w, http.StatusInternalServerError, "failed to get user URLs")
			return
		}

		response := dto.ShortenedLinkBatchGetResponseFromDomains(shortLinks, urlroot)
		httputils.WriteJSONResponse(w, http.StatusOK, response)
	}
}
