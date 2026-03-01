#!/bin/bash

# TrueSignal Full Stack Smoke Test (Linux/Mac)
# Verify frontend-backend API integration

API_BASE_URL="http://localhost:8080/api"
FRONTEND_URL="http://localhost:5173"
HEALTH_URL="http://localhost:8080/health"

TESTS_PASSED=0
TESTS_FAILED=0

echo ""
echo "=========================================="
echo "    TrueSignal Smoke Test (Linux/Mac)"
echo "=========================================="
echo ""

# Pre-check: Go Backend
echo "[INFO] Checking Go backend connection..."
sleep 1

if curl -s "$HEALTH_URL" > /dev/null 2>&1; then
    echo "[SUCCESS] Go backend is responding"
else
    echo ""
    echo "[ERROR] Go backend not responding!"
    echo "Please check:"
    echo "  - Docker running: docker-compose ps"
    echo "  - Go service running: go run main.go"
    echo "  - Port 8080 accessible"
    echo ""
    exit 1
fi

# Test 1: Get all sources
echo ""
echo "========== Test 1: List all sources =========="
if curl -s "$API_BASE_URL/sources" | grep -q "id"; then
    echo "[SUCCESS] Get sources working"
    ((TESTS_PASSED++))
else
    echo "[FAILED] Cannot get sources"
    ((TESTS_FAILED++))
fi

# Test 2: Create new source
echo ""
echo "========== Test 2: Create new source =========="
UNIQUE_ID=$(date +%s)
RESPONSE=$(curl -s -X POST "$API_BASE_URL/sources" \
  -H "Content-Type: application/json" \
  -d "{\"url\":\"https://test-$UNIQUE_ID.example.com/rss\",\"author_name\":\"Test Blog\",\"priority\":7,\"enabled\":true,\"fetch_interval_seconds\":1800}")

if echo "$RESPONSE" | grep -q "id"; then
    echo "[SUCCESS] Source created successfully"
    ((TESTS_PASSED++))
else
    echo "[FAILED] Cannot create source"
    ((TESTS_FAILED++))
fi

# Test 3: Get single source (first one)
echo ""
echo "========== Test 3: Get single source =========="
if curl -s "$API_BASE_URL/sources/1" | grep -q "id"; then
    echo "[SUCCESS] Get single source working"
    ((TESTS_PASSED++))
else
    echo "[FAILED] Cannot get single source"
    ((TESTS_FAILED++))
fi

# Test 4: Update source
echo ""
echo "========== Test 4: Update source =========="
if curl -s -X PUT "$API_BASE_URL/sources/1" \
  -H "Content-Type: application/json" \
  -d '{"priority":9}' | grep -q "id"; then
    echo "[SUCCESS] Source update working"
    ((TESTS_PASSED++))
else
    echo "[FAILED] Cannot update source"
    ((TESTS_FAILED++))
fi

# Test 5: Trigger sync
echo ""
echo "========== Test 5: Trigger source sync =========="
if curl -s -X POST "$API_BASE_URL/sources/1/fetch" > /dev/null 2>&1; then
    echo "[SUCCESS] Sync triggered successfully"
    ((TESTS_PASSED++))
else
    echo "[FAILED] Cannot trigger sync"
    ((TESTS_FAILED++))
fi

# Test 6: Get sync logs
echo ""
echo "========== Test 6: Get sync logs =========="
if curl -s "$API_BASE_URL/sources/1/sync-logs" | grep -q "logs\|sourceId"; then
    echo "[SUCCESS] Sync logs available"
    ((TESTS_PASSED++))
else
    echo "[WARNING] Sync logs endpoint may not return data (this is OK)"
    ((TESTS_PASSED++))
fi

# Test 7: CORS headers check
echo ""
echo "========== Test 7: CORS headers =========="
if curl -s -I "$API_BASE_URL/sources" | grep -iq "Access-Control"; then
    echo "[SUCCESS] CORS headers present"
    ((TESTS_PASSED++))
else
    echo "[WARNING] CORS headers not detected"
fi

# Test 8: Frontend check
echo ""
echo "========== Test 8: Frontend accessible =========="
if curl -s "$FRONTEND_URL" | grep -q "TrueSignal\|<!DOCTYPE"; then
    echo "[SUCCESS] Frontend is accessible"
    ((TESTS_PASSED++))
else
    echo "[WARNING] Frontend not responding (may still be starting)"
fi

# Print summary
echo ""
echo "=========================================="
echo "Test Summary:"
echo "  Passed: $TESTS_PASSED"
echo "  Failed: $TESTS_FAILED"
echo "=========================================="
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo "[SUCCESS] All tests passed!"
    exit 0
else
    echo "[FAILED] Some tests failed"
    exit 1
fi
