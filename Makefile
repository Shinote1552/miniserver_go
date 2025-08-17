.PHONY: db-up db-new db-down run up clean

# PostgreSQL configuration
POSTGRES_IMAGE := postgres:bookworm
POSTGRES_CONTAINER := urlshortener-db
POSTGRES_DB := gpx_test
POSTGRES_USER := postgres
POSTGRES_PASSWORD := admin
POSTGRES_PORT := 5432


# go generate (если настроены //go:generate комментарии)
mockgen:
	@echo "Generating using go generate..."
	@go generate ./...
	@echo "Generation complete"


# Start existing container
db-up:
	@echo "Starting existing PostgreSQL container..."
	@docker start $(POSTGRES_CONTAINER) || (echo "Container not found. Use 'make db-new'"; exit 1)
	@echo "PostgreSQL running on port $(POSTGRES_PORT)"

# Create new container (with automatic removal of old one)
db-new:
	@echo "Creating new PostgreSQL container..."
	@docker rm -f $(POSTGRES_CONTAINER) >/dev/null 2>&1 || true
	@docker run -d \
		--name $(POSTGRES_CONTAINER) \
		-e POSTGRES_USER=$(POSTGRES_USER) \
		-e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
		-e POSTGRES_DB=$(POSTGRES_DB) \
		-p $(POSTGRES_PORT):$(POSTGRES_PORT) \
		$(POSTGRES_IMAGE)
	@echo "New container created and running on port $(POSTGRES_PORT)"
	@sleep 2  # Allow time for initialization

# Stop container
db-down:
	@echo "Stopping PostgreSQL container..."
	@docker stop $(POSTGRES_CONTAINER) >/dev/null 2>&1 || true
	@echo "Container stopped"

# Start server
run:
	@echo "Starting server..."
	@DATABASE_DSN="postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable" \
	go run cmd/server/main.go

# Combined command: DB + server
up: db-new run

# Cleanup
clean: db-down
	@echo "Removing container..."
	@docker rm -f $(POSTGRES_CONTAINER) >/dev/null 2>&1 || true
	@rm -rf tmp
	@echo "Cleanup complete"

lines:
	@echo "Summary code lines in this project: "
	@find ./ -type f -exec cat {} + | wc -l

# in server psql -h localhost -p 5432 -U postgres -d gpx_test

# Usage examples:
# make db-new  # Create new container (old one will be removed)
# make db-up   # Start existing container
# make run     # Start server only
# make up      # Full startup (DB + server)
# make clean   # Stop and remove container




# FAST TEST
# # Сохраняем куку при первом запросе
# curl -v -c cookies.txt -X POST http://localhost:8080/api/shorten -d '{"url":"https://example.com"}'

# # Используем для всех последующих запросов
# curl -v -b cookies.txt -X POST http://localhost:8080/api/shorten -d '{"url":"https://another.com"}'
# curl -v -b cookies.txt -X POST http://localhost:8080/api/shorten/batch -d '[{"url":"https://google.com"}]'
# curl -v -b cookies.txt -X GET http://localhost:8080/api/user/urls