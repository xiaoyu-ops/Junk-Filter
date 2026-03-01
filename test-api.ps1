#!/usr/bin/env powershell
# Test API Script for JunkFilter Backend

$BaseURL = "http://localhost:8080"
$Passed = 0
$Failed = 0

Write-Host "`n===================================="
Write-Host "  JunkFilter API Test Suite"
Write-Host "====================================" -ForegroundColor Green
Write-Host ""

# Test 1: Health Check
Write-Host "[1/7] Testing Health Check..."
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
try {
    $response = Invoke-WebRequest -Uri "$BaseURL/health" -UseBasicParsing
    Write-Host $response.Content
    Write-Host "Status: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "✓ Health check PASSED" -ForegroundColor Green
    $Passed++
} catch {
    Write-Host "✗ Health check FAILED: $_" -ForegroundColor Red
    $Failed++
}

# Test 2: Get All Sources
Write-Host "`n[2/7] Testing Get All Sources..."
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
try {
    $response = Invoke-WebRequest -Uri "$BaseURL/api/sources" -UseBasicParsing
    Write-Host $response.Content
    Write-Host "Status: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "✓ Get sources PASSED" -ForegroundColor Green
    $Passed++
} catch {
    Write-Host "✗ Get sources FAILED: $_" -ForegroundColor Red
    $Failed++
}

# Test 3: Search Sources
Write-Host "`n[3/7] Testing Search Sources..."
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
try {
    $response = Invoke-WebRequest -Uri "$BaseURL/api/sources/search?q=blog" -UseBasicParsing
    Write-Host $response.Content
    Write-Host "Status: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "✓ Search sources PASSED" -ForegroundColor Green
    $Passed++
} catch {
    Write-Host "✗ Search sources FAILED: $_" -ForegroundColor Red
    $Failed++
}

# Test 4: Create New Source
Write-Host "`n[4/7] Testing Create Source..."
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
try {
    $body = @{
        url = "https://example.com/feed"
        title = "Example Blog"
        priority = 5
    } | ConvertTo-Json

    $response = Invoke-WebRequest -Uri "$BaseURL/api/sources" `
        -Method POST `
        -Headers @{"Content-Type" = "application/json"} `
        -Body $body `
        -UseBasicParsing

    Write-Host $response.Content
    Write-Host "Status: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "✓ Create source PASSED" -ForegroundColor Green
    $Passed++
} catch {
    Write-Host "✗ Create source FAILED: $_" -ForegroundColor Red
    $Failed++
}

# Test 5: Content Stats
Write-Host "`n[5/7] Testing Content Stats..."
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
try {
    $response = Invoke-WebRequest -Uri "$BaseURL/api/content/stats" -UseBasicParsing
    Write-Host $response.Content
    Write-Host "Status: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "✓ Content stats PASSED" -ForegroundColor Green
    $Passed++
} catch {
    Write-Host "✗ Content stats FAILED: $_" -ForegroundColor Red
    $Failed++
}

# Test 6: Get All Content
Write-Host "`n[6/7] Testing Get All Content..."
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
try {
    $response = Invoke-WebRequest -Uri "$BaseURL/api/content" -UseBasicParsing
    Write-Host $response.Content
    Write-Host "Status: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "✓ Get content PASSED" -ForegroundColor Green
    $Passed++
} catch {
    Write-Host "✗ Get content FAILED: $_" -ForegroundColor Red
    $Failed++
}

# Test 7: Get High Scores
Write-Host "`n[7/7] Testing High Scores..."
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
try {
    $response = Invoke-WebRequest -Uri "$BaseURL/api/evaluations/high-scores" -UseBasicParsing
    Write-Host $response.Content
    Write-Host "Status: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "✓ High scores PASSED" -ForegroundColor Green
    $Passed++
} catch {
    Write-Host "✗ High scores FAILED: $_" -ForegroundColor Red
    $Failed++
}

# Summary
Write-Host "`n===================================="
Write-Host "  Test Summary"
Write-Host "====================================" -ForegroundColor Green
Write-Host "✓ Passed: $Passed/7" -ForegroundColor Green
Write-Host "✗ Failed: $Failed/7" -ForegroundColor Red
Write-Host "===================================="
Write-Host ""

if ($Failed -eq 0) {
    Write-Host "All tests PASSED! ✓" -ForegroundColor Green
    exit 0
} else {
    Write-Host "Some tests FAILED. Check the output above." -ForegroundColor Red
    exit 1
}
