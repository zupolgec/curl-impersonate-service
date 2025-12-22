#!/bin/bash

# Example test requests for curl-impersonate-service
# Make sure to set your TOKEN and BASE_URL

TOKEN="${TOKEN:-test-token-123}"
BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "=== Testing curl-impersonate-service ==="
echo

# Test 1: Health check (no auth)
echo "1. Health check:"
curl -s "$BASE_URL/health" | jq .
echo

# Test 2: List browsers (with auth)
echo "2. List available browsers:"
curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/browsers" | jq '.browsers | length'
echo

# Test 3: Get metrics
echo "3. Service metrics:"
curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/metrics" | jq .
echo

# Test 4: Simple GET request with chrome-latest (default)
echo "4. Simple GET request (chrome-latest):"
curl -s -X POST "$BASE_URL/impersonate?token=$TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://httpbin.org/get"
  }' | jq '{success, status_code, timing}'
echo

# Test 5: GET request with specific browser
echo "5. GET with Firefox 109:"
curl -s -X POST "$BASE_URL/impersonate" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "browser": "ff109",
    "url": "https://httpbin.org/user-agent"
  }' | jq '{success, status_code, body}'
echo

# Test 6: POST request with custom headers
echo "6. POST with custom headers:"
curl -s -X POST "$BASE_URL/impersonate" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "browser": "chrome116",
    "url": "https://httpbin.org/post",
    "method": "POST",
    "headers": {
      "X-Custom-Header": "test-value"
    },
    "body": "{\"test\": \"data\"}"
  }' | jq '{success, status_code}'
echo

# Test 7: Request with query params
echo "7. Request with query params:"
curl -s -X POST "$BASE_URL/impersonate" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "browser": "edge101",
    "url": "https://httpbin.org/get",
    "query_params": {
      "key1": "value1",
      "key2": "value with spaces"
    }
  }' | jq '{success, status_code, final_url}'
echo

# Test 8: Test timeout
echo "8. Test with custom timeout:"
curl -s -X POST "$BASE_URL/impersonate" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "browser": "safari15_5",
    "url": "https://httpbin.org/delay/2",
    "timeout": 5
  }' | jq '{success, status_code, timing}'
echo

# Test 9: Test authentication failure
echo "9. Test invalid token (should fail):"
curl -s -X POST "$BASE_URL/impersonate" \
  -H "Authorization: Bearer invalid-token" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://httpbin.org/get"
  }' | jq .
echo

# Test 10: Test invalid browser
echo "10. Test invalid browser (should fail):"
curl -s -X POST "$BASE_URL/impersonate" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "browser": "invalid-browser",
    "url": "https://httpbin.org/get"
  }' | jq .
echo

echo "=== All tests completed ==="
