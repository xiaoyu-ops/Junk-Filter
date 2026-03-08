@echo off
REM JunkFilter Complete Startup Script (Windows)
REM Starts all 5 services: Docker -> Go Backend -> Python API -> Python Consumer -> Frontend

setlocal enabledelayedexpansion

REM Save project root path
set "PROJECT_ROOT=%~dp0"
set "PROJECT_ROOT=%PROJECT_ROOT:~0,-1%"

echo.
echo ========================================
echo    JunkFilter All Services Startup
echo ========================================
echo.

REM Check if in project root directory
if not exist "%PROJECT_ROOT%\docker-compose.yml" (
    echo [ERROR] docker-compose.yml not found in %PROJECT_ROOT%
    exit /b 1
)

if not exist "%PROJECT_ROOT%\.env" (
    echo [ERROR] .env file not found in project root
    echo Please create .env file first
    pause
    exit /b 1
)

echo [OK] Environment check passed
echo.

REM Initialize Conda (required for cmd.exe sessions without Conda in PATH)
set "CONDA_ROOT=C:\Users\XIAOYU\miniconda3"
if exist "%CONDA_ROOT%\condabin\conda.bat" (
    call "%CONDA_ROOT%\condabin\conda.bat" activate base >nul 2>&1
    echo [OK] Conda initialized from %CONDA_ROOT%
) else (
    echo [WARN] Conda not found at %CONDA_ROOT%, Python services may fail
)
echo.

REM Clean up old processes
echo ========== Cleaning Up Old Processes ==========
echo.

taskkill /F /IM "junkfilter-go.exe" >nul 2>&1
REM Kill services by window title (exclude current window by using specific prefixes)
taskkill /FI "WINDOWTITLE eq JF-Go*" >nul 2>&1
taskkill /FI "WINDOWTITLE eq JF-PyAPI*" >nul 2>&1
taskkill /FI "WINDOWTITLE eq JF-PyConsumer*" >nul 2>&1
taskkill /FI "WINDOWTITLE eq JF-Frontend*" >nul 2>&1

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
docker-compose -f "%PROJECT_ROOT%\docker-compose.yml" up -d

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

if not exist "%PROJECT_ROOT%\backend-go\main.go" (
    echo [ERROR] backend-go\main.go not found
    exit /b 1
)

echo [INFO] Building Go backend...
cd /d "%PROJECT_ROOT%\backend-go"
del junkfilter-go.exe >nul 2>&1
go build -o junkfilter-go.exe .
if errorlevel 1 (
    echo [ERROR] Go build failed
    cd /d "%PROJECT_ROOT%"
    exit /b 1
)

echo [INFO] Starting Go backend...
start "JF-Go Backend" cmd /k "cd /d "%PROJECT_ROOT%\backend-go" && set PYTHON_API_URL=http://localhost:8083 && junkfilter-go.exe"
cd /d "%PROJECT_ROOT%"

echo [OK] Go backend started (Port 8080)
timeout /t 3 /nobreak >nul

REM Phase 3: Python API Server
echo.
echo ========== Phase 3: Starting Python API Server (Port 8083) ==========
echo.

if not exist "%PROJECT_ROOT%\backend-python\api_server.py" (
    echo [WARN] backend-python\api_server.py not found, skipping
    goto phase4
)

start "JF-PyAPI Server" cmd /k "call "%CONDA_ROOT%\condabin\conda.bat" activate junkfilter && cd /d "%PROJECT_ROOT%\backend-python" && python api_server.py"

echo [OK] Python API server started (Port 8083)
timeout /t 2 /nobreak >nul

:phase4
REM Phase 4: Python Stream Consumer
echo.
echo ========== Phase 4: Starting Python Stream Consumer ==========
echo.

if not exist "%PROJECT_ROOT%\backend-python\main.py" (
    echo [WARN] backend-python\main.py not found, skipping
    goto phase5
)

start "JF-PyConsumer" cmd /k "call "%CONDA_ROOT%\condabin\conda.bat" activate junkfilter && cd /d "%PROJECT_ROOT%\backend-python" && python main.py"

echo [OK] Python stream consumer started
timeout /t 2 /nobreak >nul

:phase5
REM Phase 5: Frontend
echo.
echo ========== Phase 5: Starting Vue Frontend (Port 5173) ==========
echo.

if not exist "%PROJECT_ROOT%\frontend-vue\package.json" (
    echo [ERROR] frontend-vue\package.json not found
    exit /b 1
)

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
