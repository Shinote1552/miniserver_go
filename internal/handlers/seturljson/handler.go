package seturljson

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"urlshortener/internal/httputils"
	"urlshortener/internal/models"
)

type ServiceURLShortener interface {
	SetURL(ctx context.Context, url string) (string, error)
}

func HandlerSetURLJSON(svc ServiceURLShortener, urlroot string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if r.Method != http.MethodPost {
			writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		var req models.APIShortenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
			return
		}

		if req.URL == "" {
			writeJSONError(w, http.StatusBadRequest, "url is required")
			return
		}

		id, err := svc.SetURL(ctx, req.URL)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to shorten URL: %v", err))
			return
		}

		res := models.APIShortenResponse{Result: buildShortURL(urlroot, id)}
		w.Header().Set("Content-Type", httputils.MIMEApplicationJSON)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(res)
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
