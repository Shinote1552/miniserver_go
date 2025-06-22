package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"urlshortener/internal/deps"
	"urlshortener/internal/service"
)

const (
	contentTypeJSON  = "application/json"
	contentTypePlain = "text/plain"
)

type HandlerURL struct {
	service deps.URLshortener
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

func NewHandlerURL(service *service.URLshortener, baseURL string) *HandlerURL {
	return &HandlerURL{
		service: service,
		baseURL: baseURL,
	}
}

// writeErrorTP выводит ошибку в формате text/plain (TP = Text Plain)
func (h *HandlerURL) writeErrorTP(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", contentTypePlain)
	w.WriteHeader(status)
	w.Write([]byte(message))
}

// writeError выводит ошибку в JSON-формате
func (h *HandlerURL) writeErrorJSON(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func (h *HandlerURL) buildShortURL(id string) string {
	return "http://" + h.baseURL + "/" + id
}

// GET 307
func (h *HandlerURL) GetURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeErrorTP(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/")
	url, err := h.service.GetURL(id)
	if err != nil {
		h.writeErrorTP(w, http.StatusBadRequest, "GetURL Error(): "+err.Error())
		return
	}

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// POST 201
func (h *HandlerURL) SetURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeErrorTP(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.writeErrorTP(w, http.StatusBadRequest, "SetURL Error(): "+err.Error())
		return
	}
	defer r.Body.Close()

	url := string(body)
	if url == "" {
		h.writeErrorTP(w, http.StatusBadRequest, "empty request body")
		return
	}

	id, err := h.service.SetURL(url)
	if err != nil {
		h.writeErrorTP(w, http.StatusBadRequest, "SetURL Error(): "+err.Error())
		return
	}

	w.Header().Set("Content-Type", contentTypePlain)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(h.buildShortURL(id)))
}

// POST /api/shorten 201
func (h *HandlerURL) SetURLwithJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorJSON(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if req.URL == "" {
		h.writeErrorJSON(w, http.StatusBadRequest, "url is required")
		return
	}

	id, err := h.service.SetURL(req.URL)
	if err != nil {
		h.writeErrorJSON(w, http.StatusInternalServerError, "failed to shorten URL: "+err.Error())
		return
	}

	res := Response{Result: h.buildShortURL(id)}
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

// DEFAULT PAGE 400
func (h *HandlerURL) DefaultURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", contentTypePlain)
	w.WriteHeader(http.StatusBadRequest)
	response := fmt.Sprintf("Bad Request (400)\nMethod: %s\nPath: %s",
		r.Method, r.URL.Path)
	w.Write([]byte(response))
}
