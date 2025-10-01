#!/bin/bash
BASE_URL=${1:-http://localhost:8080}
PORT=${BASE_URL##*:}
PORT=${PORT%/}
TIMEOUT=10

echo "=== Comprehensive API Tests: $BASE_URL ==="
echo ""

# Очистка предыдущих файлов
rm -f cookies.txt short_url.txt response.json batch_response.json delete_ids.txt

# Проверяем что сервис доступен
echo "Checking service availability..."
if ! curl -s -o /dev/null --max-time 5 "$BASE_URL/ping"; then
    echo "✗ Service is not available at $BASE_URL"
    exit 1
fi
echo "✓ Service is available"
echo ""

# 1. Получаем JWT токен
echo "1. Getting new JWT token:"
if ! curl -s -X POST --max-time $TIMEOUT "$BASE_URL/" -c cookies.txt > /dev/null; then
    echo "✗ Failed to get JWT token"
    exit 1
fi
echo "✓ New token saved to cookies.txt"
echo ""

# 2. Тестируем публичные endpoint'ы
echo "2. Testing public endpoints:"

echo "2.1. GET /ping:"
curl -s -o /dev/null -w "Status: %{http_code}\n" --max-time $TIMEOUT -X GET "$BASE_URL/ping"
echo ""

echo "2.2. GET / (default handler):"
curl -s -o /dev/null -w "Status: %{http_code}\n" --max-time $TIMEOUT -X GET "$BASE_URL/"
echo ""

# 3. Тестируем защищённые endpoint'ы
echo "3. Testing protected endpoints:"

echo "3.1. POST / (text/plain) - NEW URL:"
if curl -s -X POST --max-time $TIMEOUT \
    -H "Content-Type: text/plain" \
    -b cookies.txt \
    -d "https://google.com/$(date +%s)" \
    "$BASE_URL/" \
    | tee short_url.txt; then
    echo "✓ Text URL shortened"
else
    echo "✗ Text URL shortening failed"
fi
echo ""

echo "3.2. POST /api/shorten (application/json) - NEW URL:"
echo "Request: {\"url\":\"https://yandex.ru/$(date +%s)\"}"
if curl -s -X POST --max-time $TIMEOUT \
    -H "Content-Type: application/json" \
    -b cookies.txt \
    -d "{\"url\":\"https://yandex.ru/$(date +%s)\"}" \
    -w "Status: %{http_code}\n" \
    "$BASE_URL/api/shorten" | tee response.json; then
    echo "✓ JSON URL shortened"
    echo "Response: $(cat response.json)"
else
    echo "✗ JSON URL shortening failed"
fi
echo ""

echo "3.3. POST /api/shorten/batch (batch create):"
BATCH_DATA='[
    {"correlation_id": "1", "original_url": "https://google.com/batch1"},
    {"correlation_id": "2", "original_url": "https://youtube.com/batch2"}
]'

echo "Batch request sent"
if curl -s -X POST --max-time $TIMEOUT \
    -H "Content-Type: application/json" \
    -b cookies.txt \
    -d "$BATCH_DATA" \
    -w "Status: %{http_code}\n" \
    "$BASE_URL/api/shorten/batch" | tee batch_response.json; then
    echo "✓ Batch URLs created"
    echo "Response: $(cat batch_response.json)"
    
    # Сохраняем short_urls для последующего удаления
    if command -v jq >/dev/null 2>&1; then
        jq -r '.[].short_url' batch_response.json | sed "s|$BASE_URL/||" > delete_ids.txt
        echo "Short IDs saved for deletion: $(cat delete_ids.txt | tr '\n' ' ')"
    fi
else
    echo "✗ Batch URL creation failed"
fi
echo ""

echo "3.4. GET /api/user/urls (before deletion):"
if curl -s --max-time $TIMEOUT -X GET \
    -b cookies.txt \
    -w "Status: %{http_code}\n" \
    "$BASE_URL/api/user/urls" | tee response.json; then
    
    RESPONSE=$(cat response.json)
    if [ -n "$RESPONSE" ]; then
        echo "✓ User URLs retrieved"
        echo "Response length: ${#RESPONSE} characters"
        
        # Простой подсчет элементов (если это JSON массив)
        if echo "$RESPONSE" | grep -q "\[.*\]"; then
            COUNT=$(echo "$RESPONSE" | tr -d '[:space:]' | grep -o ',' | wc -l)
            COUNT=$((COUNT + 1))
            echo "Number of user URLs: $COUNT"
        fi
    else
        echo "No user URLs found"
    fi
else
    echo "✗ Failed to get user URLs"
fi
echo ""

# 4. Тестируем редирект ДО удаления
echo "4. Testing redirect BEFORE deletion:"
if [ -f short_url.txt ]; then
    SHORT_URL=$(cat short_url.txt | tr -d '\n' | tr -d '\r')
    echo "Testing redirect for: $SHORT_URL"
    SHORT_ID=${SHORT_URL##*$PORT/}
    
    if [ -n "$SHORT_ID" ] && [ "$SHORT_ID" != "$SHORT_URL" ]; then
        echo "Redirect test for ID: $SHORT_ID"
        if curl -s -o /dev/null -w "Redirect: %{http_code} -> %{redirect_url}\n" --max-time $TIMEOUT -X GET "$BASE_URL/$SHORT_ID"; then
            echo "✓ Redirect successful"
        else
            echo "✗ Redirect failed"
        fi
    else
        echo "Invalid short URL format"
    fi
else
    echo "No short URL found for redirect test"
fi
echo ""

# 5. Большое пакетное создание URL
echo "5. Large batch URL creation:"
LARGE_BATCH_DATA='[
    {"correlation_id": "1", "original_url": "https://google.com12313"},
    {"correlation_id": "2", "original_url": "https://youtube.com123123qdas"},
    {"correlation_id": "3", "original_url": "https://github.comasdasdsda"},
    {"correlation_id": "4", "original_url": "https://stackoverflow.comasdasda"},
    {"correlation_id": "5", "original_url": "https://reddit.comasdasdas"},
    {"correlation_id": "6", "original_url": "https://twitter.com123132d1d"},
    {"correlation_id": "7", "original_url": "https://linkedin.comasd21d"},
    {"correlation_id": "8", "original_url": "https://amazon.comasddd21"},
    {"correlation_id": "9", "original_url": "https://netflix.comasd23232d"},
    {"correlation_id": "10", "original_url": "https://microsoft.comasd321d"}
]'

echo "Large batch request sent"
if curl -s -X POST --max-time $TIMEOUT \
    -H "Content-Type: application/json" \
    -b cookies.txt \
    -d "$LARGE_BATCH_DATA" \
    -w "Status: %{http_code}\n" \
    "$BASE_URL/api/shorten/batch" | tee response.json; then
    echo "✓ Large batch URLs created"
    echo "Response: $(cat response.json)"
    
    # Сохраняем дополнительные ID для удаления
    if command -v jq >/dev/null 2>&1; then
        jq -r '.[].short_url' response.json | sed "s|$BASE_URL/||" >> delete_ids.txt
        echo "Additional short IDs saved for deletion"
    fi
else
    echo "✗ Large batch URL creation failed"
fi
echo ""

# 6. ТЕСТИРОВАНИЕ BATCH DELETE
echo "6. Testing BATCH DELETE functionality:"

if [ -f delete_ids.txt ] && [ -s delete_ids.txt ]; then
    # Берем первые 3 ID для теста удаления
    DELETE_IDS=$(head -3 delete_ids.txt | tr '\n' ',' | sed 's/,$//')
    DELETE_JSON="[$(echo $DELETE_IDS | sed 's/,/","/g' | sed 's/^/"/' | sed 's/$/"/')]"
    
    echo "DELETE request with IDs: $DELETE_IDS"
    echo "Request body: $DELETE_JSON"
    
    echo "6.1. Sending DELETE /api/user/urls:"
    if curl -s -X DELETE --max-time $TIMEOUT \
        -H "Content-Type: application/json" \
        -b cookies.txt \
        -d "$DELETE_JSON" \
        -w "Status: %{http_code}\n" \
        "$BASE_URL/api/user/urls" | tee response.json; then
        echo "✓ DELETE request accepted"
    else
        echo "✗ DELETE request failed"
    fi
    echo ""
    
    # Ждем немного для асинхронной обработки
    echo "Waiting 2 seconds for async processing..."
    sleep 2
    echo ""
    
    echo "6.2. Testing access to deleted URLs (should return 410 Gone):"
    for id in $(echo $DELETE_IDS | tr ',' ' '); do
        echo "Testing full URL: $id"
        # Извлекаем short ID для отладки
        SHORT_ID=$(echo "$id" | sed 's|.*/||')
        echo "Extracted short ID: $SHORT_ID"
        curl -s -o /dev/null -w "Status: %{http_code}\n" --max-time $TIMEOUT -X GET "$BASE_URL/$SHORT_ID"
    done
    
    echo "6.3. GET /api/user/urls (after deletion):"
    if curl -s --max-time $TIMEOUT -X GET \
        -b cookies.txt \
        -w "Status: %{http_code}\n" \
        "$BASE_URL/api/user/urls" | tee response.json; then
        
        RESPONSE=$(cat response.json)
        if [ -n "$RESPONSE" ]; then
            echo "✓ User URLs retrieved after deletion"
            echo "Response length: ${#RESPONSE} characters"
        else
            echo "No user URLs found after deletion"
        fi
    else
        echo "✗ Failed to get user URLs after deletion"
    fi
    echo ""
    
else
    echo "✗ No short IDs available for deletion test"
    echo ""
fi

# 7. Тестируем некорректные запросы на удаление
echo "7. Testing invalid DELETE requests:"

echo "7.1. DELETE without authentication:"
curl -s -X DELETE --max-time $TIMEOUT \
    -H "Content-Type: application/json" \
    -d '["test1","test2"]' \
    -w "Status: %{http_code}\n" \
    "$BASE_URL/api/user/urls"
echo ""

echo "7.2. DELETE with empty array:"
curl -s -X DELETE --max-time $TIMEOUT \
    -H "Content-Type: application/json" \
    -b cookies.txt \
    -d '[]' \
    -w "Status: %{http_code}\n" \
    "$BASE_URL/api/user/urls"
echo ""

echo "7.3. DELETE with invalid JSON:"
curl -s -X DELETE --max-time $TIMEOUT \
    -H "Content-Type: application/json" \
    -b cookies.txt \
    -d 'invalid json' \
    -w "Status: %{http_code}\n" \
    "$BASE_URL/api/user/urls"
echo ""

# Очищаем временные файлы
rm -f cookies.txt short_url.txt response.json batch_response.json delete_ids.txt

echo "=== All tests completed for $BASE_URL ==="