#!/bin/bash
# Junk Filter — 一键关闭所有服务

set -e

echo "🛑 Stopping Junk Filter services..."

# 1. 停止前端 (Node.js on port 5173)
FRONTEND_PID=$(lsof -ti:5173 2>/dev/null || true)
if [ -n "$FRONTEND_PID" ]; then
    echo "  🌐 Frontend (PID: $FRONTEND_PID)"
    kill $FRONTEND_PID 2>/dev/null || true
else
    echo "  🌐 Frontend (not running)"
fi

# 2. 停止 Go 后端 (port 8080)
GO_PID=$(lsof -ti:8080 2>/dev/null || true)
if [ -n "$GO_PID" ]; then
    echo "  🐹 Go Backend (PID: $GO_PID)"
    kill $GO_PID 2>/dev/null || true
else
    echo "  🐹 Go Backend (not running)"
fi

# 3. 停止 Python API (api_server.py)
API_PID=$(pgrep -f "python api_server.py" || true)
if [ -n "$API_PID" ]; then
    echo "  🐍 Python API (PID: $API_PID)"
    kill $API_PID 2>/dev/null || true
else
    echo "  🐍 Python API (not running)"
fi

# 4. 停止 Python Consumer (main.py)
CONSUMER_PID=$(pgrep -f "python main.py" || true)
if [ -n "$CONSUMER_PID" ]; then
    echo "  🐍 Consumer (PID: $CONSUMER_PID)"
    kill $CONSUMER_PID 2>/dev/null || true
else
    echo "  🐍 Consumer (not running)"
fi

# 5. 停止 Telegram Bot (LaunchAgent)
echo "  🤖 Telegram Bot (LaunchAgent)"
launchctl unload ~/Library/LaunchAgents/com.junkfilter.telegrambot.plist 2>/dev/null || true

# 6. 停止 Docker 基础设施（可选，默认不停）
# docker-compose down

sleep 1

echo ""
echo "✅ All services stopped."
