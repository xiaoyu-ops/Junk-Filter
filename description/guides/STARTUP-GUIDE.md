# 🚀 TrueSignal 一键启动指南

## 快速开始

### Windows 用户
```cmd
# 进入项目目录
cd D:\TrueSignal

# 执行启动脚本（双击或命令行）
start-all.bat
```

### Linux/Mac 用户
```bash
# 进入项目目录
cd /path/to/TrueSignal

# 执行启动脚本
chmod +x start-all.sh
./start-all.sh
```

---

## 📊 启动流程

脚本会自动按顺序启动：

| 序号 | 服务 | 端口 | 启动方式 |
|------|------|------|---------|
| 1️⃣ | Docker (PostgreSQL + Redis) | 5432, 6379 | 后台容器 |
| 2️⃣ | Go Backend | 8080 | 新窗口 |
| 3️⃣ | Python Backend | 8081 | 新窗口 |
| 4️⃣ | Vue Frontend | 5173 | 新窗口 |

---

## 🔗 服务地址

启动完成后，你可以访问：

```
🌐 前端应用
   http://localhost:5173

🔌 API 接口
   Go Backend:     http://localhost:8080/health
   Python Backend: http://localhost:8081/health

💾 数据库
   PostgreSQL: localhost:5432
     用户: truesignal
     密码: truesignal123
     数据库: truesignal

   Redis: localhost:6379
```

---

## ⚙️ 服务说明

### Go Backend (8080)
- REST API 服务，提供任务、内容、评估接口
- 新的任务聊天端点：`POST /api/tasks/:task_id/chat`
- 负责与 Python 后端通信和响应转发

### Python Backend (8081)
- FastAPI 服务，提供 LLM 评估和聊天功能
- 内容评估端点：`POST /api/evaluate`
- 任务聊天端点：`POST /api/task/:task_id/chat`
- 使用 Conda junkfilter 环境运行

### Vue Frontend (5173)
- 用户交互界面
- 任务管理、配置、聊天功能
- 开发热重载支持

### PostgreSQL (5432)
- 主要数据存储
- 表：sources, content, evaluation, messages 等
- 凭证：truesignal / truesignal123

### Redis (6379)
- 缓存层
- 消息队列（Stream）
- 去重 Bloom Filter 存储

---

## 🧪 测试聊天功能

### 方式 1: 使用浏览器 UI（最直观）
```
1. 打开 http://localhost:5173
2. 导航到任务详情页面
3. 在右侧聊天面板输入问题
4. 例如："现在的执行进度如何？"
5. 观察 Agent 的自然语言回复
```

### 方式 2: 使用 curl 测试 API（开发调试）
```bash
# 获取任务列表
curl http://localhost:8080/api/sources

# 发送聊天请求
curl -X POST http://localhost:8080/api/tasks/1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "为什么这张卡片被标记为 SKIP？",
    "agent_context": {
      "task_metadata": {"id": 1, "name": "Test Task"},
      "chat_history": [],
      "recent_cards": [],
      "current_config": {"temperature": 0.7}
    }
  }'
```

---

## 🛑 停止服务

### 方式 1: 关闭窗口
- 直接关闭各个服务的窗口即可停止

### 方式 2: 优雅停止
- 在各窗口中按 `Ctrl+C` 优雅停止

### 方式 3: 完全清理（包括 Docker）
```bash
# 停止所有容器
docker-compose down

# 删除所有数据卷（谨慎）
docker-compose down -v
```

---

## ⚠️ 常见问题

### 问题 1: 脚本无法执行
**Windows:**
```powershell
# 解决方案：设置执行策略（仅第一次需要）
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

**Linux/Mac:**
```bash
# 确保脚本有执行权限
chmod +x start-all.sh
```

### 问题 2: 端口被占用
```bash
# Windows：查找占用端口的进程
netstat -ano | findstr :8080

# Linux/Mac：查找占用端口的进程
lsof -i :8080

# 杀死进程（替换 PID）
kill -9 <PID>
```

### 问题 3: Docker 容器启动失败
```bash
# 查看 Docker 日志
docker-compose logs

# 重建容器
docker-compose down -v
docker-compose up -d
```

### 问题 4: Python 环境错误
```bash
# 检查 conda 环境
conda activate junkfilter
pip list

# 重新安装依赖
pip install -r backend-python/requirements.txt
```

### 问题 5: 前端无法连接后端
```
检查事项：
1. Go Backend 是否在运行（查看 8080 端口）
2. 浏览器控制台是否有 CORS 错误
3. 尝试直接访问 http://localhost:8080/health
4. 查看网络标签页中的请求是否成功
```

---

## 📝 日志查看

每个服务的日志显示在对应的窗口中：

```
[timestamp] [服务日志信息...]
```

- **Go Backend 日志**：显示 HTTP 请求、数据库操作
- **Python Backend 日志**：显示 API 调用、LLM 处理
- **Vue Frontend 日志**：显示编译信息、热重载

---

## 🎯 验证清单

启动完成后，逐一验证：

- [ ] Docker 容器运行正常（2 个容器）
- [ ] Go Backend 显示 "Server: listening on :8080"
- [ ] Python Backend 显示 "Application running"
- [ ] Vue Frontend 显示 "Local: http://localhost:5173"
- [ ] 可以访问 http://localhost:5173
- [ ] 能获取任务列表 (GET /api/sources)
- [ ] 能发送聊天消息 (POST /api/tasks/:id/chat)
- [ ] 收到流式 SSE 响应

---

## 🚀 下一步

1. **配置 LLM**（如果还没配）
   ```bash
   # 编辑 .env，设置 OpenAI API 密钥
   OPENAI_API_KEY=sk-...
   LLM_MODEL_ID=gpt-4
   ```

2. **添加 RSS 源**（可选）
   ```bash
   curl -X POST http://localhost:8080/api/sources \
     -H "Content-Type: application/json" \
     -d '{"url": "https://example.com/rss", "author_name": "Example"}'
   ```

3. **测试完整流程**
   - 添加 RSS 源 → RSS 抓取 → 内容评估 → 聊天调优

---

## 💡 提示

- 首次启动可能需要 20-30 秒，请耐心等待
- 所有服务的日志实时显示在各自的窗口中
- 修改代码后，各服务会自动重新加载（特别是 Python/Vue）
- 数据持久化在 Docker 卷中，停止服务不会丢失数据
- 要清空数据，使用 `docker-compose down -v` 后重启

---

## 📚 相关文档

- 详细的系统架构说明：见 `description/` 文件夹
- Go 后端 API 文档：`description/guides/API.md`
- 前端使用指南：`description/guides/FRONTEND.md`
