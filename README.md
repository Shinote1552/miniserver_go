# Текущая структура проекта:
- Domain Models (domain/models/) - место для бизнес-сущностей.
- Services (domain/services/) - содержит бизнес-логику (это и есть use cases).
- Handlers (internal/http/handlers/) - это контроллеры.
- DTOs (internal/http/dto/) (internal/repository/dto/) - разделение на транспортных модели.
- Repository - слой для работы с данными.


# System configurating
## Формат указания времени в конфигурации
Важно: Все значения должны быть строкой в кавычках при использовании в env-файлах или командной строке.
Используется формат, совместимый с `time.ParseDuration` из стандартного пакета Go:
### Основные единицы измерения:
- `ns` - наносекунды
- `us` (или `µs`) - микросекунды
- `ms` - миллисекунды
- `s` - секунды
- `m` - минуты
- `h` - часы

### Примеры допустимых значений:

| Переменная              | Примеры корректных значений      | Эквивалент в time.Duration      |
|-------------------------|----------------------------------|----------------------------------|
| `JWT_ACCESS_EXPIRE`     | `"15m"`, `"1h30m"`, `"90m"`     | `15 * time.Minute`               |
| `JWT_REFRESH_EXPIRE`    | `"24h"`, `"168h"`, `"720h"`     | `24 * time.Hour`                 |

### Особенности:
1. Можно комбинировать единицы:
   ```bash
   export JWT_ACCESS_EXPIRE="1h30m"  # 1 час 30 минут
   export JWT_ACCESS_EXPIRE="1.5h"  # дробные 1.5 часа = 90 минут
   export JWT_ACCESS_EXPIRE="500ms"  # минимальные значения 500 миллисекунд

## Приоритеты настроек (от высшего к низшему)
1. Флаги командной строки (`-a`, `-d` и др.)
2. Переменные окружения (`SERVER_ADDRESS`, `DATABASE_DSN`)
3. Значения по умолчанию (хардкод в `config` пакете)

## Значения по умолчанию

| Параметр             | Значение по умолчанию                          |
|----------------------|-----------------------------------------------|
| ServerAddress        | `"localhost:8080"`                            |
| BaseURL              | `"http://localhost:8080"`                     |
| FileStoragePath      | `"tmp/short-url-db.json"`                     |
| DatabaseDSN          | `"postgres://postgres:admin@localhost:5432/gpx_test?sslmode=disable"` |
| JWTAccessExpire      | `15 * time.Minute` (15 минут)                 |
| JWTRefreshExpire     | `168 * time.Hour` (7 дней)                    |
| JWTSecretKey         | Автогенерация при запуске (32 байта в base64) |

## Доступные флаги

| Флаг               | Описание                          | Пример                   |
|--------------------|-----------------------------------|--------------------------|
| `-a`               | Адрес сервера                     | `-a=":8080"`             |
| `-b`               | Базовый URL                       | `-b="https://example.com"`|
| `-f`               | Путь к файловому хранилищу        | `-f="/data/storage.json"`|
| `-d`               | DSN для подключения к БД          | `-d="postgres://user:pass@host/db"`|
| `-jwt-access-expire`| Время жизни access-токена JWT    | `-jwt-access-expire=30m` |
| `-jwt-refresh-expire`| Время жизни refresh-токена JWT  | `-jwt-refresh-expire=168h`|


## Команда запуска со всеми кастомными параметрами

```bash
go run urlshortener/cmd/server \
  -a="0.0.0.0:9000" \
  -b="https://my-shortener.com" \
  -f="/opt/app/data/urls.json" \
  -d="postgres://postgres:admin@localhost:5432/url_shortener" \
  -jwt-access-expire="20m" \
  -jwt-refresh-expire="240h" \
  -jwt-secret-key="my-super-secret-key-must-be-32-bytes!"
```

# Environment Variables Configuration

# Полный набор переменных окружения

| Переменная              | Описание                          | Пример значения                          |
|-------------------------|-----------------------------------|------------------------------------------|
| `SERVER_ADDRESS`        | Адрес сервера                    | `:8080` или `0.0.0.0:9000`              |
| `BASE_URL`              | Базовый URL для ссылок           | `https://short.example.com`              |
| `FILE_STORAGE_PATH`     | Путь к файлу хранилища           | `/var/data/urls.json`                    |
| `DATABASE_DSN`          | PostgreSQL DSN строка            | `postgres://user:pass@host:5432/dbname`  |
| `JWT_SECRET_KEY`        | Секретный ключ для JWT           | `base64-encoded-32-byte-key`             |
| `JWT_ACCESS_EXPIRE`     | Время жизни Access токена        | `15m` (15 минут)                         |
| `JWT_REFRESH_EXPIRE`    | Время жизни Refresh токена       | `168h` (7 дней)                          |

## Пример .env файла

```ini
SERVER_ADDRESS=:8080
BASE_URL=https://short.example.com
FILE_STORAGE_PATH=/data/urls.json
DATABASE_DSN=postgres://app_user:password@db-server:5432/shortener_prod
JWT_SECRET_KEY=uV8q7z$A%D*G-KaPdSgVkYp3s6v9y/BE
JWT_ACCESS_EXPIRE=30m
JWT_REFRESH_EXPIRE=720h
```