# ==============================================================================
# URL Shortener Service Makefile
# ==============================================================================

.PHONY: help setup build test cover clean run dev logs logs-tail status

# ==============================================================================
# VARIABLES
# ==============================================================================

DOCKER_COMPOSE = docker compose
GO_TEST_FLAGS = -coverprofile=coverage.out ./...

# ==============================================================================
# MAIN COMMANDS
# ==============================================================================

## help: Show this help message
help:
	@echo "URL Shortener Service - Available commands:"
	@echo ""
	@echo "Development:"
	@echo "  make setup          Setup project (make scripts executable)"
	@echo "  make build          Build Docker images"
	@echo "  make test           Run Go tests"
	@echo "  make cover          Show test coverage"
	@echo ""
	@echo "Docker Services (Production-like):"
	@echo "  make run            Start all services in background"
	@echo "  make run-inmemory   Start in-memory version in background"
	@echo "  make run-fallback   Start fallback version in background"
	@echo ""
	@echo "Development Shortcuts (Foreground):"
	@echo "  make dev-postgres   Start Postgres + main app (foreground)"
	@echo "  make dev-inmemory   Start in-memory version (foreground)"
	@echo "  make dev-fallback   Start fallback version (foreground)"
	@echo ""
	@echo "Testing:"
	@echo "  make test-curl          Full API test (Postgres @8080)"
	@echo "  make test-curl-inmemory Full API test (InMemory @8081)"
	@echo "  make test-curl-fallback Full API test (Fallback @8082)"
	@echo "  make test-ping          Quick ping test (Postgres)"
	@echo "  make test-all           Test all versions sequentially"
	@echo ""
	@echo "Logging:"
	@echo "  make logs           Show recent logs (one-time)"
	@echo "  make logs-tail      Follow logs in real-time (interactive)"
	@echo "  make logs-inmemory  Show recent in-memory logs"
	@echo "  make logs-fallback  Show recent fallback logs"
	@echo ""
	@echo "Monitoring:"
	@echo "  make status         Show container status"
	@echo ""
	@echo "Cleanup:"
	@echo "  make clean          Stop containers and clean temp files"
	@echo "  make clean-all      Full cleanup (including volumes)"
	@echo "  make lines          Count lines of code"

# ==============================================================================
# SETUP & BUILD
# ==============================================================================

## setup: Make scripts executable
setup:
	chmod +x scripts/*.sh

## build: Build Docker images
build:
	$(DOCKER_COMPOSE) build --pull

# ==============================================================================
# TESTING
# ==============================================================================

## test: Run Go tests
test:
	@go test $(GO_TEST_FLAGS)

## cover: Show test coverage
cover: test
	@go tool cover -func=coverage.out

## test-curl: Full API test (Postgres @8080)
test-curl:
	@./scripts/test-api.sh http://localhost:8080

## test-curl-inmemory: Full API test (InMemory @8081)
test-curl-inmemory:
	@./scripts/test-api.sh http://localhost:8081

## test-curl-fallback: Full API test (Fallback @8082)
test-curl-fallback:
	@./scripts/test-api.sh http://localhost:8082

## test-ping: Quick ping test (Postgres)
test-ping:
	@./scripts/test-ping.sh http://localhost:8080

## test-ping-inmemory: Quick ping test (InMemory)
test-ping-inmemory:
	@./scripts/test-ping.sh http://localhost:8081

## test-ping-fallback: Quick ping test (Fallback)
test-ping-fallback:
	@./scripts/test-ping.sh http://localhost:8082

## test-all: Test all versions
test-all:
	@echo "=== Testing Postgres version (8080) ==="
	@./scripts/test-api.sh http://localhost:8080
	@echo ""
	@echo "=== Testing InMemory version (8081) ==="
	@./scripts/test-api.sh http://localhost:8081
	@echo ""
	@echo "=== Testing Fallback version (8082) ==="
	@./scripts/test-api.sh http://localhost:8082

# ==============================================================================
# DOCKER SERVICES (BACKGROUND)
# ==============================================================================

## run: Start all services in background
run:
	$(DOCKER_COMPOSE) up -d

## run-inmemory: Start in-memory version in background
run-inmemory:
	$(DOCKER_COMPOSE) --profile inmemory up -d urlshortener-inmemory

## run-fallback: Start fallback test version in background
run-fallback:
	$(DOCKER_COMPOSE) --profile fallback up -d urlshortener-fallback

# ==============================================================================
# DEVELOPMENT SHORTCUTS (FOREGROUND)
# ==============================================================================

## dev-postgres: Start Postgres + main app in foreground
dev-postgres:
	$(DOCKER_COMPOSE) up urlshortener-db urlshortener-service

## dev-inmemory: Start in-memory version in foreground
dev-inmemory:
	$(DOCKER_COMPOSE) --profile inmemory up urlshortener-inmemory

## dev-fallback: Start fallback version in foreground
dev-fallback:
	$(DOCKER_COMPOSE) --profile fallback up urlshortener-fallback

# ==============================================================================
# LOGGING
# ==============================================================================

## logs: Show recent logs (one-time, non-interactive)
logs:
	$(DOCKER_COMPOSE) logs --tail=50 urlshortener-service

## logs-tail: Follow logs in real-time (interactive)
logs-tail:
	$(DOCKER_COMPOSE) logs -f urlshortener-service

## logs-inmemory: Show recent in-memory logs
logs-inmemory:
	$(DOCKER_COMPOSE) --profile inmemory logs --tail=50 urlshortener-inmemory

## logs-fallback: Show recent fallback logs
logs-fallback:
	$(DOCKER_COMPOSE) --profile fallback logs --tail=50 urlshortener-fallback

# ==============================================================================
# MONITORING
# ==============================================================================

## status: Show container status
status:
	@echo "=== Container Status ==="
	@docker ps --filter "name=urlshortener" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

## lines: Count lines of code
lines:
	@echo "Lines of code: "
	@find ./ -name '*.go' -type f -exec cat {} + | wc -l

# ==============================================================================
# CLEANUP
# ==============================================================================

## clean: Stop containers and clean temp files
clean:
	$(DOCKER_COMPOSE) down
	$(DOCKER_COMPOSE) --profile inmemory down
	$(DOCKER_COMPOSE) --profile fallback down
	@rm -rf coverage.out tmp short_url.txt request cookies.txt
	@docker rm -f test-app urlshortener-inmemory urlshortener-fallback 2>/dev/null || true

## clean-all: Full cleanup with volumes
clean-all:
	$(DOCKER_COMPOSE) down --volumes --rmi local
	$(DOCKER_COMPOSE) --profile inmemory down --volumes --rmi local
	$(DOCKER_COMPOSE) --profile fallback down --volumes --rmi local

# ==============================================================================
# DEFAULT TARGET
# ==============================================================================

.DEFAULT_GOAL := help