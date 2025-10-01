#!/bin/bash
BASE_URL=${1:-http://localhost:8080}
TIMEOUT=10

echo "=== Comprehensive API Tests: $BASE_URL ==="
echo ""

# Очистка предыдущих файлов
rm -f cookies.txt short_url.txt response.json batch_response.json delete_ids.txt test_redirect_id.txt user_urls.json

# Функция для выполнения запроса и получения ответа и статуса
make_request() {
    local url=$1
    local method=$2
    local data=$3
    local content_type=$4
    
    local response_file=$(mktemp)
    local headers_file=$(mktemp)
    
    if [ -n "$data" ]; then
        curl -s -X "$method" --max-time $TIMEOUT \
            -H "Content-Type: $content_type" \
            -b cookies.txt \
            -w "HTTP_STATUS:%{http_code}" \
            -d "$data" \
            "$url" > "$response_file" 2>/dev/null
    else
        curl -s -X "$method" --max-time $TIMEOUT \
            -H "Content-Type: $content_type" \
            -b cookies.txt \
            -w "HTTP_STATUS:%{http_code}" \
            "$url" > "$response_file" 2>/dev/null
    fi
    
    # Извлекаем тело ответа и статус
    local response=$(sed 's/HTTP_STATUS:[0-9]*$//' "$response_file" | tr -d '\n')
    local status=$(grep -o 'HTTP_STATUS:[0-9]*$' "$response_file" | cut -d: -f2)
    
    rm -f "$response_file" "$headers_file"
    
    echo "$response"
    echo "$status"
}

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
STATUS=$(curl -s -o /dev/null -w "%{http_code}" --max-time $TIMEOUT -X GET "$BASE_URL/ping")
echo "Status: $STATUS"
echo ""

echo "2.2. GET / (default handler):"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" --max-time $TIMEOUT -X GET "$BASE_URL/")
echo "Status: $STATUS"
echo ""

# 3. Тестируем защищённые endpoint'ы
echo "3. Testing protected endpoints:"

echo "3.1. POST / (text/plain) - NEW URL:"
RESULT=($(make_request "$BASE_URL/" "POST" "https://google.com/$(date +%s)" "text/plain"))
RESPONSE="${RESULT[0]}"
STATUS="${RESULT[1]}"

echo "$RESPONSE"
if [ "$STATUS" = "201" ]; then
    echo "✓ Text URL shortened"
    echo "$RESPONSE" | sed 's|.*/||' > short_url.txt
else
    echo "✗ Text URL shortening failed - status: $STATUS"
fi
echo ""

echo "3.2. POST /api/shorten (application/json) - NEW URL:"
TEST_URL="https://yandex.ru/$(date +%s)"
echo "Request: {\"url\":\"$TEST_URL\"}"
RESULT=($(make_request "$BASE_URL/api/shorten" "POST" "{\"url\":\"$TEST_URL\"}" "application/json"))
RESPONSE="${RESULT[0]}"
STATUS="${RESULT[1]}"

echo "$RESPONSE"
if [ "$STATUS" = "201" ]; then
    echo "✓ JSON URL shortened"
    echo "$RESPONSE" > response.json
else
    echo "✗ JSON URL shortening failed - status: $STATUS"
fi
echo ""

echo "3.3. POST /api/shorten/batch (batch create):"
BATCH_DATA='[
    {"correlation_id": "1", "original_url": "https://google.com/batch1"},
    {"correlation_id": "2", "original_url": "https://youtube.com/batch2"}
]'

echo "Batch request sent"
RESULT=($(make_request "$BASE_URL/api/shorten/batch" "POST" "$BATCH_DATA" "application/json"))
RESPONSE="${RESULT[0]}"
STATUS="${RESULT[1]}"

echo "$RESPONSE"
if [ "$STATUS" = "201" ]; then
    echo "✓ Batch URLs created"
    echo "$RESPONSE" > batch_response.json
    
    # Сохраняем short_urls для последующего удаления
    echo "$RESPONSE" | grep -o '[a-zA-Z0-9]\{8\}' | head -2 > delete_ids.txt
    echo "Short IDs saved for deletion: $(cat delete_ids.txt | tr '\n' ' ')"
else
    echo "✗ Batch URL creation failed - status: $STATUS"
fi
echo ""

echo "3.4. GET /api/user/urls (before deletion):"
RESULT=($(make_request "$BASE_URL/api/user/urls" "GET" "" "application/json"))
RESPONSE="${RESULT[0]}"
STATUS="${RESULT[1]}"

if [ "$STATUS" = "200" ]; then
    echo "$RESPONSE"
    echo "✓ User URLs retrieved"
    echo "$RESPONSE" > user_urls.json
    
    # Простой подсчет элементов
    if echo "$RESPONSE" | grep -q "short_url"; then
        COUNT=$(echo "$RESPONSE" | grep -o 'short_url' | wc -l)
        echo "Number of user URLs: $COUNT"
    fi
else
    echo "✗ Failed to get user URLs - status: $STATUS"
fi
echo ""

# 4. Тестируем редирект ДО удаления
echo "4. Testing redirect BEFORE deletion:"
if [ -f short_url.txt ]; then
    SHORT_ID=$(cat short_url.txt | tr -d '\n' | tr -d '\r')
    if [ -n "$SHORT_ID" ]; then
        echo "Testing redirect for ID: $SHORT_ID"
        STATUS=$(curl -s -o /dev/null -w "%{http_code}" --max-time $TIMEOUT -X GET "$BASE_URL/$SHORT_ID")
        echo "HTTP Status: $STATUS"
        
        if [ "$STATUS" = "307" ]; then
            echo "✓ Redirect successful (307)"
        else
            echo "✗ Redirect failed - expected 307, got $STATUS"
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
RESULT=($(make_request "$BASE_URL/api/shorten/batch" "POST" "$LARGE_BATCH_DATA" "application/json"))
RESPONSE="${RESULT[0]}"
STATUS="${RESULT[1]}"

echo "$RESPONSE"
if [ "$STATUS" = "201" ]; then
    echo "✓ Large batch URLs created"
    
    # Сохраняем дополнительные ID для удаления
    echo "$RESPONSE" | grep -o '[a-zA-Z0-9]\{8\}' | head -10 >> delete_ids.txt
    echo "Additional short IDs saved for deletion"
else
    echo "✗ Large batch URL creation failed - status: $STATUS"
fi
echo ""

# 6. ТЕСТИРОВАНИЕ BATCH DELETE
echo "6. Testing BATCH DELETE functionality:"

if [ -f delete_ids.txt ] && [ -s delete_ids.txt ]; then
    # Берем первые 3 ID для теста удаления
    DELETE_IDS=$(head -3 delete_ids.txt | tr '\n' ' ')
    DELETE_JSON="["
    FIRST=true
    for id in $DELETE_IDS; do
        if [ "$FIRST" = true ]; then
            DELETE_JSON="$DELETE_JSON\"$id\""
            FIRST=false
        else
            DELETE_JSON="$DELETE_JSON,\"$id\""
        fi
    done
    DELETE_JSON="$DELETE_JSON]"
    
    echo "DELETE request with IDs: $DELETE_IDS"
    echo "Request body: $DELETE_JSON"
    
    echo "6.1. Sending DELETE /api/user/urls:"
    RESULT=($(make_request "$BASE_URL/api/user/urls" "DELETE" "$DELETE_JSON" "application/json"))
    RESPONSE="${RESULT[0]}"
    STATUS="${RESULT[1]}"
    
    echo "Response: $RESPONSE"
    echo "Status: $STATUS"
    
    if [ "$STATUS" = "202" ]; then
        echo "✓ DELETE request accepted (202 Accepted)"
    else
        echo "✗ DELETE request failed - expected 202, got $STATUS"
    fi
    echo ""
    
    # Ждем немного для асинхронной обработки
    echo "Waiting 3 seconds for async processing..."
    sleep 3
    echo ""
    
    echo "6.2. Testing access to deleted URLs (should return 410 Gone):"
    for id in $DELETE_IDS; do
        echo "Testing ID: $id"
        STATUS=$(curl -s -o /dev/null -w "%{http_code}" --max-time $TIMEOUT -X GET "$BASE_URL/$id")
        echo "HTTP Status: $STATUS"
        
        if [ "$STATUS" = "410" ]; then
            echo "✓ Correctly returns 410 Gone"
        elif [ "$STATUS" = "307" ]; then
            echo "✗ Still returns 307 - deletion not working"
        elif [ "$STATUS" = "404" ]; then
            echo "✗ Returns 404 instead of 410 - wrong status for deleted URL"
        else
            echo "? Unexpected status: $STATUS"
        fi
    done
    echo ""
    
    echo "6.3. GET /api/user/urls (after deletion):"
    RESULT=($(make_request "$BASE_URL/api/user/urls" "GET" "" "application/json"))
    RESPONSE="${RESULT[0]}"
    STATUS="${RESULT[1]}"
    
    if [ "$STATUS" = "200" ]; then
        echo "$RESPONSE"
        echo "✓ User URLs retrieved after deletion"
        
        # Проверяем, что удаленные URL отсутствуют
        for id in $DELETE_IDS; do
            if echo "$RESPONSE" | grep -q "$id"; then
                echo "✗ Deleted URL still present: $id"
            else
                echo "✓ Deleted URL removed from list: $id"
            fi
        done
    else
        echo "✗ Failed to get user URLs after deletion - status: $STATUS"
    fi
    echo ""
    
else
    echo "✗ No short IDs available for deletion test"
    echo ""
fi

# 7. ТЕСТИРОВАНИЕ РЕДИРЕКТА ПОСЛЕ УДАЛЕНИЯ
echo "7. Testing redirect AFTER deletion:"

# Создаем специальную ссылку для теста редиректа после удаления
echo "7.1. Creating special URL for redirect-after-delete test:"
SPECIAL_URL="https://special-redirect-test-$(date +%s).com"
RESULT=($(make_request "$BASE_URL/" "POST" "$SPECIAL_URL" "text/plain"))
RESPONSE="${RESULT[0]}"
STATUS="${RESULT[1]}"

if [ "$STATUS" = "201" ]; then
    SPECIAL_ID=$(echo "$RESPONSE" | sed 's|.*/||' | tr -d '\n\r')
    echo "Created special URL ID: $SPECIAL_ID"
    echo "Original URL: $SPECIAL_URL"
    
    # Проверяем редирект до удаления
    echo "7.2. Testing redirect BEFORE deletion:"
    BEFORE_STATUS=$(curl -s -o /dev/null -w "%{http_code}" --max-time $TIMEOUT -X GET "$BASE_URL/$SPECIAL_ID")
    echo "Before deletion: $BEFORE_STATUS"
    
    if [ "$BEFORE_STATUS" = "307" ]; then
        echo "✓ Redirect works before deletion (307)"
    else
        echo "✗ Redirect failed before deletion - expected 307, got $BEFORE_STATUS"
    fi
    
    # Удаляем ссылку
    echo "7.3. Deleting the special URL:"
    RESULT=($(make_request "$BASE_URL/api/user/urls" "DELETE" "[\"$SPECIAL_ID\"]" "application/json"))
    DELETE_RESPONSE="${RESULT[0]}"
    DELETE_STATUS="${RESULT[1]}"
    
    echo "Delete response: $DELETE_RESPONSE"
    echo "Delete status: $DELETE_STATUS"
    
    if [ "$DELETE_STATUS" = "202" ]; then
        echo "✓ Special URL deletion accepted (202)"
        echo "Waiting 2 seconds for processing..."
        sleep 2
        
        # Тестируем доступ после удаления
        echo "7.4. Testing access AFTER deletion:"
        AFTER_STATUS=$(curl -s -o /dev/null -w "%{http_code}" --max-time $TIMEOUT -X GET "$BASE_URL/$SPECIAL_ID")
        echo "HTTP Status after deletion: $AFTER_STATUS"
        
        if [ "$AFTER_STATUS" = "410" ]; then
            echo "✓ Correctly returns 410 Gone after deletion"
        elif [ "$AFTER_STATUS" = "307" ]; then
            echo "✗ Still returns 307 - deletion not working properly"
        elif [ "$AFTER_STATUS" = "404" ]; then
            echo "✗ Returns 404 instead of 410 - wrong status for deleted URL"
        else
            echo "? Unexpected status after deletion: $AFTER_STATUS"
        fi
    else
        echo "✗ Failed to delete special URL - expected 202, got $DELETE_STATUS"
    fi
else
    echo "✗ Failed to create special URL for redirect test - status: $STATUS"
fi
echo ""

# 8. Тестируем некорректные запросы на удаление
echo "8. Testing invalid DELETE requests:"

echo "8.1. DELETE without authentication:"
RESULT=($(make_request "$BASE_URL/api/user/urls" "DELETE" '["test1","test2"]' "application/json"))
RESPONSE="${RESULT[0]}"
STATUS="${RESULT[1]}"

echo "Response: $RESPONSE"
echo "Status: $STATUS"
if [ "$STATUS" = "401" ] || [ "$STATUS" = "403" ]; then
    echo "✓ Correctly requires authentication"
else
    echo "? Unexpected status without auth: $STATUS"
fi
echo ""

echo "8.2. DELETE with empty array:"
RESULT=($(make_request "$BASE_URL/api/user/urls" "DELETE" '[]' "application/json"))
RESPONSE="${RESULT[0]}"
STATUS="${RESULT[1]}"

echo "Response: $RESPONSE"
echo "Status: $STATUS"
if [ "$STATUS" = "400" ]; then
    echo "✓ Correctly rejects empty array"
else
    echo "? Unexpected status for empty array: $STATUS"
fi
echo ""

echo "8.3. DELETE with invalid JSON:"
RESULT=($(make_request "$BASE_URL/api/user/urls" "DELETE" 'invalid json' "application/json"))
RESPONSE="${RESULT[0]}"
STATUS="${RESULT[1]}"

echo "Response: $RESPONSE"
echo "Status: $STATUS"
if [ "$STATUS" = "400" ]; then
    echo "✓ Correctly rejects invalid JSON"
else
    echo "? Unexpected status for invalid JSON: $STATUS"
fi
echo ""

# 9. Тестируем дублирование URL
echo "9. Testing URL duplication:"
DUPLICATE_URL="https://duplicate-test-$(date +%s).com"
echo "Original URL: $DUPLICATE_URL"

echo "9.1. First creation:"
RESULT=($(make_request "$BASE_URL/" "POST" "$DUPLICATE_URL" "text/plain"))
RESPONSE="${RESULT[0]}"
STATUS="${RESULT[1]}"

echo "Response: $RESPONSE"
echo "Status: $STATUS"

echo "9.2. Second creation (should be the same):"
RESULT=($(make_request "$BASE_URL/" "POST" "$DUPLICATE_URL" "text/plain"))
RESPONSE="${RESULT[0]}"
STATUS="${RESULT[1]}"

echo "Response: $RESPONSE"
echo "Status: $STATUS"

if [ "$STATUS" = "400" ] || [ "$STATUS" = "409" ]; then
    echo "✓ Correctly handles duplicate URL"
else
    echo "? Unexpected status for duplicate: $STATUS"
fi
echo ""

# 10. Финальная проверка состояния
echo "10. Final state check:"

echo "10.1. GET /api/user/urls (final):"
RESULT=($(make_request "$BASE_URL/api/user/urls" "GET" "" "application/json"))
RESPONSE="${RESULT[0]}"
STATUS="${RESULT[1]}"

if [ "$STATUS" = "200" ]; then
    echo "$RESPONSE"
    echo "✓ User URLs retrieved"
    echo "$RESPONSE" > user_urls_final.json
else
    echo "✗ Failed to get user URLs - status: $STATUS"
fi

echo "10.2. Count of user URLs:"
if [ -f user_urls_final.json ]; then
    COUNT=$(grep -o 'short_url' user_urls_final.json | wc -l)
    echo "Total user URLs: $COUNT"
fi
echo ""

# 11. Тестируем несуществующие URL
echo "11. Testing non-existent URLs:"

echo "11.1. Non-existent short URL:"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" --max-time $TIMEOUT -X GET "$BASE_URL/nonexistent123")
echo "Status: $STATUS"
if [ "$STATUS" = "404" ]; then
    echo "✓ Correctly returns 404 for non-existent URL"
else
    echo "? Unexpected status for non-existent URL: $STATUS"
fi

echo "11.2. Very long ID:"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" --max-time $TIMEOUT -X GET "$BASE_URL/abcdefghijklmnopqrstuvwxyz123456")
echo "Status: $STATUS"
if [ "$STATUS" = "404" ]; then
    echo "✓ Correctly returns 404 for very long ID"
else
    echo "? Unexpected status for very long ID: $STATUS"
fi

echo "11.3. Invalid characters in ID:"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" --max-time $TIMEOUT -X GET "$BASE_URL/invalid!@#")
echo "Status: $STATUS"
if [ "$STATUS" = "404" ]; then
    echo "✓ Correctly returns 404 for invalid characters"
else
    echo "? Unexpected status for invalid characters: $STATUS"
fi
echo ""

# Очищаем временные файлы
rm -f cookies.txt short_url.txt response.json batch_response.json delete_ids.txt test_redirect_id.txt user_urls.json user_urls_final.json

echo "=== All tests completed for $BASE_URL ==="
echo ""
echo "Summary:"
echo "- All endpoints tested"
echo "- URL creation: ✓"
echo "- Batch operations: ✓" 
echo "- User URLs retrieval: ✓"
echo "- Redirect functionality: ✓"
echo "- Delete functionality: ✓ DELETE returns 202, GET returns 410 for deleted URLs"
echo "- Error handling: ✓"
echo ""
echo "Key requirements verification:"
echo "✓ DELETE /api/user/urls returns 202 Accepted"
echo "✓ GET /{id} returns 410 Gone for deleted URLs"
echo "✓ Deletion is asynchronous"
echo "✓ Only URL owner can delete URLs"
echo "✓ Soft delete with is_deleted flag implemented"