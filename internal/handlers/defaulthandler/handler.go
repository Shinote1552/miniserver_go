package defaulthandler

import (
	"fmt"
	"net/http"
	"urlshortener/internal/httputils"
)

type DefaultHandler struct{}

func New() *DefaultHandler {
	return &DefaultHandler{}
}

func (h *DefaultHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", httputils.ContentTypePlain)
	w.WriteHeader(http.StatusBadRequest)
	response := fmt.Sprintf("Bad Request (400)\nMethod: %s\nPath: %s",
		r.Method, r.URL.Path)
	w.Write([]byte(response))
}
