#!/bin/bash

# Integration tests for curl-impersonate-service
# This script tests the actual service running in Docker

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

BASE_URL="${BASE_URL:-http://localhost:8080}"
TOKEN="${TOKEN:-test-token-123}"

TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

log_test() {
    echo -e "${YELLOW}[TEST]${NC} $1"
}

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((TESTS_PASSED++))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((TESTS_FAILED++))
}

assert_eq() {
    local expected="$1"
    local actual="$2"
    local message="$3"
    
    ((TESTS_RUN++))
    
    if [ "$expected" = "$actual" ]; then
        log_pass "$message"
        return 0
    else
        log_fail "$message (expected: '$expected', got: '$actual')"
        return 1
    fi
}

assert_contains() {
    local haystack="$1"
    local needle="$2"
    local message="$3"
    
    ((TESTS_RUN++))
    
    if echo "$haystack" | grep -q "$needle"; then
        log_pass "$message"
        return 0
    else
        log_fail "$message (expected to contain: '$needle')"
        return 1
    fi
}

assert_not_empty() {
    local value="$1"
    local message="$2"
    
    ((TESTS_RUN++))
    
    if [ -n "$value" ]; then
        log_pass "$message"
        return 0
    else
        log_fail "$message (value was empty)"
        return 1
    fi
}

echo "========================================="
echo "curl-impersonate-service Integration Tests"
echo "========================================="
echo

# Test 1: Health check (no auth)
log_test "Health check endpoint (no auth required)"
HEALTH_RESPONSE=$(curl -s "$BASE_URL/health")
assert_eq "ok" "$(echo "$HEALTH_RESPONSE" | jq -r '.status')" "Health check returns ok"
assert_not_empty "$(echo "$HEALTH_RESPONSE" | jq -r '.version')" "Health check returns version"
echo

# Test 2: Browsers endpoint (with auth)
log_test "Browsers endpoint (requires auth)"
BROWSERS_RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/browsers")
assert_not_empty "$(echo "$BROWSERS_RESPONSE" | jq -r '.browsers[0].name')" "Browsers endpoint returns browser list"
CHROME_ALIAS=$(echo "$BROWSERS_RESPONSE" | jq -r '.aliases["chrome-latest"]')
assert_not_empty "$CHROME_ALIAS" "Browsers endpoint returns aliases"
echo

# Test 3: Missing auth should fail
log_test "Missing authentication"
AUTH_FAIL_RESPONSE=$(curl -s "$BASE_URL/browsers")
assert_eq "auth" "$(echo "$AUTH_FAIL_RESPONSE" | jq -r '.error_type')" "Missing auth returns error_type auth"
echo

# Test 4: Chrome impersonation with actual request
log_test "Chrome 116 impersonation"
CHROME_RESPONSE=$(curl -s -X POST "$BASE_URL/impersonate?token=$TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"browser":"chrome116","url":"https://httpbin.org/get"}')

assert_eq "true" "$(echo "$CHROME_RESPONSE" | jq -r '.success')" "Chrome request succeeds"
assert_eq "200" "$(echo "$CHROME_RESPONSE" | jq -r '.status_code')" "Chrome request returns 200"
assert_not_empty "$(echo "$CHROME_RESPONSE" | jq -r '.body')" "Chrome request returns body"

# Verify browser impersonation is working (should NOT be just "curl")
USER_AGENT=$(echo "$CHROME_RESPONSE" | jq -r '.body | fromjson | .headers."User-Agent"')
assert_not_empty "$USER_AGENT" "Chrome request has User-Agent"
echo "  User-Agent: $USER_AGENT"
echo

# Test 5: Firefox impersonation
log_test "Firefox 109 impersonation"
FF_RESPONSE=$(curl -s -X POST "$BASE_URL/impersonate?token=$TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"browser":"ff109","url":"https://httpbin.org/get"}')

assert_eq "true" "$(echo "$FF_RESPONSE" | jq -r '.success')" "Firefox request succeeds"
assert_eq "200" "$(echo "$FF_RESPONSE" | jq -r '.status_code')" "Firefox request returns 200"

FF_USER_AGENT=$(echo "$FF_RESPONSE" | jq -r '.body | fromjson | .headers."User-Agent"')
assert_not_empty "$FF_USER_AGENT" "Firefox request has User-Agent"
echo "  User-Agent: $FF_USER_AGENT"
echo

# Test 6: Browser alias (chrome-latest)
log_test "Browser alias (chrome-latest)"
ALIAS_RESPONSE=$(curl -s -X POST "$BASE_URL/impersonate?token=$TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"browser":"chrome-latest","url":"https://httpbin.org/get"}')

assert_eq "true" "$(echo "$ALIAS_RESPONSE" | jq -r '.success')" "Browser alias works"
echo

# Test 7: POST request with body
log_test "POST request with body"
POST_RESPONSE=$(curl -s -X POST "$BASE_URL/impersonate?token=$TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"browser":"chrome116","url":"https://httpbin.org/post","method":"POST","body":"{\"test\":\"data\"}"}')

assert_eq "true" "$(echo "$POST_RESPONSE" | jq -r '.success')" "POST request succeeds"
assert_eq "200" "$(echo "$POST_RESPONSE" | jq -r '.status_code')" "POST request returns 200"
echo

# Test 8: Custom headers
log_test "Custom headers"
HEADERS_RESPONSE=$(curl -s -X POST "$BASE_URL/impersonate?token=$TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"browser":"chrome116","url":"https://httpbin.org/headers","headers":{"X-Custom-Test":"test-value"}}')

assert_eq "true" "$(echo "$HEADERS_RESPONSE" | jq -r '.success')" "Custom headers request succeeds"
CUSTOM_HEADER=$(echo "$HEADERS_RESPONSE" | jq -r '.body | fromjson | .headers."X-Custom-Test"')
assert_eq "test-value" "$CUSTOM_HEADER" "Custom header is sent"
echo

# Test 9: Query parameters
log_test "Query parameters"
QUERY_RESPONSE=$(curl -s -X POST "$BASE_URL/impersonate?token=$TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"browser":"chrome116","url":"https://httpbin.org/get","query_params":{"foo":"bar","test":"123"}}')

assert_eq "true" "$(echo "$QUERY_RESPONSE" | jq -r '.success')" "Query params request succeeds"
assert_contains "$(echo "$QUERY_RESPONSE" | jq -r '.body | fromjson | .url')" "foo=bar" "Query param foo added"
assert_contains "$(echo "$QUERY_RESPONSE" | jq -r '.body | fromjson | .url')" "test=123" "Query param test added"
echo

# Test 10: Invalid browser
log_test "Invalid browser name"
INVALID_RESPONSE=$(curl -s -X POST "$BASE_URL/impersonate?token=$TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"browser":"invalid-browser","url":"https://httpbin.org/get"}')

assert_eq "false" "$(echo "$INVALID_RESPONSE" | jq -r '.success')" "Invalid browser fails"
assert_eq "validation" "$(echo "$INVALID_RESPONSE" | jq -r '.error_type')" "Invalid browser returns validation error"
echo

# Test 11: Timing information
log_test "Timing information"
TIMING_RESPONSE=$(curl -s -X POST "$BASE_URL/impersonate?token=$TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"browser":"chrome116","url":"https://httpbin.org/get"}')

TOTAL_TIME=$(echo "$TIMING_RESPONSE" | jq -r '.timing.total')
assert_not_empty "$TOTAL_TIME" "Timing total is present"
echo "  Total time: ${TOTAL_TIME}s"
echo

# Test 12: Metrics endpoint
log_test "Metrics endpoint"
METRICS_RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/metrics")
REQUESTS_TOTAL=$(echo "$METRICS_RESPONSE" | jq -r '.requests_total')
assert_not_empty "$REQUESTS_TOTAL" "Metrics returns request count"
echo "  Total requests: $REQUESTS_TOTAL"
echo

echo "========================================="
echo "Test Results"
echo "========================================="
echo "Total tests:  $TESTS_RUN"
echo -e "Passed:       ${GREEN}$TESTS_PASSED${NC}"
echo -e "Failed:       ${RED}$TESTS_FAILED${NC}"
echo

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
fi
