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
	w.Write([]byte(url))
	w.Write([]byte("\n"))
	w.Write([]byte("\n--GetURL 307 SUCCESS--\n ID: " + id))
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

	// Тело запроса как строка
	text := string(body)
	if text == "" {
		msg := "empty request body"
		w.Write([]byte(msg))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := h.service.SetURL(text)
	if err != nil || id == "" {
		msg := "SetURL Error(): " + err.Error()
		w.Write([]byte(msg))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	msg := "\nhttp://localhost:8080/" + id
	w.Write([]byte(msg))
	w.Write([]byte("\n"))
	w.Write([]byte("\n--SetURL 201--\n"))
	w.WriteHeader(http.StatusCreated)
}

// DEFAULT PAGE 400
func (h *HandlderURL) DefaultURL(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("--DefaultURL 400--"))
	w.WriteHeader(http.StatusBadRequest)
}
