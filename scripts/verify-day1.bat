@echo off
REM TrueSignal Day 1 验证脚本 (Windows)

echo.
echo ==========================================
echo TrueSignal Day 1 - 环境验证
echo ==========================================
echo.

REM 1. 检查 Docker
echo 1. 检查 Docker...
docker --version >nul 2>&1
if errorlevel 1 (
    echo    x Docker 未安装
    exit /b 1
)
echo    √ Docker 已安装

docker-compose --version >nul 2>&1
if errorlevel 1 (
    echo    x Docker Compose 未安装
    exit /b 1
)
echo    √ Docker Compose 已安装

echo.
echo 2. 启动 Docker 容器...
docker-compose up -d
if errorlevel 1 (
    echo    x 启动失败
    exit /b 1
)

echo    等待容器就绪... (30 秒)
timeout /t 30 /nobreak

echo.
echo 3. 验证 PostgreSQL...
docker exec truesignal-db pg_isready -U truesignal >nul 2>&1
if errorlevel 1 (
    echo    x PostgreSQL 未就绪
    exit /b 1
)
echo    √ PostgreSQL 就绪

echo.
echo 4. 验证 Redis...
for /f "tokens=*" %%a in ('docker exec truesignal-redis redis-cli ping 2^>nul') do set REDIS_PING=%%a
if "%REDIS_PING%"=="PONG" (
    echo    √ Redis 就绪
) else (
    echo    x Redis 未就绪
    exit /b 1
)

echo.
echo 5. 检查数据库表...
for /f "tokens=*" %%a in ('docker exec truesignal-db psql -U truesignal -d truesignal -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" 2^>nul') do set TABLE_COUNT=%%a
echo    √ 数据库表数量: %TABLE_COUNT%

echo.
echo 6. 检查初始数据...
for /f "tokens=*" %%a in ('docker exec truesignal-db psql -U truesignal -d truesignal -t -c "SELECT COUNT(*) FROM sources;" 2^>nul') do set SOURCE_COUNT=%%a
echo    √ RSS 源数量: %SOURCE_COUNT%

echo.
echo ==========================================
echo √√ Day 1 环境验证完成！
echo ==========================================
echo.
echo Docker 容器状态：
docker-compose ps
echo.
echo 下一步：
echo 1. 查看 PostgreSQL 初始数据:
echo    docker exec -it truesignal-db psql -U truesignal -d truesignal
echo    ^> SELECT * FROM sources;
echo.
echo 2. 测试 Redis 连接:
echo    docker exec -it truesignal-redis redis-cli
echo    ^> PING
echo.
echo 3. 本地运行 Go 应用:
echo    cd backend-go
echo    go mod download
echo    go run main.go
echo.
echo 4. 本地运行 Python 应用:
echo    cd backend-python
echo    pip install -r requirements.txt
echo    python main.py
echo.
pause
