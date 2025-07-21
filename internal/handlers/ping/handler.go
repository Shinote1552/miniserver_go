package ping

import (
	"context"
	"net/http"
	"urlshortener/internal/httputils"
)

type Service interface {
	PingDataBase(context.Context) error
}

func HandlerPing(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set(httputils.HeaderContentType, httputils.MIMETextPlain)

		if err := svc.PingDataBase(r.Context()); err != nil {
			http.Error(w, "Database unavailable", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}
