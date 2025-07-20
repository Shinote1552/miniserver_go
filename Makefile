.PHONY: postgres-up postgres-new postgres-down run

# Переменные
POSTGRES_IMAGE := postgres:bookworm
POSTGRES_CONTAINER := urlshortener-db
POSTGRES_DB := gpx_test
POSTGRES_USER := postgres
POSTGRES_PASSWORD := admin
POSTGRES_PORT := 5432
VOLUME_NAME := pg-data-urlshortener

# Поднять существующий контейнер PostgreSQL
postgres-up:
	@echo "Starting existing PostgreSQL container..."
	@docker start $(POSTGRES_CONTAINER) || (echo "Container not found, use 'make postgres-new' to create new one"; exit 1)
	@echo "PostgreSQL is running on port $(POSTGRES_PORT)"

# Создать новый контейнер PostgreSQL
postgres-new:
	@echo "Creating new PostgreSQL container..."
	@docker run -d \
		--name $(POSTGRES_CONTAINER) \
		-e POSTGRES_USER=$(POSTGRES_USER) \
		-e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
		-e POSTGRES_DB=$(POSTGRES_DB) \
		-v $(VOLUME_NAME):/var/lib/postgresql/data \
		-p $(POSTGRES_PORT):$(POSTGRES_PORT) \
		$(POSTGRES_IMAGE)
	@echo "New PostgreSQL container created and running on port $(POSTGRES_PORT)"
	@sleep 2
	@docker exec $(POSTGRES_CONTAINER) psql -U $(POSTGRES_USER) -c "CREATE DATABASE $(POSTGRES_DB);" || echo "Database already exists"

# Остановить PostgreSQL
postgres-down:
	@echo "Stopping PostgreSQL container..."
	@docker stop $(POSTGRES_CONTAINER) || true
	@echo "PostgreSQL container stopped"

# Запустить сервер
run:
	@echo "Starting server..."
	@DATABASE_DSN="postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable" \
	go run urlshortener/cmd/server

# Комбинация: поднять PostgreSQL и запустить сервер
up: postgres-new run

# Полная очистка
clean: postgres-down
	@echo "Removing volume..."
	@docker volume rm $(VOLUME_NAME) || true
	@echo "Cleanup complete"