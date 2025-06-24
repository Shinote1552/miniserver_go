package geturl

import (
	"net/http"
	"strings"
	"urlshortener/internal/deps"
	"urlshortener/internal/httputils"
)

type GetURLHandler struct {
	service deps.ServiceURLShortener
}

func New(service deps.ServiceURLShortener) *GetURLHandler {
	return &GetURLHandler{
		service: service,
	}
}

func (h *GetURLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/")
	url, err := h.service.GetURL(id)
	if err != nil {
		writeTextPlainError(w, http.StatusBadRequest, "GetURL Error(): "+err.Error())
		return
	}

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func writeTextPlainError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", httputils.ContentTypePlain)
	w.WriteHeader(status)
	w.Write([]byte(message))
}
