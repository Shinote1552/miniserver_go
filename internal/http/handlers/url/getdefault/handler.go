package getdefault

import (
	"fmt"
	"net/http"
	"urlshortener/internal/http/httputils"
)

func HandlerGetDefault() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		select {
		case <-ctx.Done():
			return
		default:
			details := fmt.Sprintf("Method: %s\nPath: %s", r.Method, r.URL.Path)
			httputils.WriteBadRequestError(w, details)
		}
	}
}
