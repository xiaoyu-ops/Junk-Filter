@echo off
REM JunkFilter Local Dev Startup Script (Windows)
REM Docker: PostgreSQL + Redis only
REM Native: Go backend, Python API, Python evaluator, Vue frontend

setlocal enabledelayedexpansion

REM Save project root path
set "PROJECT_ROOT=%~dp0"
set "PROJECT_ROOT=%PROJECT_ROOT:~0,-1%"

echo.
echo ========================================
echo    JunkFilter Startup (Local Dev)
echo ========================================
echo.

REM Check prerequisites
if not exist "%PROJECT_ROOT%\docker-compose.yml" (
    echo [ERROR] docker-compose.yml not found
    exit /b 1
)

if not exist "%PROJECT_ROOT%\.env" (
    echo [ERROR] .env file not found
    pause
    exit /b 1
)

docker --version >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Docker is not installed
    exit /b 1
)

echo [OK] All prerequisites checked
echo.

REM Clean up old processes
echo ========== Cleaning Up ==========
echo.

taskkill /FI "WINDOWTITLE eq JF-*" >nul 2>&1

for /f "tokens=5" %%p in ('netstat -ano ^| findstr ":5173 " ^| findstr "LISTENING"') do (
    taskkill /F /PID %%p >nul 2>&1
)
for /f "tokens=5" %%p in ('netstat -ano ^| findstr ":8080 " ^| findstr "LISTENING"') do (
    taskkill /F /PID %%p >nul 2>&1
)
for /f "tokens=5" %%p in ('netstat -ano ^| findstr ":8083 " ^| findstr "LISTENING"') do (
    taskkill /F /PID %%p >nul 2>&1
)

timeout /t 2 /nobreak >nul

REM Phase 1: Start PostgreSQL + Redis in Docker
echo.
echo ========== Phase 1: Starting DB + Redis (Docker) ==========
echo.

echo [INFO] Starting PostgreSQL and Redis containers...
docker-compose -f "%PROJECT_ROOT%\docker-compose.yml" up -d postgres redis

if errorlevel 1 (
    echo [ERROR] Docker startup failed
    exit /b 1
)

echo [OK] PostgreSQL + Redis containers started
echo.

REM Wait for DB to be healthy
echo [INFO] Waiting for PostgreSQL to be ready...
:wait_pg
docker exec junkfilter-db pg_isready -U junkfilter >nul 2>&1
if errorlevel 1 (
    timeout /t 2 /nobreak >nul
    goto wait_pg
)
echo [OK] PostgreSQL ready (Port 5432)

docker exec junkfilter-redis redis-cli ping >nul 2>&1
echo [OK] Redis ready (Port 6379)
echo.

REM Phase 2: Start Go backend
echo ========== Phase 2: Starting Go Backend (Port 8080) ==========
echo.

start "JF-Go" cmd /k "cd /d "%PROJECT_ROOT%\backend-go" && go run main.go"
echo [OK] Go backend starting...
echo.

timeout /t 3 /nobreak >nul

REM Phase 3: Start Python API
echo ========== Phase 3: Starting Python API (Port 8083) ==========
echo.

start "JF-PyAPI" cmd /k "cd /d "%PROJECT_ROOT%\backend-python" && call %USERPROFILE%\miniconda3\condabin\conda.bat activate junkfilter && python api_server.py"
echo [OK] Python API starting...
echo.

timeout /t 3 /nobreak >nul

REM Phase 4: Start Python evaluator
echo ========== Phase 4: Starting Python Evaluator ==========
echo.

start "JF-PyEval" cmd /k "cd /d "%PROJECT_ROOT%\backend-python" && call %USERPROFILE%\miniconda3\condabin\conda.bat activate junkfilter && python main.py"
echo [OK] Python evaluator starting...
echo.

timeout /t 2 /nobreak >nul

REM Phase 5: Start Web Frontend
echo ========== Phase 5: Starting Web Frontend (Port 5173) ==========
echo.

if not exist "%PROJECT_ROOT%\frontend-vue\node_modules" (
    echo [INFO] Installing npm dependencies...
    cd /d "%PROJECT_ROOT%\frontend-vue"
    call npm install
    cd /d "%PROJECT_ROOT%"
)

start "JF-Frontend" cmd /k "cd /d "%PROJECT_ROOT%\frontend-vue" && npm run dev"

timeout /t 5 /nobreak >nul

REM Summary
echo.
echo ========================================
echo          JunkFilter Running
echo ========================================
echo.
echo   [Docker]  PostgreSQL      :5432
echo   [Docker]  Redis           :6379
echo   [Native]  Go Backend      :8080  (window: JF-Go)
echo   [Native]  Python API      :8083  (window: JF-PyAPI)
echo   [Native]  Python Evaluator       (window: JF-PyEval)
echo   [Native]  Web Frontend    http://localhost:5173  (window: JF-Frontend)
echo.
echo   Open browser: http://localhost:5173
echo.
echo   Each service runs in its own cmd window.
echo   Close a window to stop that service.
echo   To stop DB/Redis: docker-compose stop postgres redis
echo.

pause
