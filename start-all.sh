#!/bin/bash
# TrueSignal 一键启动脚本 (Linux/Mac)
# 启动所有服务：Docker (PostgreSQL + Redis) + Go 后端 + Python 后端 + Vue 前端

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${CYAN}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[!]${NC} $1"
}

log_error() {
    echo -e "${RED}[✗]${NC} $1"
}

# 清屏
clear

echo ""
echo "========================================"
echo "TrueSignal 一键启动脚本"
echo "========================================"
echo ""

# 检查前置条件
log_info "检查前置环境..."

if ! command -v docker &> /dev/null; then
    log_error "Docker 未安装"
    exit 1
fi
log_success "Docker 已安装"

if ! command -v docker-compose &> /dev/null; then
    log_error "docker-compose 未安装"
    exit 1
fi
log_success "docker-compose 已安装"

if ! command -v go &> /dev/null; then
    log_error "Go 未安装"
    exit 1
fi
log_success "Go 已安装"

if ! command -v node &> /dev/null; then
    log_error "Node.js 未安装"
    exit 1
fi
log_success "Node.js 已安装"

# 1. Docker & Database
echo ""
log_info "启动 Docker 容器 (PostgreSQL + Redis)..."

if docker-compose ps 2>/dev/null | grep -q "Up"; then
    log_success "Docker 容器已运行"
else
    log_warning "启动容器中..."
    docker-compose down -v 2>/dev/null || true
    docker-compose up -d
    log_warning "等待数据库初始化 (15 秒)..."
    sleep 15
    log_success "Docker 容器已启动"
fi

# 2. Go Backend
echo ""
log_info "启动 Go 后端 (localhost:8080)..."

if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    open -a Terminal "$SCRIPT_DIR/backend-go" --args "cd '$SCRIPT_DIR/backend-go' && go run main.go"
else
    # Linux
    gnome-terminal -- bash -c "cd '$SCRIPT_DIR/backend-go' && go run main.go" 2>/dev/null || \
    xterm -e "cd '$SCRIPT_DIR/backend-go' && go run main.go" &
fi

log_success "Go 后端启动命令已发送"
sleep 3

# 3. Python Backend
echo ""
log_info "启动 Python 后端 (localhost:8081)..."

if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    open -a Terminal --args "bash -c 'cd \"$SCRIPT_DIR/backend-python\" && conda activate junkfilter && python api_server.py'"
else
    # Linux
    gnome-terminal -- bash -c "cd '$SCRIPT_DIR/backend-python' && conda activate junkfilter && python api_server.py" 2>/dev/null || \
    xterm -e "cd '$SCRIPT_DIR/backend-python' && conda activate junkfilter && python api_server.py" &
fi

log_success "Python 后端启动命令已发送"
sleep 3

# 4. Vue Frontend
echo ""
log_info "启动 Vue 前端 (localhost:5173)..."

if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    open -a Terminal --args "bash -c 'cd \"$SCRIPT_DIR/frontend-vue\" && npm run dev'"
else
    # Linux
    gnome-terminal -- bash -c "cd '$SCRIPT_DIR/frontend-vue' && npm run dev" 2>/dev/null || \
    xterm -e "cd '$SCRIPT_DIR/frontend-vue' && npm run dev" &
fi

log_success "Vue 前端启动命令已发送"
sleep 3

# 验证和摘要
echo ""
log_warning "等待服务初始化 (5 秒)..."
sleep 5

echo ""
echo "========================================"
log_success "所有服务已启动！"
echo "========================================"
echo ""

echo "📍 访问地址："
echo "   • 前端应用:      http://localhost:5173"
echo "   • Go 后端 API:   http://localhost:8080/health"
echo "   • Python 后端:   http://localhost:8081/health"
echo ""

echo "🗄️  数据库连接："
echo "   • PostgreSQL:    localhost:5432 (user: truesignal / pass: truesignal123)"
echo "   • Redis:         localhost:6379"
echo ""

echo "📋 服务说明："
echo "   • Go Backend (8080):     REST API + 任务聊天网关"
echo "   • Python Backend (8081): LLM 评估和聊天"
echo "   • Vue Frontend (5173):   用户界面"
echo "   • PostgreSQL (5432):     数据存储"
echo "   • Redis (6379):          缓存和消息队列"
echo ""

echo "⚠️  使用说明："
echo "   • 各服务运行在独立的终端窗口中"
echo "   • 关闭窗口可停止相应服务"
echo "   • 使用 Ctrl+C 优雅停止服务"
echo ""

echo "🧪 测试方法："
echo "   1. 打开浏览器访问 http://localhost:5173"
echo "   2. 进入任务详情页面"
echo "   3. 在聊天面板输入消息，测试 Agent 调优功能"
echo ""

echo "========================================"
echo ""
