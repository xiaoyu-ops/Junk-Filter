@echo off
REM JunkFilter Complete Startup Script (Windows)
REM Start in correct order: Docker - Go Backend - Python Backend - Frontend

setlocal enabledelayedexpansion

echo.
echo ========================================
echo    JunkFilter All Services Startup
echo    System: Windows
echo ========================================
echo.

REM Check if in project root directory
if not exist "docker-compose.yml" (
    echo [ERROR] Please run this script from JunkFilter project root directory
    exit /b 1
)

REM Check if .env file exists
if not exist ".env" (
    echo [ERROR] .env file not found in project root
    echo Please create .env file by copying .env.example
    echo   copy .env.example .env
    pause
    exit /b 1
)

echo [OK] Environment check passed
echo.

REM Kill Old Processes
echo.
echo ========== Cleaning Up Old Processes ==========
echo.

taskkill /F /IM "junkfilter-go.exe" >nul 2>&1
taskkill /F /IM "python.exe" >nul 2>&1
taskkill /F /IM "node.exe" >nul 2>&1

timeout /t 2 /nobreak >nul

REM Phase 1: Start Docker
echo.
echo ========== Phase 1: Starting Docker Containers ==========
echo.

docker --version >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Docker is not installed or not in PATH
    exit /b 1
)

echo [INFO] Stopping old containers...
docker-compose down 2>nul

echo [INFO] Starting Docker containers...
docker-compose up -d

if errorlevel 1 (
    echo [ERROR] Docker startup failed
    exit /b 1
)

echo [SUCCESS] Docker containers started
timeout /t 3 /nobreak >nul

REM Phase 2: Start Go Backend
echo.
echo ========== Phase 2: Starting Go Backend (Port 8080) ==========
echo.

if not exist "backend-go\main.go" (
    echo [ERROR] backend-go\main.go not found
    exit /b 1
)

echo [INFO] Building Go backend...
cd backend-go
del junkfilter-go.exe >nul 2>&1
go build -o junkfilter-go.exe main.go
if errorlevel 1 (
    echo [ERROR] Go build failed
    cd ..
    exit /b 1
)

echo [INFO] Starting Go backend...
start "JunkFilter Go Backend" cmd /k "junkfilter-go.exe"
cd ..

echo [SUCCESS] Go backend started (Port 8080)
timeout /t 3 /nobreak >nul

REM Phase 3: Python Backend
if exist "backend-python\main.py" (
    echo.
    echo ========== Phase 3: Starting Python Backend (Port 8081) ==========
    echo.
    cd backend-python
    start "JunkFilter Python Backend" cmd /k "conda activate junkfilter && python main.py"
    cd ..
    echo [INFO] Python backend started in new window
)

REM Phase 4: Frontend
echo.
echo ========== Phase 4: Starting Vue Frontend (Port 5173) ==========
echo.

if not exist "frontend-vue\package.json" (
    echo [ERROR] frontend-vue\package.json not found
    exit /b 1
)

cd frontend-vue
if not exist "node_modules" (
    echo [INFO] Installing npm dependencies...
    call npm install
)

start "JunkFilter Vue Frontend" cmd /k "npm run dev"
cd ..

timeout /t 10 /nobreak >nul

REM Summary
echo.
echo ========================================
echo            Startup Complete
echo ========================================
echo.
echo Services:
echo   * Docker:   http://localhost:5432 (PostgreSQL), 6379 (Redis)
echo   * Go:       http://localhost:8080
echo   * Python:   http://localhost:8081
echo   * Frontend: http://localhost:5173
echo.
echo Open browser: http://localhost:5173
echo.

pause
