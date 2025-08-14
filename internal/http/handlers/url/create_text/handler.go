package create_text

import (
	"context"
	"errors"
	"io"
	"net/http"
	"urlshortener/domain/models"
	"urlshortener/internal/http/dto"
	"urlshortener/internal/http/httputils"
)

type ServiceURLShortener interface {
	SetURL(ctx context.Context, model models.ShortenedLink) (models.ShortenedLink, error)
}

func HandlerSetURLText(svc ServiceURLShortener, urlroot string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if r.Method != http.MethodPost {
			httputils.WriteTextError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		userID, ok := ctx.Value("user_id").(int64)
		if !ok || userID == 0 {
			httputils.WriteTextError(w, http.StatusUnauthorized, "authentication required")
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			httputils.WriteTextError(w, http.StatusBadRequest, httputils.ErrInvalidData.Error())
			return
		}
		defer r.Body.Close()

		req := dto.ShortenedLinkTextRequest{
			URL: string(body),
		}

		if req.URL == "" {
			httputils.WriteTextError(w, http.StatusBadRequest, httputils.ErrInvalidData.Error())
			return
		}

		model := dto.ShortenedLinkTextRequestToDomain(req, userID)
		urlModel, err := svc.SetURL(ctx, model)

		if err != nil {
			if errors.Is(err, httputils.ErrConflict) {
				resp := dto.ShortenedLinkTextResponseFromDomain(urlModel, urlroot)
				httputils.WriteTextResponse(w, http.StatusConflict, resp.ShortURL)
				return
			}
			httputils.WriteTextError(w, http.StatusBadRequest, err.Error())
			return
		}

		resp := dto.ShortenedLinkTextResponseFromDomain(urlModel, urlroot)
		httputils.WriteTextResponse(w, http.StatusCreated, resp.ShortURL)
	}
}
