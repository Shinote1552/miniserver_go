#!/bin/bash

BASE_URL=${1:-http://localhost:8080}
TIMEOUT=5  # 5 секунд таймаут

echo "=== Simple curl tests: $BASE_URL ==="

# Тестируем /ping endpoint
echo "Testing /ping endpoint:"
if curl -s -o /dev/null -w "Status: %{http_code}\n" --max-time $TIMEOUT "$BASE_URL/ping"; then
    echo "✓ Ping successful"
else
    echo "✗ Ping failed or timed out"
    exit 1
fi

echo ""

# Пропускаем batch creation если сервис не отвечает нормально
echo "Testing quick API call (skip batch if fails):"
if curl -s -o /dev/null -w "API status: %{http_code}\n" --max-time $TIMEOUT "$BASE_URL/"; then
    echo "✓ API is responsive"
else
    echo "✗ API not responsive, skipping batch test"
    exit 1
fi

echo "=== Simple tests completed ==="