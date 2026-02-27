#!/bin/bash

# TrueSignal Day 1 验证脚本

echo "=========================================="
echo "TrueSignal Day 1 - 环境验证"
echo "=========================================="
echo ""

# 1. 检查 Docker
echo "1. 检查 Docker..."
if command -v docker &> /dev/null; then
    echo "   ✓ Docker 已安装"
else
    echo "   ✗ Docker 未安装"
    exit 1
fi

if command -v docker-compose &> /dev/null; then
    echo "   ✓ Docker Compose 已安装"
else
    echo "   ✗ Docker Compose 未安装"
    exit 1
fi

echo ""
echo "2. 启动 Docker 容器..."
docker-compose up -d
if [ $? -ne 0 ]; then
    echo "   ✗ 启动失败"
    exit 1
fi

echo "   等待容器就绪... (30 秒)"
sleep 30

echo ""
echo "3. 验证 PostgreSQL..."
docker exec truesignal-db pg_isready -U truesignal &> /dev/null
if [ $? -eq 0 ]; then
    echo "   ✓ PostgreSQL 就绪"
else
    echo "   ✗ PostgreSQL 未就绪"
    exit 1
fi

echo ""
echo "4. 验证 Redis..."
REDIS_PING=$(docker exec truesignal-redis redis-cli ping)
if [ "$REDIS_PING" = "PONG" ]; then
    echo "   ✓ Redis 就绪"
else
    echo "   ✗ Redis 未就绪"
    exit 1
fi

echo ""
echo "5. 检查数据库表..."
TABLE_COUNT=$(docker exec truesignal-db psql -U truesignal -d truesignal -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" | tr -d ' ')
echo "   ✓ 数据库表数量: $TABLE_COUNT"

echo ""
echo "6. 检查初始数据..."
SOURCE_COUNT=$(docker exec truesignal-db psql -U truesignal -d truesignal -t -c "SELECT COUNT(*) FROM sources;" | tr -d ' ')
echo "   ✓ RSS 源数量: $SOURCE_COUNT"

echo ""
echo "=========================================="
echo "✅ Day 1 环境验证完成！"
echo "=========================================="
echo ""
echo "Docker 容器状态："
docker-compose ps
echo ""
echo "下一步："
echo "1. 查看 PostgreSQL 初始数据:"
echo "   docker exec -it truesignal-db psql -U truesignal -d truesignal"
echo "   > SELECT * FROM sources;"
echo ""
echo "2. 测试 Redis 连接:"
echo "   docker exec -it truesignal-redis redis-cli"
echo "   > PING"
echo ""
echo "3. 本地运行 Go 应用:"
echo "   cd backend-go"
echo "   go mod download"
echo "   go run main.go"
echo ""
echo "4. 本地运行 Python 应用:"
echo "   cd backend-python"
echo "   pip install -r requirements.txt"
echo "   python main.py"
echo ""
