#!/bin/bash
# TrueSignal 适配层冒烟测试 - 环境检查脚本
# 自动验证所有后端服务是否正常运行

echo "🧪 TrueSignal 适配层冒烟测试 - 环境检查"
echo "=========================================="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查函数
check_service() {
  local service=$1
  local url=$2
  local port=$3

  echo -n "检查 $service ($port)... "

  if curl -s -f "$url" > /dev/null 2>&1; then
    echo -e "${GREEN}✅ 正常${NC}"
    return 0
  else
    echo -e "${RED}❌ 无法连接${NC}"
    return 1
  fi
}

# 检查各个服务
echo "📊 服务状态检查"
echo "---"

services_ok=true

# Go 后端
check_service "Go 后端" "http://localhost:8080/health" "8080" || services_ok=false

# Mock 后端
check_service "Mock 后端" "http://localhost:3000/api/tasks" "3000" || services_ok=false

# 前端
check_service "前端应用" "http://localhost:5173/" "5173" || services_ok=false

# PostgreSQL
echo -n "检查 PostgreSQL (5432)... "
if docker exec -it truesignal-db psql -U truesignal -d truesignal -c "SELECT 1" > /dev/null 2>&1; then
  echo -e "${GREEN}✅ 正常${NC}"
else
  echo -e "${RED}❌ 无法连接${NC}"
  services_ok=false
fi

# Redis
echo -n "检查 Redis (6379)... "
if docker exec -it truesignal-redis redis-cli ping > /dev/null 2>&1; then
  echo -e "${GREEN}✅ 正常${NC}"
else
  echo -e "${RED}❌ 无法连接${NC}"
  services_ok=false
fi

echo ""
echo "📋 环境变量检查"
echo "---"

# 检查 .env.local
ENV_FILE="frontend-vue/.env.local"
if [ -f "$ENV_FILE" ]; then
  echo -e "${GREEN}✅ .env.local 存在${NC}"

  api_url=$(grep "VITE_API_URL" "$ENV_FILE" | cut -d '=' -f2)
  mock_url=$(grep "VITE_MOCK_URL" "$ENV_FILE" | cut -d '=' -f2)

  echo "   VITE_API_URL=$api_url"
  echo "   VITE_MOCK_URL=$mock_url"

  if [ "$api_url" = "http://localhost:8080" ] && [ "$mock_url" = "http://localhost:3000" ]; then
    echo -e "   ${GREEN}✅ 配置正确${NC}"
  else
    echo -e "   ${YELLOW}⚠️  配置可能需要调整${NC}"
  fi
else
  echo -e "${RED}❌ .env.local 不存在${NC}"
  services_ok=false
fi

echo ""
echo "📊 数据库数据检查"
echo "---"

# 检查源表
source_count=$(docker exec -it truesignal-db psql -U truesignal -d truesignal -t -c "SELECT COUNT(*) FROM sources" 2>/dev/null | tr -d ' ')
if [ ! -z "$source_count" ] && [ "$source_count" -gt 0 ]; then
  echo -e "✅ sources 表有 ${GREEN}$source_count${NC} 条记录"
else
  echo -e "${YELLOW}⚠️  sources 表为空或无法查询${NC}"
fi

echo ""
echo "=========================================="

if [ "$services_ok" = true ]; then
  echo -e "${GREEN}✅ 所有服务都已就绪，可以开始冒烟测试！${NC}"
  echo ""
  echo "📖 请参考以下文档进行测试："
  echo "   1. description/SMOKE_TEST_COMPLETE_GUIDE.md"
  echo ""
  echo "🚀 开始测试步骤："
  echo "   1. 打开浏览器访问 http://localhost:5173"
  echo "   2. 按 F12 打开开发工具"
  echo "   3. 按照测试指南逐步操作"
  exit 0
else
  echo -e "${RED}❌ 有服务未正常运行，请检查以下内容：${NC}"
  echo ""
  echo "排查步骤："
  echo "1. 确认 Docker 容器已启动："
  echo "   docker-compose ps"
  echo ""
  echo "2. 确认 Go 后端已启动（在新终端中）："
  echo "   cd backend-go && go run main.go"
  echo ""
  echo "3. 确认 Mock 后端已启动（在新终端中）："
  echo "   cd backend-mock && node server.js"
  echo ""
  echo "4. 确认前端已启动（在新终端中）："
  echo "   cd frontend-vue && npm run dev"
  exit 1
fi
