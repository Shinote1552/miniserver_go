package storage

type InMemoryStorage interface {
	Set(url string) (uint64, error)
	Get(id uint64) (string, error)
}
