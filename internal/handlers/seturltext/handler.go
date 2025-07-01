package seturltext

import (
	"io"
	"net/http"
	"urlshortener/internal/httputils"
)

type ServiceURLShortener interface {
	GetURL(token string) (string, error)
	SetURL(url string) (string, error)
}

func HandlerSetURLText(svc ServiceURLShortener, urlroot string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		id, err := svc.SetURL(url)
		if err != nil {
			writeTextPlainError(w, http.StatusBadRequest, "SetURL Error(): "+err.Error())
			return
		}

		w.Header().Set("Content-Type", httputils.ContentTypePlain)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(buildShortURL(urlroot, id)))
	}
}

// EXAMPLE: http://localhost:8080/bzwVcXmW
func buildShortURL(urlroot string, id string) string {
	return "http://" + urlroot + "/" + id
}

func writeTextPlainError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", httputils.ContentTypePlain)
	w.WriteHeader(status)
	w.Write([]byte(message))
}
