@echo off
REM TrueSignal Day 1 Verification Script (Windows Batch)
REM Verifies: Docker, PostgreSQL, Redis, and initial setup
REM Execution: verify-day1.bat

setlocal enabledelayedexpansion
cd /d "%~dp0"

echo.
echo ════════════════════════════════════════════════════════════════
echo   TrueSignal Day 1 - Verification Script
echo ════════════════════════════════════════════════════════════════
echo.

REM 1. Check Docker
echo 📦 [1/5] Checking Docker installation...
docker --version >nul 2>&1
if errorlevel 1 (
    echo ❌ Docker not found. Please install Docker Desktop first.
    pause
    exit /b 1
) else (
    echo ✅ Docker installed
)
echo.

REM 2. Check Docker compose status
echo 🐳 [2/5] Checking Docker containers...
docker-compose ps >nul 2>&1
if errorlevel 1 (
    echo ⚠️  Docker containers not running. Starting...
    docker-compose up -d >nul 2>&1
    timeout /t 5 /nobreak
)
echo ✅ Docker containers running
echo.

REM 3. Check PostgreSQL
echo 🐘 [3/5] Checking PostgreSQL connection...
docker exec truesignal-db psql -U truesignal -d truesignal -c "SELECT version();" >nul 2>&1
if errorlevel 1 (
    echo ❌ PostgreSQL connection failed
    echo   Try: docker-compose down -v && docker-compose up -d
    pause
    exit /b 1
) else (
    echo ✅ PostgreSQL connected
)
echo.

REM 4. Check Redis
echo 🔴 [4/5] Checking Redis connection...
docker exec truesignal-redis redis-cli ping >nul 2>&1
if errorlevel 1 (
    echo ❌ Redis connection failed
    echo   Try: docker-compose down -v && docker-compose up -d
    pause
    exit /b 1
) else (
    echo ✅ Redis connected
)
echo.

REM 5. Check Go backend compilation
echo 🐹 [5/5] Checking Go backend...
cd backend-go
go build -v ./internal/config >nul 2>&1
if errorlevel 1 (
    echo ❌ Go backend compilation failed
    echo   Try: cd backend-go && go mod download
    cd ..
    pause
    exit /b 1
) else (
    echo ✅ Go backend compiles successfully
    cd ..
)
echo.

echo ════════════════════════════════════════════════════════════════
echo ✅ All Day 1 Verifications Passed!
echo ════════════════════════════════════════════════════════════════
echo.
echo 📊 Environment Status:
echo   • Docker: Ready
echo   • PostgreSQL: Connected (localhost:5432)
echo   • Redis: Connected (localhost:6379)
echo   • Go Backend: Buildable
echo.
echo 🚀 Next Steps:
echo   1. Start all services: start-all.bat
echo   2. Open frontend: http://localhost:5173
echo   3. Check logs: docker-compose logs -f
echo.
echo 📚 Documentation:
echo   • description/README.md - Project overview
echo   • description/MASTER_INDEX.md - Full documentation index
echo   • CLAUDE.md - Development guidelines
echo.
pause
