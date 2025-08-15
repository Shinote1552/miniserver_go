package list_user_urls

import (
	"context"
	"errors"
	"net/http"
	"urlshortener/domain/models"
	"urlshortener/internal/http/dto"
	"urlshortener/internal/http/httputils"

	"github.com/rs/zerolog"
)

type ServiceURLShortener interface {
	GetUserLinks(ctx context.Context, userID int64) ([]models.ShortenedLink, error)
}

func HandlerGetURLJsonBatch(svc ServiceURLShortener, urlroot string, logger zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger = logger.With().Str("handler", "HandlerGetURLJsonBatch").Logger()

		// Логируем входящий запрос
		logger.Debug().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("Incoming request")

		// Получаем userID из контекста
		userID, ok := ctx.Value("user_id").(int64)
		if !ok || userID == 0 {
			logger.Error().
				Msg("Failed to get user_id from context or user_id is 0")
			httputils.WriteJSONError(w, http.StatusUnauthorized, "authentication required")
			return
		}

		logger = logger.With().Int64("user_id", userID).Logger()
		logger.Debug().Msg("Getting user links")

		// Получаем ссылки пользователя
		shortLinks, err := svc.GetUserLinks(ctx, userID)
		if err != nil {
			logger.Error().
				Err(err).
				Msg("Failed to get user links from service")

			if errors.Is(err, models.ErrUnfound) || errors.Is(err, models.ErrEmpty) {
				logger.Debug().Msg("No content found for user")
				w.WriteHeader(http.StatusNoContent)
				return
			}

			httputils.WriteJSONError(w, http.StatusInternalServerError, "failed to get user URLs")
			return
		}

		if len(shortLinks) == 0 {
			logger.Debug().Msg("User has no shortened links")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		logger.Debug().
			Int("links_count", len(shortLinks)).
			Msg("Successfully retrieved user links")

		// Формируем и отправляем ответ
		response := dto.ShortenedLinkBatchGetResponseFromDomains(shortLinks, urlroot)
		httputils.WriteJSONResponse(w, http.StatusOK, response)
	}
}
