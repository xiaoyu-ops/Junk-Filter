# TrueSignal 分发任务模块 - 完整实现计划

**日期**: 2026-02-26
**模块**: Task Distribution (分发任务)
**状态**: 规划阶段
**优先级**: 高

---

## 📋 需求概览

### 业务流程
```
用户界面
  ↓
点击"添加任务"按钮
  ↓
打开创建任务模态框
  ├─ 输入: 任务名称
  ├─ 输入: 自然语言指令 (带 Sparkles 图标)
  └─ 高级: 调度频率、执行时间、通知渠道
  ↓
点击"立即创建"
  ↓
后端 LLM 解析指令 (转换为结构化任务元数据)
  ↓
任务保存到数据库
  ↓
左侧任务列表实时更新 (新任务出现)
  ↓
右侧对话窗口切换到新任务
  ↓
AI 自动分析关联 RSS 源
  ↓
展示结构化执行总结卡片
  └─ 卡片包含: 摘要、状态开关、配置链接
  ↓
用户可继续与 AI 对话或配置任务
```

---

## 🏗️ 技术架构

### 前端框架选择
```
Vue 3 (Composition API + <script setup>)
├─ 响应式状态管理
├─ 组件化模块化
└─ 完整类型支持

Pinia Store
├─ TaskStore: 任务列表、选中任务、创建表单
├─ ChatStore: 消息历史、实时流状态
└─ UIStore: 模态框状态、加载动画

Tailwind CSS (已有)
├─ Dark Mode 支持
├─ 自定义颜色与组件样式
└─ 响应式布局

SSE / WebSocket
├─ 流式消息输出 (Task Creation LLM 解析)
├─ 实时对话响应
└─ 任务执行状态更新
```

### API 数据流
```
【创建任务流程】
POST /api/tasks
Request:
{
  name: "Twitter AI 新闻早报",
  command: "每天早上9点总结Twitter上关于AI的新闻，并发送到我的邮箱",
  schedule: "0 9 * * *",
  frequency: "daily",
  execution_time: "09:00",
  notification_channels: ["email"],
  context_rss_ids: ["rss_001", "rss_002"]
}

Response:
{
  id: "task_123",
  name: "Twitter AI 新闻早报",
  status: "active",
  created_at: "2026-02-26T10:30:00Z",
  first_execution_stream: true  // 触发 SSE 流
}

【获取任务对话历史】
GET /api/tasks/{id}/messages

Response:
[
  {
    id: "msg_1",
    type: "system_execution",
    timestamp: "2026-02-26T10:30:00Z",
    content: {
      summary: "OpenAI 发布 GPT-4 Turbo...",
      status: "success",
      items_count: 3,
      next_execution: "2026-02-27T09:00:00Z"
    }
  },
  {
    id: "msg_2",
    type: "user",
    content: "总结一下今天关于 OpenAI 的动态",
    timestamp: "2026-02-26T10:35:00Z"
  },
  {
    id: "msg_3",
    type: "ai",
    content: "OpenAI 动态追踪...",
    timestamp: "2026-02-26T10:36:00Z"
  }
]

【更新任务状态】
PATCH /api/tasks/{id}/status
Request:
{
  status: "paused"  // active | paused
}
```

---

## 📁 项目结构规划

```
frontend-vue/src/
├── components/
│   ├── TaskDistribution.vue       # 主页面容器
│   ├── TaskModal.vue              # 创建任务模态框
│   ├── TaskSidebar.vue            # 左侧任务列表
│   ├── TaskChat.vue               # 右侧对话区域
│   ├── ChatMessage.vue            # 单条消息组件
│   ├── ExecutionCard.vue          # 执行总结卡片
│   └── ChatInput.vue              # 底部消息输入框
│
├── stores/
│   ├── useTaskStore.ts            # 任务状态管理
│   ├── useChatStore.ts            # 对话状态管理
│   └── useUIStore.ts              # UI 状态管理
│
├── composables/
│   ├── useTaskAPI.ts              # 任务 API 调用
│   ├── useChatSSE.ts              # SSE 流处理
│   └── useTaskValidation.ts       # 表单验证逻辑
│
├── types/
│   └── task.ts                    # TypeScript 类型定义
│
└── pages/
    └── TaskDistribution.vue       # 页面入口
```

---

## 🎯 核心功能模块

### Module 1: 任务创建模态框 (TaskModal.vue)

**功能需求**:
- [x] 模态框开启/关闭动画 (Transition scale-95 → scale-100)
- [x] 表单字段:
  - 任务名称 (text input)
  - 任务指令 (textarea, 右下角 Sparkles 图标)
  - 高级设置 (折叠菜单):
    - 执行频率 (select: 每日/每周/每小时/仅一次)
    - 执行时间 (time input)
    - 通知渠道 (checkbox: Email/Slack/Telegram)
- [x] 表单验证 (非空检查、长度限制)
- [x] 提交按钮 (深色背景、右侧箭头图标、禁用态)
- [x] 取消按钮 (文字或浅色背景)

**State Management**:
```typescript
// useTaskStore.ts
interface CreateTaskForm {
  name: string
  command: string
  frequency: 'daily' | 'weekly' | 'hourly' | 'once'
  execution_time: string
  notification_channels: string[]  // ['email', 'slack', 'telegram']
}

const taskForm = ref<CreateTaskForm>({
  name: '',
  command: '',
  frequency: 'daily',
  execution_time: '09:00',
  notification_channels: ['email']
})

const isModalOpen = ref(false)
const isCreating = ref(false)
```

**API Integration**:
```typescript
const createTask = async () => {
  isCreating.value = true
  try {
    const response = await fetch('/api/tasks', {
      method: 'POST',
      body: JSON.stringify(taskForm.value)
    })
    const task = await response.json()
    tasks.value.push(task)
    selectedTaskId.value = task.id
    isModalOpen.value = false
    // 触发 AI 首次分析
    await fetchTaskMessages(task.id)
  } finally {
    isCreating.value = false
  }
}
```

---

### Module 2: 任务侧边栏 (TaskSidebar.vue)

**功能需求**:
- [x] 任务列表循环渲染
- [x] 选中态样式:
  - 背景色: 浅灰 (light) / 深灰 (dark)
  - 左侧 4px 深色指示条
- [x] 悬停态过渡动画
- [x] 点击任务切换右侧对话上下文
- [x] "添加任务"按钮:
  - 点击打开 TaskModal
  - 悬停显示 Tooltip (提示文案)

**State Binding**:
```typescript
const selectedTaskId = ref<string | null>(null)

const handleSelectTask = (taskId: string) => {
  selectedTaskId.value = taskId
  // 触发右侧对话加载
  fetchTaskMessages(taskId)
}
```

---

### Module 3: 对话区域 (TaskChat.vue)

**功能需求**:
- [x] 消息流渲染 (ChatMessage 组件)
- [x] 消息类型:
  - `user`: 用户消息 (左侧头像 + 消息气泡)
  - `ai`: AI 消息 (右侧头像 + 消息气泡)
  - `system_execution`: 执行总结卡片 (ExecutionCard)
- [x] 流式消息输出 (SSE):
  - 消息逐字符显示
  - "AI 正在分析..." 动态指示器
- [x] 自动滚动到最新消息
- [x] 底部消息输入框 (ChatInput)

**State**:
```typescript
interface Message {
  id: string
  type: 'user' | 'ai' | 'system_execution'
  content: string | ExecutionContent
  timestamp: string
  stream_complete?: boolean
}

interface ExecutionContent {
  summary: string
  status: 'success' | 'error' | 'pending'
  items_count: number
  next_execution: string
  tags: string[]  // ['Active', 'RSS']
}

const messages = ref<Message[]>([])
const isAILoading = ref(false)
```

---

### Module 4: 执行总结卡片 (ExecutionCard.vue)

**组件结构**:
```vue
<template>
  <div class="bg-gray-50 dark:bg-gray-800/60 rounded-lg p-5">
    <!-- 标题 + 状态开关 -->
    <div class="flex justify-between items-start mb-4">
      <div>
        <h4>OpenAI 动态追踪</h4>
        <p>实时 RSS 监控</p>
      </div>
      <toggle-switch v-model="taskStatus" @change="updateTaskStatus" />
    </div>

    <!-- 摘要内容 -->
    <div class="bg-white dark:bg-gray-900/50 rounded p-3 mb-4">
      <bullet-list :items="executionItems" />
    </div>

    <!-- 标签 + 配置链接 -->
    <div class="flex justify-between items-center">
      <div class="flex gap-2">
        <tag>Active</tag>
        <tag>RSS</tag>
      </div>
      <button @click="configureSource">配置来源</button>
    </div>
  </div>
</template>

<script setup>
const props = defineProps<{ execution: ExecutionContent }>()
const taskStatus = ref(true)

const updateTaskStatus = async () => {
  await fetch(`/api/tasks/${props.execution.task_id}/status`, {
    method: 'PATCH',
    body: JSON.stringify({ status: taskStatus.value ? 'active' : 'paused' })
  })
}
</script>
```

---

### Module 5: 消息输入框 (ChatInput.vue)

**功能需求**:
- [x] 文本输入框
- [x] Placeholder: "输入消息... (Shift+Enter 换行)"
- [x] 发送按钮 (右侧，旋转 -45° 的 send 图标)
- [x] 快捷键:
  - Enter: 发送消息
  - Shift+Enter: 换行
- [x] 自动清空输入框
- [x] 禁用态 (AI 加载中)

**实现**:
```typescript
const inputValue = ref('')

const handleSendMessage = async (event: KeyboardEvent) => {
  if (event.key === 'Enter' && !event.shiftKey) {
    event.preventDefault()

    if (!inputValue.value.trim()) return

    // 添加用户消息
    messages.value.push({
      type: 'user',
      content: inputValue.value,
      timestamp: new Date().toISOString()
    })

    // 发送到 API (SSE 流处理)
    const query = inputValue.value
    inputValue.value = ''

    await sendChatMessage(selectedTaskId.value, query)
  }
}
```

---

## 🎨 UI 动画与样式

### 动画 1: 模态框淡入缩放
```vue
<Transition
  enter-active-class="transition-all duration-300 ease-out"
  enter-from-class="opacity-0 scale-95"
  enter-to-class="opacity-100 scale-100"
  leave-active-class="transition-all duration-200 ease-in"
  leave-from-class="opacity-100 scale-100"
  leave-to-class="opacity-0 scale-95"
>
  <div v-if="isModalOpen" class="modal">...</div>
</Transition>
```

### 动画 2: 任务选中态过渡
```vue
<div
  class="transition-all duration-200"
  :class="isSelected ? 'bg-gray-200 dark:bg-gray-600 border-l-4 border-primary' : 'bg-white dark:bg-gray-700'"
>
  {{ task.name }}
</div>
```

### 动画 3: AI 加载指示器
```vue
<!-- 动态点动画 -->
<div class="flex items-center gap-2">
  <span class="relative flex h-2 w-2">
    <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-blue-400 opacity-75"></span>
    <span class="relative inline-flex rounded-full h-2 w-2 bg-blue-500"></span>
  </span>
  <span>AI 正在分析 RSS 源...</span>
</div>
```

---

## 🔌 API 集成计划

### useTaskAPI.ts (Composable)

```typescript
export function useTaskAPI() {
  // 创建任务
  const createTask = async (form: CreateTaskForm) => {
    return await fetch('/api/tasks', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(form)
    }).then(r => r.json())
  }

  // 获取任务列表
  const fetchTasks = async () => {
    return await fetch('/api/tasks').then(r => r.json())
  }

  // 获取任务消息历史
  const fetchTaskMessages = async (taskId: string) => {
    return await fetch(`/api/tasks/${taskId}/messages`).then(r => r.json())
  }

  // 更新任务状态
  const updateTaskStatus = async (taskId: string, status: string) => {
    return await fetch(`/api/tasks/${taskId}/status`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ status })
    }).then(r => r.json())
  }

  return { createTask, fetchTasks, fetchTaskMessages, updateTaskStatus }
}
```

### useChatSSE.ts (Composable)

```typescript
export function useChatSSE() {
  const streamMessage = (taskId: string, query: string, onChunk: (text: string) => void) => {
    const eventSource = new EventSource(
      `/api/tasks/${taskId}/chat/stream?query=${encodeURIComponent(query)}`
    )

    eventSource.onmessage = (event) => {
      const chunk = event.data
      if (chunk === '[DONE]') {
        eventSource.close()
        return
      }
      onChunk(chunk)
    }

    eventSource.onerror = () => {
      eventSource.close()
    }
  }

  return { streamMessage }
}
```

---

## 📊 Pinia Store 设计

### useTaskStore.ts

```typescript
import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useTaskStore = defineStore('task', () => {
  // State
  const tasks = ref<Task[]>([])
  const selectedTaskId = ref<string | null>(null)
  const isModalOpen = ref(false)
  const isCreating = ref(false)

  const taskForm = ref<CreateTaskForm>({
    name: '',
    command: '',
    frequency: 'daily',
    execution_time: '09:00',
    notification_channels: ['email']
  })

  // Getters
  const selectedTask = computed(() =>
    tasks.value.find(t => t.id === selectedTaskId.value)
  )

  // Actions
  const openModal = () => { isModalOpen.value = true }
  const closeModal = () => { isModalOpen.value = false }

  const selectTask = (id: string) => { selectedTaskId.value = id }

  const resetForm = () => {
    taskForm.value = { name: '', command: '', ... }
  }

  return {
    tasks, selectedTaskId, isModalOpen, isCreating, taskForm,
    selectedTask,
    openModal, closeModal, selectTask, resetForm
  }
})
```

### useChatStore.ts

```typescript
export const useChatStore = defineStore('chat', () => {
  const messages = ref<Message[]>([])
  const isAILoading = ref(false)
  const streamingText = ref('')

  const addMessage = (message: Message) => {
    messages.value.push(message)
  }

  const updateLastMessage = (content: string) => {
    if (messages.value.length > 0) {
      messages.value[messages.value.length - 1].content = content
    }
  }

  const clearMessages = () => {
    messages.value = []
  }

  return {
    messages, isAILoading, streamingText,
    addMessage, updateLastMessage, clearMessages
  }
})
```

---

## 🔄 交互流程时序图

```
用户                        前端                        后端
 │                          │                           │
 ├─ 点击"添加任务"────────→ │                           │
 │                          ├─ 打开 Modal              │
 │                          │  (Transition)            │
 │                          │                           │
 ├─ 输入表单数据────────→ │                           │
 │                          ├─ v-model 双向绑定        │
 │                          │                           │
 ├─ 点击"立即创建"────────→ │                           │
 │                          ├─ 验证表单────────────────→│
 │                          │                           ├─ 解析指令 (LLM)
 │                          │                           ├─ 生成 Task
 │                          │                           │
 │                          │← 201 Created, SSE Stream │
 │                          │  (首次分析开始)           │
 │                          │                           │
 │                          ├─ 关闭 Modal              │
 │                          ├─ 添加任务到列表          │
 │                          ├─ 切换选中任务            │
 │                          │                           │
 │← 任务出现在列表          │                           │
 │   右侧对话更新           │                           │
 │   AI 分析中...          │                           │
 │                          │← SSE: 执行总结卡片     │
 │                          ├─ 渲染 ExecutionCard     │
 │                          │                           │
 ├─ 点击任务/输入问题──────→ │                           │
 │                          ├─ 获取消息历史────────────→│
 │                          │                           ├─ 查询消息
 │                          │← 历史消息                │
 │                          ├─ 渲染对话               │
 │                          │                           │
 │                          ├─ 发送聊天消息────────────→│
 │                          │                           ├─ LLM 生成回复
 │                          │← SSE: AI 流式输出      │
 │                          ├─ 逐字符显示             │
 │                          │                           │
 └─ 继续对话────────────────→ └─────────────────────────┘
```

---

## ✅ 实现检查清单

### Phase 1: 组件框架搭建
- [ ] TaskDistribution.vue (主容器)
- [ ] TaskModal.vue (模态框)
- [ ] TaskSidebar.vue (侧边栏)
- [ ] TaskChat.vue (对话区)
- [ ] ChatMessage.vue (消息组件)
- [ ] ExecutionCard.vue (卡片)
- [ ] ChatInput.vue (输入框)

### Phase 2: 状态管理
- [ ] useTaskStore.ts (任务状态)
- [ ] useChatStore.ts (对话状态)
- [ ] useUIStore.ts (UI 状态)

### Phase 3: API 集成
- [ ] useTaskAPI.ts (任务 API)
- [ ] useChatSSE.ts (SSE 处理)
- [ ] useTaskValidation.ts (表单验证)

### Phase 4: 动画与样式
- [ ] 模态框过渡动画
- [ ] 任务选中态过渡
- [ ] 加载指示器动画
- [ ] 暗黑模式适配
- [ ] 响应式布局

### Phase 5: 测试与优化
- [ ] 功能测试
- [ ] 性能优化
- [ ] 浏览器兼容性
- [ ] 可访问性 (A11y)

---

## 📝 类型定义 (types/task.ts)

```typescript
export interface Task {
  id: string
  name: string
  command: string
  schedule: string  // Cron 表达式
  frequency: 'daily' | 'weekly' | 'hourly' | 'once'
  execution_time: string
  notification_channels: string[]
  status: 'active' | 'paused'
  rss_ids?: string[]
  created_at: string
  updated_at: string
}

export interface Message {
  id: string
  task_id: string
  type: 'user' | 'ai' | 'system_execution'
  content: string | ExecutionContent
  timestamp: string
  stream_complete?: boolean
}

export interface ExecutionContent {
  summary: string
  status: 'success' | 'error' | 'pending'
  items_count: number
  next_execution: string
  tags: string[]
}

export interface CreateTaskForm {
  name: string
  command: string
  frequency: 'daily' | 'weekly' | 'hourly' | 'once'
  execution_time: string
  notification_channels: string[]
}
```

---

## 🚀 实现优先级

1. **P0 - 核心功能** (Week 1)
   - TaskModal 创建任务
   - TaskSidebar 任务列表
   - TaskChat 对话展示
   - API 基础集成

2. **P1 - 交互优化** (Week 2)
   - SSE 流式输出
   - ExecutionCard 卡片渲染
   - 状态管理完善
   - 动画过渡效果

3. **P2 - 增强功能** (Week 3)
   - 消息搜索/筛选
   - 批量操作
   - 导出/分享任务
   - 高级配置面板

---

## 💡 技术要点

### SSE 流式处理
```typescript
// 前端处理 SSE 流
const eventSource = new EventSource('/api/stream')
eventSource.onmessage = (event) => {
  const chunk = JSON.parse(event.data)
  updateUIWithChunk(chunk)
}
```

### 实时滚动到底部
```typescript
const messageContainer = ref<HTMLElement>()
watch(() => messages.value.length, () => {
  nextTick(() => {
    messageContainer.value?.scrollTo({
      top: messageContainer.value.scrollHeight,
      behavior: 'smooth'
    })
  })
})
```

### 表单验证
```typescript
const validateForm = () => {
  const errors: Record<string, string> = {}
  if (!form.name.trim()) errors.name = '任务名称不能为空'
  if (!form.command.trim()) errors.command = '任务指令不能为空'
  return Object.keys(errors).length === 0
}
```

---

## 📌 关键决策

1. **为什么选 Vue 3 + Pinia？**
   - 项目已使用 Vue 3
   - Pinia 是官方推荐的状态管理库
   - 更好的 TypeScript 支持

2. **为什么使用 SSE 而非 WebSocket？**
   - SSE 单向推送足以满足需求
   - 实现更简单，浏览器原生支持
   - 减少服务器复杂度

3. **为什么分离 useTaskStore 和 useChatStore？**
   - 职责清晰（任务元数据 vs 对话数据）
   - 便于数据隔离和缓存
   - 支持多任务并行管理

---

## 📚 参考链接

- [Vue 3 Composition API](https://vuejs.org/guide/extras/composition-api-faq.html)
- [Pinia 官方文档](https://pinia.vuejs.org/)
- [Server-Sent Events MDN](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events)
- [Tailwind Transitions](https://tailwindcss.com/docs/transition-property)

