/*

	// fixme: здесь должны быть внешние зависимости относительно всего приложение.
	Пока мне лень все инетерфейсы перераспределять, поэтому буду держать здесь.
	Разделил пока что только ендпоинты


*/

package deps

// Интерфейс для хранилища
type InMemoryStorage interface {
	Set(key string, value string) error
	Get(token string) (string, error)
	GetAll() ([]string, error)
}
