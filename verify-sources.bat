@echo off
chcp 65001 >nul
REM Verify sources hiding feature

echo.
echo ========================================
echo   Verify Default Sources Hidden
echo ========================================
echo.

REM Check 1: Database sources disabled
echo [Check 1] Database sources status...
docker exec junkfilter-db psql -U truesignal -d truesignal -c "SELECT id, author_name, enabled FROM sources ORDER BY id;"
echo.
echo Expected: 3 sources with enabled=false
echo.

REM Check 2: Docker containers
echo [Check 2] Docker containers status...
docker-compose ps
echo.

REM Check 3: Go backend connection
echo [Check 3] Try connecting to Go backend (http://localhost:8080)...
timeout /t 2 /nobreak >nul
curl -s -o nul -w "HTTP Status: %%{http_code}\n" http://localhost:8080/api/sources

if errorlevel 1 (
    echo.
    echo WARNING: Go backend not running
    echo Please start Go backend: cd backend-go ^&^& go run main.go
    echo.
) else (
    echo.
    echo OK: Go backend is running
    echo.
    echo [Check 4] API /api/sources response...
    curl -s http://localhost:8080/api/sources
    echo.
    echo.
    echo Expected: empty array [] or { "data": [] }
    echo.
)

REM Frontend check
echo [Check 5] Frontend verification (manual)
echo.
echo   1. Start frontend: cd frontend-vue ^&^& npm run dev
echo   2. Open browser: http://localhost:5173
echo   3. Verify: Right sidebar shows "No tasks"
echo   4. Create task: Click "Add Task" button
echo   5. Test delete: Select task, click red "Delete" button
echo.

echo ========================================
echo   Verification Complete!
echo ========================================
echo.
