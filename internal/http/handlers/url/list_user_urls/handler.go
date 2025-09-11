package list_user_urls

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"urlshortener/internal/domain/models"
	"urlshortener/internal/http/dto"
	"urlshortener/internal/http/httputils"
)

type ServiceURLShortener interface {
	GetUserLinks(ctx context.Context, userID int64) ([]models.ShortenedLink, error)
}

func HandlerGetURLJsonBatch(svc ServiceURLShortener, urlroot string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		userID, ok := ctx.Value("user_id").(int64)
		if !ok || userID == 0 {
			httputils.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
			return
		}

		shortLinks, err := svc.GetUserLinks(ctx, userID)

		if err != nil {
			if errors.Is(err, models.ErrUnfound) || errors.Is(err, models.ErrEmpty) {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			if strings.Contains(err.Error(), "failed to validate userID") {
				httputils.WriteJSONError(w, http.StatusBadRequest, "invalid user ID")
				return
			}
			httputils.WriteJSONError(w, http.StatusInternalServerError,
				fmt.Sprintf("failed to get user URLs: %v", err))
			return
		}

		if len(shortLinks) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		response := dto.ShortenedLinkBatchGetResponseFromDomains(shortLinks, urlroot)
		httputils.WriteJSONResponse(w, http.StatusOK, response)
	}
}
