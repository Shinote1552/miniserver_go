package handlers

import (
	"io"
	"net/http"
	"strings"
	"urlshortener/internal/service"
)

type HandlderURL struct {
	service *service.URLshortener
}

func NewHandlderURL(service *service.URLshortener) *HandlderURL {
	return &HandlderURL{
		service: service,
	}
}

// GET 307
func (h *HandlderURL) GetURL(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	id := strings.TrimPrefix(path, "/")
	url, err := h.service.GetURL(id)

	if url == "" || err != nil {
		msg := "GetURL Error(): " + err.Error()
		w.Write([]byte(msg))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// POST 201
func (h *HandlderURL) SetURL(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)

	if err != nil {
		msg := "io.ReadAll(r.Body) Error(): " + err.Error()
		w.Write([]byte(msg))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	text := string(body)
	if text == "" {
		msg := "empty request body"
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(msg))
		return
	}

	id, err := h.service.SetURL(text)
	if err != nil || id == "" {
		msg := "SetURL Error(): " + err.Error()
		w.Write([]byte(msg))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	shortURL := h.service.BaseURL + id

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(shortURL))
	w.WriteHeader(http.StatusCreated)
}

// DEFAULT PAGE 400
func (h *HandlderURL) DefaultURL(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("--DefaultURL 400--"))
	w.WriteHeader(http.StatusBadRequest)
}
