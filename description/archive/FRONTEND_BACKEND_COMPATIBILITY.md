# 前端与真实后端兼容性分析

**分析日期**: 2026-02-27
**前端版本**: Vue 3 + Vite + Pinia
**目标**: 评估前端能否与 backend-go 和 backend-python 对接

---

## 📊 快速结论

**兼容性评分**: ⚠️ **70% 兼容**

前端的 API 层设计良好，支持灵活的后端切换，但需要小幅适配。无需大改，仅需调整：
1. API 端点映射
2. 数据格式转换
3. 新增少量接口补充

---

## 🔍 详细兼容性分析

### 1️⃣ 任务管理 (Task CRUD)

**前端期望的 API** (useTaskStore.js):
```javascript
tasks.list()              // GET /api/tasks
tasks.get(id)             // GET /api/tasks/:id
tasks.create(data)        // POST /api/tasks
tasks.update(id, data)    // PUT /api/tasks/:id
tasks.delete(id)          // DELETE /api/tasks/:id
tasks.execute(id)         // POST /api/tasks/:id/execute
```

**兼容性分析**:

| 端点 | Mock | Go | Python | 状态 |
|------|------|----|---------|----|
| POST /api/tasks | ✅ | ❌ | ❌ | **需补充** |
| GET /api/tasks | ✅ | ✅ sources | ⚠️ 可映射 | **可用，需映射** |
| GET /api/tasks/:id | ✅ | ✅ sources/:id | ⚠️ 可映射 | **可用，需映射** |
| PUT /api/tasks/:id | ✅ | ✅ sources/:id | ⚠️ 可映射 | **可用，需映射** |
| DELETE /api/tasks/:id | ✅ | ✅ sources/:id | ⚠️ 可映射 | **可用，需映射** |
| POST /api/tasks/:id/execute | ✅ | ❌ | ❌ | **需补充** |

**结论**:
- ✅ 可直接用 Go 的 `/sources` 端点替代 `/tasks`
- ❌ 缺少 `/execute` 端点（可实现为手动触发 RSS 抓取）

---

### 2️⃣ 消息管理 (Messages)

**前端期望的 API** (TaskChat.vue):
```javascript
messages.list(taskId, { limit, offset })   // GET /api/tasks/:taskId/messages
messages.save(data)                         // POST /api/messages
```

**前端发送的消息格式**:
```javascript
{
  task_id: string,
  role: 'user' | 'ai',
  type: 'text' | 'error' | 'execution',
  content: string,
  timestamp: ISO8601
}
```

**兼容性分析**:

| 方法 | Mock 支持 | Go 支持 | 说明 |
|------|---------|--------|------|
| 消息保存 | ✅ JSON | ❌ | Go 无消息存储，需新增 API |
| 消息查询 | ✅ 按任务 | ❌ | Go 无此端点 |
| SSE 流式 | ✅ | ❌ | Go 无聊天接口 |

**结论**:
- ⚠️ Go 后端无消息存储能力
- ⚠️ 需创建新的消息表和 API
- ⚠️ SSE 端点需新增实现

---

### 3️⃣ SSE 流式对话

**前端期望的端点**:
```javascript
GET /api/chat/stream?taskId=xxx&message=yyy
// 返回 text/event-stream 格式
// 事件: delta, execution, done, error
```

**兼容性分析**:

| 功能 | Mock | Go | Python | 状态 |
|------|------|----|---------|----|
| SSE 端点 | ✅ | ❌ | ❌ | **需补充** |
| Delta 流式 | ✅ | ❌ | ❌ | **需补充** |
| Execution 卡片 | ✅ | ❌ | ❌ | **需补充** |
| Done 事件 | ✅ | ❌ | ❌ | **需补充** |

**结论**:
- ❌ Go 和 Python 都无此功能
- 选项 A: 保留 Mock SSE（推荐短期）
- 选项 B: 在 Go 中实现 SSE（长期）

---

### 4️⃣ 认证接口

**前端期望的 API**:
```javascript
auth.login(email, password)              // POST /api/auth/login
auth.register(email, password, name)     // POST /api/auth/register
```

**兼容性分析**:
- ❌ 三个后端都未实现
- ⚠️ 当前前端未使用（可选功能）
- 📝 后续集成时需补充

---

## 📈 数据格式对比

### 任务对象格式

**前端期望** (Mock 返回):
```json
{
  "id": "task-001",
  "name": "每日新闻摘要",
  "command": "fetch_rss",
  "frequency": "daily",
  "status": "active",
  "last_execution": "2026-02-27T10:00:00Z",
  "created_at": "2026-02-26T14:30:00Z"
}
```

**Go 返回** (sources 表):
```json
{
  "id": 1,
  "url": "https://example.com/rss",
  "name": "Example RSS",
  "priority": 5,
  "enabled": true,
  "last_fetch_time": "2026-02-27T10:00:00Z",
  "created_at": "2026-02-26T14:30:00Z",
  "updated_at": "2026-02-27T10:00:00Z"
}
```

**适配方案**:
```javascript
// 在 useAPI.js 中添加转换器
const adaptSourceToTask = (source) => ({
  id: `source-${source.id}`,
  name: source.name,
  command: 'fetch_rss',
  frequency: 'hourly',
  status: source.enabled ? 'active' : 'paused',
  last_execution: source.last_fetch_time,
  created_at: source.created_at
})
```

---

### 消息对象格式

**前端使用** (统一格式):
```json
{
  "id": "msg-123",
  "task_id": "task-001",
  "role": "user" | "ai",
  "type": "text" | "error" | "execution",
  "content": "消息内容",
  "timestamp": "2026-02-27T10:00:00Z"
}
```

**需新建的数据库表**:
```sql
CREATE TABLE messages (
  id UUID PRIMARY KEY,
  task_id INT REFERENCES sources(id),
  role VARCHAR(10),
  type VARCHAR(20),
  content TEXT,
  timestamp TIMESTAMP
);
```

---

## 🎯 迁移方案

### 方案 A: 混合模式（推荐）

**配置**: 前端同时连接两个后端

```javascript
// .env
VITE_API_URL=http://localhost:8080        # Go 后端（主数据）
VITE_CHAT_URL=http://localhost:3000       # Mock（仅 SSE 聊天）
VITE_EVAL_URL=http://localhost:5000       # Python（仅评估结果查询）
```

**前端修改**:
```javascript
// useAPI.js 中区分调用
const taskApi = 'http://localhost:8080'   // 任务和消息 → Go
const chatApi = 'http://localhost:3000'   // SSE 聊天 → Mock
const evalApi = 'http://localhost:5000'   // 评估查询 → Python
```

**优点**:
- ✅ 最小化改动
- ✅ 充分利用各后端能力
- ✅ 低风险过渡

**缺点**:
- ⚠️ 需维护多个后端
- ⚠️ API 分散，不够统一

---

### 方案 B: 完整迁移到 Go

**前提**: Go 后端补充实现
- 消息表和 API
- SSE 流式聊天端点
- 与 Python 的集成

**前端改动**:
```javascript
// .env
VITE_API_URL=http://localhost:8080        # 所有 API 都走 Go
```

**优点**:
- ✅ 统一后端
- ✅ 接口清晰

**缺点**:
- ⚠️ Go 需较大改动（新增消息管理、SSE 等）
- ⚠️ 工作量中等（1-2 天）

---

## 🛠️ 所需的实现工作

### 短期（保持现状 + 小幅适配）

**1. useAPI.js 适配**
```javascript
// 添加数据转换层
const adaptSourceToTask = (source) => {
  return {
    id: `source-${source.id}`,
    name: source.name || source.url,
    frequency: 'hourly',
    status: source.enabled ? 'active' : 'paused',
    last_execution: source.last_fetch_time,
  }
}

// 在 tasks.list() 中应用
list: async () => {
  const sources = await request('/api/sources')
  return sources.map(adaptSourceToTask)
}
```

**2. 消息持久化选项**
- 保留 Mock 的 JSON 存储
- 或新建 Go 消息 API（见下）
- 或使用 Python 查询最近的评估内容

**3. SSE 聊天端点**
- 保留 Mock `/api/chat/stream`
- 或在 Go 后端补充实现（推荐长期）

**工作量**: 2-4 小时

---

### 中期（Go 后端扩展）

**在 Go 后端新增**:

```go
// handlers/message_handler.go
type MessageHandler struct {
    messageRepo *repositories.MessageRepository
}

func (mh *MessageHandler) SaveMessage(c *gin.Context) {
    var req models.CreateMessageRequest
    // 保存消息到 messages 表
}

func (mh *MessageHandler) ListMessages(c *gin.Context) {
    taskId := c.Param("taskId")
    // 从 messages 表查询
}

// handlers/chat_handler.go
func (ch *ChatHandler) StreamChat(c *gin.Context) {
    taskId := c.Query("taskId")
    message := c.Query("message")

    // 1. 保存用户消息
    // 2. 调用 Python 评估服务（RPC 或 HTTP）
    // 3. 流式返回结果 (text/event-stream)
}
```

**工作量**: 1-2 天

---

### 长期（完整集成）

- [ ] Go 消息存储和 API
- [ ] Go SSE 聊天端点
- [ ] Python 评估 API 端点
- [ ] 用户认证体系
- [ ] 前端配置中心（RSS 源、模型参数）

---

## ✅ 兼容性清单

### 关键路径检查

- [x] 任务列表加载 (可用，需适配)
- [x] 任务创建/删除 (可用，需适配)
- [ ] 消息历史查询 (需新增 Go API)
- [ ] 消息保存 (保留 Mock 或新增 Go API)
- [x] SSE 流式对话 (保留 Mock)
- [ ] 内容评估查询 (可新增 Python API)

---

## 💡 立即行动建议

### 今天（第 1 步）
1. ✅ 启动 Docker + Go + Python
2. ✅ 验证 Go `/sources` 端点
3. ✅ 在 useAPI.js 添加转换函数
4. ✅ 修改 `VITE_API_URL=http://localhost:8080`
5. ✅ 测试任务列表加载（会看到 Go 的 RSS 源作为任务）

### 明天（第 2 步）
6. ⚠️ 处理消息存储问题
   - 选项 A: 保留 Mock JSON（简单）
   - 选项 B: 新建 Go 消息 API（完整）

7. ⚠️ 处理 SSE 聊天
   - 选项 A: 保留 Mock `/api/chat/stream`（简单）
   - 选项 B: 在 Go 实现新的端点（完整）

### 一周后（第 3 步）
8. [ ] Go 后端完整扩展（消息 + SSE）
9. [ ] 前端配置迁移（VITE_API_URL 统一指向 Go）
10. [ ] 弃用 Mock 后端

---

## 📊 风险评估

| 风险 | 概率 | 影响 | 缓解方案 |
|------|------|------|---------|
| Go API 数据格式不匹配 | 中 | 中 | 在 useAPI 添加转换层 |
| 消息丢失 | 低 | 高 | 保留 Mock 作为临时存储 |
| SSE 性能差 | 低 | 中 | 混合模式：聊天用 Mock |
| 认证缺失 | 低 | 中 | 后续补充，非阻塞 |

---

## 🎯 最终建议

**立即启动的方案**: **方案 A - 混合模式**

理由：
1. ✅ 最小化前端改动（符合约束）
2. ✅ 立即可用（今天开始）
3. ✅ 低风险过渡
4. ✅ 为完整迁移争取时间

**一周后迁移**: 完整移到 Go（阶段化集成）

---

**分析完成**: 2026-02-27
**建议采纳**: 方案 A + 逐步升级到方案 B

