.PHONY: run test cover 

run:
	docker compose up

build:
	docker compose build --pull
	docker compose up --build

test:
	@go test -coverprofile=coverage.out ./...

cover: test
	@go tool cover -func=coverage.out

clean:
	docker compose down
	@rm -rf coverage.out tmp short_url.txt request cookies.txt
	docker rm -f test-app 2>/dev/null || true








clean-all:
	docker compose down --volumes --rmi local
	docker pull postgres:15.5-bookworm


run-inmemory: 
	docker rm -f test-app 2>/dev/null || true
	docker run -p 8080:8080 --name test-app \
		-e STORAGE_TYPE=memory \
		-e FILE_STORAGE_PATH=/app/tmp/short-url-db.json \
		-e DATABASE_DSN="" \
		-e SERVER_ADDRESS=:8080 \
		-e BASE_URL=http://localhost:8080 \
		urlshortener-service


# EXPERIMENTAL!!!

test_curl: test_ping
	@echo "=== Starting comprehensive curl tests ==="
	
	# 1. Получаем JWT токен и сохраняем cookie
	@rm -f cookies.txt short_url.txt
	@echo "1. Getting JWT token:"
	@curl -s -X POST http://localhost:8080/ -c cookies.txt
	@echo "Token saved to cookies.txt"
	@echo ""
	
	# 2. Тестируем публичные endpoint'ы
	@echo "2. Testing public endpoints:"
	@echo "GET /ping:"
	@curl -s -o /dev/null -w "Status: %{http_code}\n" -X GET http://localhost:8080/ping
	@echo ""
	
	@echo "GET / (default handler):"
	@curl -s -o /dev/null -w "Status: %{http_code}\n" -X GET http://localhost:8080/
	@echo ""
	
	# 3. Тестируем защищённые endpoint'ы
	@echo "3. Testing protected endpoints:"
	
	@echo "3.1. POST / (text/plain):"
	@curl -s -X POST \
		-H "Content-Type: text/plain" \
		-b cookies.txt \
		-d "https://google.com" \
		http://localhost:8080/ \
		| tee short_url.txt
	@echo ""
	
	@echo "3.2. POST /api/shorten (application/json):"
	@curl -s -X POST \
		-H "Content-Type: application/json" \
		-b cookies.txt \
		-d '{"url":"https://yandex.ru"}' \
		http://localhost:8080/api/shorten | python3 -m json.tool
	@echo ""
	
	@echo "3.3. POST /api/shorten/batch (batch create):"
	@curl -s -X POST \
		-H "Content-Type: application/json" \
		-b cookies.txt \
		-d '[{"correlation_id": "1", "original_url": "https://google.com"}, {"correlation_id": "2", "original_url": "https://youtube.com"}]' \
		http://localhost:8080/api/shorten/batch | python3 -m json.tool
	@echo ""
	
	@echo "3.4. GET /api/user/urls:"
	@curl -s -X GET \
		-b cookies.txt \
		http://localhost:8080/api/user/urls | python3 -m json.tool
	@echo ""
	
	# 4. Тестируем редирект (ИСПРАВЛЕННЫЙ ВАРИАНТ)
	@echo "4. Testing redirect:"
	@if [ -f short_url.txt ]; then \
		SHORT_URL=$$(cat short_url.txt); \
		echo "Testing redirect for: $${SHORT_URL}"; \
		SHORT_ID=$${SHORT_URL##*:8080/}; \
		if [ "$${SHORT_ID}" != "$${SHORT_URL}" ]; then \
			echo "Redirect test for ID: $${SHORT_ID}"; \
			curl -s -o /dev/null -w "Redirect: %{http_code} -> %{redirect_url}\n" -X GET "http://localhost:8080/$${SHORT_ID}"; \
		else \
			echo "Invalid short URL format: $${SHORT_URL}"; \
		fi; \
	else \
		echo "No short URL found for redirect test"; \
	fi
	@echo ""
	
	# 5. Большое пакетное создание URL
	@echo "5. Large batch URL creation:"
	@curl -s -X POST \
		-H "Content-Type: application/json" \
		-b cookies.txt \
		-d '[ \
			{"correlation_id": "1", "original_url": "https://google.com12313"}, \
			{"correlation_id": "2", "original_url": "https://youtube.com123123qdas"}, \
			{"correlation_id": "3", "original_url": "https://github.comasdasdsda"}, \
			{"correlation_id": "4", "original_url": "https://stackoverflow.comasdasda"}, \
			{"correlation_id": "5", "original_url": "https://reddit.comasdasdas"}, \
			{"correlation_id": "6", "original_url": "https://twitter.com123132d1d"}, \
			{"correlation_id": "7", "original_url": "https://linkedin.comasd21d"}, \
			{"correlation_id": "8", "original_url": "https://amazon.comasddd21"}, \
			{"correlation_id": "9", "original_url": "https://netflix.comasd23232d"}, \
			{"correlation_id": "10", "original_url": "https://microsoft.comasd321d"} \
		]' \
		http://localhost:8080/api/shorten/batch | python3 -m json.tool
	@echo ""
	
	# Очищаем временные файлы
	@rm -f cookies.txt short_url.txt request cookies.txt
	
	@echo "=== All tests completed ==="

test_ping:
	@echo "=== Simple curl tests ==="
	@echo "Testing /ping endpoint:"
	@curl -s -o /dev/null -w "Status: %{http_code}\n" http://localhost:8080/ping
	
	@echo ""
	@echo "Testing batch URL creation:"
	@curl -s -X POST \
		-H "Content-Type: application/json" \
		-H "Cookie: auth_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTc0MzI4MjAsImlhdCI6MTc1NzQzMTkyMCwiVXNlcklEIjoxfQ.39o8PRoB-OALTSKc-F3WGO-MOkVCjkP8iL6DMZ2Lo0Y" \
		-d '[{"correlation_id": "1", "original_url": "https://example.com"}]' \
		http://localhost:8080/api/shorten/batch | python3 -m json.tool 2>/dev/null || echo "Test completed"
	@echo "=== Simple tests completed ==="



# EXPERIMENTAL!!!


# # Основные команды
# make run          # Запуск сервисов
# make build        # Пересборка и запуск
# make test         # Запуск тестов Go
# make cover        # Покрытие кода тестами

# # Очистка
# make clean        # Остановка контейнеров
# make clean-all    # Полная очистка

# # Тестирование
# make test-ping    # Быстрая проверка работы
# make test-curl    # Полное тестирование API
# make test-verbose # Подробное тестирование