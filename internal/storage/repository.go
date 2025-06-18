package storage

type InMemoryStorage interface {
	Set(url string) (string, error)
	Get(token string) (string, error)
}
