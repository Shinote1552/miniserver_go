.PHONY: db-up db-new db-down run up clean

# PostgreSQL configuration
POSTGRES_IMAGE := postgres:bookworm
POSTGRES_CONTAINER := urlshortener-db
POSTGRES_DB := gpx_test
POSTGRES_USER := postgres
POSTGRES_PASSWORD := admin
POSTGRES_PORT := 5432

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




# EXPERIMENTAL!!!
SERVER_ADDRESS := localhost:8080
COOKIE_FILE := /tmp/curl_cookie.txt
SHORT_URL_FILE := /tmp/short_url.txt

.PHONY: test_curl
test_curl:
	echo "=== Starting curl tests ==="
	
	# 1. Получаем JWT токен и сохраняем cookie
	rm -f $(COOKIE_FILE) $(SHORT_URL_FILE)
	curl -v -X POST http://$(SERVER_ADDRESS)/ -c $(COOKIE_FILE)
	
	# 2. Тестируем публичные endpoint'ы
	echo "=== Testing public endpoints ==="
	echo "GET /ping"
	curl -v -X GET http://$(SERVER_ADDRESS)/ping
	echo ""
	
	echo "GET / (default handler)"
	curl -v -X GET http://$(SERVER_ADDRESS)/
	echo ""
	
	# 3. Тестируем защищённые endpoint'ы
	echo "=== Testing protected endpoints ==="
	
	# 3.1. Создаём URL через text/plain
	echo "POST / (text/plain)"
	curl -v -X POST \
		-H "Content-Type: text/plain" \
		-b $(COOKIE_FILE) \
		-d "https://google.com" \
		http://$(SERVER_ADDRESS)/ \
		| tee $(SHORT_URL_FILE)
	echo ""
	
	# 3.2. Создаём URL через application/json
	echo "POST /api/shorten (application/json)"
	curl -v -X POST \
		-H "Content-Type: application/json" \
		-b $(COOKIE_FILE) \
		-d '{"url":"https://yandex.ru"}' \
		http://$(SERVER_ADDRESS)/api/shorten
	echo ""
	
	# 3.3. Пакетное создание URL
	echo "POST /api/shorten/batch (batch create)"
	curl -v -X POST \
		-H "Content-Type: application/json" \
		-b $(COOKIE_FILE) \
		-d '[{"correlation_id": "1", "original_url": "https://google.com"}, {"correlation_id": "2", "original_url": "https://youtube.com"}]' \
		http://$(SERVER_ADDRESS)/api/shorten/batch
	echo ""
	
	# 3.4. Получаем список URL пользователя
	echo "GET /api/user/urls"
	curl -v -X GET \
		-b $(COOKIE_FILE) \
		http://$(SERVER_ADDRESS)/api/user/urls
	echo ""
	
	# 4. Тестируем редирект
	echo "=== Testing redirect ==="
	echo "Testing redirect for: $$(cat $(SHORT_URL_FILE))"
	curl -v -X GET $$(cat $(SHORT_URL_FILE))
	echo ""
	
	# Очищаем временные файлы
	rm -f $(COOKIE_FILE) $(SHORT_URL_FILE)
	
	echo "=== All tests completed ==="
# EXPERIMENTAL!!!













	















	










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
	@rm -rf tmp $() $() $() $()
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

#make test_curl 

