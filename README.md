# General System configurating
## Структура проекта

- **Dependencies/versions**: Все зависимости проекта и их версии указаны в файле `go.mod`
###
- **Domain Models**: `domain/models/` - бизнес-сущности
- **Services**: `domain/services/` - бизнес-логика (use cases)
- **Handlers**: `internal/http/handlers/` - HTTP контроллеры
- **DTOs**: 
  - `internal/http/dto/` - транспортные модели
  - `internal/repository/dto/` - модели репозитория
- **Repository** - слой работы с данными

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
| `JWT_ACCESS_EXPIRE`    | `"24h"`, `"168h"`, `"720h"`     | `24 * time.Hour`                 |

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

### Значения по умолчанию
| Параметр             | Значение по умолчанию                          |
|----------------------|-----------------------------------------------|
| ServerAddress        | `"localhost:8080"`                           |
| BaseURL              | `"http://localhost:8080"`                    |
| FileStoragePath      | `"tmp/short-url-db.json"`                    |
| DatabaseDSN          | `"postgres://postgres:admin@localhost:5432/gpx_test?sslmode=disable"` |
| JWTAccessExpire      | `15m` (15 минут)                             |
| JWTSecretKey         | Автогенерация (32 байта base64)              |

## Доступные флаги командной строки

| Флаг                  | Описание                          | Пример значения                  |
|-----------------------|-----------------------------------|----------------------------------|
| `-server-address`     | Адрес и порт сервера              | `-server-address=":8080"`        |
| `-base-url`           | Базовый URL для ссылок            | `-base-url="https://example.com"`|
| `-file-storage-path`  | Путь к файлу хранилища URL        | `-file-storage-path="/data/urls.json"`|
| `-database-dsn`       | DSN для подключения к PostgreSQL  | `-database-dsn="postgres://user:pass@localhost:5432/db"`|
| `-jwt-access-expire`  | Время жизни Access токена         | `-jwt-access-expire="15m"`       |

## Команда запуска со всеми кастомными параметрами

```bash
go run urlshortener/cmd/server \
  -server-address="0.0.0.0:80" \
  -database-dsn="postgres://prod_user:pass@db:5432/production" \
  -jwt-access-expire="30m" 

```

# Полный набор переменных окружения

| Переменная              | Описание                          | Пример значения                          |
|-------------------------|-----------------------------------|------------------------------------------|
| `SERVER_ADDRESS`        | Адрес сервера                    | `:8080` или `0.0.0.0:9000`              |
| `BASE_URL`              | Базовый URL для ссылок           | `https://short.example.com`              |
| `FILE_STORAGE_PATH`     | Путь к файлу хранилища           | `/var/data/urls.json`                    |
| `DATABASE_DSN`          | PostgreSQL DSN строка            | `postgres://user:pass@host:5432/dbname`  |
| `JWT_SECRET_KEY`        | Секретный ключ для JWT           | `base64-encoded-32-byte-key`             |
| `JWT_ACCESS_EXPIRE`     | Время жизни Access токена        | `15m` (15 минут)                         |

## Пример .env файла

```ini
SERVER_ADDRESS=:8080
BASE_URL=https://short.example.com
FILE_STORAGE_PATH=/data/urls.json
DATABASE_DSN=postgres://app_user:password@db-server:5432/shortener_prod
JWT_SECRET_KEY=uV8q7z$A%D*G-KaPdSgVkYp3s6v9y/BE
JWT_ACCESS_EXPIRE=30m
```

## Основные возможности

### Особенности
Автоматическая JWT-аутентификация через куки
Поддержка PostgreSQL и файлового хранилища Inmmeory
Проверка дубликатов URL, shortCode (409 Conflict)
Сжатие ответов (gzip) по хэдеру "Accept-Encoding:gzip" и логирование запросов

## Публичные обработчики

### `GET /ping`
- Проверяет соединение с базой данных
- Возвращает 200 OK если БД доступна, 500 если нет
- Используется для мониторинга работоспособности

### `GET /{id}`
- Перенаправляет на оригинальный URL по короткому идентификатору
- Возвращает HTTP 307 (Temporary Redirect)
- Если URL не найден - возвращает 400 Bad Request

### `GET /`
- Дефолтный обработчик для корневого пути
- Всегда возвращает 400 Bad Request
- Используется для обработки некорректных запросов

## Защищенные обработчики (требуют JWT, если нету выдает)

### `POST /api/shorten`
- Создает короткую версию URL из JSON-запроса
- Возвращает 201 Created с коротким URL
- При дубликате URL возвращает 409 Conflict
- Валидирует входные данные

### `POST /api/shorten/batch` 
- Пакетное создание коротких URL
- Принимает массив URL с correlation_id
- Возвращает массив результатов с сохранением correlation_id
- Обрабатывает до 10 URL за один запрос

### `POST /`
- Альтернативная версия сокращения URL
- Принимает URL в теле запроса (text/plain)
- Возвращает короткий URL как текст
- Аналогичная логика обработки дубликатов

### `GET /api/user/urls`
- Возвращает все URL текущего пользователя
- Формат: массив объектов {short_url, original_url}
- Пустой результат - 204 No Content
- Требует валидной JWT куки
### Примеры использования

```bash
# Проверка сервиса
curl http://localhost:8080/ping

# Сокращение URL, в ответ возвращается JWT токен, нужно в каждый POST запрос прикладывать свой JWT токен для идентификации, иначе при `GET /api/user/urls` мы не сможем получить все данные
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com"}' \
  -d "auth_token=USER_JWT"


curl -X POST http://localhost:8080/api/shorten/batch \
  -H "Content-Type: application/json" \
  -d '[{"correlation_id": "1", "original_url": "https://ya.ru"}]'
  -d "auth_token=USER_JWT"

# Получение своих URL по токену который выдан уникальному пользователю/JWT
curl http://localhost:8080/api/user/urls \
  -H "Cookie: auth_token=your_token" \
  -d "auth_token=USER_JWT"

```
