/*

	// fixme: здесь должны быть внешние зависимости относительно всего приложение.
	Пока мне лень все инетерфейсы перераспределять, поэтому буду держать здесь.
	Разделил пока что только ендпоинты


*/

package deps

import (
	"net/http"

	"github.com/rs/zerolog"
)

// Слои для логирования(пока ограничился двумя)(хз насколько праивльно так делать)
type Logger interface {
	Info() *zerolog.Event
	Error() *zerolog.Event
}

// Интерфейс для хранилища
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

// Интерфейс для middleware(для нестандартного роутера gorila/mux)
type Middleware interface {
	Handler(next http.Handler) http.Handler
}
