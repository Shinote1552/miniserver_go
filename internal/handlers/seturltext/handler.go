package seturltext

import (
	"io"
	"net/http"
	"urlshortener/internal/deps"
	"urlshortener/internal/httputils"
)

type SetURLTextHandler struct {
	service deps.ServiceURLShortener
	baseURL string
}

func New(service deps.ServiceURLShortener, baseURL string) *SetURLTextHandler {
	return &SetURLTextHandler{
		service: service,
		baseURL: baseURL,
	}
}

func (h *SetURLTextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeTextPlainError(w, http.StatusBadRequest, "SetURL Error(): "+err.Error())
		return
	}
	defer r.Body.Close()

	url := string(body)
	if url == "" {
		writeTextPlainError(w, http.StatusBadRequest, "empty request body")
		return
	}

	id, err := h.service.SetURL(url)
	if err != nil {
		writeTextPlainError(w, http.StatusBadRequest, "SetURL Error(): "+err.Error())
		return
	}

	w.Header().Set("Content-Type", httputils.ContentTypePlain)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(h.buildShortURL(id)))
}

func (h *SetURLTextHandler) buildShortURL(id string) string {
	// return "http://" + h.baseURL + "/" + id
	return h.baseURL + "/" + id
}

func writeTextPlainError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", httputils.ContentTypePlain)
	w.WriteHeader(status)
	w.Write([]byte(message))
}
