# ==============================================================================
# URL Shortener Service - Main Makefile (Postgres @8080)
# ==============================================================================

.PHONY: help setup build test cover clean run dev logs logs-tail status lines

DOCKER_COMPOSE = docker compose
GO_TEST_FLAGS = -coverprofile=coverage.out ./...

# ==============================================================================
# MAIN HELP
# ==============================================================================

## help: Show main help message
help:
	@echo "URL Shortener Service - Main Commands (Postgres @8080):"
	@echo ""
	@echo "Development:"
	@echo "  make setup          Setup project (make scripts executable)"
	@echo "  make build          Build Docker images"
	@echo "  make test           Run Go tests"
	@echo "  make cover          Show test coverage"
	@echo ""
	@echo "Docker Services:"
	@echo "  make run            Start Postgres + main app in background"
	@echo "  make dev            Start Postgres + main app in foreground"
	@echo ""
	@echo "Testing:"
	@echo "  make test-curl      Full API test"
	@echo "  make test-ping      Quick ping test"
	@echo ""
	@echo "Logging & Monitoring:"
	@echo "  make logs           Show recent logs"
	@echo "  make logs-tail      Follow logs in real-time"
	@echo "  make status         Show container status"
	@echo ""
	@echo "Cleanup:"
	@echo "  make clean          Stop containers and clean temp files"
	@echo "  make clean-all      Full cleanup with volumes"
	@echo ""
	@echo "Utils:"
	@echo "  make lines          Count lines of code"
	@echo ""
	@echo "For test environments (InMemory @8081, Fallback @8082):"
	@echo "  make -f debug/Makefile.test help"

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

## test-curl: Full API test
test-curl:
	@./scripts/test-api.sh http://localhost:8080

## test-ping: Quick ping test
test-ping:
	@./scripts/test-ping.sh http://localhost:8080

# ==============================================================================
# DOCKER SERVICES
# ==============================================================================

## run: Start all services in background
run:
	$(DOCKER_COMPOSE) up -d

## dev: Start in foreground (development)
dev:
	$(DOCKER_COMPOSE) up urlshortener-db urlshortener-service

# ==============================================================================
# LOGGING & MONITORING
# ==============================================================================

## logs: Show recent logs
logs:
	$(DOCKER_COMPOSE) logs --tail=50 urlshortener-service

## logs-tail: Follow logs in real-time
logs-tail:
	$(DOCKER_COMPOSE) logs -f urlshortener-service

## status: Show container status
status:
	@echo "=== Container Status ==="
	@docker ps --filter "name=urlshortener" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# ==============================================================================
# CLEANUP
# ==============================================================================

## clean: Stop containers and clean temp files
clean:
	$(DOCKER_COMPOSE) down
	@rm -rf coverage.out tmp short_url.txt request cookies.txt
	@docker rm -f test-app 2>/dev/null || true

## clean-all: Full cleanup with volumes
clean-all: clean
	$(DOCKER_COMPOSE) down --volumes --rmi local

# ==============================================================================
# UTILS
# ==============================================================================

## lines: Count lines of code
lines:
	@echo "Lines of code: "
	@find ./ -name '*.go' -type f -exec cat {} + | wc -l

.DEFAULT_GOAL := help