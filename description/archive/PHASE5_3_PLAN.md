# Phase 5.3 前端全量接入真实后端 - 详细执行计划

**目标**：将 TrueSignal 前端从 Mock 模式切换到真实后端，实现两大核心功能：
1. 配置中心的配置能够流向后端并被使用
2. 分发任务处能够调用 AI 进行实时对话评估

---

## 📋 执行计划概览

### 第一阶段：环境和配置切换 (15min)
- [ ] 检查并更新 `frontend-vue/.env.local` - API 端点指向 localhost:8080
- [ ] 禁用所有 Mock 模式相关代码
- [ ] 验证网络连接正常

### 第二阶段：API 层适配 (30min)
- [ ] 分析现有 `useAPI.js` 的 Mock 数据结构
- [ ] 分析真实后端返回的数据格式（已验证过的 API）
- [ ] 创建数据转换函数以处理格式差异
- [ ] 更新所有 API 调用逻辑

### 第三阶段：配置中心改造 (45min)
**核心需求**：用户在配置中心设置的配置能流向后端并被使用

**涉及文件**：
- `src/components/Config.vue` - 配置界面
- `src/stores/useConfigStore.js` - 配置状态管理
- 后端 `/api/evaluations` 调用时需要传入这些配置

**任务列表**：
- [ ] 检查 Config.vue 现有的配置字段
- [ ] 验证 useConfigStore 中的配置项（model、temperature、topP 等）
- [ ] 修改 useAPI.js 中的评估请求，将 store 中的配置作为参数发送到后端
- [ ] 在 Python 后端的 ContentEvaluationAgent 中使用这些配置参数
- [ ] 测试：修改配置 → 发起评估 → 验证后端是否使用了新配置

### 第四阶段：分发任务聊天流改造 (60min)
**核心需求**：在分发任务页面，用户可以实时与 AI 对话并获得评估

**涉及文件**：
- `src/components/TaskChat.vue` - 聊天界面
- `src/composables/useAPI.js` - API 调用逻辑
- 后端 `/api/chat/stream` SSE 流

**任务列表**：
- [ ] 分析 TaskChat.vue 现有的消息显示逻辑
- [ ] 更新 `useAPI.js` 中的 `chatStream()` 函数：
  - 改为调用真实的 `/api/chat/stream` 端点
  - 正确处理 SSE 事件流
  - 解析事件中的 `status`, `result`, `error` 字段
- [ ] 修改消息保存逻辑：
  - 用户消息 → POST 到 `/api/tasks/:id/messages`
  - AI 响应 → 也保存到 `/api/tasks/:id/messages`
- [ ] 更新 UI 以展示：
  - 聊天过程中的处理状态（processing → completed）
  - 评估结果（innovation_score, depth_score, decision）
  - 历史消息列表（从 `/api/tasks/:id/messages` 加载）
- [ ] 测试完整流程：发送消息 → 流式接收评估 → 显示结果

### 第五阶段：数据流验证 (30min)
- [ ] 验证配置创建任务时是否正确流向后端
- [ ] 验证 SSE 流式数据是否正确解析并显示
- [ ] 验证历史消息是否正确加载
- [ ] 检查控制台网络请求（F12 开发者工具）

### 第六阶段：Bug 修复和优化 (30min)
- [ ] 修复任何格式不匹配的 Bug
- [ ] 优化 SSE 流处理性能
- [ ] 测试边界情况（网络断开、超时、错误响应）

---

## 🔄 数据流映射

### 配置流向：
```
Config.vue (用户修改配置)
    ↓
useConfigStore (保存配置状态)
    ↓
useAPI.js (调用 /api/evaluate 时，传入 temperature, topP 等)
    ↓
Go Backend (接收配置参数)
    ↓
Python FastAPI (传入 ContentEvaluationAgent)
    ↓
LLM API (gpt-5.2，使用指定的 temperature 等参数)
```

### 聊天评估流向：
```
TaskChat.vue (用户输入消息)
    ↓
useAPI.js (POST /api/tasks/:id/messages 保存用户消息)
    ↓
Go Backend (保存到 PostgreSQL)
    ↓
useAPI.js (GET /api/chat/stream 启动 SSE)
    ↓
Go Backend (调用 Python /api/evaluate/stream)
    ↓
Python FastAPI (流式返回评估进度)
    ↓
Go Backend (转发 SSE 事件)
    ↓
TaskChat.vue (实时显示 processing → completed)
    ↓
useAPI.js (POST /api/tasks/:id/messages 保存 AI 响应)
    ↓
Go Backend (保存到 PostgreSQL)
```

---

## 📊 关键 API 端点验证清单

**已验证** ✅：
- [x] POST /api/evaluate - 返回评估结果
- [x] GET /api/chat/stream - SSE 流式评估
- [x] POST /api/tasks/:id/messages - 保存消息
- [x] GET /api/tasks/:id/messages - 获取消息列表

**前端需要新增调用**：
- [ ] 从 useConfigStore 中读取配置并传入评估请求
- [ ] 正确处理 SSE 事件中的不同状态
- [ ] 处理消息历史加载时的格式转换

---

## ⚠️ 已知风险点

1. **SSE 事件格式差异**
   - Mock 可能返回 `result` 字段包含完整评估对象
   - 真实返回可能是分步骤的事件流
   - **解决**：需要在 useAPI.js 中实现事件聚合逻辑

2. **消息字段映射**
   - Mock 消息可能有不同的字段名或结构
   - 真实 API 返回的字段：`id`, `task_id`, `role`, `type`, `content`, `created_at`, `updated_at`
   - **解决**：创建适配函数转换字段

3. **配置参数传递**
   - 需要确保 useConfigStore 中的参数名与后端期望的参数名一致
   - **解决**：检查 api_server.py 的 EvaluationRequest 模型

4. **网络错误处理**
   - SSE 连接断开时如何恢复
   - 超时处理
   - **解决**：添加重试逻辑和错误提示

---

## 🎯 成功标志

1. ✅ 用户在配置中心修改 temperature = 0.9
2. ✅ 用户在分发任务发送消息 "测试"
3. ✅ 前端实时显示 "处理中..." → "完成"
4. ✅ 后端日志显示使用了 temperature=0.9 调用 LLM
5. ✅ PostgreSQL 中保存了用户消息和 AI 响应
6. ✅ 刷新页面后，历史消息正常加载

---

## 📝 实施顺序

**强烈建议按此顺序执行**（依赖关系）：
1. 阶段一（环境切换）- 必须首先完成
2. 阶段二（API 适配层）- 为后续提供基础
3. 阶段四（聊天流）- 相对独立，可以先验证 SSE
4. 阶段三（配置流）- 依赖于聊天流的成功
5. 阶段五（数据验证）- 综合测试
6. 阶段六（Bug 修复）- 根据实际情况调整

---

## 💡 预期工作量

- **总耗时**：约 3 小时
- **最复杂部分**：SSE 事件解析 + 配置参数传递
- **最可能出现 Bug**：数据格式转换、事件处理顺序

