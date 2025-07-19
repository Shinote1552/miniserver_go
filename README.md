.PHONY: run build test docker-up docker-down

# Переменные
APP_NAME := urlshortener
POSTGRES_IMAGE := postgres:bookworm
POSTGRES_CONTAINER := postgres
POSTGRES_DB := gpx_test
POSTGRES_USER := postgres
POSTGRES_PASSWORD := admin
POSTGRES_PORT := 5432

# Запуск сервера в development режиме
run:
	go run cmd/server/main.go

# Сборка сервера
build:
	go build -o bin/$(APP_NAME) cmd/server/main.go

# Запуск тестов
test:
	go test -v ./...

# Запуск PostgreSQL в Docker
docker-up:
	docker run \
		--name $(POSTGRES_CONTAINER) \
		--env POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
		--env POSTGRES_USER=$(POSTGRES_USER) \
		--env POSTGRES_DB=$(POSTGRES_DB) \
		--volume pg-data:/var/lib/postgresql/data \
		--publish $(POSTGRES_PORT):$(POSTGRES_PORT) \
		-d \
		$(POSTGRES_IMAGE)

# Остановка PostgreSQL
docker-down:
	docker stop $(POSTGRES_CONTAINER)
	docker rm $(POSTGRES_CONTAINER)

# Запуск сервера с PostgreSQL
run-with-db: docker-up
	DATABASE_DSN="postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable" \
	go run cmd/server/main.go

# Очистка
clean:
	rm -rf bin/
	docker volume rm pg-data