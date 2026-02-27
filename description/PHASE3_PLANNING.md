# Phase 3 规划文档 - 消息持久化 & 前后端集成

**日期**: 2026-02-26
**阶段**: Phase 3 - 生产就绪的完整系统
**状态**: 🎯 规划阶段

---

## 📊 前端完成度评估

### ✅ Phase 1-2.5 已完成的功能

#### 核心 UI 组件 (100%)
- ✅ TaskDistribution.vue (主容器)
- ✅ TaskSidebar.vue (左侧任务列表)
- ✅ TaskModal.vue (创建任务)
- ✅ TaskChat.vue (消息对话)
- ✅ ChatMessage.vue (消息气泡)
- ✅ ExecutionCard.vue (执行卡片)

#### 状态管理 (80%)
- ✅ useTaskStore.js (基础任务管理)
- ⏳ 需要扩展：消息持久化、任务详情

#### Composables (90%)
- ✅ useMarkdown.js (Markdown 渲染)
- ✅ useScrollLock.js (智能滚动)
- ✅ useSSE.js (流式回复)
- ⏳ 需要新增：useAPI.js (API 封装)

#### 功能完整性
- ✅ 创建任务 (模式)
- ✅ 选中任务
- ✅ 发送消息
- ✅ 流式回复 (SSE)
- ✅ Markdown 渲染
- ✅ 暗黑模式
- ⏳ 需要：消息持久化
- ⏳ 需要：任务编辑/删除
- ⏳ 需要：消息搜索

#### 工程质量 (85%)
- ✅ Tailwind CSS 样式完整
- ✅ 完整的错误处理
- ✅ JSDoc 注释
- ✅ 响应式设计
- ⏳ 需要：单元测试
- ⏳ 需要：E2E 测试

---

## 🎯 Phase 3 的核心目标

Phase 3 不是为前端添加新功能，而是为**前后端集成和生产就绪**做准备。

### 三个核心方向

```
Phase 3 = 数据持久化 + API 集成 + 测试覆盖
```

---

## 📋 Phase 3 详细规划

### 1️⃣ 消息持久化 (20% 工作量)

#### 数据库设计
```sql
-- 任务表 (tasks)
CREATE TABLE tasks (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL,
  name VARCHAR(255) NOT NULL,
  command TEXT NOT NULL,
  frequency VARCHAR(50),
  execution_time VARCHAR(5),
  notification_channels JSONB,
  status VARCHAR(50),
  created_at TIMESTAMP,
  updated_at TIMESTAMP
)

-- 消息表 (messages)
CREATE TABLE messages (
  id UUID PRIMARY KEY,
  task_id UUID NOT NULL REFERENCES tasks(id),
  role VARCHAR(50),        -- 'user' | 'ai'
  type VARCHAR(50),        -- 'text' | 'error' | 'execution'
  content TEXT,
  execution_data JSONB,    -- 用于 type='execution'
  created_at TIMESTAMP
)

-- 执行记录表 (executions)
CREATE TABLE executions (
  id UUID PRIMARY KEY,
  task_id UUID NOT NULL REFERENCES tasks(id),
  status VARCHAR(50),
  item_count INT,
  summary TEXT,
  error_message TEXT,
  started_at TIMESTAMP,
  completed_at TIMESTAMP
)
```

#### 前端改动 (useTaskStore.js 扩展)
```javascript
// 新增方法
const loadTaskMessages = async (taskId) => {
  // 从后端加载消息历史
  const response = await fetch(`/api/tasks/${taskId}/messages`)
  const messages = await response.json()
  return messages
}

const saveMessage = async (taskId, message) => {
  // 保存用户/AI 消息
  const response = await fetch(`/api/messages`, {
    method: 'POST',
    body: JSON.stringify({ taskId, ...message })
  })
  return response.json()
}

// 当任务切换时加载消息
watch(() => taskStore.selectedTaskId, async (taskId) => {
  if (taskId) {
    messages.value = await loadTaskMessages(taskId)
  }
})
```

---

### 2️⃣ 前后端 API 集成 (60% 工作量)

#### 后端 API 规范设计

| 端点 | 方法 | 功能 | 参数 |
|------|------|------|------|
| `/api/auth/login` | POST | 用户登录 | email, password |
| `/api/auth/register` | POST | 用户注册 | email, password |
| `/api/tasks` | GET | 获取任务列表 | - |
| `/api/tasks` | POST | 创建任务 | name, command, frequency, ... |
| `/api/tasks/:id` | GET | 获取任务详情 | - |
| `/api/tasks/:id` | PUT | 更新任务 | name, command, ... |
| `/api/tasks/:id` | DELETE | 删除任务 | - |
| `/api/tasks/:id/messages` | GET | 获取消息历史 | limit, offset |
| `/api/messages` | POST | 保存消息 | taskId, role, content, type |
| `/api/chat/stream` | GET | 流式聊天 | taskId, message |
| `/api/tasks/:id/execute` | POST | 手动执行任务 | - |

#### 创建 useAPI Composable (新文件)

**文件**: `src/composables/useAPI.js`

```javascript
/**
 * useAPI Composable
 * 统一的 API 调用接口，处理认证、错误、重试等
 */
export const useAPI = () => {
  const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:3000'

  /**
   * 发送 API 请求
   */
  const request = async (path, options = {}) => {
    const url = `${apiUrl}${path}`
    const token = localStorage.getItem('authToken')

    const response = await fetch(url, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...(token && { Authorization: `Bearer ${token}` }),
        ...options.headers,
      },
    })

    if (!response.ok) {
      if (response.status === 401) {
        // 认证失败，清除 token 并重定向
        localStorage.removeItem('authToken')
        window.location.href = '/login'
      }
      throw new Error(`API 错误: ${response.status}`)
    }

    return response.json()
  }

  // 任务相关 API
  const tasks = {
    list: () => request('/api/tasks'),
    get: (id) => request(`/api/tasks/${id}`),
    create: (data) => request('/api/tasks', { method: 'POST', body: JSON.stringify(data) }),
    update: (id, data) => request(`/api/tasks/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    delete: (id) => request(`/api/tasks/${id}`, { method: 'DELETE' }),
    execute: (id) => request(`/api/tasks/${id}/execute`, { method: 'POST' }),
  }

  // 消息相关 API
  const messages = {
    list: (taskId, options = {}) => {
      const params = new URLSearchParams(options)
      return request(`/api/tasks/${taskId}/messages?${params}`)
    },
    save: (data) => request('/api/messages', { method: 'POST', body: JSON.stringify(data) }),
  }

  // 认证相关 API
  const auth = {
    login: (email, password) => request('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password })
    }),
    register: (email, password) => request('/api/auth/register', {
      method: 'POST',
      body: JSON.stringify({ email, password })
    }),
  }

  return {
    request,
    tasks,
    messages,
    auth,
  }
}
```

#### 修改 TaskChat.vue 集成 API

```javascript
// 从 API 加载任务消息
const loadMessages = async (taskId) => {
  try {
    const { messages: loadedMessages } = await api.messages.list(taskId)
    messages.value = loadedMessages
  } catch (error) {
    console.error('加载消息失败:', error)
  }
}

// 保存用户消息到后端
const handleSendMessage = async (e) => {
  // ... 前面的代码 ...

  // 保存用户消息到数据库
  await api.messages.save({
    taskId: taskStore.selectedTaskId,
    role: 'user',
    type: 'text',
    content: trimmedText,
  })

  // 然后处理 SSE 流式回复
  await handleSSEResponse(trimmedText)
}

// 任务切换时加载消息
watch(() => taskStore.selectedTaskId, (taskId) => {
  if (taskId) {
    loadMessages(taskId)
  }
})
```

#### 修改 useTaskStore.js 集成 API

```javascript
const api = useAPI()

// 创建任务时调用后端 API
const createTask = async () => {
  try {
    const newTask = await api.tasks.create(taskForm.value)
    tasks.value.push(newTask)
    selectedTaskId.value = newTask.id
    closeModal()
    return newTask
  } catch (error) {
    console.error('创建任务失败:', error)
    throw error
  }
}

// 删除任务
const deleteTask = async (taskId) => {
  try {
    await api.tasks.delete(taskId)
    tasks.value = tasks.value.filter(t => t.id !== taskId)
  } catch (error) {
    console.error('删除任务失败:', error)
    throw error
  }
}

// 更新任务状态
const updateTaskStatus = async (taskId, status) => {
  try {
    await api.tasks.update(taskId, { status })
    const task = tasks.value.find(t => t.id === taskId)
    if (task) task.status = status
  } catch (error) {
    console.error('更新任务失败:', error)
    throw error
  }
}
```

---

### 3️⃣ 测试覆盖 (20% 工作量)

#### 单元测试 (Vitest)

**文件**: `src/__tests__/composables/useAPI.test.js`

```javascript
import { describe, it, expect, vi } from 'vitest'
import { useAPI } from '@/composables/useAPI'

describe('useAPI', () => {
  it('应该成功创建任务', async () => {
    global.fetch = vi.fn(() =>
      Promise.resolve({
        ok: true,
        json: () => Promise.resolve({ id: 'task-1', name: '测试任务' })
      })
    )

    const { tasks } = useAPI()
    const result = await tasks.create({ name: '测试任务' })

    expect(result.id).toBe('task-1')
    expect(result.name).toBe('测试任务')
  })

  it('应该处理 API 错误', async () => {
    global.fetch = vi.fn(() =>
      Promise.resolve({
        ok: false,
        status: 500
      })
    )

    const { tasks } = useAPI()
    await expect(tasks.create({})).rejects.toThrow()
  })
})
```

#### E2E 测试 (Playwright)

**文件**: `tests/e2e/task-flow.spec.ts`

```typescript
import { test, expect } from '@playwright/test'

test.describe('任务流程', () => {
  test('应该完成完整的任务创建和消息发送流程', async ({ page }) => {
    // 1. 导航到任务页面
    await page.goto('http://localhost:5173/task')

    // 2. 创建任务
    await page.click('button:has-text("添加任务")')
    await page.fill('input[placeholder*="任务名称"]', '测试任务')
    await page.fill('textarea[placeholder*="任务指令"]', '这是测试指令')
    await page.click('button:has-text("立即创建")')

    // 3. 验证任务创建成功
    await expect(page.locator('text=测试任务')).toBeVisible()

    // 4. 发送消息
    await page.fill('textarea[placeholder*="输入消息"]', '你好')
    await page.click('button[aria-label*="send"]')

    // 5. 验证消息显示
    await expect(page.locator('text=你好')).toBeVisible()
    await expect(page.locator('text=.*AI助手.*')).toBeVisible()
  })
})
```

---

## 🔄 Phase 3 的实施路线

### 第一周：数据库和 API 设计

**后端工作**:
1. 设计数据库 schema
2. 实现认证系统 (JWT)
3. 实现任务 CRUD API
4. 实现消息保存 API
5. 调试 SSE 端点

**前端准备**:
1. 审查 API 规范
2. 准备环境变量
3. 测试 API 端点

### 第二周：前端集成

**前端工作**:
1. 创建 useAPI Composable
2. 修改 useTaskStore 集成 API
3. 修改 TaskChat 集成 API
4. 修改 TaskModal 集成 API
5. 测试所有 API 端点

**测试**:
1. 手动测试各功能
2. 添加错误处理
3. 测试边界情况

### 第三周：测试和优化

**测试**:
1. 编写单元测试
2. 编写 E2E 测试
3. 性能测试
4. 浏览器兼容性测试

**优化**:
1. 缓存策略
2. 错误恢复
3. 用户体验优化

---

## 📦 Phase 3 需要修改的文件

### 新建文件
```
src/
├── composables/
│   └── useAPI.js                  ✅ (新建)
│
└── __tests__/
    ├── composables/
    │   └── useAPI.test.js         ✅ (新建)
    └── e2e/
        └── task-flow.spec.ts      ✅ (新建)
```

### 修改文件
```
src/
├── stores/
│   └── useTaskStore.js            ✏️  (集成 API)
│
└── components/
    ├── TaskChat.vue               ✏️  (集成 API)
    ├── TaskModal.vue              ✏️  (集成 API)
    └── TaskSidebar.vue            ✏️  (集成 API)

frontend-vue/
└── .env.local                     ✏️  (配置 API 端点)
```

---

## 🎯 前端完成度总体评估

### ✅ 已完成 (90%)

#### UI/UX 层 (95%)
- ✅ 所有核心页面设计完成
- ✅ 响应式布局完成
- ✅ 暗黑模式完成
- ✅ 动画和过渡完成
- ✅ 表单验证完成

#### 功能层 (85%)
- ✅ 任务创建/选中/显示
- ✅ 消息发送/显示
- ✅ Markdown 渲染
- ✅ SSE 流式回复
- ✅ 执行卡片显示
- ⏳ 消息持久化 (需要后端)
- ⏳ 任务编辑/删除 (需要后端)

#### 代码质量 (80%)
- ✅ 代码结构清晰
- ✅ 注释完整
- ✅ 错误处理
- ⏳ 单元测试 (需要完善)
- ⏳ E2E 测试 (需要添加)

#### 性能 (85%)
- ✅ 虚拟滚动考虑 (可选优化)
- ✅ 懒加载支持
- ✅ 连接复用
- ⏳ 缓存策略 (后端集成后)

---

## 📊 前后端集成的依赖关系

```
Phase 1-2.5: 前端独立开发完成 ✅
    ↓
Phase 3: 前后端集成
    ├─→ 后端: API 实现
    ├─→ 前端: API 调用
    ├─→ 联调: 端到端测试
    └─→ 优化: 性能和稳定性
```

---

## 💡 可以立即着手的工作

### 前端可以做的 (不需要后端)

✅ **现在就可以做**:
1. 创建 useAPI Composable (假数据)
2. 编写单元测试框架
3. 编写 E2E 测试框架
4. 优化性能 (虚拟滚动等)
5. 添加更多功能 (消息搜索、导出等)

### 需要后端配合的工作

⏳ **等待后端**:
1. 认证系统 (登录/注册)
2. 数据库存储
3. API 端点实现
4. SSE 流式实现

---

## 🔌 后端集成清单

### 后端需要实现的

| 功能 | 优先级 | 预估工期 |
|------|--------|---------|
| 数据库设计 | P0 | 2天 |
| 认证系统 (JWT) | P0 | 2天 |
| 任务 CRUD API | P0 | 3天 |
| 消息保存 API | P0 | 2天 |
| SSE 流式端点 | P0 | 2天 |
| 错误处理和日志 | P1 | 2天 |
| 速率限制 | P1 | 1天 |
| 缓存策略 | P2 | 2天 |

**总计**: ~16 天

### 后端的 API 规范

需要确定:
- [ ] 请求/响应格式
- [ ] 错误代码
- [ ] 认证方式 (JWT vs Session)
- [ ] 速率限制
- [ ] CORS 配置
- [ ] 日志级别

---

## 🚀 建议的集成顺序

### 第 1 阶段：基础集成 (1 周)

1. **后端**: 实现基本的任务 CRUD API
2. **前端**: 创建 useAPI 并集成任务 API
3. **联调**: 测试任务的创建/读取/更新/删除

### 第 2 阶段：消息集成 (1 周)

1. **后端**: 实现消息保存和查询 API
2. **前端**: 集成消息 API，保存用户消息
3. **联调**: 测试消息持久化

### 第 3 阶段：认证集成 (3 天)

1. **后端**: 实现 JWT 认证系统
2. **前端**: 添加登录/注册页面，集成认证
3. **联调**: 完整的登录-使用流程

### 第 4 阶段：优化和测试 (1 周)

1. **性能优化**: 缓存、虚拟滚动等
2. **测试覆盖**: 单元测试和 E2E 测试
3. **错误处理**: 完善错误提示和恢复

---

## ✅ Phase 3 完成标志

Phase 3 完成当满足以下条件:

- [ ] 所有任务都能持久化到数据库
- [ ] 所有消息都能持久化到数据库
- [ ] 用户认证工作正常
- [ ] 完整的前后端联调完成
- [ ] 单元测试覆盖率 > 80%
- [ ] E2E 测试覆盖主要流程
- [ ] 性能达到生产标准
- [ ] 所有边界情况都能正确处理

---

## 📈 整体项目进度

```
Phase 1 (核心功能)          ████████████ 100% ✅
Phase 2 (消息交互)          ████████████ 100% ✅
Phase 2.5 (SSE 流式)        ████████████ 100% ✅
Phase 3 (持久化和集成)      ░░░░░░░░░░░░   0% ⏳

前端完成度:                 ██████████░░  90%
后端准备度:                 ░░░░░░░░░░░░   0%
集成准备度:                 ████░░░░░░░░  30%

总体项目进度:               ██████░░░░░░  50%
```

---

## 🎓 总结

### 前端现状
✅ **前端已经 90% 完成**
- 所有 UI 组件完成
- 所有交互逻辑完成
- 所有样式和动画完成
- SSE 流式支持完成

### 能否开始后端集成
✅ **完全可以！**
- 前端代码已稳定
- API 接口已明确
- 后端可独立开发
- 前后端可并行开发

### 后端需要做什么
需要后端实现:
1. 数据库设计和初始化
2. REST API 端点
3. SSE 流式端点
4. 认证和授权
5. 错误处理和日志

### 最短完成时间
- **前端调整**: 2-3 天 (集成 API)
- **后端实现**: 2-3 周 (API + DB)
- **联调优化**: 1 周
- **总计**: 约 1 个月达到生产就绪

---

## 📞 建议的下一步

### 立即行动

1. **规范化 API 设计**
   - 确定所有端点
   - 确定请求/响应格式
   - 确定错误代码

2. **启动后端开发**
   - 数据库设计
   - 框架选择
   - API 模板

3. **前端准备**
   - 创建 useAPI Composable
   - 准备测试框架
   - 配置环境变量

### 并行进行

- 前端开发测试框架
- 后端实现 API
- 两周后开始联调

---

**前端已准备就绪，随时可以开始前后端集成！** 🚀
