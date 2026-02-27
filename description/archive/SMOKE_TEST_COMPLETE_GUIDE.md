# 🧪 适配层冒烟测试完整指南

**目标**: 验证前端与 Go/Mock 后端的数据适配层是否完美工作

**预计时间**: 15-20 分钟

**测试覆盖**:
- ✅ 任务发现（Go 后端数据映射）
- ✅ 写操作验证（前端创建→后端保存）
- ✅ 混合链路测试（Go + Mock 并行）
- ✅ 异常边界测试（后端故障降级）

---

## 第 0 步：环境准备（5 分钟）

### 0.1 启动所有后端服务（4 个独立终端）

**终端 A - Docker 基础设施**
```bash
cd D:\TrueSignal
docker-compose up -d

# 验证
docker-compose ps
# 应该看到 truesignal-db 和 truesignal-redis 都是 "Up" 状态
```

**终端 B - Go 后端（端口 8080）**
```bash
cd D:\TrueSignal\backend-go
go run main.go
```

**预期输出**:
```
✓ Configuration loaded
✓ Database connected
✓ Redis connected

========== TrueSignal Backend ==========
Database: localhost:5432/truesignal
Redis: localhost:6379
Server: listening on :8080
========================================
```

**终端 C - Mock 后端（端口 3000）**
```bash
cd D:\TrueSignal\backend-mock
node server.js
```

**预期输出**:
```
🚀 TrueSignal Mock 后端服务器已启动
📡 监听端口: 3000
```

**终端 D - 前端（端口 5173）**
```bash
cd D:\TrueSignal\frontend-vue
npm run dev
```

**预期输出**:
```
  Local: http://localhost:5173/
```

### 0.2 验证环境配置

检查 `.env.local` 文件：

```bash
cat D:\TrueSignal\frontend-vue\.env.local | grep VITE_API
```

**预期结果**:
```
VITE_API_URL=http://localhost:8080
VITE_MOCK_URL=http://localhost:3000
```

### 0.3 打开浏览器开发工具

```
访问: http://localhost:5173
按 F12 打开开发者工具
切换到 Console 标签（用于查看日志和错误）
切换到 Network 标签（用于监控 HTTP 请求）
```

---

## 第 1 步：任务发现测试 (5 分钟)

### 目标
验证前端能否正确显示 Go 后端的 RSS 源，且数据适配正确。

### 1.1 检查现有数据

**在终端 A 中执行**:
```bash
# 连接到 PostgreSQL
docker exec -it truesignal-db psql -U truesignal -d truesignal

# 在 psql 提示符中
SELECT id, url, author_name, priority, enabled FROM sources LIMIT 5;
```

**预期结果**:
```
 id |                              url                               |   author_name    | priority | enabled
----+------------------------------------------------------------------+------------------+----------+---------
  1 | https://feeds.arstechnica.com/arstechnica/index                 | Ars Technica     |        8 | t
  2 | https://news.ycombinator.com/rss                                | Hacker News      |        9 | t
  3 | https://feeds.medium.com/tag/technology/latest                  | Medium Tech      |        7 | t
```

记录下这些源的 ID（1, 2, 3）。

### 1.2 前端验证

**在浏览器中**:
1. 打开 http://localhost:5173
2. 左侧边栏应该显示 3 个任务
3. 每个任务应该显示：
   - 名称（来自 author_name）
   - 状态为 "active" 或 "paused"（来自 enabled 字段）

### 1.3 验证数据映射

**在浏览器 Console 中输入**:
```javascript
// 查看前端加载的任务数据
const { tasks } = await useAPI().tasks.list()
console.log('任务列表:', tasks)
```

**预期结果**:
```javascript
[
  {
    id: "source-1",                    // ✅ ID 格式正确：source-{id}
    name: "Ars Technica",
    command: "https://feeds.arstechnica.com/arstechnica/index",
    frequency: "hourly",               // ✅ priority:8 → hourly
    status: "active",                  // ✅ enabled:true → active
    created_at: "2026-02-27T...",
    _source: { ... }                   // ✅ 保留原始 Go 数据
  },
  ...
]
```

### ✅ 成功指标

- [ ] 左侧任务列表显示 3 个任务
- [ ] 任务名称正确显示
- [ ] 状态映射正确（active/paused）
- [ ] 没有网络错误（Network 标签中无红色状态码）
- [ ] Console 中无红色错误日志
- [ ] useAPI 返回的数据格式完全符合预期

**如果失败**:
```bash
# 检查 Go 后端是否正常运行
curl http://localhost:8080/api/sources

# 检查前端是否能连接
curl http://localhost:5173/
```

---

## 第 2 步：写操作验证 (5 分钟)

### 目标
验证前端创建任务时，adaptTaskToSource 正确转换数据格式。

### 2.1 在前端创建任务

**在浏览器中**:
1. 点击 "创建任务" 按钮（或类似的 UI 元素）
2. 填写表单：
   - 任务名称: `🧪 测试 RSS 源`
   - 命令: `https://example.com/test-feed.xml`
   - 频率: `daily`
3. 点击 "创建" 按钮

### 2.2 监控网络请求

**在浏览器 Network 标签中观察**:

找到 `POST http://localhost:8080/api/sources` 请求，检查：

**Request Body** (应该是转换后的 Source 格式):
```json
{
  "name": "🧪 测试 RSS 源",
  "url": "https://example.com/test-feed.xml",
  "priority": 6,           // ✅ daily → priority:6
  "enabled": true          // ✅ status:active → enabled:true
}
```

**不应该看到**:
```json
{
  "id": "source-xxx",      // ❌ 不应该发送 ID
  "frequency": "daily",    // ❌ 应该转换为 priority
  "status": "active"       // ❌ 应该转换为 enabled
}
```

### 2.3 验证后端保存

**在 Console 中执行**:
```javascript
// 重新加载任务列表
const tasks = await useAPI().tasks.list()
console.log('最新任务列表:', tasks)
```

**预期结果**:
- 任务列表应该包含 4 个任务（原来 3 个 + 新建 1 个）
- 最后一个任务是：
  ```javascript
  {
    id: "source-4",
    name: "🧪 测试 RSS 源",
    command: "https://example.com/test-feed.xml",
    frequency: "daily"
  }
  ```

### 2.4 数据库验证

**在 psql 中执行**:
```bash
docker exec -it truesignal-db psql -U truesignal -d truesignal

# 在 psql 提示符中
SELECT id, author_name, url, priority FROM sources ORDER BY id DESC LIMIT 1;
```

**预期结果**:
```
 id |      author_name      |                       url                       | priority
----+-----------------------+-------------------------------------------------+----------
  4 | 🧪 测试 RSS 源        | https://example.com/test-feed.xml               |        6
```

### ✅ 成功指标

- [ ] POST 请求发送到 `http://localhost:8080/api/sources`
- [ ] Request Body 是正确转换的 Source 格式（priority/enabled 而不是 frequency/status）
- [ ] Response 状态码为 201 或 200
- [ ] 刷新前端后，新任务出现在列表中
- [ ] 数据库中能查到新插入的源

**如果失败**:
```bash
# 检查 Go 后端的 POST 处理
curl -X POST http://localhost:8080/api/sources \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","url":"https://example.com","priority":6,"enabled":true}'

# 查看响应
```

---

## 第 3 步：混合链路测试 (5 分钟)

### 目标
验证前端能同时维持对 Go (8080) 的业务请求和对 Mock (3000) 的消息/SSE 请求，且互不干扰。

### 3.1 选择刚创建的任务

**在浏览器中**:
1. 左侧任务列表中点击 `🧪 测试 RSS 源` 任务

**预期**:
- 右侧聊天框切换到该任务
- Console 应该显示消息加载请求（向 Mock 发送）

### 3.2 监控双链路请求

**在浏览器 Network 标签中观察**:

创建任务后，应该看到：

**Request 1** (业务数据 → Go):
```
GET http://localhost:8080/api/sources
GET http://localhost:8080/api/sources/4
```

**Request 2** (消息数据 → Mock):
```
GET http://localhost:3000/api/tasks/4/messages
```

两个请求应该**几乎同时发送**，说明适配层正确分发请求。

### 3.3 发送消息测试

**在浏览器聊天框中**:
1. 输入消息: `🧪 测试消息`
2. 按 Enter 发送

**在 Network 标签中观察两个并行请求**:

**Request A** (保存消息到 Mock):
```
POST http://localhost:3000/api/messages
Body: {
  "task_id": "4",           // ✅ 自动转换 source-4 → 4
  "role": "user",
  "type": "text",
  "content": "🧪 测试消息"
}
Status: 201 或 200
```

**Request B** (SSE 聊天连接到 Mock):
```
GET http://localhost:3000/api/chat/stream?taskId=4&message=...
Status: 200
Type: text/event-stream
```

### 3.4 验证流式回复

**在浏览器聊天框中观察**:
1. 用户消息应该立即显示
2. AI 加载动画出现（3 个点跳动）
3. AI 回复逐字显示（流式效果）
4. 执行卡片显示（如果有）

**在 Console 中查看日志**:
```javascript
// 应该看到 [SSE] 开头的日志
[SSE] 连接成功
[SSE] Delta: "你"
[SSE] Delta: "好"
[SSE] Done 事件收到
```

**验证消息保存到 Mock**:
```bash
# 检查 Mock 数据文件
cat backend-mock/data/messages.json | jq '.[] | select(.task_id == "4")'
```

**预期结果**:
```json
{
  "id": "msg-xxx",
  "task_id": "4",
  "role": "user",
  "type": "text",
  "content": "🧪 测试消息",
  "timestamp": "2026-02-27T10:00:00Z"
},
{
  "id": "msg-yyy",
  "task_id": "4",
  "role": "ai",
  "type": "text",
  "content": "你好！👋 我是 TrueSignal AI 助手...",
  "timestamp": "2026-02-27T10:00:01Z"
}
```

### ✅ 成功指标

- [ ] 消息发送到 `http://localhost:3000/api/messages` (Mock)
- [ ] SSE 连接到 `http://localhost:3000/api/chat/stream` (Mock)
- [ ] task_id 自动从 "source-4" 转换为 "4"
- [ ] 同时支持 Go (8080) 和 Mock (3000) 的请求
- [ ] 两个服务的请求互不干扰
- [ ] SSE 流式回复正常工作
- [ ] 消息正确保存到 Mock 的 JSON 文件

**如果失败**:
```bash
# 检查 Mock 的消息端点
curl -X POST http://localhost:3000/api/messages \
  -H "Content-Type: application/json" \
  -d '{"task_id":"4","role":"user","type":"text","content":"test"}'

# 检查 SSE 流式端点
curl -N "http://localhost:3000/api/chat/stream?taskId=4&message=hello"
```

---

## 第 4 步：异常边界测试 (3 分钟)

### 目标
验证当 Go 后端故障时，前端的错误处理机制是否正确。

### 4.1 停止 Go 后端

**在 Go 后端终端中**:
```bash
# 按 Ctrl+C 停止 Go 后端
# 或者用另一个终端杀死进程
```

### 4.2 尝试创建任务

**在浏览器中**:
1. 点击 "创建任务"
2. 填写表单并点击 "创建"

**预期行为**:
- 应该在 3 秒内显示错误提示
- 错误消息: `请求超时` 或 `请求失败 (0)` 或 `无法连接到服务器`
- 创建操作应该失败，不会添加新任务

**在 Console 中查看日志**:
```
[API] 请求失败: Error: 请求超时
```

### 4.3 验证 Mock 仍可用

**在浏览器中**:
1. 选择任何任务
2. 发送消息

**预期行为**:
- 消息仍然可以保存（到 Mock 的 3000）
- SSE 聊天仍然可用
- 不会影响 Mock 后端的功能

**验证**:
```bash
# Mock 应该仍在运行
curl http://localhost:3000/api/tasks
# 应该返回任务列表（不受 Go 后端状态影响）
```

### 4.4 重启 Go 后端

**在新的终端中**:
```bash
cd D:\TrueSignal\backend-go
go run main.go
```

### 4.5 验证恢复

**在浏览器中**:
1. 刷新页面（或等待自动重试）
2. 任务列表应该重新加载

**预期行为**:
- 任务列表恢复显示
- 可以再次创建任务
- 系统完全恢复

### ✅ 成功指标

- [ ] Go 后端停止时，创建任务显示错误（不是无限等待）
- [ ] 错误消息清晰明了
- [ ] Mock 后端仍然可用（消息和 SSE）
- [ ] Go 后端重启后自动恢复
- [ ] 没有 UI 挂起或崩溃

---

## 完整测试清单

使用此清单验证所有功能：

### 环境就绪
- [ ] Docker 容器运行（PostgreSQL + Redis）
- [ ] Go 后端启动在 8080
- [ ] Mock 后端启动在 3000
- [ ] 前端启动在 5173
- [ ] .env.local 配置正确

### 第 1 步 - 任务发现
- [ ] 数据库有初始 RSS 源数据
- [ ] 前端左侧显示 3 个任务
- [ ] 任务名称正确
- [ ] 状态映射正确（active/paused）
- [ ] 无网络错误

### 第 2 步 - 写操作验证
- [ ] 创建新任务表单可用
- [ ] POST 请求发送到 8080 (Go)
- [ ] Request Body 格式正确（priority/enabled）
- [ ] Response 状态码正确（201/200）
- [ ] 新任务立即显示在列表
- [ ] 数据库中能查到新源

### 第 3 步 - 混合链路
- [ ] 任务加载请求到 8080
- [ ] 消息加载请求到 3000
- [ ] 消息发送到 3000 (Mock)
- [ ] SSE 连接到 3000 (Mock)
- [ ] 流式回复正常显示
- [ ] task_id 自动转换
- [ ] 消息保存到 Mock JSON

### 第 4 步 - 异常边界
- [ ] Go 后端停止时提示错误
- [ ] Mock 后端仍可用
- [ ] 错误提示清晰明了
- [ ] 重启后自动恢复

---

## 故障排查快速参考

| 问题 | 原因 | 解决方案 |
|------|------|---------|
| 任务列表为空 | Go 后端无数据或未连接 | 检查 Go 启动日志，验证 `docker exec curl http://localhost:8080/api/sources` |
| 创建任务报错 | Go 后端不可达 | 检查 8080 是否开放，`netstat -ano \| findstr :8080` |
| 消息无法保存 | Mock 后端不可达 | 检查 3000 是否开放，`curl http://localhost:3000/api/tasks` |
| SSE 无流式效果 | Mock 后端 SSE 端点问题 | 直接测试 `curl -N "http://localhost:3000/api/chat/stream?taskId=1&message=hi"` |
| 前端显示错误 | 后端返回 500 | 查看后端日志，检查是否有 SQL 错误或连接问题 |
| 网络请求到错误的 URL | .env.local 配置错误 | 检查 `VITE_API_URL` 和 `VITE_MOCK_URL` |

---

## 成功完成的标志

✅ **冒烟测试通过** 意味着：

1. **数据适配层正常工作**
   - Task ↔ Source 转换无缝
   - ID 格式自动处理

2. **混合架构可用**
   - Go 后端处理业务数据 (8080)
   - Mock 后端处理消息/SSE (3000)
   - 两者互不干扰

3. **错误处理完善**
   - 网络故障有明确提示
   - 部分故障不影响整体
   - 自动恢复机制有效

4. **UI 体验流畅**
   - 无卡顿或无限等待
   - 流式回复平滑
   - 交互响应快速

---

## 继续下一步

如果冒烟测试全部通过，可以：

1. **性能测试**: 创建 100 个任务，验证列表加载速度
2. **并发测试**: 同时在多个任务中发送消息
3. **完整集成**: 启动 Python 后端，验证评估结果查询
4. **生产就绪检查**: 配置 HTTPS、认证、日志等

---

**现在开始测试吧！** 🚀

