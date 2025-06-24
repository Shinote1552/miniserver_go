// fixme: здесь должны быть внешние зависимости относительно всего приложение

package deps

import (
	"net/http"

	"github.com/rs/zerolog"
)

type Logger interface {
	// logging Layers
	Info() *zerolog.Event
	Error() *zerolog.Event
}

type InMemoryStorage interface {
	Set(key string, value string) error
	Get(token string) (string, error)
	GetAll() ([]string, error)
}

//go:generate mockgen -destination=mocks/url_shortener_mock.go -package=mocks urlshortener/internal/deps ServiceURLShortener
type ServiceURLShortener interface {
	GetURL(token string) (string, error)
	SetURL(url string) (string, error)
}

type Middleware interface {
	Handler(next http.Handler) http.Handler
}
