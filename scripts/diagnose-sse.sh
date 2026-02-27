#!/bin/bash
# SSE 连接诊断脚本

echo "=========================================="
echo "SSE 连接诊断"
echo "=========================================="
echo ""

# 检查 Mock 服务器是否运行
echo "1️⃣ 检查 Mock 服务器连接..."
curl -s -o /dev/null -w "HTTP 状态码: %{http_code}\n" http://localhost:3000/api/tasks
echo ""

# 测试 SSE 端点
echo "2️⃣ 测试 SSE 端点（5 秒超时）..."
timeout 5 curl -N \
  -H "Accept: text/event-stream" \
  "http://localhost:3000/api/chat/stream?taskId=task-1&message=你好" 2>/dev/null | head -20 || echo "⚠️ SSE 连接超时或失败"
echo ""

# 检查前端是否运行
echo "3️⃣ 检查前端运行状态..."
curl -s -o /dev/null -w "HTTP 状态码: %{http_code}\n" http://localhost:5173
echo ""

echo "=========================================="
echo "诊断完成"
echo "=========================================="
