# Phase 3 详细执行计划 - 数据持久化与集成

**日期**: 2026-02-26
**版本**: Phase 3 - 完整的前后端集成方案
**状态**: 🎯 执行计划阶段

---

## 📋 Phase 3 核心目标

将前端从"演示状态"升级到"生产就绪"，实现完整的前后端一体化系统。

### 三个阶段

```
阶段 1: 后端 Mock 服务 (1-2 天)
├─ 创建本地 JSON 数据存储
├─ 实现 REST API 端点
├─ 实现 SSE 流式端点
└─ 目的: 前端可以立即测试

↓

阶段 2: 前端 API 集成 (2-3 天)
├─ 创建 useAPI Composable
├─ 修改 useTaskStore 集成 API
├─ 修改 TaskChat 集成 API
└─ 目的: 前端完全依赖 API 数据

↓

阶段 3: 真实后端集成 (可选)
├─ 替换 Mock 为真实数据库
├─ 集成用户认证
├─ 完整的生产部署
└─ 目的: 生产环境就绪
```

---

## 🎯 我的完整计划

### 📍 阶段 1: 后端 Mock 服务 (我来做)

#### 1.1 创建 Mock 服务器
**文件**: `backend-mock/server.js`

功能:
- 提供完整的 REST API 端点
- 提供 SSE 流式端点
- 使用 JSON 文件模拟数据库
- 支持 CRUD 操作

#### 1.2 数据结构设计
**文件**: `backend-mock/data/tasks.json`、`messages.json`

```json
// tasks.json
[
  {
    "id": "task-1",
    "name": "Twitter AI 早报",
    "command": "每天早上9点总结Twitter上关于AI的新闻",
    "frequency": "daily",
    "execution_time": "09:00",
    "notification_channels": ["email"],
    "status": "active",
    "created_at": "2026-02-26T10:00:00Z",
    "updated_at": "2026-02-26T10:00:00Z"
  }
]

// messages.json
[
  {
    "id": "msg-1",
    "task_id": "task-1",
    "role": "user",
    "type": "text",
    "content": "你好",
    "timestamp": "2026-02-26T10:00:00Z"
  }
]
```

#### 1.3 Mock API 端点
```javascript
// GET /api/tasks - 获取任务列表
GET /api/tasks
Response: { data: [...tasks] }

// POST /api/tasks - 创建任务
POST /api/tasks
Body: { name, command, frequency, execution_time, notification_channels }
Response: { data: newTask }

// GET /api/tasks/:id - 获取任务详情
GET /api/tasks/:id
Response: { data: task }

// PUT /api/tasks/:id - 更新任务
PUT /api/tasks/:id
Body: { name, command, ... }
Response: { data: updatedTask }

// DELETE /api/tasks/:id - 删除任务
DELETE /api/tasks/:id
Response: { success: true }

// GET /api/tasks/:id/messages - 获取消息历史
GET /api/tasks/:id/messages?limit=50&offset=0
Response: { data: [...messages] }

// POST /api/messages - 保存消息
POST /api/messages
Body: { task_id, role, type, content }
Response: { data: newMessage }

// GET /api/chat/stream - SSE 流式端点
GET /api/chat/stream?taskId=xxx&message=hello
Response: text/event-stream
```

#### 1.4 SSE 流式实现
```javascript
// 后端发送流式数据
app.get('/api/chat/stream', (req, res) => {
  res.setHeader('Content-Type', 'text/event-stream')

  // 模拟 AI 回复（逐字发送）
  const text = "我已收到你的消息..."
  for (const char of text) {
    res.write(`event: delta\ndata: ${JSON.stringify({type:'delta',content:char})}\n\n`)
    await sleep(50)
  }

  // 发送完成
  res.write(`event: done\ndata: ${JSON.stringify({type:'done'})}\n\n`)
})
```

---

### 📍 阶段 2: 前端 API 集成 (我来做)

#### 2.1 创建 useAPI Composable
**文件**: `src/composables/useAPI.js`

```javascript
export const useAPI = () => {
  const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:3000'

  const request = async (path, options = {}) => {
    const url = `${apiUrl}${path}`
    const response = await fetch(url, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
    })

    if (!response.ok) {
      throw new Error(`API Error: ${response.status}`)
    }

    const data = await response.json()
    return data.data  // 返回真实数据
  }

  return {
    tasks: {
      list: () => request('/api/tasks'),
      get: (id) => request(`/api/tasks/${id}`),
      create: (data) => request('/api/tasks', {
        method: 'POST',
        body: JSON.stringify(data)
      }),
      update: (id, data) => request(`/api/tasks/${id}`, {
        method: 'PUT',
        body: JSON.stringify(data)
      }),
      delete: (id) => request(`/api/tasks/${id}`, {
        method: 'DELETE'
      }),
    },
    messages: {
      list: (taskId, { limit = 50, offset = 0 } = {}) =>
        request(`/api/tasks/${taskId}/messages?limit=${limit}&offset=${offset}`),
      save: (data) => request('/api/messages', {
        method: 'POST',
        body: JSON.stringify(data)
      }),
    },
  }
}
```

#### 2.2 修改 useTaskStore
```javascript
const { tasks: tasksAPI } = useAPI()

const loadTasks = async () => {
  try {
    tasks.value = await tasksAPI.list()
  } catch (error) {
    console.error('加载任务失败:', error)
  }
}

const createTask = async () => {
  try {
    const newTask = await tasksAPI.create(taskForm.value)
    tasks.value.push(newTask)
    selectedTaskId.value = newTask.id
    closeModal()
    return newTask
  } catch (error) {
    console.error('创建任务失败:', error)
    throw error
  }
}

const deleteTask = async (taskId) => {
  try {
    await tasksAPI.delete(taskId)
    tasks.value = tasks.value.filter(t => t.id !== taskId)
  } catch (error) {
    console.error('删除任务失败:', error)
    throw error
  }
}

// 在 setup 中初始化
onMounted(() => {
  loadTasks()
})
```

#### 2.3 修改 TaskChat 集成 API
```javascript
const { messages: messagesAPI } = useAPI()

// 加载任务消息
const loadMessages = async (taskId) => {
  try {
    messages.value = await messagesAPI.list(taskId)
  } catch (error) {
    console.error('加载消息失败:', error)
  }
}

// 保存用户消息
const handleSendMessage = async (e) => {
  // ... 前面的验证代码 ...

  // 保存用户消息到后端
  const userMessage = {
    task_id: taskStore.selectedTaskId,
    role: 'user',
    type: 'text',
    content: trimmedText,
  }

  await messagesAPI.save(userMessage)
  messages.value.push({...userMessage, id: `msg-${Date.now()}`})

  // 处理 SSE 回复
  await handleSSEResponse(trimmedText)
}

// 任务切换时加载消息
watch(() => taskStore.selectedTaskId, (taskId) => {
  if (taskId) {
    loadMessages(taskId)
  }
})
```

---

### 📍 阶段 3: 测试和验证 (我来做)

#### 3.1 本地集成测试
- 启动 Mock 服务器
- 测试所有 CRUD 操作
- 测试 SSE 流式
- 验证数据持久化

#### 3.2 前端测试
- 任务加载和创建
- 消息发送和接收
- SSE 流式回复
- 错误处理

---

## 🔧 具体实施步骤

### 第一天：创建 Mock 服务器

#### 步骤 1.1: 初始化项目结构
```bash
mkdir backend-mock
cd backend-mock
npm init -y
npm install express cors
```

#### 步骤 1.2: 创建服务器文件
创建 `backend-mock/server.js`：
- 基本 Express 服务器
- CORS 配置
- 静态数据加载

#### 步骤 1.3: 实现 CRUD API
- GET /api/tasks
- POST /api/tasks
- PUT /api/tasks/:id
- DELETE /api/tasks/:id
- GET /api/tasks/:id/messages
- POST /api/messages

#### 步骤 1.4: 实现 SSE 端点
- GET /api/chat/stream
- 流式发送 delta 事件
- 完成时发送 done 事件

#### 步骤 1.5: 本地测试
```bash
node server.js
# 验证所有端点
```

---

### 第二天：集成前端 API

#### 步骤 2.1: 创建 useAPI Composable
`src/composables/useAPI.js`

#### 步骤 2.2: 修改 useTaskStore
集成 `api.tasks` 方法：
- loadTasks()
- createTask()
- deleteTask()
- updateTask()

#### 步骤 2.3: 修改 TaskChat
集成 `api.messages` 方法：
- loadMessages()
- saveMessage()

#### 步骤 2.4: 配置环境变量
`.env.local`:
```
VITE_API_URL=http://localhost:3000
```

#### 步骤 2.5: 前端测试
```bash
npm run dev
# 验证各功能
```

---

### 第三天：完整测试和优化

#### 步骤 3.1: 功能测试
- [ ] 创建任务成功
- [ ] 任务持久化成功
- [ ] 加载任务列表成功
- [ ] 删除任务成功
- [ ] 发送消息成功
- [ ] 消息持久化成功
- [ ] 加载消息历史成功
- [ ] SSE 流式回复成功

#### 步骤 3.2: 错误处理测试
- [ ] 网络错误处理
- [ ] 超时处理
- [ ] 无效数据处理
- [ ] API 错误处理

#### 步骤 3.3: 性能测试
- [ ] 大量任务加载
- [ ] 大量消息加载
- [ ] 并发请求处理

#### 步骤 3.4: 优化
- [ ] 添加加载状态
- [ ] 添加错误提示
- [ ] 优化数据结构

---

## 📁 文件清单

### 需要创建的文件

```
backend-mock/
├── server.js                       ✅ 主服务器
├── data/
│   ├── tasks.json                  ✅ 任务数据
│   └── messages.json               ✅ 消息数据
├── routes/
│   ├── tasks.js                    ✅ 任务路由
│   ├── messages.js                 ✅ 消息路由
│   └── chat.js                     ✅ SSE 路由
├── package.json
└── .gitignore

frontend-vue/src/
├── composables/
│   └── useAPI.js                   ✅ API 封装

frontend-vue/
└── .env.local                      ✅ 环境配置
```

### 需要修改的文件

```
frontend-vue/src/
├── stores/
│   └── useTaskStore.js             ✏️  集成 API
├── components/
│   ├── TaskChat.vue                ✏️  集成 API
│   ├── TaskSidebar.vue             ✏️  加载状态
│   └── TaskModal.vue               ✏️  错误处理
```

---

## 🎯 验收标准

### ✅ Mock 服务器验收
- [x] 所有 API 端点都能正常响应
- [x] 数据能正确保存到 JSON 文件
- [x] SSE 端点能正确流式发送数据
- [x] CORS 配置正确
- [x] 错误处理完善

### ✅ 前端集成验收
- [x] useAPI Composable 完成
- [x] useTaskStore 使用 API 数据
- [x] TaskChat 保存和加载消息
- [x] 所有 CRUD 操作成功
- [x] 错误处理完善
- [x] 加载状态正确显示

### ✅ 集成测试验收
- [x] 完整的任务生命周期测试
- [x] 完整的消息生命周期测试
- [x] SSE 流式测试
- [x] 网络错误处理
- [x] 边界情况处理

---

## 📊 时间分配

```
第 1 天 (8 小时):
├─ Mock 服务器设置 (2h)
├─ API 端点实现 (4h)
├─ 本地测试 (2h)
└─ 目标: Mock 服务器就绪

第 2 天 (8 小时):
├─ useAPI Composable (2h)
├─ useTaskStore 集成 (2h)
├─ TaskChat 集成 (2h)
├─ 前端测试 (2h)
└─ 目标: 前端 API 集成完成

第 3 天 (8 小时):
├─ 完整集成测试 (3h)
├─ 错误处理优化 (2h)
├─ 性能优化 (2h)
├─ 文档整理 (1h)
└─ 目标: Phase 3 完成验收
```

---

## 🚀 推荐的启动顺序

### Step 1: Mock 服务器（今天完成）
创建完整的后端 Mock 服务，包括：
- Express 服务器
- CRUD API 端点
- SSE 流式端点
- JSON 文件数据存储

### Step 2: 前端 API 集成（明天完成）
修改前端以使用真实 API：
- 创建 useAPI Composable
- 修改 useTaskStore
- 修改 TaskChat 组件
- 集成错误处理

### Step 3: 完整测试（后天完成）
验证所有功能：
- 功能测试
- 错误处理
- 性能测试
- 文档完善

---

## ⚠️ 注意事项

### Mock 服务器
- 使用 JSON 文件存储数据（简单快速）
- 数据在服务器重启后会重置
- 适合开发和测试
- 不适合长期存储

### 前端集成
- 使用 VITE_API_URL 配置 API 基础 URL
- 所有 API 调用都通过 useAPI 进行
- 添加完善的错误处理
- 使用 Toast 显示错误信息

### 过渡到真实后端
- Mock 服务器的 API 格式与真实后端保持一致
- 只需替换 VITE_API_URL 即可切换到真实后端
- 无需修改前端代码逻辑

---

## 📈 预期成果

### 完成后
✅ 所有任务都能持久化
✅ 所有消息都能持久化
✅ 前端完全依赖 API 数据
✅ 完整的 CRUD 操作
✅ SSE 流式回复正常
✅ 错误处理完善
✅ 可以随时切换到真实后端

### 项目进度
```
Phase 1-2.5: ████████████ 100%
Phase 3 Mock:  ████████████ 100%  ← 我们的目标
Phase 3 后端:  ░░░░░░░░░░░░   0%  ← 真实后端（可选）
Total:         █████████░░░  80%
```

---

## 💡 我的建议

### 立即开始
1. **第 1 天**: 我创建 Mock 服务器
2. **第 2 天**: 我集成前端 API
3. **第 3 天**: 我完整测试验证

### 完成后的选项
- **Option A**: 使用 Mock 继续开发其他功能
- **Option B**: 基于 Mock API 规范实现真实后端
- **Option C**: 混合模式（部分 Mock，部分真实）

### 关键优势
✅ **快速**：3 天完成整个 Phase 3
✅ **完整**：包含所有 CRUD 操作
✅ **灵活**：可随时切换到真实后端
✅ **可测试**：前端可完全独立开发

---

## ✅ 最终检查清单

- [ ] Mock 服务器实现完整
- [ ] 所有 API 端点测试成功
- [ ] useAPI Composable 创建完成
- [ ] useTaskStore 集成 API
- [ ] TaskChat 集成 API
- [ ] 完整功能测试通过
- [ ] 错误处理完善
- [ ] 文档更新完成

---

**准备好开始 Phase 3 了吗？** 🚀

我会按照以下顺序进行：
1. **今天**: 创建 Mock 服务器
2. **明天**: 前端 API 集成
3. **后天**: 完整测试验证

让我们在 3 天内完成 Phase 3！
