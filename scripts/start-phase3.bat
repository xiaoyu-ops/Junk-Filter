@echo off
REM TrueSignal Phase 3 启动脚本 - 同时启动 Mock 服务器和前端

echo.
echo ========================================================================
echo  TrueSignal Phase 3 - Mock 服务器 + 前端启动
echo ========================================================================
echo.

REM 检查 Node.js
node --version >nul 2>&1
if errorlevel 1 (
    echo ✗ 错误: 未找到 Node.js, 请先安装 Node.js
    pause
    exit /b 1
)

echo ✓ Node.js 版本:
node --version

echo.
echo 启动步骤:
echo.
echo 1. 启动 Mock 服务器 (端口 3000)...
start "TrueSignal Mock Server" cmd /k "cd backend-mock && node server.js"

echo ✓ Mock 服务器启动中 (3 秒后启动前端)
timeout /t 3 /nobreak

echo.
echo 2. 启动前端开发服务器 (端口 5173)...
start "TrueSignal Frontend" cmd /k "cd frontend-vue && npm run dev -- --host"

echo.
echo ========================================================================
echo  ✓ 两个服务都在启动中！
echo ========================================================================
echo.
echo 访问地址:
echo   前端: http://localhost:5173
echo   Mock 服务器: http://localhost:3000
echo.
echo 停止服务:
echo   - 关闭任意一个命令窗口即可停止对应服务
echo   - 或按 Ctrl+C 在命令窗口中停止
echo.
echo ========================================================================
echo.
pause
