package getdefault

import (
	"fmt"
	"net/http"
	"urlshortener/internal/httputils"
)

// HandlerGetDefault возвращает HTTP хендлер для обработки запросов к корневому пути
func HandlerGetDefault() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", httputils.MIMETextPlain)
		w.WriteHeader(http.StatusBadRequest)
		response := fmt.Sprintf("Bad Request (400)\nMethod: %s\nPath: %s",
			r.Method, r.URL.Path)
		w.Write([]byte(response))
	}
}
