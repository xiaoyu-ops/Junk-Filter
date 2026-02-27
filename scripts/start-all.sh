#!/bin/bash

# TrueSignal 完整应用启动脚本
# 启动前端、Go 后端、Python 后端、数据库、Redis

echo "============================================"
echo "TrueSignal - 完整应用启动"
echo "============================================"
echo ""

# 检查 Docker
if ! command -v docker &> /dev/null; then
    echo "ERROR: Docker 未安装"
    echo "请从 https://www.docker.com/products/docker-desktop 下载安装"
    exit 1
fi

echo "✓ Docker 已安装"

# 进入项目根目录
cd "$(dirname "$0")"

# 启动 Docker 容器
echo ""
echo "正在启动 Docker 服务..."
docker-compose up -d

if [ $? -ne 0 ]; then
    echo "ERROR: Docker Compose 启动失败"
    exit 1
fi

echo "✓ Docker 服务已启动"
echo ""

# 等待数据库就绪
echo "等待数据库初始化..."
sleep 5

# 启动 Go 后端
echo ""
echo "正在启动 Go 后端服务 (http://localhost:8080)..."
cd backend-go
go run main.go &
GO_PID=$!
cd ..

# 启动 Python 后端
echo ""
echo "正在启动 Python 评估引擎..."
cd backend-python
python main.py &
PYTHON_PID=$!
cd ..

# 显示访问信息
echo ""
echo "============================================"
echo "所有服务已启动！"
echo "============================================"
echo ""
echo "访问地址："
echo "  • 前端应用:    http://localhost:5173"
echo "  • Go API:     http://localhost:8080"
echo "  • PostgreSQL: localhost:5432"
echo "  • Redis:      localhost:6379"
echo ""
echo "进程 ID:"
echo "  • Go 后端:     $GO_PID"
echo "  • Python 后端: $PYTHON_PID"
echo ""
echo "需要停止所有服务？运行："
echo "  kill $GO_PID $PYTHON_PID"
echo "  docker-compose down"
echo ""

# 打开浏览器（macOS）
if [[ "$OSTYPE" == "darwin"* ]]; then
    open http://localhost:5173
fi

# 等待进程
wait
