# TrueSignal API 集成指南
## Config.vue ↔ Go 后端 API 对接方案

**日期**: 2026-02-28
**版本**: 1.0
**状态**: 实施中

---

## 📋 API 契约概览

### 基础信息
- **Go 后端基础 URL**: `http://localhost:8080`
- **前端请求源**: `http://localhost:5173` (Vite 开发服务器)
- **CORS 支持**: ✅ 已在 Go 主.go 中配置
- **认证**: 暂无（Demo 阶段）

### 核心 API 端点

#### 1. RSS 源管理 (`/api/sources`)

##### 1.1 获取所有 RSS 源
```http
GET /api/sources?enabled=true
```

**Request Parameters**:
- `enabled` (optional, boolean): 仅返回启用的源。默认: null (返回全部)

**Response (200 OK)**:
```json
[
  {
    "id": 1,
    "url": "https://news.ycombinator.com/rss",
    "platform": "blog",
    "author_name": "Hacker News",
    "author_id": null,
    "priority": 9,
    "enabled": true,
    "last_fetch_time": "2026-02-28T10:30:00Z",
    "fetch_interval_seconds": 1800,
    "created_at": "2026-02-26T00:00:00Z",
    "updated_at": "2026-02-28T10:30:00Z"
  },
  {
    "id": 2,
    "url": "https://feeds.arstechnica.com/arstechnica/index",
    "platform": "blog",
    "author_name": "Ars Technica",
    "author_id": null,
    "priority": 8,
    "enabled": true,
    "last_fetch_time": "2026-02-28T09:45:00Z",
    "fetch_interval_seconds": 3600,
    "created_at": "2026-02-26T00:00:00Z",
    "updated_at": "2026-02-28T09:45:00Z"
  }
]
```

**Error Response (500)**:
```json
{
  "error": "Failed to list sources"
}
```

---

##### 1.2 创建新 RSS 源
```http
POST /api/sources
Content-Type: application/json
```

**Request Body**:
```json
{
  "url": "https://example.com/feed",
  "author_name": "Example Blog",
  "platform": "blog",
  "priority": 7,
  "enabled": true,
  "fetch_interval_seconds": 1800
}
```

**Response (201 Created)**:
```json
{
  "id": 4,
  "url": "https://example.com/feed",
  "platform": "blog",
  "author_name": "Example Blog",
  "author_id": null,
  "priority": 7,
  "enabled": true,
  "last_fetch_time": null,
  "fetch_interval_seconds": 1800,
  "created_at": "2026-02-28T11:00:00Z",
  "updated_at": "2026-02-28T11:00:00Z"
}
```

**前端 useConfigStore 映射**:
```javascript
newRssForm = {
  name: "author_name"
  url: "url"
  frequency: "fetch_interval_seconds" (需转换: "hourly"→3600, "30min"→1800 等)
  filterRules: "(暂不支持)"  // 可扩展
}
```

**Error Response (400)**:
```json
{
  "error": "Invalid request body"
}
```

**Error Response (500)**:
```json
{
  "error": "Failed to create source"
}
```

---

##### 1.3 获取单个 RSS 源
```http
GET /api/sources/:id
```

**Response (200 OK)**:
```json
{
  "id": 1,
  "url": "https://news.ycombinator.com/rss",
  "platform": "blog",
  "author_name": "Hacker News",
  "author_id": null,
  "priority": 9,
  "enabled": true,
  "last_fetch_time": "2026-02-28T10:30:00Z",
  "fetch_interval_seconds": 1800,
  "created_at": "2026-02-26T00:00:00Z",
  "updated_at": "2026-02-28T10:30:00Z"
}
```

**Error Response (404)**:
```json
{
  "error": "Source not found"
}
```

---

##### 1.4 更新 RSS 源
```http
PUT /api/sources/:id
Content-Type: application/json
```

**Request Body** (部分更新):
```json
{
  "priority": 10,
  "enabled": false,
  "fetch_interval_seconds": 7200
}
```

**Response (200 OK)**:
```json
{
  "id": 1,
  "url": "https://news.ycombinator.com/rss",
  "platform": "blog",
  "author_name": "Hacker News",
  "author_id": null,
  "priority": 10,
  "enabled": false,
  "last_fetch_time": "2026-02-28T10:30:00Z",
  "fetch_interval_seconds": 7200,
  "created_at": "2026-02-26T00:00:00Z",
  "updated_at": "2026-02-28T11:05:00Z"
}
```

---

##### 1.5 删除 RSS 源
```http
DELETE /api/sources/:id
```

**Response (200 OK)**:
```json
{
  "message": "Source deleted"
}
```

**Error Response (404)**:
```json
{
  "error": "Source not found"
}
```

---

##### 1.6 手动同步 RSS 源（关键！）
```http
POST /api/sources/:id/fetch
```

**Response (200 OK)**:
```json
{
  "message": "Source fetch triggered"
}
```

**后续流程**:
1. Go handler 调用 `rssService.FetchSourceOnDemand(ctx, id)`
2. RSSService 内部：
   - 获取源的 URL
   - 发起 HTTP GET 请求获取 RSS Feed
   - 解析 XML，提取 `<item>` 元素
   - 对每个 item，生成 content_hash (URL 或 title 的 MD5)
   - 检查 Redis 和 PostgreSQL 中是否已存在 (L1/L2/L3 去重)
   - 新 item → INSERT 到 content 表，status='PENDING'
   - 新 item → XADD 到 Redis Stream `ingestion_queue`
3. Python Consumer 订阅 Redis Stream，使用 SmartEvaluator 评估
4. 评估完成 → UPDATE content 表, status='EVALUATED'
5. 前端 Timeline 刷新或自动推送更新

**UI 反馈**:
- 同步前: 按钮显示 sync icon，可点击
- 同步中: 按钮显示 animate-spin，disabled
- 同步完后:
  - 成功: Toast "已同步 N 条新文章"，展开行日志更新
  - 失败: Toast "同步失败"，展开行显示错误原因

---

#### 2. 配置管理（AI 模型配置）

> 注：这部分暂时使用 localStorage，可选进一步对接数据库

##### 2.1 获取配置（可选）
```http
GET /api/config
```

**Response (200 OK)**:
```json
{
  "modelName": "gpt-4-turbo",
  "apiKey": "***", // 不返回真实 key
  "baseUrl": "https://api.openai.com/v1",
  "temperature": 0.7,
  "topP": 0.9,
  "maxTokens": 2048
}
```

---

##### 2.2 更新配置（可选，暂不实现）
```http
PUT /api/config
Content-Type: application/json
```

> 当前 useConfigStore 使用 localStorage，无需调用此 API

---

### 3. 同步日志实时推送（关键功能！）

#### 方案 A: Server-Sent Events (SSE) - **推荐**

```http
GET /api/sources/:id/sync-logs/stream
Accept: text/event-stream
```

**Response Headers**:
```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
```

**Response Body** (服务器持续推送):
```
data: {"timestamp":"2026-02-28T11:10:00Z","status":"started","message":"开始同步..."}

data: {"timestamp":"2026-02-28T11:10:05Z","status":"fetching","itemsCount":5,"message":"已获取 5 条新文章"}

data: {"timestamp":"2026-02-28T11:10:30Z","status":"success","itemsCount":5,"message":"同步完成：获取 5 条新文章"}
```

**前端实现示例**:
```javascript
const startSyncStream = (sourceId) => {
  const eventSource = new EventSource(`/api/sources/${sourceId}/sync-logs/stream`)

  eventSource.onmessage = (event) => {
    const log = JSON.parse(event.data)
    // 实时更新 UI
    source.syncLogs.unshift(log)
  }

  eventSource.addEventListener('done', () => {
    eventSource.close()
    // 同步完成
  })

  eventSource.onerror = () => {
    eventSource.close()
    showToast('同步中断', 'error')
  }
}
```

#### 方案 B: 轮询（简单但低效）

```http
GET /api/sources/:id/sync-logs
```

**Query Parameters**:
- `limit` (default: 10): 返回最近的 N 条日志
- `offset` (default: 0): 分页偏移

**Response (200 OK)**:
```json
{
  "sourceId": 1,
  "logs": [
    {
      "timestamp": "2026-02-28T11:10:30Z",
      "status": "success",
      "itemsCount": 5,
      "message": "同步完成：获取 5 条新文章"
    },
    {
      "timestamp": "2026-02-28T11:10:05Z",
      "status": "fetching",
      "itemsCount": 5,
      "message": "已获取 5 条新文章"
    }
  ],
  "total": 15
}
```

---

## 🔄 数据流映射

### 前端 Form → Go API Request

**Config.vue 中的数据**:
```javascript
newRssForm = {
  name: "TechCrunch",           // → author_name
  url: "https://...",            // → url
  frequency: "hourly",           // → fetch_interval_seconds (转换)
  filterRules: ""                // → (暂不支持)
}
```

**频率转换表**:
```
"hourly"   → 3600 (seconds)
"30min"    → 1800
"2hours"   → 7200
"daily"    → 86400
```

**发送给 Go 的 Request**:
```json
POST /api/sources
{
  "url": "https://...",
  "author_name": "TechCrunch",
  "platform": "blog",
  "priority": 5,  // 默认
  "enabled": true,
  "fetch_interval_seconds": 3600
}
```

**Go 返回 Response**:
```json
{
  "id": 4,
  "url": "https://...",
  "author_name": "TechCrunch",
  "platform": "blog",
  "priority": 5,
  "enabled": true,
  "last_fetch_time": null,
  "fetch_interval_seconds": 3600,
  "created_at": "2026-02-28T11:00:00Z",
  "updated_at": "2026-02-28T11:00:00Z"
}
```

**前端转换为 Store 格式**:
```javascript
const source = {
  id: response.id,
  name: response.author_name,     // ← name
  url: response.url,
  frequency: (() => {              // ← 频率转换
    const map = { 3600: 'hourly', 1800: '30min', 7200: '2hours', 86400: 'daily' }
    return map[response.fetch_interval_seconds] || 'hourly'
  })(),
  status: response.enabled ? 'active' : 'paused',  // ← enabled → status
  statusClass: response.enabled ? 'bg-green-...' : 'bg-gray-...',
  lastSyncTime: response.last_fetch_time,
  lastSyncStatus: 'success',  // ← 初始值
  syncLogs: [],               // ← 初始为空
  filterRules: ''
}
```

---

## 🐛 错误处理

### HTTP Status Codes

| 状态码 | 含义 | 前端处理 |
|--------|------|---------|
| 200 | OK | ✅ 成功，显示 success toast |
| 201 | Created | ✅ 创建成功，刷新列表 |
| 400 | Bad Request | ❌ 验证失败，显示错误信息 |
| 404 | Not Found | ❌ 资源不存在，提示删除或刷新 |
| 500 | Internal Server Error | ❌ 服务器错误，提示重试 |

### 前端统一错误处理

```javascript
const handleApiCall = async (fetchFn, successMsg, errorMsg) => {
  try {
    const response = await fetchFn()
    if (!response.ok) {
      const errData = await response.json()
      showToast(errData.error || errorMsg, 'error')
      return null
    }
    showToast(successMsg, 'success')
    return await response.json()
  } catch (err) {
    showToast(err.message || errorMsg, 'error')
    return null
  }
}
```

---

## ✅ 数据一致性保证

### 前端状态更新顺序

```
用户操作
  ↓
发送 API 请求 (loading = true)
  ↓
等待响应 (timeout = 30s)
  ↓
收到响应
  ├─ 成功 (200/201)
  │   ├─ 更新本地 store 状态
  │   ├─ 显示 success toast
  │   └─ 刷新列表（重新 GET /api/sources）
  │
  └─ 失败 (400/404/500)
      ├─ 保持原状态不变
      ├─ 显示 error toast
      └─ 提示用户重试

  完成 (loading = false)
```

### 数据库事务一致性

> Go 后端处理

```
1. 创建 source → INSERT sources 表
   ↓
2. 成功 → 返回 201，前端获取新 id
   ↓
3. 前端添加到 store.sources 列表
```

**关键点**:
- Go 数据库操作必须是原子性的
- INSERT 失败 → 返回 500，前端不更新状态
- DELETE 操作级联删除相关 content/evaluation 记录

---

## 🚀 快速开始

### 1. 启动 Go 后端
```bash
cd backend-go
go run main.go
# ✓ Database connected
# ✓ Redis connected
# ✓ Server: listening on :8080
```

### 2. 验证 CORS
```bash
curl -i -X OPTIONS http://localhost:8080/api/sources \
  -H "Access-Control-Request-Method: GET" \
  -H "Access-Control-Request-Headers: Content-Type"
# 应该看到 204 No Content
```

### 3. 测试 API
```bash
# 获取所有源
curl http://localhost:8080/api/sources

# 创建新源
curl -X POST http://localhost:8080/api/sources \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com/feed",
    "author_name": "Example",
    "priority": 7,
    "enabled": true,
    "fetch_interval_seconds": 1800
  }'

# 手动同步
curl -X POST http://localhost:8080/api/sources/1/fetch
```

### 4. 前端访问
```
http://localhost:5173
→ Config 页面
→ 点击"添加订阅源"
→ 填写表单，提交
→ 请求应该发送到 http://localhost:8080/api/sources
→ 检查浏览器 DevTools Network 标签
```

---

## 📊 故障排查

### 问题 1: CORS 错误

**错误**:
```
Access to XMLHttpRequest at 'http://localhost:8080/api/sources'
from origin 'http://localhost:5173' has been blocked by CORS policy
```

**排查**:
- ✅ 已在 Go main.go 中配置了 CORS 中间件
- 检查 Go 是否真的在监听 8080
- 检查前端是否真的发送了请求

**解决**:
- 确保 Go 中间件在最前面（在所有路由之前）
- 确认已设置 `Access-Control-Allow-Origin: *`

---

### 问题 2: 404 Not Found

**错误**:
```json
{
  "error": "Source not found"
}
```

**排查**:
- 确认使用了正确的源 ID
- 检查是否真的在数据库中创建了记录
- 检查是否有级联删除

---

### 问题 3: 500 Internal Server Error

**错误**:
```json
{
  "error": "Failed to create source"
}
```

**排查**:
- 检查 Go 后端日志: `docker-compose logs go`
- 检查是否有数据库连接问题
- 检查是否有 URL UNIQUE 约束冲突

---

## 📝 关键改动清单

- [ ] ✅ Go 后端 CORS 配置（已完成）
- [ ] ✅ Go 后端 sources CRUD API（已完成）
- [ ] ✅ Go 后端 FetchSourceNow handler（已完成）
- [ ] ⏳ 实现 `/api/sources/:id/sync-logs/stream` (SSE)
- [ ] ⏳ 修改 useConfigStore 的 fetch 调用
- [ ] ⏳ 修改 Config.vue 的频率转换逻辑
- [ ] ⏳ 实现冒烟测试脚本

---

## 📚 下一步

1. **实现 SSE 流** - Go 后端添加 `/api/sources/:id/sync-logs/stream`
2. **更新 useConfigStore** - 替换所有 fetch 调用
3. **全链路冒烟测试** - 端到端验证
4. **性能优化** - 必要时添加缓存和索引

---

**版本历史**:
- v1.0 (2026-02-28): 初始版本，API 契约定义完整

