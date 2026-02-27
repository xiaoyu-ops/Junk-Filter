# 分发任务模块实现计划 - 审阅版

**文档**: TASK_DISTRIBUTION_IMPLEMENTATION_PLAN.md
**位置**: `D:\TrueSignal\description\`
**状态**: 待审阅

---

## 📌 核心概览

### 业务目标
实现一个 **"指令即任务"** 的 AI 信息过滤系统，用户通过自然语言描述需求，系统自动化处理 RSS 信息流。

### 两个主要界面

**Page A: 创建任务模态框**
```
用户点击"添加任务"
  → 模态框打开(淡入缩放动画)
  → 输入: 任务名称 + 自然语言指令 + 高级设置
  → 点击"立即创建"
  → 后端 LLM 解析指令
  → 任务创建成功
```

**Page B: 任务管理对话界面**
```
左侧: 任务列表 (可选中，显示选中态)
右侧: 对话区域
  ├─ 消息历史 (用户/AI/系统执行总结卡片)
  ├─ 执行状态 (AI 正在分析...)
  └─ 底部输入框 (支持 Shift+Enter 换行)
```

---

## 🏗️ 技术架构

### 前端技术栈
```
Vue 3 (Composition API + <script setup>)
├─ Pinia (状态管理)
├─ Tailwind CSS (样式)
├─ TypeScript (类型安全)
└─ SSE/EventSource (实时流)
```

### 核心组件
```
TaskDistribution.vue (主容器)
├── TaskModal.vue (创建任务模态框)
├── TaskSidebar.vue (任务列表)
└── TaskChat.vue (对话区域)
    ├── ChatMessage.vue (消息)
    ├── ExecutionCard.vue (执行卡片)
    └── ChatInput.vue (输入框)
```

### 状态管理 (Pinia)
```
useTaskStore
├─ tasks[] (任务列表)
├─ selectedTaskId (选中任务)
├─ isModalOpen (模态框状态)
└─ taskForm (创建表单)

useChatStore
├─ messages[] (对话历史)
├─ isAILoading (AI 加载态)
└─ streamingText (流式输出)

useUIStore
├─ darkMode (暗黑模式)
└─ sidebarCollapsed (侧边栏折叠)
```

---

## 📊 API 数据流

### 1. 创建任务
```
POST /api/tasks
Request: {
  name: "Twitter AI 新闻早报",
  command: "每天早上9点总结Twitter上关于AI的新闻，并发送到我的邮箱",
  frequency: "daily",
  execution_time: "09:00",
  notification_channels: ["email"]
}

Response: {
  id: "task_123",
  status: "active",
  first_execution_stream: true
}
```

### 2. 获取任务消息
```
GET /api/tasks/{id}/messages

Response: [
  {
    type: "system_execution",
    content: { summary: "...", status: "success", items_count: 3 }
  },
  {
    type: "user",
    content: "总结一下今天关于 OpenAI 的动态"
  },
  {
    type: "ai",
    content: "OpenAI 动态追踪..."
  }
]
```

### 3. 实时消息流 (SSE)
```
GET /api/tasks/{id}/chat/stream?query=...

Event Stream:
data: "OpenAI 发布了"
data: "新的模型..."
data: "[DONE]"
```

### 4. 更新任务状态
```
PATCH /api/tasks/{id}/status
Request: { status: "paused" }
```

---

## 🎯 功能矩阵

### TaskModal (创建任务)
| 功能 | 实现 | 优先级 |
|------|------|--------|
| 模态框打开/关闭 | Vue ref + Transition | P0 |
| 表单字段 (名称/指令/设置) | v-model 双向绑定 | P0 |
| Sparkles 图标 (指令框) | Material Icons | P0 |
| 高级设置折叠菜单 | `<details>` + 动画 | P0 |
| 表单验证 | useTaskValidation | P0 |
| 提交 & API 调用 | useTaskAPI | P0 |
| 错误提示 | Toast 通知 | P1 |

### TaskSidebar (任务列表)
| 功能 | 实现 | 优先级 |
|------|------|--------|
| 列表渲染 | v-for 循环 | P0 |
| 选中态样式 | 左侧 4px 条 + 背景 | P0 |
| 悬停过渡 | transition-all | P0 |
| 任务切换 | @click selectTask | P0 |
| 添加任务按钮 | 打开 Modal | P0 |
| Tooltip 提示 | 悬停显示 | P1 |

### TaskChat (对话区域)
| 功能 | 实现 | 优先级 |
|------|------|--------|
| 消息列表 | ChatMessage 组件 | P0 |
| 消息类型 (user/ai/system) | 条件渲染 | P0 |
| ExecutionCard 卡片 | 专用组件 | P0 |
| SSE 流式输出 | useChatSSE | P0 |
| AI 加载指示器 | 动态点动画 | P0 |
| 自动滚动底部 | watch + nextTick | P0 |

### ChatInput (输入框)
| 功能 | 实现 | 优先级 |
|------|------|--------|
| 文本输入 | `<input>` | P0 |
| Enter 发送 | @keydown.enter | P0 |
| Shift+Enter 换行 | event.shiftKey 检查 | P0 |
| 发送按钮 | 旋转 send 图标 | P0 |
| 输入框禁用 (AI 加载中) | :disabled 绑定 | P0 |

### ExecutionCard (执行卡片)
| 功能 | 实现 | 优先级 |
|------|------|--------|
| 卡片布局 | Tailwind 网格 | P0 |
| 标题 + 状态开关 | Toggle 组件 | P0 |
| 摘要内容 | 子弹列表 | P0 |
| 标签 (Active/RSS) | Badge 组件 | P0 |
| 配置链接 | @click 事件 | P1 |
| 状态同步 | PATCH API | P0 |

---

## 🎨 动画与交互

### 1. 模态框 (Transition)
```
淡入缩放: scale-95 → scale-100 (300ms)
背景遮罩: bg-black/40 + backdrop-blur-sm
```

### 2. 任务选中 (Transition)
```
背景色变化: white → gray-200 (light mode)
左侧指示条: 4px 深灰色边框
过渡时间: 200ms ease-in-out
```

### 3. AI 加载 (Animation)
```
动态点动画: scale(0) → scale(1)
脉冲效果: ping + dot 组合
```

### 4. 消息滚动
```
新消息到达时自动平滑滚动到底部
使用 behavior: 'smooth'
```

---

## 📋 实现步骤 (优先级顺序)

### Phase 1 (Week 1) - 核心功能
```
Step 1: 创建组件框架
  └─ TaskDistribution.vue (主容器)
  └─ TaskModal.vue
  └─ TaskSidebar.vue
  └─ TaskChat.vue
  └─ ChatMessage.vue
  └─ ChatInput.vue

Step 2: 搭建 Pinia Store
  └─ useTaskStore.ts (任务状态)
  └─ useChatStore.ts (对话状态)

Step 3: 基础 API 集成
  └─ useTaskAPI.ts (fetch 调用)
  └─ useChatSSE.ts (EventSource)

Step 4: 样式与布局
  └─ Tailwind 类名应用
  └─ Dark Mode 适配
```

### Phase 2 (Week 2) - 增强交互
```
Step 5: ExecutionCard 卡片
  └─ 渲染执行总结
  └─ 状态开关同步

Step 6: 动画效果
  └─ Transition 组件
  └─ 加载指示器动画

Step 7: 表单验证
  └─ useTaskValidation.ts
  └─ 错误提示

Step 8: 性能优化
  └─ 虚拟滚动 (长消息列表)
  └─ 消息缓存
```

### Phase 3 (Week 3+) - 增强功能
```
Step 9: 高级功能
  └─ 消息搜索
  └─ 批量操作
  └─ 任务导出/分享

Step 10: 可访问性
  └─ ARIA 标签
  └─ 键盘导航
  └─ 屏幕阅读器适配
```

---

## ⚙️ 技术决策理由

| 决策 | 原因 |
|------|------|
| **Vue 3 Composition API** | 项目已使用，更好的代码组织 |
| **Pinia 状态管理** | 官方推荐，比 Vuex 轻量 |
| **SSE vs WebSocket** | 单向推送够用，实现更简单 |
| **分离 useTaskStore & useChatStore** | 职责清晰，便于隔离和缓存 |
| **Tailwind CSS** | 项目已使用，快速开发 |
| **TypeScript** | 类型安全，更好的开发体验 |

---

## 🔍 关键实现细节

### 1. 任务创建流程
```typescript
// 1. 用户提交表单
const createTask = async () => {
  // 2. 验证表单
  if (!validateForm()) return

  // 3. 调用 API
  const task = await taskAPI.createTask(taskForm.value)

  // 4. 添加到列表
  taskStore.tasks.push(task)

  // 5. 自动选中
  taskStore.selectedTaskId = task.id

  // 6. 获取消息历史 (首次分析结果)
  const messages = await taskAPI.fetchTaskMessages(task.id)
  chatStore.messages = messages

  // 7. 关闭模态框
  taskStore.isModalOpen = false
}
```

### 2. SSE 流式处理
```typescript
const sendMessage = async (query: string) => {
  // 1. 添加用户消息
  chatStore.addMessage({ type: 'user', content: query })

  // 2. 开始 AI 加载
  chatStore.isAILoading = true

  // 3. 创建 AI 消息占位符
  const aiMessageId = chatStore.addMessage({
    type: 'ai',
    content: ''
  })

  // 4. SSE 流处理
  chatSSE.streamMessage(taskId, query, (chunk) => {
    // 逐字符追加
    chatStore.updateLastMessage(chunk)
  })

  // 5. 流结束
  chatStore.isAILoading = false
}
```

### 3. 任务选中态切换
```typescript
const handleSelectTask = (taskId: string) => {
  // 1. 更新选中 ID
  taskStore.selectedTaskId = taskId

  // 2. 清空当前对话 (可选)
  chatStore.clearMessages()

  // 3. 获取该任务的消息历史
  const messages = await taskAPI.fetchTaskMessages(taskId)
  chatStore.messages = messages

  // 4. 自动滚动到底部
  nextTick(() => {
    messageContainer.value?.scrollTo({
      top: messageContainer.value.scrollHeight,
      behavior: 'smooth'
    })
  })
}
```

---

## 📝 文件结构概览

```
frontend-vue/src/
├── components/
│   ├── TaskDistribution.vue          (主容器)
│   ├── TaskModal.vue                 (创建模态框)
│   ├── TaskSidebar.vue               (任务列表)
│   ├── TaskChat.vue                  (对话区)
│   ├── ChatMessage.vue               (消息组件)
│   ├── ExecutionCard.vue             (执行卡片)
│   └── ChatInput.vue                 (输入框)
│
├── stores/
│   ├── useTaskStore.ts               (任务状态)
│   ├── useChatStore.ts               (对话状态)
│   └── useUIStore.ts                 (UI 状态)
│
├── composables/
│   ├── useTaskAPI.ts                 (API 调用)
│   ├── useChatSSE.ts                 (SSE 处理)
│   └── useTaskValidation.ts          (表单验证)
│
├── types/
│   └── task.ts                       (类型定义)
│
└── pages/
    └── TaskDistribution.vue          (页面入口)
```

---

## ✅ 验证检查清单

### 核心功能
- [ ] 创建任务模态框正常打开/关闭
- [ ] 表单字段正确绑定
- [ ] 任务创建 API 调用成功
- [ ] 任务出现在左侧列表
- [ ] 点击任务切换右侧对话
- [ ] SSE 流式消息正确展示
- [ ] ExecutionCard 卡片正确渲染

### 交互体验
- [ ] 模态框淡入缩放动画流畅
- [ ] 任务选中态过渡自然
- [ ] AI 加载指示器清晰
- [ ] 消息自动滚动到底部
- [ ] 输入框快捷键 (Enter/Shift+Enter) 正确

### 样式与适配
- [ ] Light Mode 显示正确
- [ ] Dark Mode 显示正确
- [ ] 响应式布局 (桌面/平板/手机)
- [ ] 所有组件边框/圆角/阴影一致
- [ ] 暗黑模式下文字对比度合理

### 性能与兼容性
- [ ] 大列表 (100+任务) 不卡顿
- [ ] Chrome/Firefox/Safari 兼容
- [ ] 消息历史加载速度 < 2s
- [ ] SSE 连接稳定

---

## 🚀 交付物

1. **完整的组件代码** (7 个 Vue 文件)
2. **状态管理代码** (3 个 Pinia Store)
3. **API & Composable** (3 个)
4. **TypeScript 类型定义**
5. **单元测试** (可选)
6. **使用文档** (README)

---

## 📞 审阅要点

请确认以下几点：

1. **架构选择是否合理？**
   - Vue 3 + Pinia + TypeScript
   - SSE 流式处理

2. **组件划分是否清晰？**
   - 是否需要进一步拆分？
   - Props 传递方式是否合理？

3. **API 数据格式是否正确？**
   - 与后端接口是否匹配？
   - 是否需要调整？

4. **UI/UX 交互是否符合预期？**
   - 动画效果是否满足？
   - 是否需要额外的交互？

5. **实现优先级是否合理？**
   - 是否需要调整 Phase 划分？
   - 是否有其他 blocking items？

---

**完整详细计划请查看**: `TASK_DISTRIBUTION_IMPLEMENTATION_PLAN.md`

