@echo off
REM JunkFilter Desktop App Startup Script (Windows)
REM All backends run in Docker, only the Tauri desktop app runs natively

setlocal enabledelayedexpansion

REM Save project root path
set "PROJECT_ROOT=%~dp0"
set "PROJECT_ROOT=%PROJECT_ROOT:~0,-1%"

echo.
echo ========================================
echo    JunkFilter Desktop App Startup
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

REM Check Rust/Cargo
set "PATH=%USERPROFILE%\.cargo\bin;%PATH%"
where cargo >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Rust/Cargo not found. Please install Rust: https://rustup.rs
    pause
    exit /b 1
)

REM Check MSVC Build Tools
set "VCVARS=D:\visual studio\VC\Auxiliary\Build\vcvars64.bat"
if not exist "!VCVARS!" (
    echo [ERROR] Visual Studio C++ Build Tools not found
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

REM Clean up
echo ========== Cleaning Up ==========
echo.

taskkill /F /IM "junk-filter.exe" >nul 2>&1
taskkill /FI "WINDOWTITLE eq JF-Desktop*" >nul 2>&1

for /f "tokens=5" %%p in ('netstat -ano ^| findstr ":5173 " ^| findstr "LISTENING"') do (
    taskkill /F /PID %%p >nul 2>&1
)

timeout /t 2 /nobreak >nul

REM Phase 1: Start all backends in Docker
echo.
echo ========== Phase 1: Starting All Backends (Docker) ==========
echo.

echo [INFO] Building and starting containers...
docker-compose -f "%PROJECT_ROOT%\docker-compose.yml" up -d

if errorlevel 1 (
    echo [ERROR] Docker startup failed
    exit /b 1
)

echo [OK] All backend services started in Docker
echo.

REM Wait for services to be healthy
echo [INFO] Waiting for services to be ready...
timeout /t 8 /nobreak >nul

REM Quick health check
docker exec junkfilter-db pg_isready -U junkfilter >nul 2>&1
if errorlevel 1 (
    echo [WARN] PostgreSQL may not be ready yet, waiting...
    timeout /t 5 /nobreak >nul
)
echo [OK] PostgreSQL ready
echo [OK] Redis ready

timeout /t 3 /nobreak >nul
echo [OK] Go backend ready (Port 8080)
echo [OK] Python API ready (Port 8083)
echo [OK] Python evaluator ready
echo.

REM Phase 2: Start Tauri Desktop App
echo ========== Phase 2: Starting Desktop App ==========
echo.

if not exist "%PROJECT_ROOT%\frontend-vue\node_modules" (
    echo [INFO] Installing npm dependencies...
    cd /d "%PROJECT_ROOT%\frontend-vue"
    call npm install
    cd /d "%PROJECT_ROOT%"
)

echo [INFO] Launching Tauri desktop app...
echo     First launch may take 3-5 min for Rust compilation...
echo.
start "JF-Desktop App" cmd /k "call "!VCVARS!" >nul 2>&1 && set "PATH=%USERPROFILE%\.cargo\bin;%PATH%" && cd /d "%PROJECT_ROOT%\frontend-vue" && npm run tauri:dev"

timeout /t 5 /nobreak >nul

REM Summary
echo.
echo ========================================
echo          JunkFilter Running
echo ========================================
echo.
echo   [Docker]  PostgreSQL      :5432
echo   [Docker]  Redis           :6379
echo   [Docker]  Go Backend      :8080
echo   [Docker]  Python API      :8083
echo   [Docker]  Python Evaluator (background)
echo   [Native]  Tauri Desktop App
echo.
echo   To stop backends: docker-compose down
echo   To view logs:     docker-compose logs -f
echo.

pause
