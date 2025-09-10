.PHONY: run run-inmemory run-fallback test cover clean clean-all test-curl test-ping test-all

# Основные команды
run:
	docker compose up

run-inmemory:
	docker compose --profile inmemory up urlshortener-inmemory

run-fallback:
	docker compose --profile fallback up urlshortener-fallback

build:
	docker compose build --pull

test:
	@go test -coverprofile=coverage.out ./...

cover: test
	@go tool cover -func=coverage.out

clean:
	docker compose down
	docker compose --profile inmemory down
	docker compose --profile fallback down
	@rm -rf coverage.out tmp short_url.txt request cookies.txt
	docker rm -f test-app urlshortener-inmemory urlshortener-fallback 2>/dev/null || true

clean-all:
	docker compose down --volumes --rmi local
	docker compose --profile inmemory down --volumes --rmi local
	docker compose --profile fallback down --volumes --rmi local

# Тестовые команды через скрипты
test-curl:
	@./scripts/test-api.sh http://localhost:8080

test-curl-inmemory:
	@./scripts/test-api.sh http://localhost:8081

test-curl-fallback:
	@./scripts/test-api.sh http://localhost:8082

test-ping:
	@./scripts/test-ping.sh http://localhost:8080

test-ping-inmemory:
	@./scripts/test-ping.sh http://localhost:8081

test-ping-fallback:
	@./scripts/test-ping.sh http://localhost:8082

test-all:
	@echo "=== Testing Postgres version (8080) ==="
	@./scripts/test-api.sh http://localhost:8080
	@echo ""
	@echo "=== Testing InMemory version (8081) ==="
	@./scripts/test-api.sh http://localhost:8081
	@echo ""
	@echo "=== Testing Fallback version (8082) ==="
	@./scripts/test-api.sh http://localhost:8082

# Команды для разработки
dev-postgres: clean
	docker compose up urlshortener-db urlshortener-service

dev-inmemory: clean
	docker compose --profile inmemory up urlshortener-inmemory

dev-fallback: clean
	docker compose --profile fallback up urlshortener-fallback

logs:
	docker compose logs -f urlshortener-service

logs-inmemory:
	docker compose --profile inmemory logs -f urlshortener-inmemory

logs-fallback:
	docker compose --profile fallback logs -f urlshortener-fallback

# Мониторинг
status:
	@echo "=== Container Status ==="
	@docker ps --filter "name=urlshortener" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

lines:
	@echo "Summary code lines in this project: "
	@find ./ -type f -exec cat {} + | wc -l

# Создаем скрипты исполняемыми
setup-scripts:
	chmod +x scripts/*.sh

# Помощь
help:
	@echo "Available commands:"
	@echo "  make run              - Запуск с Postgres (8080)"
	@echo "  make run-inmemory     - Запуск inmemory версии (8081)"
	@echo "  make run-fallback     - Запуск fallback теста (8082)"
	@echo "  make test-curl        - Полный тест Postgres версии"
	@echo "  make test-curl-inmemory - Полный тест inmemory версии"
	@echo "  make test-all         - Тест всех версий"
	@echo "  make test-ping        - Быстрый тест ping"
	@echo "  make dev-postgres     - Только Postgres + сервис"
	@echo "  make status           - Статус контейнеров"