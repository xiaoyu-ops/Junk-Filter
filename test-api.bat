@echo off
REM Test API Script for JunkFilter Backend
REM This script tests all main API endpoints

setlocal enabledelayedexpansion
chcp 65001 > nul

set "BASE_URL=http://localhost:8080"
set "FAILED=0"
set "PASSED=0"

echo.
echo ====================================
echo   JunkFilter API Test Suite
echo ====================================
echo.

REM 1. Health Check
echo.
echo [1/7] Testing Health Check...
echo ========================================
curl -s -w "\nStatus: %%{http_code}\n" %BASE_URL%/health
if !errorlevel! equ 0 (
    echo PASSED: Health check
    set /a PASSED+=1
) else (
    echo FAILED: Health check
    set /a FAILED+=1
)

REM 2. Get All Sources
echo.
echo [2/7] Testing Get All Sources...
echo ========================================
curl -s -w "\nStatus: %%{http_code}\n" %BASE_URL%/api/sources
if !errorlevel! equ 0 (
    echo PASSED: Get sources
    set /a PASSED+=1
) else (
    echo FAILED: Get sources
    set /a FAILED+=1
)

REM 3. Search Sources
echo.
echo [3/7] Testing Search Sources...
echo ========================================
curl -s -w "\nStatus: %%{http_code}\n" "%BASE_URL%/api/sources/search?q=blog"
if !errorlevel! equ 0 (
    echo PASSED: Search sources
    set /a PASSED+=1
) else (
    echo FAILED: Search sources
    set /a FAILED+=1
)

REM 4. Create New Source
echo.
echo [4/7] Testing Create Source...
echo ========================================
curl -s -X POST %BASE_URL%/api/sources ^
  -H "Content-Type: application/json" ^
  -d "{\"url\":\"https://example.com/feed\",\"title\":\"Example Blog\",\"priority\":5}" ^
  -w "\nStatus: %%{http_code}\n"
if !errorlevel! equ 0 (
    echo PASSED: Create source
    set /a PASSED+=1
) else (
    echo FAILED: Create source
    set /a FAILED+=1
)

REM 5. Content Stats
echo.
echo [5/7] Testing Content Stats...
echo ========================================
curl -s -w "\nStatus: %%{http_code}\n" %BASE_URL%/api/content/stats
if !errorlevel! equ 0 (
    echo PASSED: Content stats
    set /a PASSED+=1
) else (
    echo FAILED: Content stats
    set /a FAILED+=1
)

REM 6. Get All Content
echo.
echo [6/7] Testing Get All Content...
echo ========================================
curl -s -w "\nStatus: %%{http_code}\n" %BASE_URL%/api/content
if !errorlevel! equ 0 (
    echo PASSED: Get content
    set /a PASSED+=1
) else (
    echo FAILED: Get content
    set /a FAILED+=1
)

REM 7. Get High Scores
echo.
echo [7/7] Testing High Scores...
echo ========================================
curl -s -w "\nStatus: %%{http_code}\n" %BASE_URL%/api/evaluations/high-scores
if !errorlevel! equ 0 (
    echo PASSED: High scores
    set /a PASSED+=1
) else (
    echo FAILED: High scores
    set /a FAILED+=1
)

REM Summary
echo.
echo ====================================
echo   Test Summary
echo ====================================
echo Passed: %PASSED%/7
echo Failed: %FAILED%/7
echo ====================================
echo.

if %FAILED% equ 0 (
    echo All tests PASSED!
    exit /b 0
) else (
    echo Some tests FAILED.
    exit /b 1
)
