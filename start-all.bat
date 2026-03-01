@echo off
REM TrueSignal Complete Startup Script (Windows)
REM Start in correct order: Docker - Go Backend - Python Backend - Frontend

setlocal enabledelayedexpansion

echo.
echo ========================================
echo    TrueSignal All Services Startup
echo    System: Windows
echo ========================================
echo.

REM Check if in project root directory
if not exist "docker-compose.yml" (
    echo [ERROR] Please run this script from TrueSignal project root directory
    exit /b 1
)

REM ========== Kill Old Processes ==========
echo.
echo ========== Cleaning Up Old Processes ==========
echo.

echo [INFO] Terminating old processes (if any)...
taskkill /F /IM "truesignal-go.exe" >nul 2>&1
taskkill /F /IM "python.exe" >nul 2>&1
taskkill /F /IM "node.exe" >nul 2>&1

timeout /t 2 /nobreak >nul

REM ========== Phase 1: Start Docker ==========
echo.
echo ========== Phase 1: Starting Docker Containers ==========
echo.

echo [INFO] Checking Docker status...
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

REM ========== Phase 2: Start Go Backend ==========
echo.
echo ========== Phase 2: Starting Go Backend (Port 8080) ==========
echo.

if not exist "backend-go\main.go" (
    echo [ERROR] backend-go\main.go not found
    exit /b 1
)

echo [INFO] Building Go backend...
cd backend-go
del main_refactored.go >nul 2>&1
go build -o truesignal-go.exe main.go
if errorlevel 1 (
    echo [ERROR] Go build failed
    cd ..
    exit /b 1
)

echo [INFO] Starting Go backend...
start "TrueSignal Go Backend" cmd /k "truesignal-go.exe"

cd ..

echo [INFO] Waiting for Go backend to start...
set attempt=0
:wait_go_backend
timeout /t 1 /nobreak >nul
curl -s http://localhost:8080/api/content >nul 2>&1
if errorlevel 1 (
    set /a attempt+=1
    if !attempt! lss 30 (
        goto wait_go_backend
    ) else (
        echo [WARNING] Go backend health check timeout (continuing anyway)
    )
)

echo [SUCCESS] Go backend started (Port 8080)

REM ========== Phase 3: Optional Python Backend ==========
if exist "backend-python\main.py" (
    echo.
    echo ========== Phase 3: Starting Python Backend (Port 8081) ==========
    echo.

    echo [INFO] Starting Python backend with conda junkfilter environment...
    cd backend-python

    REM Start Python in new window using conda junkfilter environment
    start "TrueSignal Python Backend" cmd /k "conda activate junkfilter && python main.py"

    cd ..

    echo [INFO] Python backend started in new window
) else (
    echo.
    echo [INFO] backend-python\main.py not found, skipping Python backend
)

REM ========== Phase 4: Verify Services ==========
echo.
echo ========== Verifying Services ==========
echo.

echo [CHECK] PostgreSQL...
docker exec truesignal-db psql -U truesignal -d truesignal -c "SELECT 1" >nul 2>&1
if errorlevel 1 (
    echo [WARNING] PostgreSQL connection failed
) else (
    echo [SUCCESS] PostgreSQL is running
)

echo [CHECK] Redis...
docker exec truesignal-redis redis-cli ping >nul 2>&1
if errorlevel 1 (
    echo [WARNING] Redis connection failed
) else (
    echo [SUCCESS] Redis is running
)

echo [CHECK] Go Backend...
curl -s http://localhost:8080/api/content >nul 2>&1
if errorlevel 1 (
    echo [WARNING] Go backend is not responding yet
) else (
    echo [SUCCESS] Go backend is running (Port 8080)
)

REM ========== Phase 5: Start Frontend ==========
echo.
echo ========== Phase 5: Starting Vue Frontend (Port 5173) ==========
echo.

if not exist "frontend-vue\package.json" (
    echo [ERROR] frontend-vue\package.json not found
    exit /b 1
)

echo [INFO] Checking frontend dependencies...
cd frontend-vue

REM Check if node_modules exists
if not exist "node_modules" (
    echo [INFO] Installing npm dependencies...
    call npm install
    if errorlevel 1 (
        echo [ERROR] npm install failed
        cd ..
        exit /b 1
    )
)

REM Start frontend in new window
echo [INFO] Starting Vue development server...
start "TrueSignal Vue Frontend" cmd /k "npm run dev"

cd ..

echo [INFO] Waiting for frontend to start (10 seconds)...
timeout /t 10 /nobreak >nul

REM ========== Summary ==========
echo.
echo ========================================
echo            Startup Summary
echo ========================================
echo.

echo [COMPLETE] All services have started!
echo.
echo Services Running:
echo   * Docker:        PostgreSQL + Redis (Port 5432, 6379)
echo   * Go Backend:    http://localhost:8080/api (Port 8080)
echo   * Python:        http://localhost:8081 (Port 8081, optional)
echo   * Vue Frontend:  http://localhost:5173 (Port 5173)
echo.
echo Next Steps:
echo   1. Open browser: http://localhost:5173
echo   2. You should see the TrueSignal application
echo   3. Test API: Run smoke_test.bat in new terminal
echo.
echo Quick Commands:
echo   * Test API:     smoke_test.bat
echo   * View logs:    docker-compose logs -f
echo   * Stop all:     docker-compose down
echo.

REM Keep window open
pause
