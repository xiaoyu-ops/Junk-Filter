@echo off
REM JunkFilter Smoke Test - Environment Check Script (Windows)
REM Automatically verifies all backend services are running

setlocal enabledelayedexpansion

echo.
echo ========== JunkFilter Smoke Test - Environment Check ==========
echo.

set "services_ok=true"

REM Check Go backend
echo Checking Go backend (8080)...
powershell -NoProfile -Command "(Invoke-WebRequest -Uri 'http://localhost:8080/health' -UseBasicParsing -TimeoutSec 2).StatusCode" >nul 2>&1
if %ERRORLEVEL% equ 0 (
  echo   [OK] Go backend is running
) else (
  echo   [ERROR] Cannot connect to Go backend
  set "services_ok=false"
)

REM Check Mock backend
echo Checking Mock backend (3000)...
powershell -NoProfile -Command "(Invoke-WebRequest -Uri 'http://localhost:3000/api/tasks' -UseBasicParsing -TimeoutSec 2).StatusCode" >nul 2>&1
if %ERRORLEVEL% equ 0 (
  echo   [OK] Mock backend is running
) else (
  echo   [ERROR] Cannot connect to Mock backend
  set "services_ok=false"
)

REM Check frontend
echo Checking frontend application (5173)...
powershell -NoProfile -Command "(Invoke-WebRequest -Uri 'http://localhost:5173/' -UseBasicParsing -TimeoutSec 2).StatusCode" >nul 2>&1
if %ERRORLEVEL% equ 0 (
  echo   [OK] Frontend is running
) else (
  echo   [ERROR] Cannot connect to frontend
  set "services_ok=false"
)

REM Check PostgreSQL
echo Checking PostgreSQL (5432)...
docker exec junkfilter-db psql -U junkfilter -d junkfilter -c "SELECT 1" >nul 2>&1
if %ERRORLEVEL% equ 0 (
  echo   [OK] PostgreSQL is running
) else (
  echo   [ERROR] Cannot connect to PostgreSQL
  set "services_ok=false"
)

REM Check Redis
echo Checking Redis (6379)...
docker exec junkfilter-redis redis-cli ping >nul 2>&1
if %ERRORLEVEL% equ 0 (
  echo   [OK] Redis is running
) else (
  echo   [ERROR] Cannot connect to Redis
  set "services_ok=false"
)

echo.
echo Environment Variables Check
echo ---

if exist "frontend-vue\.env.local" (
  echo   [OK] .env.local exists

  REM Read variables from .env.local
  for /f "tokens=2 delims==" %%a in ('findstr "VITE_API_URL" frontend-vue\.env.local') do set "api_url=%%a"
  for /f "tokens=2 delims==" %%a in ('findstr "VITE_MOCK_URL" frontend-vue\.env.local') do set "mock_url=%%a"

  echo   VITE_API_URL=!api_url!
  echo   VITE_MOCK_URL=!mock_url!
) else (
  echo   [ERROR] .env.local does not exist
  set "services_ok=false"
)

echo.
echo Database Data Check
echo ---

REM Check sources table data
for /f %%a in ('docker exec junkfilter-db psql -U junkfilter -d junkfilter -t -c "SELECT COUNT(*) FROM sources"') do set "source_count=%%a"

if "%source_count%" gtr "0" (
  echo   [OK] sources table has %source_count% records
) else (
  echo   [WARNING] sources table is empty or cannot query
)

echo.
echo ==================================================
echo.

if "%services_ok%"=="true" (
  echo [SUCCESS] All services are ready! Begin smoke testing.
  echo.
  echo Documentation:
  echo    description/SMOKE_TEST_QUICK_START.md
  echo.
  echo Testing Steps:
  echo    1. Open browser and go to http://localhost:5173
  echo    2. Press F12 to open developer tools
  echo    3. Follow the test guide
) else (
  echo [ERROR] Some services are not running. Check the following:
  echo.
  echo Troubleshooting Steps:
  echo 1. Verify Docker containers are running:
  echo    docker-compose ps
  echo.
  echo 2. Verify Go backend is running in new terminal:
  echo    cd backend-go
  echo    go run main.go
  echo.
  echo 3. Verify Mock backend is running in new terminal:
  echo    cd backend-mock
  echo    node server.js
  echo.
  echo 4. Verify frontend is running in new terminal:
  echo    cd frontend-vue
  echo    npm run dev
)

endlocal
