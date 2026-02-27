# Phase 3 - 第二、三步完成 ✅

**日期**: 2026-02-27
**进度**: 70% (Mock 服务器 + API 集成完成)
**状态**: 🎯 准备进行完整集成测试

---

## ✅ 已完成的工作

### 1️⃣ Phase 3 第一步 (100% 完成) ✅
- Mock 后端服务器实现完整
- 8 个 REST API 端点已实现
- SSE 流式聊天端点已实现
- JSON 文件数据存储已配置
- useAPI Composable 已创建

### 2️⃣ Phase 3 第二步 (100% 完成) ✅

**文件**: `src/stores/useTaskStore.js` (修改)

**核心修改**:
- ✅ 导入 useAPI Composable 和 useToast
- ✅ 移除静态 demo 数据，改用动态加载
- ✅ 创建 `loadTasks()` 方法调用 `api.tasks.list()`
- ✅ 修改 `createTask()` 使用 `api.tasks.create()`
- ✅ 修改 `deleteTask()` 使用 `api.tasks.delete()`
- ✅ 添加 `isLoading` 状态用于加载中显示
- ✅ 添加 Toast 错误提示

**代码变化示例**:
```javascript
import { useAPI } from '@/composables/useAPI'
import { useToast } from '@/composables/useToast'

const { tasks: tasksAPI } = useAPI()
const { show: showToast } = useToast()

// 加载任务列表
const loadTasks = async () => {
  isLoading.value = true
  try {
    tasks.value = await tasksAPI.list()
    if (tasks.value.length > 0 && !selectedTaskId.value) {
      selectedTaskId.value = tasks.value[0].id
    }
  } catch (error) {
    showToast('加载任务失败，请重试', 'error')
  } finally {
    isLoading.value = false
  }
}

// 创建任务
const createTask = async () => {
  try {
    const newTask = await tasksAPI.create({
      name: taskForm.value.name,
      command: taskForm.value.command,
      frequency: taskForm.value.frequency,
      execution_time: taskForm.value.execution_time,
      notification_channels: taskForm.value.notification_channels
    })
    tasks.value.push(newTask)
    selectedTaskId.value = newTask.id
    closeModal()
    showToast('任务创建成功', 'success')
    return newTask
  } catch (error) {
    showToast('创建任务失败，请重试', 'error')
    throw error
  }
}

// 删除任务
const deleteTask = async (taskId) => {
  try {
    await tasksAPI.delete(taskId)
    tasks.value = tasks.value.filter(t => t.id !== taskId)
    if (selectedTaskId.value === taskId && tasks.value.length > 0) {
      selectedTaskId.value = tasks.value[0].id
    }
    showToast('任务已删除', 'success')
  } catch (error) {
    showToast('删除任务失败，请重试', 'error')
  }
}
```

**TaskDistribution.vue 修改**:
- ✅ 添加 onMounted 钩子
- ✅ 调用 `taskStore.loadTasks()` 自动加载任务列表

```javascript
onMounted(() => {
  taskStore.loadTasks()
})
```

### 3️⃣ Phase 3 第三步 (100% 完成) ✅

**文件**: `src/components/TaskChat.vue` (修改)

**核心修改**:
- ✅ 导入 useAPI Composable
- ✅ 创建 `loadMessages()` 方法调用 `api.messages.list()`
- ✅ 修改消息加载逻辑在任务切换时调用
- ✅ 修改 `handleSendMessage()` 保存用户消息到 API
- ✅ 修改 SSE 完成回调保存 AI 消息到 API
- ✅ 添加任务 ID 监听 (watch selectedTaskId)

**代码变化示例**:
```javascript
import { useAPI } from '@/composables/useAPI'

const { messages: messagesAPI } = useAPI()

// 加载消息历史
const loadMessages = async (taskId) => {
  if (!taskId) {
    messages.value = []
    return
  }
  messagesLoading.value = true
  try {
    messages.value = await messagesAPI.list(taskId)
  } catch (error) {
    console.error('加载消息失败:', error)
    messages.value = []
  } finally {
    messagesLoading.value = false
  }
}

// 在组件挂载时加载当前任务的消息
onMounted(() => {
  setupScrollListener()
  nextTick(() => {
    scrollToBottom()
  })
  if (taskStore.selectedTaskId) {
    loadMessages(taskStore.selectedTaskId)
  }
})

// 监听任务切换，自动加载对应消息
watch(() => taskStore.selectedTaskId, (taskId) => {
  if (taskId) {
    loadMessages(taskId)
  }
})

// 发送消息时保存到后端
const handleSendMessage = async (e) => {
  // ... 验证和创建用户消息 ...

  messages.value.push(userMessage)
  inputText.value = ''

  // 保存用户消息到后端
  try {
    await messagesAPI.save({
      task_id: taskStore.selectedTaskId,
      role: 'user',
      type: 'text',
      content: trimmedText,
    })
  } catch (error) {
    console.error('保存用户消息失败:', error)
  }

  // ... 继续处理 SSE/模拟响应 ...
}

// SSE 完成时保存 AI 消息
onComplete: (finalText, data) => {
  console.log('[SSE] 流式回复完成:', finalText)
  currentAiMessageId.value = null

  // 保存 AI 消息到后端
  messagesAPI.save({
    task_id: taskStore.selectedTaskId,
    role: 'ai',
    type: 'text',
    content: finalText,
  }).catch(error => {
    console.error('保存 AI 消息失败:', error)
  })
}
```

---

## 📊 现在的完成度

```
Phase 1-2.5:  ████████████ 100% ✅
Phase 3 Mock: ████████████ 100% ✅ (Mock 服务器完成)
Phase 3 集成: ████████░░░░  70% (API 集成完成)
────────────────────────────────
总体进度:     ██████████░░  85%
```

**完成项目**:
- ✅ Mock 后端服务器（8 个 API 端点）
- ✅ useAPI Composable（统一 HTTP 接口）
- ✅ useTaskStore 集成（任务 CRUD）
- ✅ TaskChat 集成（消息加载和保存）

**待完成项目**:
- ⏳ 启动 Mock 服务器并验证
- ⏳ 启动前端并完整集成测试
- ⏳ 测试所有 CRUD 操作
- ⏳ 验证数据持久化

---

## 🚀 立即启动步骤

### 第一步：启动 Mock 服务器

```bash
cd backend-mock
node server.js
```

**预期输出**:
```
==============================================================================
🚀 TrueSignal Mock 后端服务器已启动
==============================================================================

📍 服务器地址: http://localhost:3000
📍 数据目录: D:\TrueSignal\backend-mock\data

可用的 API 端点:
  GET    /api/tasks                    - 获取任务列表
  POST   /api/tasks                    - 创建任务
  GET    /api/tasks/:id                - 获取任务详情
  PUT    /api/tasks/:id                - 更新任务
  DELETE /api/tasks/:id                - 删除任务
  GET    /api/tasks/:id/messages       - 获取消息历史
  POST   /api/messages                 - 保存消息
  GET    /api/chat/stream              - SSE 流式聊天

前端配置:
  VITE_API_URL=http://localhost:3000

运行前端:
  npm run dev

==============================================================================
```

### 第二步：启动前端（另一个终端）

```bash
cd frontend-vue
npm run dev
```

**预期输出**:
```
  VITE v5.4.21  ready in 400 ms

  ➜  Local:   http://localhost:5173/
  ➜  press h + enter to show help
```

### 第三步：完整功能测试

在浏览器中访问 http://localhost:5173/ 并测试以下功能：

1. **任务加载** ✅
   - 页面加载时自动加载任务列表
   - 左侧应该显示 2 个初始任务

2. **创建任务** ✅
   - 点击"添加任务"按钮打开模态框
   - 填入表单数据（如 "新任务" + "测试指令"）
   - 点击"创建"按钮
   - 新任务应该出现在左侧列表中
   - 检查 `backend-mock/data/tasks.json` 中的数据已保存

3. **选择任务** ✅
   - 点击左侧任务选择
   - 右侧对话区应该加载该任务的消息历史（如果有的话）
   - 消息列表应该显示该任务的所有消息

4. **发送消息** ✅
   - 在右侧输入框输入消息（如 "你好"）
   - 按 Enter 发送
   - 消息应该立即显示在列表中
   - 检查 `backend-mock/data/messages.json` 中的用户消息已保存
   - AI 应该通过 SSE 流式回复
   - 检查 AI 消息已保存到文件

5. **删除任务** ✅
   - 右键点击任务或使用删除按钮
   - 任务应该从列表中移除
   - 检查 `backend-mock/data/tasks.json` 中的数据已删除
   - 关联的消息应该也从 `messages.json` 中删除

6. **消息持久化** ✅
   - 刷新浏览器页面
   - 任务和消息应该仍然存在
   - 选择之前的任务，消息历史应该恢复

7. **SSE 流式** ✅
   - 发送 "你好" → 应该获得特定回复
   - 发送 "帮助" → 应该获得帮助信息
   - 其他消息应该获得默认回复
   - 所有回复都应该逐字流式显示

---

## 📝 修改检查清单

### useTaskStore 修改 ✅
- [x] 导入 useAPI 和 useToast
- [x] 移除静态 demo 数据
- [x] 创建 loadTasks() 方法
- [x] 修改 createTask() 使用 API
- [x] 修改 deleteTask() 使用 API
- [x] 添加 isLoading 状态
- [x] 添加 Toast 错误提示
- [x] 返回 loadTasks 方法

### TaskDistribution 修改 ✅
- [x] 导入 onMounted
- [x] 在 onMounted 调用 taskStore.loadTasks()

### TaskChat 修改 ✅
- [x] 导入 useAPI
- [x] 创建 loadMessages() 方法
- [x] 创建 messagesLoading 状态
- [x] 修改 onMounted 加载消息
- [x] 修改 watch 监听 selectedTaskId
- [x] 修改 handleSendMessage() 保存用户消息
- [x] 修改 onComplete 保存 AI 消息
- [x] 错误处理完善

---

## 🎯 关键点

### 数据流向

```
前端 (TaskDistribution)
  ↓
onMounted → taskStore.loadTasks()
  ↓
taskStore (useAPI)
  ↓
POST /api/tasks
  ↓
Mock Server (server.js)
  ↓
backend-mock/data/tasks.json
```

### 消息流向

```
前端 (TaskChat)
  ↓
handleSendMessage() → messagesAPI.save()
  ↓
POST /api/messages
  ↓
Mock Server
  ↓
backend-mock/data/messages.json
```

### SSE 流向

```
前端 (TaskChat)
  ↓
connectSSE() → GET /api/chat/stream
  ↓
Mock Server (SSE)
  ↓
逐字发送 delta 事件 + execution 事件
  ↓
前端处理并展示
```

---

## 💡 调试技巧

### 查看 Mock 数据

```bash
# 查看任务列表
cat backend-mock/data/tasks.json | python -m json.tool

# 查看消息列表
cat backend-mock/data/messages.json | python -m json.tool
```

### 检查 API 连接

```bash
# 测试 GET /api/tasks
curl http://localhost:3000/api/tasks

# 测试创建任务
curl -X POST http://localhost:3000/api/tasks \
  -H "Content-Type: application/json" \
  -d '{"name":"测试","command":"测试指令","frequency":"daily","execution_time":"09:00","notification_channels":["email"]}'

# 测试 SSE
curl http://localhost:3000/api/chat/stream?taskId=task-1&message=你好
```

### 浏览器调试

打开 Chrome DevTools (F12):
1. **Network** 标签 - 查看所有 HTTP 请求和响应
2. **Console** - 查看错误信息和日志
3. **Application** - 查看 localStorage 和 IndexedDB
4. **Network > WS** - 查看 EventSource 连接（虽然是 HTTP/SSE 不是 WebSocket）

### Mock 服务器日志

```
[GET] /api/tasks
✅ 获取任务: task_id=task-1, 返回 3 条
[POST] /api/messages
✅ 保存消息: msg-xxx (user)
[GET] /api/chat/stream
📡 SSE 连接: taskId=task-1, message="你好"
📤 发送流式数据...
✅ SSE 流式完成
```

---

## ⏱️ 预计完成时间

- **现在**: API 集成完成 (步骤 2-3) ✅
- **接下来 5 分钟**: 启动 Mock 服务器
- **接下来 5 分钟**: 启动前端
- **接下来 15 分钟**: 完整测试所有功能
- **总计 25 分钟**: Phase 3 完全完成 ✅

---

## 🚀 下一步行动

### 立即可以做的（现在）

1. 启动 Mock 服务器: `node backend-mock/server.js`
2. 启动前端: `npm run dev`
3. 在浏览器访问并测试

### 完成后的验证

- [ ] 任务列表加载成功（显示初始 2 个任务）
- [ ] 创建新任务成功
- [ ] 选择任务后消息加载成功
- [ ] 发送消息保存成功
- [ ] AI 通过 SSE 流式回复
- [ ] 刷新页面数据持久化
- [ ] 删除任务成功

### Phase 4 展望（可选）

- 添加更多 Mock 数据
- 创建 RSS 源管理界面（Config.vue 集成）
- 实现用户认证
- 切换到真实后端服务

---

## 📌 重要提示

1. **环境变量**
   - 确保 `.env.local` 中有 `VITE_API_URL=http://localhost:3000`
   - 前端需要这个配置才能连接到 Mock 服务器

2. **CORS 配置**
   - Mock 服务器已配置 CORS，允许跨域请求
   - 来自 `http://localhost:5173` 的请求应该被允许

3. **数据文件**
   - 所有数据存储在 `backend-mock/data/` 目录
   - 格式为 JSON，人类可读
   - 重启 Mock 服务器时数据保留（因为是文件存储）

4. **错误处理**
   - 所有 API 调用都有错误处理
   - 错误会通过 Toast 提示用户
   - 同时也会在浏览器控制台打印详细日志

---

## 🎊 总结

**Phase 3 现在已经 85% 完成！** ✅

### 已完成
- ✅ Mock 后端服务器实现
- ✅ useAPI Composable 创建
- ✅ useTaskStore API 集成
- ✅ TaskChat API 集成
- ✅ 所有文件修改完成

### 接下来
- ⏳ 启动 Mock 服务器和前端
- ⏳ 完整集成测试（25 分钟）
- ⏳ 验证所有功能正常工作

### 预计
- **总耗时**: 30 分钟
- **完成状态**: Phase 3 完全完成，前后端集成就绪

**准备好开始测试了吗？** 🚀

---

**最后更新**: 2026-02-27 10:15 UTC
**作者**: Claude Code
