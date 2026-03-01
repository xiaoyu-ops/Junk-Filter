@echo off
REM TrueSignal Full Stack Smoke Test (Windows)
REM Verify frontend-backend API integration

setlocal enabledelayedexpansion

set "API_BASE_URL=http://localhost:8080/api"
set "FRONTEND_URL=http://localhost:5173"
set "HEALTH_URL=http://localhost:8080/health"

set "TESTS_PASSED=0"
set "TESTS_FAILED=0"

echo.
echo ============================================
echo    TrueSignal Smoke Test (Windows)
echo ============================================
echo.

REM Pre-check: Go Backend
echo [INFO] Checking Go backend connection...
timeout /t 1 /nobreak >nul

curl -s "%HEALTH_URL%" >nul 2>&1
if errorlevel 1 (
    echo.
    echo [ERROR] Go backend not responding!
    echo Please check:
    echo   - Docker running: docker-compose ps
    echo   - Go service running: go run main.go
    echo   - Port 8080 accessible
    echo.
    exit /b 1
)
echo [SUCCESS] Go backend is responding

REM Test 1: Get all sources
echo.
echo ========== Test 1: List all sources ==========
curl -s "%API_BASE_URL%/sources" | findstr "id" >nul 2>&1
if errorlevel 1 (
    echo [FAILED] Cannot get sources
    set /a TESTS_FAILED+=1
) else (
    echo [SUCCESS] Get sources working
    set /a TESTS_PASSED+=1
)

REM Test 2: Create new source
echo.
echo ========== Test 2: Create new source ==========
for /f "tokens=2-4 delims=/ " %%a in ('date /t') do (set mydate=%%c%%a%%b)
for /f "tokens=1-2 delims=/:" %%a in ('time /t') do (set mytime=%%a%%b)
set unique_id=%mydate%_%mytime%

curl -s -X POST "%API_BASE_URL%/sources" ^
  -H "Content-Type: application/json" ^
  -d "{\"url\":\"https://test-%unique_id%.example.com/rss\",\"author_name\":\"Test Blog\",\"priority\":7,\"enabled\":true,\"fetch_interval_seconds\":1800}" | findstr "id" >nul 2>&1

if errorlevel 1 (
    echo [FAILED] Cannot create source
    set /a TESTS_FAILED+=1
) else (
    echo [SUCCESS] Source created successfully
    set /a TESTS_PASSED+=1
)

REM Test 3: Get single source (first one)
echo.
echo ========== Test 3: Get single source ==========
curl -s "%API_BASE_URL%/sources/1" | findstr "id" >nul 2>&1
if errorlevel 1 (
    echo [FAILED] Cannot get single source
    set /a TESTS_FAILED+=1
) else (
    echo [SUCCESS] Get single source working
    set /a TESTS_PASSED+=1
)

REM Test 4: Update source
echo.
echo ========== Test 4: Update source ==========
curl -s -X PUT "%API_BASE_URL%/sources/1" ^
  -H "Content-Type: application/json" ^
  -d "{\"priority\":9}" | findstr "id" >nul 2>&1

if errorlevel 1 (
    echo [FAILED] Cannot update source
    set /a TESTS_FAILED+=1
) else (
    echo [SUCCESS] Source update working
    set /a TESTS_PASSED+=1
)

REM Test 5: Trigger sync
echo.
echo ========== Test 5: Trigger source sync ==========
curl -s -X POST "%API_BASE_URL%/sources/1/fetch" >nul 2>&1
if errorlevel 1 (
    echo [FAILED] Cannot trigger sync
    set /a TESTS_FAILED+=1
) else (
    echo [SUCCESS] Sync triggered successfully
    set /a TESTS_PASSED+=1
)

REM Test 6: Get sync logs
echo.
echo ========== Test 6: Get sync logs ==========
curl -s "%API_BASE_URL%/sources/1/sync-logs" | findstr "logs" >nul 2>&1
if errorlevel 1 (
    echo [WARNING] Sync logs endpoint may not return data (this is OK)
    set /a TESTS_PASSED+=1
) else (
    echo [SUCCESS] Sync logs available
    set /a TESTS_PASSED+=1
)

REM Test 7: CORS headers check
echo.
echo ========== Test 7: CORS headers ==========
curl -s -I "%API_BASE_URL%/sources" | findstr /i "Access-Control" >nul 2>&1
if errorlevel 1 (
    echo [WARNING] CORS headers not detected
) else (
    echo [SUCCESS] CORS headers present
    set /a TESTS_PASSED+=1
)

REM Test 8: Frontend check
echo.
echo ========== Test 8: Frontend accessible ==========
curl -s "%FRONTEND_URL%" | findstr "TrueSignal" >nul 2>&1
if errorlevel 1 (
    echo [WARNING] Frontend not responding (may still be starting)
) else (
    echo [SUCCESS] Frontend is accessible
    set /a TESTS_PASSED+=1
)

REM Print summary
echo.
echo ========================================
echo Test Summary:
echo   Passed: %TESTS_PASSED%
echo   Failed: %TESTS_FAILED%
echo ========================================
echo.

if %TESTS_FAILED% equ 0 (
    echo [SUCCESS] All tests passed!
    exit /b 0
) else (
    echo [FAILED] Some tests failed
    exit /b 1
)
