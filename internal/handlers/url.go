package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"urlshortener/internal/service"
)

type HandlderURL struct {
	service URLshortener
	servurl string
}

func NewHandlerURL(service *service.URLshortener, url string) *HandlderURL {
	return &HandlderURL{
		service: service,
		servurl: url,
	}
}

// GET 307
func (h *HandlderURL) GetURL(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	id := strings.TrimPrefix(path, "/")
	url, err := h.service.GetURL(id)

	if url == "" || err != nil {
		msg := "GetURL Error(): " + err.Error()
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(msg))
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
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(msg))
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
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(msg))
		return
	}
	msg := "http://" + h.servurl + "/" + id

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(msg))
}

// DEFAULT PAGE 400
func (h *HandlderURL) DefaultURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusBadRequest)
	response := fmt.Sprintf("Bad Request (400)\nMethod: %s\nPath: %s",
		r.Method, r.URL.Path)
	w.Write([]byte(response))
}
