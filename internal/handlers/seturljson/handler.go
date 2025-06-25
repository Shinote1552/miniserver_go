package seturljson

import (
	"encoding/json"
	"net/http"
	"urlshortener/internal/deps"
	"urlshortener/internal/httputils"
	"urlshortener/internal/models"
)

type SetURLJSONHandler struct {
	service deps.ServiceURLShortener
	baseURL string
}

func New(service deps.ServiceURLShortener, baseURL string) *SetURLJSONHandler {
	return &SetURLJSONHandler{
		service: service,
		baseURL: baseURL,
	}
}

func (h *SetURLJSONHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req models.SetURLJSONRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSONError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if req.URL == "" {
		h.writeJSONError(w, http.StatusBadRequest, "url is required")
		return
	}

	id, err := h.service.SetURL(req.URL)
	if err != nil {
		h.writeJSONError(w, http.StatusInternalServerError, "failed to shorten URL: "+err.Error())
		return
	}

	res := models.SetURLJSONResponse{URLShort: h.buildShortURL(id)}
	w.Header().Set("Content-Type", httputils.ContentTypeJSON)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

// EXAMPLE: http://http://localhost:8080/bzwVcXmW
func (h *SetURLJSONHandler) buildShortURL(id string) string {
	return h.baseURL + "/" + id
}

func (h *SetURLJSONHandler) writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", httputils.ContentTypeJSON)
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(models.SetURLJSONErrorResponse{Error: message})
}
