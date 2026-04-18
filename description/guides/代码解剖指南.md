# Junk Filter — 代码解剖指南

> 目标：像解剖一样，从数据流入到数据出，搞清楚每一行代码在系统中的位置和意义。
> 建议顺序：先跑通系统，再按本文的顺序阅读代码。

---

## 一、整体数据流（先建立心智模型）

```
互联网 RSS 源
      │
      ▼
[Go] rss_service.go — 协程池并发 HTTP 抓取
      │
      ├─ L1: Bloom Filter（内存，快速拒绝重复）
      ├─ L2: Redis Set（精确去重，7天TTL）
      ├─ L3: PostgreSQL UNIQUE（并发兜底）
      └─ 短内容/黑名单过滤
      │
      ▼
PostgreSQL: content 表（status=PENDING）
      +
Redis Stream: ingestion_queue（XADD）
      │
      ▼
[Python] stream_consumer.py — XREADGROUP 消费
      │
      ▼
[Python] content_evaluator.py（LangGraph图）
   evaluate_node → parse_node → [retry] → 最多3次
      │
      ▼
PostgreSQL: evaluation 表（scores / decision / tldr）
      │
      ├─► push_service.py → Telegram Bot API（高分推送）
      │
      ├─► Vue 3 前端（Timeline 卡片展示）
      │
      └─► react_agent.py（自然语言查询/管理）
               ↑
        来自前端对话页 或 Telegram Bot
```

---

## 二、Go 后端（`backend-go/`）

### 入口：`main.go`

**第一步读这个文件。**

```
main.go
  loadConfig()     ← 读 config.yaml + 环境变量覆盖
  initDB()         ← 建 PostgreSQL 连接池
  initRedis()      ← 建 Redis 客户端
  initBloomFilter()← 从DB预热近7天URL hash
  setupRouter()    ← 注册所有 Gin 路由
  startRSSScheduler() ← 后台协程，每30分钟触发抓取
```

关键结构体：`Config`（main.go 里内联定义，非 internal/config）

### 路由注册：`handlers/routes.go`

所有 HTTP 路由都在这里注册，是查「哪个 URL 对应哪个 handler」的索引。

```go
// 读这个文件等于看到了所有 API 的目录
r.GET("/api/sources", sourceHandler.GetSources)
r.POST("/api/sources", sourceHandler.CreateSource)
// ...
```

### Handler 层（`handlers/`）

每个文件对应一类资源，职责：**解析 HTTP 请求 → 调 Service/Repo → 返回 JSON**。

| 文件 | 职责 |
|------|------|
| `source_handler.go` | RSS 源 CRUD + 手动触发抓取 |
| `content_handler.go` | 文章列表、详情、stats |
| `evaluation_handler.go` | 评估结果查询 |
| `config_handler.go` | LLM 配置读写、RSS 代理热更新 |
| `task_chat_handler.go` | SSE 转发 Python Agent |
| `notification_handler.go` | 通知 SSE 推流 |
| `sse_helpers.go` | SSE 公用写入辅助函数 |

### Service 层（`services/`）

| 文件 | 核心逻辑 |
|------|---------|
| `rss_service.go` | 协程池抓取、解析 Feed、调 content_service 去重入队 |
| `content_service.go` | 三层去重 + 写 content 表 + XADD 到 Redis Stream |
| `dedup_service.go` | Bloom Filter 封装（初始化、查询、添加）|

**关键路径**（手动触发抓取）：

```
POST /api/sources/:id/fetch
  → source_handler.go: FetchSource()
  → rss_service.go: FetchAndProcessSource()
  → content_service.go: ProcessContent()  ← 三级去重
  → PostgreSQL INSERT + Redis XADD
```

### Repository 层（`repositories/`）

纯 DB 操作，不含业务逻辑。

| 文件 | 负责的表 |
|------|---------|
| `source_repo.go` | sources 表（含 favicon 自动推导逻辑）|
| `content_repo.go` | content 表 |
| `evaluation_repo.go` | evaluation 表 |
| `thread_repository.go` + `message_repository.go` | 对话历史 |

**注意**：`source_repo.go` 的 `Create()` 中有 favicon 自动推导：
```go
// 从 RSS URL 解析域名，构造 favicon.ico 地址
if parsed, err := url.Parse(req.URL); err == nil {
    faviconURL := fmt.Sprintf("%s://%s/favicon.ico", parsed.Scheme, parsed.Host)
    source.FaviconURL = &faviconURL
}
```

---

## 三、Python 后端（`backend-python/`）

### 进程架构（三个独立进程）

```
api_server.py      ← FastAPI，响应前端 HTTP 请求（端口 8083）
main.py            ← Redis Stream 消费者（自动评估后台）
telegram_bot.py    ← Telegram Bot（命令+自然语言控制）
```

三者共享同一个 PostgreSQL，但各自独立运行，互不依赖进程通信。

### 配置入口：`config.py`

```python
class Settings(BaseModel):
    db_host, db_port, db_user, db_password, db_name  # 基础设施
    # LLM 配置不在这里——在 DB 的 ai_config 表

async def load_llm_config_from_db(pool) -> dict:
    # 从 ai_config 表读取 model_name / api_key / base_url
    # 被 stream_consumer 和 telegram_bot 调用
```

### `main.py`（Stream 消费者）

```python
asyncio.run(main())
  → asyncpg.create_pool()      # DB 连接池
  → load_llm_config_from_db()  # 读 LLM 配置
  → StreamConsumer(pool, llm_config).run()
      # 每 60 秒热加载一次 LLM 配置（relay/model 改了不用重启）
```

### `services/stream_consumer.py`

```python
class StreamConsumer:
    async def run(self):
        while True:
            msgs = await redis.xreadgroup(...)  # 从 ingestion_queue 取消息
            for msg in msgs:
                await self._process(msg)

    async def _process(self, msg):
        content = await fetch_content_from_db(content_id)
        result = await evaluate_content(content, llm_config, user_prefs)
        await save_evaluation(result)
        if result.decision in ('INTERESTING', 'BOOKMARK'):
            await push_to_channels(...)  # Telegram 推送
```

### `agents/content_evaluator.py`（LangGraph 评估图）

这是 LLM 评估的核心，读它要先理解 LangGraph 的 State Graph 概念：

```python
# 状态定义
class EvaluationState(TypedDict):
    title, content, url          # 输入
    raw_output, parsed_result    # 中间/输出
    retry_count, error           # 控制流

# 图的节点
evaluate_node(state)  →  调用 ChatOpenAI（streaming=True，兼容中继）
parse_node(state)     →  解析 JSON 输出，验证字段
should_retry(state)   →  条件边：parse 失败且未超限 → 回到 evaluate

# 编译成可执行图
graph = StateGraph(EvaluationState)
graph.add_node("evaluate", evaluate_node)
graph.add_node("parse", parse_node)
graph.add_conditional_edges("parse", should_retry, {...})
```

**重要细节**：`ChatOpenAI(streaming=True)` 是为了兼容部分中继服务（非流式时 content 为 null）。

### `agents/react_agent.py`（ReAct 推理循环）

自行实现的 ReAct，不依赖 Agent SDK，核心循环：

```python
async def run_react(user_message, history, llm_config, db_pool):
    messages = [system_prompt] + history + [user_message]
    
    for _ in range(MAX_ITERATIONS):  # 最多 8 次
        response = await llm.ainvoke(messages, tools=TOOL_DEFINITIONS)
        
        if response.tool_calls:
            # 执行工具，结果追加到 messages
            tool_result = await execute_tool(response.tool_calls[0])
            messages.append(tool_result)
            yield f"data: {json.dumps({'type':'tool_call', ...})}\n\n"
        else:
            # 无工具调用 = 最终回答，流式 yield
            async for chunk in response:
                yield f"data: {json.dumps({'type':'chunk','content':chunk})}\n\n"
            break
    
    yield f"data: {json.dumps({'type':'done'})}\n\n"
```

### `agents/tools.py`（5 个 Agent 工具）

| 工具函数 | 实现方式 | 功能 |
|---------|---------|------|
| `query_articles` | asyncpg 直连 DB | 按关键词/评分/时间检索，返回 tldr+reasoning |
| `get_pipeline_status` | aiohttp → Go API | 各状态文章数量 |
| `add_source` | aiohttp → Go API | 添加 RSS 源 |
| `remove_source` | aiohttp → Go API | 删除 RSS 源 |
| `update_preferences` | asyncpg 直连 DB | 更新偏好，实时影响后续评估 |

### `services/push_service.py`（Telegram 推送）

```python
async def push_to_channels(db_pool, title, summary, scores, decision, url):
    row = await db_pool.fetchrow("SELECT push_channels FROM notification_settings WHERE id=1")
    for channel in row["push_channels"]:
        if channel["type"] == "telegram" and channel["enabled"]:
            await _push_telegram(channel, ...)

async def _push_telegram(channel, ...):
    # MarkdownV2 格式，需转义特殊字符（_escape() 函数）
    await session.post(f"https://api.telegram.org/bot{token}/sendMessage", ...)
```

### `telegram_bot.py`（双向控制 Bot）

```python
async def main():
    pool = await asyncpg.create_pool(...)
    bot_token, chat_id = await _load_config(pool)  # 从 DB 读
    llm_config = await load_llm_config_from_db(pool)
    
    app = Application.builder().token(bot_token).build()
    app.bot_data = {"db_pool": pool, "chat_id": chat_id, "llm_config": llm_config}
    
    # 命令处理器
    app.add_handler(CommandHandler("status", cmd_status))
    app.add_handler(CommandHandler("fetch", cmd_fetch))
    app.add_handler(CommandHandler("recent", cmd_recent))
    
    # 自然语言 → ReAct Agent
    app.add_handler(MessageHandler(filters.TEXT, handle_message))
    
    await app.run_polling()

# 鉴权装饰器（白名单 chat_id）
def _auth(func):
    async def wrapper(update, context):
        if str(update.effective_chat.id) != context.bot_data["chat_id"]:
            return  # 静默丢弃陌生人消息
        return await func(update, context)
    return wrapper
```

---

## 四、前端（`frontend-vue/src/`）

### 页面组织

```
App.vue
  ├── AppNavbar.vue         # 顶部导航栏
  ├── Timeline.vue          # 主页面：文章卡片流
  ├── Task.vue              # Agent 对话页
  ├── Config.vue            # 配置页（RSS源/LLM/推送）
  └── Home.vue              # 首页（统计+概览）
```

### Store 层（Pinia，`src/stores/`）

| Store | 管理的状态 |
|-------|---------|
| `useConfigStore.js` | LLM 配置、RSS 代理、推送渠道设置 |
| `useTimelineStore.js` | 文章列表、过滤条件、分页 |
| `useTaskStore.js` | Agent 任务、对话线程列表 |
| `useReaderStore.js` | 文章详情抽屉状态 |
| `useThemeStore.js` | 暗色/亮色主题 |

**重要**：`useConfigStore.js` 的 `loadConfig()` 在 App 启动时调用，会同时 `loadLLMConfigFromBackend()`（从 Go API 拉最新 LLM 配置，覆盖 localStorage）。

### 关键 Composables（`src/composables/`）

| 文件 | 职责 |
|------|------|
| `useAPI.js` | 封装所有 HTTP 请求，Go 和 Python 的 base URL 在这里 |
| `useSSE.js` | 封装 EventSource / fetch+ReadableStream，处理 SSE 连接 |

**改 API 地址只需改 `useAPI.js`** 里的 `API_BASE` 和 `PYTHON_BASE` 常量。

### 核心组件读法

**`Timeline.vue`**：最复杂的页面
- 顶部 4 个统计卡片（`/api/content/stats`）
- 趋势折线图（`/api/evaluations` 聚合）
- 卡片列表（`/api/content`，支持 decision 过滤）
- 点击卡片 → `Reader.vue` 抽屉（详情+图片+Reasoning）

**`Task.vue`** + **`TaskChat.vue`**：Agent 对话
- Thread 列表（左侧）
- 消息输入 → POST `/api/tasks/:id/chat`（Go 转发 Python SSE）
- `AgentSteps.vue`：工具调用可折叠面板（执行中展开，完成后1.5s折叠）

**`Config.vue`**：配置管理
- RSS 源管理（增删改 + 推荐源弹窗）
- LLM 配置（model/api_key/base_url）
- 推送渠道（Telegram bot_token + chat_id）

---

## 五、数据库（`sql/`）

按文件编号顺序执行，理解表结构演变：

| 文件 | 内容 |
|------|------|
| `01_junk_filter_schema.sql` | 核心表：sources, content, evaluation |
| `02_schema.sql` | ai_config 表（LLM 配置）|
| `03_migration_chat_v2.sql` | threads, messages（对话历史）|
| `05_add_favicon.sql` | sources.favicon_url 列 |
| `08_add_author_filter.sql` | sources.author_filter JSONB |
| `09_add_user_preferences.sql` | user_preferences 表 |
| `10~12_add_notifications.sql` | notification_settings + push_channels JSONB |
| `13_add_eval_attempts.sql` | content.eval_attempts（防死循环）|

**最关键的几个字段**：

```sql
-- content 表（文章生命周期）
status: PENDING → PROCESSING → EVALUATED / DISCARDED
eval_attempts: 失败累计，>= 3 自动 DISCARDED

-- evaluation 表（LLM 评估结果）
innovation_score, depth_score  -- 各 0-10
decision: INTERESTING / BOOKMARK / SKIP
tldr, reasoning, key_concepts  -- 文本字段

-- notification_settings 表
push_channels: [{"type":"telegram","bot_token":"...","chat_id":"...","enabled":true}]
```

---

## 六、推荐阅读顺序

### 理解「一篇文章从抓取到推送」的完整旅程：

1. `backend-go/main.go` — 系统如何启动
2. `backend-go/services/rss_service.go` — 如何抓取 RSS
3. `backend-go/services/content_service.go` — 三级去重如何工作
4. `backend-go/repositories/source_repo.go` — DB 操作 + favicon 推导
5. `backend-python/services/stream_consumer.py` — 如何消费队列
6. `backend-python/agents/content_evaluator.py` — LangGraph 评估图
7. `backend-python/services/push_service.py` — 如何推送 Telegram

### 理解「用户说一句话，Agent 如何响应」：

1. `frontend-vue/src/components/TaskChat.vue` — 消息如何发送
2. `frontend-vue/src/composables/useSSE.js` — SSE 如何消费
3. `backend-go/handlers/task_chat_handler.go` — Go 如何转发
4. `backend-python/api_server.py` — FastAPI 路由入口
5. `backend-python/agents/react_agent.py` — ReAct 推理循环
6. `backend-python/agents/tools.py` — 5 个工具的实现

### 理解「Telegram Bot 如何工作」：

1. `backend-python/telegram_bot.py` — Bot 主体（命令+自然语言）
2. `backend-python/agents/react_agent.py` — 自然语言转 Agent
3. `backend-python/services/push_service.py` — 主动推送评估结果

### 理解「前端如何展示数据」：

1. `frontend-vue/src/composables/useAPI.js` — 所有 API 调用的封装
2. `frontend-vue/src/stores/useTimelineStore.js` — 文章状态管理
3. `frontend-vue/src/components/Timeline.vue` — 卡片展示逻辑

---

## 七、常见疑问速查

**Q: LLM 配置在哪里？**
A: 数据库 `ai_config` 表，不在 `.env`。前端 Config 页面填写 → POST `/api/config/llm` → Go 写 DB → Python 每 60 秒热加载。

**Q: 新加的 RSS 源为什么有头像？**
A: `source_repo.go` 的 `Create()` 自动解析 URL 域名，构造 `scheme://host/favicon.ico` 写入 `favicon_url`。

**Q: 评估失败为什么文章会消失？**
A: `content.eval_attempts >= 3` 时标记为 `DISCARDED`，不再出现在 Timeline。可以通过 SQL 重置 `status='PENDING', eval_attempts=0` 重新评估。

**Q: Telegram Bot 收不到消息？**
A: 检查：①Config 页面是否配置了 `bot_token` + `chat_id` 并保存；②`telegram_bot.py` 进程是否启动；③与 Bot 的第一条消息是否发 `/start`。Bot 会静默忽略非白名单 chat_id 的消息。

**Q: 换了 relay/模型后评估还是用旧的？**
A: `stream_consumer.py` 每 60 秒热加载 LLM 配置，最多等 60 秒生效。Telegram Bot 的 LLM 配置在启动时加载，需重启 `telegram_bot.py`。

**Q: SSE 在代码里怎么找？**
A: Go 端看 `handlers/sse_helpers.go` 和 `task_chat_handler.go`；Python 端看 `react_agent.py` 里的 `yield f"data: ..."`；前端看 `composables/useSSE.js`。
