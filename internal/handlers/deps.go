package handlers

//go:generate mockgen -destination=mocks/url_shortener_mock.go -package=mocks urlshortener/internal/handlers URLshortener
type URLshortener interface {
	GetURL(token string) (string, error)
	SetURL(url string) (string, error)
}
