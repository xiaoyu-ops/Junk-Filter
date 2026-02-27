# ✅ 数据适配层实施完成指南

**完成日期**: 2026-02-27
**修改文件**:
- `src/composables/useAPI.js` ✅ 已修改
- `.env.local` ✅ 已更新

---

## 📋 修改内容总结

### 1️⃣ useAPI.js - 适配层实现 (约 80 行新增)

#### 新增函数：

**adaptSourceToTask(source)**
- 将 Go 后端的 Source 对象转换为前端的 Task 对象
- 字段映射：
  - `id` (int) → `id` ("source-{id}" string)
  - `name` → `name`
  - `url` → `command`
  - `priority` (1-10) → `frequency` (hourly/daily/weekly)
  - `enabled` (bool) → `status` ("active"/"paused")
  - `last_fetch_time` → `last_execution`
  - `created_at` → `created_at`

**adaptTaskToSource(task)**
- 反向适配：将前端的 Task 转换为 Go 的 Source 格式
- 用于创建/更新操作
- 逆向映射所有字段

#### 修改的 API 方法：

**tasks.list()**
- 调用 Go 后端 `/api/sources` 而非 Mock `/api/tasks`
- 自动通过 `adaptSourceToTask()` 转换所有数据
- 对前端透明：返回格式完全一致

**tasks.get(id)**
- 提取 "source-" 前缀，获取原始 Go source ID
- 调用 `/api/sources/{id}` 并转换结果

**tasks.create(data)**
- 接收前端的 Task 格式
- 转换为 Source 格式发送给 Go
- 返回适配后的 Task 对象

**tasks.update(id, data)**
- 提取原始 source ID
- 转换数据后发送给 Go `/api/sources/{id}`
- 返回适配后的 Task 对象

**tasks.delete(id)**
- 提取原始 source ID
- 调用 Go `/api/sources/{id}` 删除

**tasks.execute(id)**
- 当前 Go 后端暂无此端点
- 预留接口，可日后补充
- 控制台打印警告信息

#### 消息 API (messages)：

**messages.list(taskId, options)**
- 继续使用 Mock 后端 `VITE_MOCK_URL`
- 自动处理 "source-" 格式的 task ID 转换
- 后续可无缝切换到 Go 后端（仅改 baseUrl）

**messages.save(data)**
- 继续使用 Mock 后端
- 自动转换 task_id 格式
- 与现有前端代码兼容

#### 认证 API (auth)：
- 保持原样（未使用，后续补充）

---

### 2️⃣ .env.local - 环境变量配置 (已更新)

```env
# 业务 API（Go 后端）
VITE_API_URL=http://localhost:8080

# Mock 后端（消息和 SSE）
VITE_MOCK_URL=http://localhost:3000
```

**说明**：
- 原有 `VITE_API_URL=http://localhost:3000` 改为 `http://localhost:8080`
- 新增 `VITE_MOCK_URL=http://localhost:3000`
- 其他配置保持不变

---

## 🚀 立即启动步骤

### 前置条件（必须）

**终端 1 - Go 后端启动**
```bash
cd backend-go
go run main.go
# 输出: "✓ Database connected", "✓ Redis connected", listening on :8080
```

**终端 2 - Python 后端启动** (可选，暂无前端交互)
```bash
cd backend-python
python main.py
# 输出: "✓ Database connected", "✓ Redis connected"
```

**终端 3 - Mock 后端启动**
```bash
cd backend-mock
node server.js
# 输出: "Mock 后端启动在 3000"
```

**终端 4 - 前端启动**
```bash
cd frontend-vue
npm run dev
# 输出: "Local: http://localhost:5173"
```

### 功能验证

#### ✅ 测试 1: 任务列表加载 (5 分钟)

```javascript
// 前端应该显示 Go 后端的 RSS 源作为任务
// 访问 http://localhost:5173
// 左侧任务列表会显示 RSS 源（已适配为 Task 格式）
```

**预期结果**:
- 左侧任务列表显示多个任务（来自 Go 的 sources 表）
- 任务名称正确显示
- 任务状态显示为 "active" 或 "paused"

**排查问题**:
```bash
# 1. 检查 Go 后端是否有源数据
curl http://localhost:8080/api/sources

# 2. 检查前端网络请求（F12 → Network）
# 应该看到 http://localhost:8080/api/sources 请求
# 状态码 200
```

#### ✅ 测试 2: 创建任务 (5 分钟)

```javascript
// 前端创建新任务
// 应该通过 Go 后端的 POST /api/sources 保存
// 刷新页面后仍然存在
```

**步骤**:
1. 点击 "创建任务" 按钮
2. 填写表单：
   - 任务名称: "测试 RSS 源"
   - 命令: "https://example.com/feed.xml"
   - 频率: "daily"
3. 点击 "创建"
4. 验证任务出现在列表中
5. 刷新页面，任务仍在列表中

**排查问题**:
```bash
# 1. 检查 Go 后端数据库
docker exec -it truesignal-db psql -U truesignal -d truesignal
\d sources
SELECT * FROM sources;

# 2. 检查前端网络请求
# POST http://localhost:8080/api/sources 状态应为 201
```

#### ✅ 测试 3: 消息功能 (5 分钟)

```javascript
// 选中任务，发送消息
// 消息应该保存到 Mock 后端（暂时）
// 切换任务后再选回，消息仍存在
```

**步骤**:
1. 选中一个任务
2. 在消息输入框输入 "你好"
3. 按 Enter 发送
4. 观察 SSE 流式回复（来自 Mock）
5. 选择另一个任务，再选回原任务
6. 消息历史应该加载显示

**排查问题**:
```bash
# 1. 检查 Mock 后端消息数据
cat backend-mock/data/messages.json | jq .

# 2. 检查前端网络请求
# 消息列表: GET http://localhost:3000/api/tasks/{taskId}/messages
# 消息保存: POST http://localhost:3000/api/messages
# SSE 聊天: GET http://localhost:3000/api/chat/stream
```

#### ✅ 测试 4: 删除任务 (3 分钟)

```javascript
// 删除任务应该调用 Go 后端的 DELETE /api/sources/{id}
// 前端列表立即移除
```

**步骤**:
1. 右键点击任务或找到删除按钮
2. 确认删除
3. 任务从列表消失
4. 刷新页面，任务仍然消失

---

## 🔍  技术细节

### ID 格式转换

**前端使用的 ID 格式**:
```javascript
"source-1"
"source-2"
"source-123"
```

**为什么?**
- Go 后端的 source ID 是数字 (int64)
- 前端的 task ID 习惯是字符串
- 添加前缀避免冲突和类型混淆

**自动转换流程**:
```javascript
// 前端调用 tasks.delete("source-1")
↓
// useAPI.js 自动提取
sourceId = "1"
↓
// 调用 Go 后端
DELETE http://localhost:8080/api/sources/1
```

### 数据流向

```
┌─────────────────────────────────┐
│      前端 TaskDistribution      │
│    使用 Task 格式的数据          │
└──────────────┬──────────────────┘
               │
               ↓
┌─────────────────────────────────┐
│    useAPI.js 适配层              │
│  转换 Task ↔ Source 格式         │
└──────────────┬──────────────────┘
               │
       ┌───────┴────────┐
       ↓                ↓
   Go 后端         Mock 后端
   (8080)          (3000)
   任务/内容        消息/SSE
```

### 后续迁移无缝切换

当 Go 后端实现消息 API 后，只需修改 `useAPI.js` 中一行：

```javascript
// 当前
const messages = {
  list: async (taskId, options) => {
    return request(..., { baseUrl: mockUrl })  // ← 这一行
  },
  save: async (data) => {
    return request(..., { baseUrl: mockUrl })  // ← 这一行
  },
}

// 迁移后
const messages = {
  list: async (taskId, options) => {
    return request(..., { baseUrl: apiUrl })  // ← 改为 apiUrl
  },
  save: async (data) => {
    return request(..., { baseUrl: apiUrl })  // ← 改为 apiUrl
  },
}
```

**无需修改任何 UI 组件！**

---

## ⚠️ 常见问题排查

### Q1: 任务列表为空或加载失败

**原因 1**: Go 后端未启动
```bash
# 检查 8080 端口是否监听
netstat -ano | findstr :8080  # Windows
lsof -i :8080                 # Mac/Linux
```

**原因 2**: VITE_API_URL 配置错误
```bash
# .env.local 中确保设置正确
VITE_API_URL=http://localhost:8080
```

**原因 3**: Go 后端没有源数据
```bash
# 添加初始 RSS 源
curl -X POST http://localhost:8080/api/sources \
  -H "Content-Type: application/json" \
  -d '{"name":"测试","url":"https://example.com/feed"}'
```

### Q2: 消息保存失败

**原因**: Mock 后端未启动或 VITE_MOCK_URL 错误
```bash
# 检查 3000 端口
netstat -ano | findstr :3000

# .env.local 确保设置
VITE_MOCK_URL=http://localhost:3000
```

### Q3: SSE 流式对话不工作

**原因**: Mock 后端的 SSE 端点问题
```bash
# 测试 SSE 连接
curl -N "http://localhost:3000/api/chat/stream?taskId=1&message=hello"
```

### Q4: 刷新页面任务消失

**可能是正常行为**：
- 新创建的任务可能还未同步到数据库
- 检查 Go 后端日志是否有保存成功
- 检查数据库中是否真的保存了

---

## 📊 配置检查清单

使用此清单验证所有配置正确：

- [ ] Go 后端启动在 8080 端口
- [ ] Mock 后端启动在 3000 端口
- [ ] `.env.local` 中 `VITE_API_URL=http://localhost:8080`
- [ ] `.env.local` 中 `VITE_MOCK_URL=http://localhost:3000`
- [ ] 前端 `npm run dev` 运行在 5173 端口
- [ ] 访问 http://localhost:5173 能正常加载
- [ ] 左侧任务列表显示 Go 后端的源

---

## 🎯 验收标准

✅ **通过条件**:
1. 任务列表正确显示 Go 后端的 RSS 源（已适配为 Task 格式）
2. 可创建新任务（通过 Go 后端 `/api/sources` 保存）
3. 可删除任务（调用 Go 后端删除）
4. 消息功能正常（使用 Mock 后端）
5. SSE 流式聊天正常工作
6. 刷新页面后数据持久化正确

❌ **失败条件**:
1. 任务列表无法加载或显示空列表
2. 创建任务报错或不保存
3. 消息无法保存或加载
4. SSE 聊天断连或无流式效果

---

## 📝 后续工作

### 短期（1-2 天）

- [ ] 验证上述所有功能正常
- [ ] Go 后端根据反馈调整 API 格式
- [ ] 监控数据库和 API 响应性能

### 中期（1 周）

- [ ] Go 后端实现消息存储 API
- [ ] Python 后端集成评估结果查询
- [ ] 迁移消息 API 从 Mock 到 Go（1 行修改）

### 长期（2 周）

- [ ] Go 后端实现 SSE 聊天端点
- [ ] 弃用 Mock 后端
- [ ] 添加用户认证系统

---

**实施完成**：✅ 所有代码修改已完成
**下一步**: 按照"立即启动步骤"测试验证

