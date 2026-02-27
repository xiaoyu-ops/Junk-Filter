#!/bin/bash

# TrueSignal Phase 3 启动脚本 - 同时启动 Mock 服务器和前端

echo ""
echo "========================================================================"
echo " TrueSignal Phase 3 - Mock 服务器 + 前端启动"
echo "========================================================================"
echo ""

# 检查 Node.js
if ! command -v node &> /dev/null; then
    echo "✗ 错误: 未找到 Node.js, 请先安装 Node.js"
    exit 1
fi

echo "✓ Node.js 版本:"
node --version

echo ""
echo "启动步骤:"
echo ""
echo "1. 启动 Mock 服务器 (端口 3000)..."
cd "$(dirname "$0")/backend-mock"
node server.js &
MOCK_PID=$!

echo "✓ Mock 服务器启动中 (PID: $MOCK_PID)"
sleep 3

echo ""
echo "2. 启动前端开发服务器 (端口 5173)..."
cd "$(dirname "$0")/frontend-vue"
npm run dev -- --host &
FRONTEND_PID=$!

echo "✓ 前端启动中 (PID: $FRONTEND_PID)"

echo ""
echo "========================================================================"
echo " ✓ 两个服务都在启动中！"
echo "========================================================================"
echo ""
echo "访问地址:"
echo "  前端: http://localhost:5173"
echo "  Mock 服务器: http://localhost:3000"
echo ""
echo "停止服务:"
echo "  按 Ctrl+C 停止所有服务"
echo ""
echo "========================================================================"
echo ""

# 等待 Ctrl+C 信号
trap "kill $MOCK_PID $FRONTEND_PID; echo ''; echo '✓ 所有服务已关闭'; exit 0" SIGINT

wait
