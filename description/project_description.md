# Junk Filter — 项目介绍文档

---

## 简历分点介绍（STAR，AI Agent 视角）

> **项目背景**：个人信息订阅系统。面对多源 RSS 海量资讯、信噪比极低、人工筛选耗时且偏好难以规则化表达的痛点，独立设计并全栈实现了一套由 AI Agent 驱动的智能内容过滤与交互系统。

---

**S（背景）** 订阅源每日产出数百条资讯，其中大量为低价值内容，传统关键词规则维护成本高且无法捕捉语义层面的质量差异；用户偏好高度个性化，难以静态编码。

**T（目标）** 独立设计并实现一套端到端系统：自动抓取、去重、LLM 语义评估、结果可交互查询，最终以自然语言对话替代全部配置界面操作。

**A（行动）** 分三层构建系统：

- **抓取与去重层（Go）**：利用 Go 协程池实现多源并发 RSS 抓取，设计内存 Bloom Filter + Redis Set + PostgreSQL 三级去重机制，并在入队前过滤内容长度不足 200 字的低质量条目，通过 Redis Stream 与评估层异步解耦。

- **LLM 评估层（Python + LangGraph）**：基于 LangGraph 构建有状态评估图（evaluate → parse → retry），对每篇文章输出创新分、深度分、决策（INTERESTING / BOOKMARK / SKIP）、TLDR 及推理过程；将用户偏好标签从数据库动态注入 system prompt，实现无需重训练的零样本个性化评估。

- **ReAct 对话 Agent 层（Python + OpenAI Function Calling）**：自行实现基于 OpenAI Function Calling 协议的 ReAct 推理循环（兼容 GLM、DeepSeek 等任意 OpenAI 兼容接口），Agent 可自主决策调度 5 类工具（文章检索、管道状态查询、RSS 源增删、偏好更新），借助 AsyncOpenAI 实现全链路 SSE 流式推理，前端以 DeepSeek 风格可折叠执行面板实时展示工具调用过程与推理链路。

**R（结果）** 系统支持多源并发抓取与三级去重，Agent 能以自然语言完成原本需要多个配置页面才能完成的操作；评估结果含 TLDR 与推理过程，可通过对话直接检索；完整交付 Go + Python + Vue 3 + Tauri 全栈闭环，含桌面端与 Telegram 双向推送控制（高分文章主动推送 + 命令调度 + 自然语言对话）。

---

## 项目整体架构

### 技术栈概览

| 层次 | 技术 | 职责 |
|------|------|------|
| 抓取层 | Go + Gin | RSS 高并发抓取、三级去重、HTTP API 网关 |
| 消息总线 | Redis Stream | 生产者-消费者异步解耦（XADD / XREADGROUP）|
| 评估层 | Python + LangGraph + LangChain | LLM 内容评估 Agent、偏好动态注入 |
| Agent 对话层 | Python + OpenAI Function Calling | ReAct 推理循环、工具调度、SSE 流式输出 |
| 持久化 | PostgreSQL | 文章、评估结果、偏好、对话历史 |
| 缓存 / 去重 | Redis | Bloom Filter、去重 Set、Stream 队列 |
| 前端 | Vue 3 + Vite + Tauri | Web 应用 + 桌面端原生应用 |
| 通知 / 控制 | Telegram Bot + python-telegram-bot | 双向控制：高分推送 + 命令 + 自然语言 Agent |

---

### 数据流

```
多个 RSS 订阅源
        │
        ▼
  Go RSS Service（协程池并发抓取，每 30 分钟轮询）
        │  ① Bloom Filter 快速拒绝（内存，< 0.1% 误判）
        │  ② Redis Set 精确去重（7 天 TTL）
        │  ③ PostgreSQL UNIQUE 约束（捕获并发竞态）
        │  ④ 短内容过滤（< 200 字符直接丢弃）
        │  ⑤ 作者白 / 黑名单过滤
        ▼
  Redis Stream（ingestion_queue）
        │  XADD 写入 / XREADGROUP 消费（消费者组 evaluators）
        ▼
  Python Stream Consumer
        │  从 DB 读取用户偏好 → 注入 system prompt
        │  LangGraph 评估 Agent（evaluate → parse → retry，最多 3 次）
        │  输出：innovation_score / depth_score / decision / tldr / reasoning
        │  高分文章触发 Telegram 推送通知
        ▼
  PostgreSQL（content + evaluation 表）
        │
        ├─► 前端 Timeline 页（卡片展示 + 详情抽屉 + 图片预览）
        │
        ├─► ReAct Agent（自然语言查询 / 管理操作）
        │       │  tool: query_articles / get_pipeline_status
        │       │  tool: add_source / remove_source / update_preferences
        │       │  SSE 流式输出 → 前端 Agent 对话页实时渲染
        │
        └─► Telegram Bot（双向控制）
                │  /status /fetch /recent 快捷命令
                │  自然语言消息 → 转发 ReAct Agent → 回复结果
                │  白名单 chat_id 鉴权（私人专属）
```

---

### 核心模块详解

#### 1. Go 抓取服务（port 8080）

**三级去重机制**（从快到慢，逐层兜底）：
- **L1 内存 Bloom Filter**：启动时从 DB 预热 7 天内的 URL hash，新请求先过此关，误判率 < 0.1%，无 I/O 开销
- **L2 Redis Set**：精确校验，7 天 TTL，单次 O(1) 查询
- **L3 PostgreSQL UNIQUE**：最终防线，捕获高并发下 L1/L2 均通过的极端竞态

**入队前质量过滤**：
- `len([]rune(content)) < 200` → 跳过（避免摘要太短的预告文章消耗 LLM token）
- 作者过滤：每个 RSS 源支持 JSONB 格式白名单 / 黑名单

#### 2. LLM 评估 Agent（LangGraph）

```
EvaluationState
    title / content / url
    → evaluate_node（LLM 调用，输出 JSON）
    → parse_node（解析并验证字段）
    → 失败 retry（最多 max_retries 次）
    → 超限抛出异常 → 调用方标记 DISCARDED
```

- 评估维度：`innovation_score`（0-10）、`depth_score`（0-10）、`decision`、`tldr`、`reasoning`、`key_concepts`
- **偏好零样本注入**：每次评估前从 `user_preferences` 表拉取 liked_topics / disliked_topics / min_score，拼入 system prompt，无需重训练即可个性化
- 串行批处理（每批 10 篇，间隔 2 秒），规避 LLM API 限速
- `eval_attempts >= 3` 自动标记 DISCARDED，防止卡队列

#### 3. ReAct 对话 Agent（port 8083）

**ReAct 推理循环**（自行实现，不依赖 Agent SDK）：
```
用户输入
  → LLM（携带 TOOL_DEFINITIONS，stream=True）
  → 返回 tool_call？
      是 → 执行工具 → 结果追加 messages → 回到 LLM（最多 8 次迭代）
      否 → 流式输出最终回答 → done 事件
```

**5 个工具**：

| 工具 | 实现方式 | 功能 |
|------|---------|------|
| `query_articles` | 直连 PostgreSQL（asyncpg）| 按关键词 / 评分 / 状态检索，返回 tldr + reasoning |
| `get_pipeline_status` | HTTP 调 Go API | 返回各状态文章数量 |
| `add_source` | HTTP 调 Go API | 添加 RSS 订阅源 |
| `remove_source` | HTTP 调 Go API | 删除 RSS 订阅源 |
| `update_preferences` | 直连 PostgreSQL | 更新偏好，实时影响后续评估 |

**全链路 SSE 流式**：
- `AsyncOpenAI(stream=True)` → 每个 text token 立即 yield `{"type":"chunk","content":"..."}`
- 工具调用过程 yield `{"type":"tool_call","tool":"...","args":{},"result":{}}`
- 完成 yield `{"type":"done"}`
- 兼容任意 OpenAI Function Calling 接口（GLM、DeepSeek、gpt-5.4 等）
- **中继兼容**：部分第三方 relay 仅在 streaming 模式下返回 content（非流式时 content 为 null），LLM 客户端统一强制 `streaming=True`

#### 4. 前端（Vue 3 + Tauri）

**三个主要页面**：
- **Timeline**：评估结果卡片流，支持 All / Interesting / Bookmark / Skip 过滤，详情抽屉含评分、TLDR、Reasoning、原文图片预览；顶部4个统计卡片 + 评估趋势折线图 + 最近活跃 RSS 源面板
- **Agent 对话页**：SSE 流式打字机效果 + DeepSeek 风格 AgentSteps（执行中展开、完成后 1.5s 自动折叠）+ 历史消息持久化（存入 PostgreSQL）
- **Config**：RSS 源管理（含推荐源弹窗）、LLM 配置（model / api_key / base_url 热加载，60s 内生效）、用户评估偏好设置

**Tauri 桌面端**：原生通知，跨平台打包，包体积远小于 Electron

**RSS 源 Favicon 自动推导**：在 Go 后端添加 RSS 源时自动从 `url` 字段解析域名，构造 `scheme://host/favicon.ico` 并写入 `sources.favicon_url`，Timeline 卡片无需手动维护图标。

#### 5. Telegram Bot（双向控制层）

私人专属 Bot，通过 `notification_settings.push_channels` JSONB 字段存储 `bot_token` + `chat_id`，启动时从 DB 读取配置，无需重启即可热更换。

**主动推送**：评估 Agent 对高分文章（decision=INTERESTING/BOOKMARK）调用 Telegram Bot API，MarkdownV2 格式发送标题、评分、TLDR 及原文链接。

**命令控制**：
| 命令 | 功能 |
|------|------|
| `/status` | 查看各状态文章数量（调 Go API） |
| `/fetch` | 触发所有 RSS 源立即抓取（调 Go API） |
| `/recent` | 展示最近 5 篇高分文章 |

**自然语言**：任意文字消息转发给 ReAct Agent，支持与前端 Agent 对话页相同的全部查询和管理能力，Bot 内即可完成原本需要打开前端才能做的操作。

**安全**：`_auth` 装饰器校验 `update.effective_chat.id` 是否在白名单，陌生人消息直接丢弃。

---

### 数据库 Schema（核心表）

| 表 | 关键字段 |
|----|---------|
| `sources` | url, priority, fetch_interval_seconds, favicon_url, author_filter (JSONB) |
| `content` | status (PENDING→PROCESSING→EVALUATED/DISCARDED), eval_attempts, image_urls, clean_content |
| `evaluation` | innovation_score, depth_score, decision, tldr, reasoning, key_concepts |
| `user_preferences` | liked_topics, disliked_topics, min_innovation_score, min_depth_score (全局 + 按源) |
| `ai_config` | default_model, api_key, base_url（热加载，60s TTL）|
| `threads` / `messages` | Agent 对话历史持久化，支持多 thread |
| `notification_settings` | score_threshold, push_channels (JSONB) |

---

### 关键设计决策

| 决策 | 原因 |
|------|------|
| Go 做抓取层 | 协程天然适配 I/O 密集的并发 HTTP 抓取；内存 Bloom Filter 性能极佳，无需外部依赖 |
| Python 做评估层 | LangGraph / LangChain 生态成熟；asyncpg 异步驱动适配 LLM 长等待场景 |
| Redis Stream 解耦 | 抓取速率与 LLM 评估速率差距悬殊（网络 I/O vs. 秒级 API 调用），Stream 天然缓冲并支持消费者组横向扩展 |
| 自行实现 ReAct 而非使用 Agent SDK | 需兼容 OpenAI Function Calling 协议的第三方接口（GLM、DeepSeek），Anthropic Agent SDK 不适用 |
| SSE 而非 WebSocket | 推理输出为单向流，SSE 更简单；无需维护双向连接状态，服务端实现零额外复杂度 |
| Tauri 而非 Electron | 包体积小 10 倍以上，Rust 安全边界更明确，原生通知支持更佳 |
| Telegram 替代 Bark | Bark 仅支持 iOS 单向推送；Telegram Bot 跨平台、双向、可挂载 ReAct Agent，一个渠道覆盖推送 + 命令 + 自然语言控制 |
| streaming=True 强制流式 | 第三方 relay 非流式响应 content 字段为 null；强制 streaming 保证与任意兼容 API 的可靠对接 |
