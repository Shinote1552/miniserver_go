package seturljson

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"urlshortener/internal/httputils"
	"urlshortener/internal/models"
)

type ServiceURLShortener interface {
	SetURL(ctx context.Context, url string) (string, error)
}

func HandlerSetURLJson(svc ServiceURLShortener, urlroot string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if r.Method != http.MethodPost {
			writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		var req models.APIShortenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, models.ErrInvalidData.Error())
			return
		}

		if req.URL == "" {
			writeJSONError(w, http.StatusBadRequest, models.ErrInvalidData.Error())
			return
		}

		id, err := svc.SetURL(ctx, req.URL)
		if err != nil {
			if errors.Is(err, models.ErrConflict) {
				res := models.APIShortenResponse{
					Result: buildShortURL(urlroot, id),
				}
				w.Header().Set("Content-Type", httputils.MIMEApplicationJSON)
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(res)
				return
			}
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}

		res := models.APIShortenResponse{Result: buildShortURL(urlroot, id)}

		w.Header().Set("Content-Type", httputils.MIMEApplicationJSON)
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(res); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to encode response")
			return
		}
	}
}

func buildShortURL(urlroot, id string) string {
	return fmt.Sprintf("http://%s/%s", urlroot, id)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", httputils.MIMEApplicationJSON)
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(models.APIErrorResponse{Error: message})
}
