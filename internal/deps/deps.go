// fixme: здесь должны быть внешние зависимости относительно всего приложение

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
type URLshortener interface { // fixme: 2 слово с большой буквы
	// fixme: Необязательно, но я бы добавил, что это сервис
	// fixme: Либо Вообще переименовал URLshortener -> Service
	GetURL(token string) (string, error) // fixme: best pactice: Заранне определи нотацию в которой будешь именовать CRUD'ы: get/fetch, update/set, delete/destroy
	SetURL(url string) (string, error)
}

type Handler interface {
	SetURLwithJSON(w http.ResponseWriter, r *http.Request) // fixme: Отра
	GetURL(w http.ResponseWriter, r *http.Request)
	SetURL(w http.ResponseWriter, r *http.Request)
	DefaultURL(w http.ResponseWriter, r *http.Request)
}

type Middleware interface {
	Handler(next http.Handler) http.Handler
}
