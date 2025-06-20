package deps

import (
	"net/http"

	"github.com/rs/zerolog"
)

type Logger interface {
	Info() *zerolog.Event
	Error() *zerolog.Event
}

type InMemoryStorage interface {
	Set(url string) (string, error)
	Get(token string) (string, error)
}

//go:generate mockgen -destination=mocks/url_shortener_mock.go -package=mocks urlshortener/internal/deps URLshortener
type URLshortener interface {
	GetURL(token string) (string, error)
	SetURL(url string) (string, error)
}

type Handler interface {
	GetURL(w http.ResponseWriter, r *http.Request)
	SetURL(w http.ResponseWriter, r *http.Request)
	DefaultURL(w http.ResponseWriter, r *http.Request)
}

type Middleware interface {
	Handler(next http.Handler) http.Handler
}
