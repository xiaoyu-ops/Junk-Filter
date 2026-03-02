#!/bin/bash

# Linux/Mac 验证脚本：隐藏默认源 & 添加删除按钮

echo ""
echo "========================================"
echo "  隐藏默认源 & 添加删除按钮 - 验证脚本"
echo "========================================"
echo ""

# 验证 1: 数据库中的源被禁用
echo "[验证 1] 检查数据库中的源状态..."
docker exec junkfilter-db psql -U truesignal -d truesignal -c "SELECT id, author_name, enabled FROM sources ORDER BY id;"
echo ""
echo "✓ 预期：3 个源，enabled 都是 false"
echo ""

# 检查容器是否运行
echo "[验证 2] 检查容器状态..."
docker-compose ps
echo ""

# 检查 Go 后端是否运行
echo "[验证 3] 尝试连接 Go 后端 (http://localhost:8080)..."
sleep 2
curl -s -o /dev/null -w "HTTP Status: %{http_code}\n" http://localhost:8080/api/sources

if [ $? -ne 0 ]; then
    echo ""
    echo "⚠ Go 后端未运行或无法连接"
    echo "  请启动 Go 后端: cd backend-go && go run main.go"
    echo ""
else
    echo ""
    echo "✓ Go 后端已运行"
    echo ""
    echo "[验证 4] 获取 /api/sources 端点响应..."
    curl -s http://localhost:8080/api/sources
    echo ""
    echo ""
    echo "✓ 预期：[] 或 { \"data\": [] }（空列表）"
    echo ""
fi

# 前端验证提示
echo "[验证 5] 前端验证（需要手动操作）"
echo ""
echo "  1. 启动前端: cd frontend-vue && npm run dev"
echo "  2. 打开浏览器: http://localhost:5173"
echo "  3. 验证：右侧侧边栏显示\"还没有任务\""
echo "  4. 创建任务：点击\"添加任务\"按钮"
echo "  5. 验证删除功能：选中任务后，点击红色\"删除\"按钮"
echo ""

echo "========================================"
echo "  验证完成！"
echo "========================================"
echo ""
