@echo off
REM JunkFilter Complete Startup Script (Windows)
REM Starts all 5 services: Docker -> Go Backend -> Python API -> Python Consumer -> Frontend

setlocal enabledelayedexpansion

echo.
echo ========================================
echo    JunkFilter All Services Startup
echo ========================================
echo.

REM Check if in project root directory
if not exist "docker-compose.yml" (
    echo [ERROR] Please run this script from JunkFilter project root directory
    exit /b 1
)

if not exist ".env" (
    echo [ERROR] .env file not found in project root
    echo Please create .env file first
    pause
    exit /b 1
)

echo [OK] Environment check passed
echo.

REM Clean up old processes
echo ========== Cleaning Up Old Processes ==========
echo.

taskkill /F /IM "junkfilter-go.exe" >nul 2>&1
REM Only kill python/node started by us (by window title)
taskkill /FI "WINDOWTITLE eq JunkFilter*" >nul 2>&1

timeout /t 2 /nobreak >nul

REM Phase 1: Docker
echo.
echo ========== Phase 1: Starting Docker Containers ==========
echo.

docker --version >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Docker is not installed or not in PATH
    exit /b 1
)

echo [INFO] Starting Docker containers...
docker-compose up -d

if errorlevel 1 (
    echo [ERROR] Docker startup failed
    exit /b 1
)

echo [OK] Docker containers started
timeout /t 3 /nobreak >nul

REM Phase 2: Go Backend
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

echo [OK] Go backend started (Port 8080)
timeout /t 3 /nobreak >nul

REM Phase 3: Python API Server
echo.
echo ========== Phase 3: Starting Python API Server (Port 8083) ==========
echo.

if not exist "backend-python\api_server.py" (
    echo [WARN] backend-python\api_server.py not found, skipping
    goto phase4
)

cd backend-python
start "JunkFilter Python API" cmd /k "conda activate junkfilter && python api_server.py"
cd ..

echo [OK] Python API server started (Port 8083)
timeout /t 2 /nobreak >nul

:phase4
REM Phase 4: Python Stream Consumer
echo.
echo ========== Phase 4: Starting Python Stream Consumer ==========
echo.

if not exist "backend-python\main.py" (
    echo [WARN] backend-python\main.py not found, skipping
    goto phase5
)

cd backend-python
start "JunkFilter Python Consumer" cmd /k "conda activate junkfilter && python main.py"
cd ..

echo [OK] Python stream consumer started
timeout /t 2 /nobreak >nul

:phase5
REM Phase 5: Frontend
echo.
echo ========== Phase 5: Starting Vue Frontend (Port 5173) ==========
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

timeout /t 5 /nobreak >nul

REM Summary
echo.
echo ========================================
echo          All Services Started
echo ========================================
echo.
echo   Docker:          PostgreSQL :5432, Redis :6379
echo   Go Backend:      http://localhost:8080
echo   Python API:      http://localhost:8083
echo   Python Consumer: Redis Stream consumer
echo   Frontend:        http://localhost:5173
echo.
echo Open browser: http://localhost:5173
echo.

pause
