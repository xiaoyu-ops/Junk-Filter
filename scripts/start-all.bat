@echo off
REM JunkFilter - Start All Services
REM This script starts Docker, Go backend, Mock backend, and frontend

echo.
echo ============================================
echo JunkFilter - Application Startup
echo ============================================
echo.

REM Check if Docker is installed
docker --version >nul 2>&1
if errorlevel 1 (
    echo ERROR: Docker not installed or not in PATH
    echo Please download from https://www.docker.com/products/docker-desktop
    pause
    exit /b 1
)

echo OK - Docker is installed
echo.

REM Navigate to project root
cd /d "%~dp0"

REM Start Docker containers
echo Starting Docker services...
docker-compose up -d

if errorlevel 1 (
    echo ERROR: Docker Compose startup failed
    pause
    exit /b 1
)

echo OK - Docker services started
echo.
timeout /t 3 /nobreak

REM Set JunkFilter database credentials for Go backend
set DB_HOST=localhost
set DB_PORT=5432
set DB_USER=junkfilter
set DB_PASSWORD=junkfilter123
set DB_NAME=junkfilter

REM Start Go backend (new window)
echo Starting Go backend (http://localhost:8080)...
start "JunkFilter-Go" cmd /k "cd backend-go && go run main.go"

REM Start Mock backend (new window)
echo Starting Mock backend (http://localhost:3000)...
start "JunkFilter-Mock" cmd /k "cd backend-mock && node server.js"

REM Start frontend (new window)
echo Starting frontend (http://localhost:5173)...
start "JunkFilter-Frontend" cmd /k "cd frontend-vue && npm run dev"

echo.
echo ============================================
echo All services started!
echo ============================================
echo.
echo Access URLs:
echo   - Frontend:     http://localhost:5173
echo   - Go backend:   http://localhost:8080
echo   - Mock backend: http://localhost:3000
echo   - PostgreSQL:   localhost:5432
echo   - Redis:        localhost:6379
echo.
echo Note: New terminal windows opened. Please wait for services to start (about 10-15 seconds)
echo.
pause

