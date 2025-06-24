package seturljson

import (
	"encoding/json"
	"net/http"
	"urlshortener/internal/deps"
	"urlshortener/internal/httputils"
)

type SetURLJSONHandler struct {
	service deps.ServiceURLShortener
	baseURL string
}

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

type ErrorResponse struct {
	Error string `json:"error"`
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

	var req Request
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

	res := Response{Result: h.buildShortURL(id)}
	w.Header().Set("Content-Type", httputils.ContentTypeJSON)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (h *SetURLJSONHandler) buildShortURL(id string) string {
	return h.baseURL + "/" + id
}

func (h *SetURLJSONHandler) writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", httputils.ContentTypeJSON)
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}
