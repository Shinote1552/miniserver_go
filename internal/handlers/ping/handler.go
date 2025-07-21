package ping

import (
	"net/http"
	"urlshortener/internal/httputils"
)

type Service interface {
	PingDataBase() error
}

func HandlerPing(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(httputils.HeaderContentType, httputils.MIMETextPlain)

		if err := svc.PingDataBase(); err != nil {
			http.Error(w, "Database unavailable", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}
