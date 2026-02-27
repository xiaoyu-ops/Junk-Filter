# 🚀 冒烟测试快速启动指南

**目标**: 5 分钟内启动完整测试环境，20 分钟内完成全链路冒烟测试

---

## ⚡ 30 秒快速启动

### 方式 1: 一键启动所有服务 (推荐)

如果你有 `start-all.bat/sh` 脚本（包含所有后端和前端启动），运行：

**Windows**:
```bash
start-all.bat
```

**Linux/Mac**:
```bash
chmod +x start-all.sh
./start-all.sh
```

**预期结果**: 5 个终端会自动打开，运行：
- Docker 容器
- Go 后端 (8080)
- Mock 后端 (3000)
- Python 后端 (可选)
- 前端 (5173)

### 方式 2: 手动启动 (4 个独立终端)

**终端 1 - 启动 Docker** (如果还未启动):
```bash
cd D:\TrueSignal
docker-compose up -d
```

**终端 2 - Go 后端**:
```bash
cd D:\TrueSignal\backend-go
go run main.go
```

**终端 3 - Mock 后端**:
```bash
cd D:\TrueSignal\backend-mock
node server.js
```

**终端 4 - 前端**:
```bash
cd D:\TrueSignal\frontend-vue
npm run dev
```

---

## 🔍 即时检查环境

### 验证所有服务启动成功

**选项 A: 自动检查** (推荐)

**Windows**:
```bash
cd D:\TrueSignal
smoke-test-check.bat
```

**Linux/Mac**:
```bash
cd D:\TrueSignal
chmod +x smoke-test-check.sh
./smoke-test-check.sh
```

**预期输出**:
```
✅ Go 后端 (8080) - 正常
✅ Mock 后端 (3000) - 正常
✅ 前端应用 (5173) - 正常
✅ PostgreSQL - 正常
✅ Redis - 正常
✅ .env.local 配置正确
✅ sources 表有 3 条记录

✅ 所有服务都已就绪，可以开始冒烟测试！
```

### 选项 B: 手动检查

**在浏览器中访问**:
```
Go 后端:   http://localhost:8080/health
Mock 后端: http://localhost:3000/api/tasks
前端:     http://localhost:5173
```

**在命令行检查**:
```bash
# 检查 8080 (Go)
curl http://localhost:8080/health

# 检查 3000 (Mock)
curl http://localhost:3000/api/tasks

# 检查 5173 (前端)
curl http://localhost:5173
```

---

## 🧪 4 步完整冒烟测试

### 步骤 1: 打开前端应用 (1 分钟)

1. **打开浏览器**
   ```
   访问: http://localhost:5173
   ```

2. **打开开发工具** (便于监控)
   ```
   按 F12 或 Ctrl+Shift+I (Windows/Linux)
   按 Cmd+Option+I (Mac)
   ```

3. **切换标签页**
   - **Console** - 查看日志和错误
   - **Network** - 监控 HTTP 请求
   - **Application** - 查看存储数据

### 步骤 2: 任务发现测试 (3 分钟)

**验证左侧任务列表**:

1. 左侧应该显示 3 个任务（来自 Go 后端的 RSS 源）
2. 每个任务显示名称和状态 (active/paused)

**在 Console 中验证数据适配**:
```javascript
// 复制粘贴到 Console，查看任务数据
const taskList = document.querySelectorAll('[data-test="task-item"]');
console.log('任务数量:', taskList.length);

// 或者在 Network 标签中检查
// 应该看到 GET http://localhost:8080/api/sources
```

### 步骤 3: 创建新任务 (5 分钟)

**在前端创建任务**:

1. 点击 "创建任务" 按钮
2. 填写表单:
   ```
   任务名称: 🧪 测试 RSS
   命令:     https://example.com/feed.xml
   频率:     daily
   ```
3. 点击 "创建"

**在 Network 标签中验证**:

找到这个请求:
```
POST http://localhost:8080/api/sources
```

查看 **Request** 部分 (Payload):
```json
{
  "name": "🧪 测试 RSS",
  "url": "https://example.com/feed.xml",
  "priority": 6,
  "enabled": true
}
```

✅ **成功标志**:
- Request 去往 8080 (Go 后端)
- Payload 格式是 Source 格式（priority/enabled，不是 frequency/status）
- Response 状态码 201 或 200
- 新任务立即显示在列表中

### 步骤 4: 消息和 SSE 测试 (5 分钟)

**选择刚创建的任务，发送消息**:

1. 点击左侧的 "🧪 测试 RSS" 任务
2. 在聊天框输入: `🧪 测试消息`
3. 按 Enter 发送

**在 Network 标签中观察两个并行请求**:

**请求 A** (消息保存到 Mock):
```
POST http://localhost:3000/api/messages
Payload: {
  "task_id": "4",
  "role": "user",
  "content": "🧪 测试消息"
}
Status: 201
```

**请求 B** (SSE 流式连接到 Mock):
```
GET http://localhost:3000/api/chat/stream?taskId=4&message=...
Status: 200 (长连接)
Type: text/event-stream
```

**在聊天框中观察**:
1. 用户消息立即显示
2. AI 加载动画（3 个点）
3. AI 回复逐字显示（流式效果）
4. 执行卡片显示（如果有）

✅ **成功标志**:
- 消息发送到 3000 (Mock)
- SSE 连接到 3000 (Mock)
- 两个请求互不干扰
- 流式回复正常工作
- Console 中无红色错误

---

## 🎯 快速验收清单

完成上述 4 步后，检查以下清单：

### 任务发现 ✅
- [ ] 左侧显示 3 个初始任务
- [ ] 任务名称正确
- [ ] 状态显示正确

### 创建任务 ✅
- [ ] 创建表单可用
- [ ] POST 请求发送到 8080
- [ ] Request 格式正确（Source 格式）
- [ ] 新任务添加到列表
- [ ] 刷新后仍存在

### 消息和 SSE ✅
- [ ] 消息发送到 3000
- [ ] SSE 连接到 3000
- [ ] 流式回复显示
- [ ] 没有网络错误
- [ ] 两个后端互不干扰

### 整体体验 ✅
- [ ] UI 流畅，无卡顿
- [ ] 响应迅速
- [ ] 没有红色错误
- [ ] 数据正确保存

---

## 🔧 如果出现问题

### 问题 1: 任务列表为空

**排查**:
```bash
# 检查 Go 后端是否有数据
curl http://localhost:8080/api/sources

# 检查数据库
docker exec -it truesignal-db psql -U truesignal -d truesignal
SELECT * FROM sources;
```

**解决**:
```bash
# 插入测试数据
docker exec -it truesignal-db psql -U truesignal -d truesignal << EOF
INSERT INTO sources (url, author_name, priority, enabled) VALUES
  ('https://example1.com/feed', 'Example 1', 8, TRUE),
  ('https://example2.com/feed', 'Example 2', 7, TRUE);
EOF
```

### 问题 2: 创建任务失败

**排查**:
- 检查 Network 标签中 POST 请求的错误
- 查看 Go 后端的日志
- 确认 .env.local 中 VITE_API_URL=http://localhost:8080

**解决**:
```bash
# 直接测试 Go API
curl -X POST http://localhost:8080/api/sources \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","url":"https://example.com","priority":6,"enabled":true}'
```

### 问题 3: 消息无法保存或 SSE 不工作

**排查**:
- 检查 Network 标签，消息请求是否发到 3000
- 确认 .env.local 中 VITE_MOCK_URL=http://localhost:3000
- 检查 Mock 后端是否还在运行

**解决**:
```bash
# 重启 Mock 后端
cd backend-mock
node server.js
```

### 问题 4: 出现网络错误

**排查**:
```bash
# 检查所有端口是否开放
netstat -ano | findstr :8080    # Windows
netstat -ano | findstr :3000
netstat -ano | findstr :5173

lsof -i :8080   # Mac/Linux
lsof -i :3000
lsof -i :5173
```

**解决**:
- 关闭占用这些端口的其他应用
- 重启对应的服务

---

## 📊 使用 Console 诊断脚本

如果想要自动化测试所有连接，可以使用 Console 脚本：

1. 打开 http://localhost:5173
2. 按 F12 打开 Console
3. 复制 `browser-console-test.js` 的全部内容
4. 粘贴到 Console，按 Enter

脚本会自动输出：
- ✅ 环境变量检查
- ✅ API 连接测试
- ✅ 数据适配验证
- ✅ 完整的测试总结

---

## 🎉 测试完成

如果所有步骤都通过，说明：

✅ **适配层工作完美**
- Source ↔ Task 数据转换无缝
- 前端与 Go 后端完全兼容
- Mock 后端消息和 SSE 功能正常

✅ **混合架构可用**
- Go 处理业务数据 (8080)
- Mock 处理消息和聊天 (3000)
- 两者互不干扰，协作无缝

✅ **可以进入下一阶段**
- 集成 Python 评估引擎
- 添加用户认证
- 部署到生产环境

---

## 📚 详细文档

完整的冒烟测试指南和更多细节：
```
description/SMOKE_TEST_COMPLETE_GUIDE.md
```

---

**祝你测试顺利！** 🚀

