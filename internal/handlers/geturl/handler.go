package geturl

import (
	"net/http"
	"strings"
	"urlshortener/internal/httputils"
)

type ServiceURLShortener interface {
	GetURL(token string) (string, error)
	SetURL(url string) (string, error)
}

func writeTextPlainError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", httputils.MIMETextPlain)
	w.WriteHeader(status)
	w.Write([]byte(message))
}

func HandlerGetURLWithID(svc ServiceURLShortener) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/")
		url, err := svc.GetURL(id)
		if err != nil {
			writeTextPlainError(w, http.StatusBadRequest, "GetURL Error(): "+err.Error())
			return
		}

		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}
