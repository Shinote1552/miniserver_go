package getdefault

import (
	"fmt"
	"net/http"
	"urlshortener/internal/httputils"
)

func HandlerGetDefault() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		select {
		case <-ctx.Done():
			return
		default:
			w.Header().Set("Content-Type", httputils.MIMETextPlain)
			w.WriteHeader(http.StatusBadRequest)
			response := fmt.Sprintf("Bad Request (400)\nMethod: %s\nPath: %s",
				r.Method, r.URL.Path)
			w.Write([]byte(response))
		}
	}
}
