package handlers

import (
	"io"
	"net/http"
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

	if r.Header.Get("Content-Type") != "text/plain" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Читаем тело запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		// w.WriteHeader(http.StatusBadRequest)
		// return
	}
	originalURL := string(body)
	w.Write(body)
	w.Write([]byte("\n\n\n"))

	w.Write([]byte(originalURL))

	// url, err := h.service.GetURL()
	// if err != nil {
	// 	w.WriteHeader(http.StatusBadRequest)
	// }

	w.Write([]byte("\n--GetURL 307 SUCCESS--\n"))
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// POST 201
func (h *HandlderURL) SetURL(w http.ResponseWriter, r *http.Request) {
	url := r.URL.String()
	// id, err := h.service.SetURL(url)
	// if err != nil {
	// 	w.WriteHeader(http.StatusBadRequest)
	// }
	// w.Write([]byte("URL/id"))

	w.Write([]byte("url: " + url))
	w.Write([]byte("--SetURL 201--"))
	w.WriteHeader(http.StatusCreated)
}

// DEFAULT PAGE 400
func (h *HandlderURL) DefaultURL(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("--DefaultURL 400--"))
	w.WriteHeader(http.StatusBadRequest)
}
