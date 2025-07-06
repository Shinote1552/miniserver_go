package seturljson

import (
	"encoding/json"
	"net/http"
	"urlshortener/internal/httputils"
	"urlshortener/internal/models"
)

type ServiceURLShortener interface {
	GetURL(token string) (string, error)
	SetURL(url string) (string, error)
}

func HandlerSetURLJSON(svc ServiceURLShortener, urlroot string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		var req models.ShortenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
			return
		}

		if req.URL == "" {
			writeJSONError(w, http.StatusBadRequest, "url is required")
			return
		}

		id, err := svc.SetURL(req.URL)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to shorten URL: "+err.Error())
			return
		}

		res := models.ShortenResponse{Result: buildShortURL(urlroot, id)}
		w.Header().Set("Content-Type", httputils.MIMEApplicationJSON)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(res)
	}
}

// EXAMPLE: http://localhost:8080/bzwVcXmW
func buildShortURL(urlroot string, id string) string {
	return "http://" + urlroot + "/" + id
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", httputils.MIMEApplicationJSON)
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(models.ErrorResponse{Error: message})
}
