@echo off
REM TrueSignal Startup Script (Windows Batch)
REM Start all services: Docker + Go Backend + Python Backend + Vue Frontend

setlocal enabledelayedexpansion
set SCRIPT_DIR=%~dp0
cd /d "%SCRIPT_DIR%"

cls
echo.
echo ========================================
echo TrueSignal Startup Script
echo ========================================
echo.

REM Check if Docker is running
echo [1/5] Checking Docker...
docker --version >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Docker is not installed or not in PATH
    echo Please install Docker Desktop first
    pause
    exit /b 1
)

echo [OK] Docker found
echo [!] Starting Docker containers...
docker-compose up -d 2>&1
if errorlevel 1 (
    echo [WARNING] Docker compose failed, but continuing...
) else (
    echo [OK] Docker containers started
    timeout /t 10 /nobreak
)

echo.
echo [2/5] Starting Go Backend (localhost:8080)...
start "Go Backend" cmd /k "cd /d %SCRIPT_DIR%backend-go && go run main.go"
timeout /t 3 /nobreak

echo.
echo [3/5] Starting Python Backend (localhost:8081)...
start "Python Backend" cmd /k "cd /d %SCRIPT_DIR%backend-python && conda activate junkfilter && uvicorn api_server:app --host 0.0.0.0 --port 8081"
timeout /t 3 /nobreak

echo.
echo [4/5] Starting Vue Frontend (localhost:5173)...
start "Vue Frontend" cmd /k "cd /d %SCRIPT_DIR%frontend-vue && npm run dev"
timeout /t 3 /nobreak

echo.
echo ========================================
echo [OK] All services started
echo ========================================
echo.
echo Service Addresses:
echo.
echo   Frontend:  http://localhost:5173
echo   Go API:    http://localhost:8080/health
echo   Python:    http://localhost:8081/health
echo.
echo   Database:  localhost:5432 (truesignal/truesignal123)
echo   Redis:     localhost:6379
echo.
echo ========================================
echo.
echo LLM/Agent Setup (Optional):
echo To enable AI-powered Agent responses:
echo 1. Get OpenAI API Key: https://platform.openai.com/api-keys
echo 2. Edit .env and set: OPENAI_API_KEY=sk-proj-YOUR_KEY
echo 3. Restart Python Backend to load new API key
echo See description/guides/LLM_SETUP_GUIDE.md for details
echo.
echo ========================================
echo.
pause
