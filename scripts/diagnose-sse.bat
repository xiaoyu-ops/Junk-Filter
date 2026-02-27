@echo off
REM SSE 连接诊断脚本

echo.
echo ==========================================
echo SSE 连接诊断
echo ==========================================
echo.

REM 检查 Mock 服务器是否运行
echo 1️⃣ 检查 Mock 服务器连接...
curl -s -o nul -w "HTTP 状态码: %%{http_code}\n" http://localhost:3000/api/tasks
echo.

REM 测试 SSE 端点
echo 2️⃣ 测试 SSE 端点...
echo (应该看到流式数据，按 Ctrl+C 停止)
echo.
curl -N "http://localhost:3000/api/chat/stream?taskId=task-1&message=你好"
echo.

REM 检查前端是否运行
echo 3️⃣ 检查前端运行状态...
curl -s -o nul -w "HTTP 状态码: %%{http_code}\n" http://localhost:5173
echo.

echo ==========================================
echo 诊断完成
echo ==========================================
echo.

pause
