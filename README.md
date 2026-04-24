# Junk Filter

智能信息聚合与价值评估系统。从多个 RSS 源自动抓取内容，利用 LLM 评估文章的创新度和深度，帮助你从信息洪流中筛选出真正有价值的内容。

## 功能概览

- **RSS 源管理** — 添加、删除、手动同步 RSS 订阅源，支持自定义抓取频率
- **AI 内容评估** — 基于 LLM 自动评分（创新度 / 深度 0-10），生成 TLDR 摘要和关键概念
- **Timeline 时间轴** — 卡片式瀑布流展示已评估内容，支持查看评估详情和 AI 推理过程
- **Reader 阅读器** — 三栏式阅读界面（源列表 → 文章列表 → 阅读面板），按源/作者筛选
- **Task 任务对话** — 与 AI Agent 对话咨询，支持子对话（Threads）组织不同话题
- **配置中心** — 管理 RSS 源和 AI 模型参数（模型、API Key、Base URL、Temperature 等）
- **推送通知** — 高分文章触发 Telegram Bot 推送 + 前端 SSE 实时通知
- **暗黑模式** — 全局亮色/暗色主题切换

## 技术架构

```
                          ┌─────────────────┐
                          │  PostgreSQL     │
                          │  :5432          │
                          └────────▲────────┘
                                   │
┌────────────┐     ┌───────────────┼─────────────┐
│  Vue 3     │◀───▶│   Go 后端     │             │
│  前端       │     │   API 网关    │             │
│  :5173     │     │   :8080       │             │
└────────────┘     └───────┬───────┘             │
                           │                     │
              Redis Stream │                     │
                           ▼                     │
                    ┌─────────────┐              │
                    │  Redis      │              │
                    │  :6379      │              │
                    └──────┬──────┘              │
                           │                     │
              ┌────────────┼────────────┐       │
              │            │            │       │
              ▼            ▼            ▼       │
    ┌─────────────┐ ┌──────────┐ ┌──────────┐  │
    │  Python     │ │ Telegram │ │ 通知 SSE │  │
    │  评估服务    │ │   Bot    │ │  推流    │──┘
    │  :8083      │ │          │ │ :8080    │
    └─────────────┘ └──────────┘ └──────────┘
```

| 组件 | 技术栈 | 职责 |
|------|--------|------|
| 前端 | Vue 3 + Pinia + Tailwind CSS | UI、状态管理、SSE 流式交互 |
| Go 后端 | Gin + go-redis + lib/pq | API 网关、RSS 抓取、三级去重、消息分发 |
| Python 服务 | FastAPI + LangGraph + asyncpg | LLM 内容评估、ReAct Agent 对话 |
| Telegram Bot | python-telegram-bot v21+ | 推送通知、命令控制、自然语言交互 |
| 数据库 | PostgreSQL 15 | 源、内容、评估结果、消息持久化 |
| 消息队列 | Redis 7 (Stream + Pub/Sub) | Go → Python 异步消息传递、通知广播 |

## 前置要求

- [Docker](https://www.docker.com/) & Docker Compose
- [Go](https://go.dev/) 1.21+
- [Python](https://www.python.org/) 3.10+（推荐 Conda 管理环境）
- [Node.js](https://nodejs.org/) 18+
- 一个 OpenAI 兼容的 API Key（支持中转站）

## 快速开始

### 1. 克隆项目

```bash
git clone <repo-url>
cd JunkFilter
```

### 2. 配置环境变量

在项目根目录创建 `.env` 文件（仅用于 Docker 容器变量替换）：

```env
# 数据库（默认值即可，本地开发通常不需要改）
DB_USER=junkfilter
DB_PASSWORD=junkfilter123
DB_NAME=junkfilter
```

> **注意**：LLM 配置（API Key、模型、Base URL）**不在 `.env` 中配置**。启动后通过前端 **配置中心** 页面填写，保存到 PostgreSQL 的 `ai_config` 表。Consumer 和 Agent 会自动从 DB 读取，支持热加载（无需重启进程）。

### 3. 启动服务

#### 方式一：一键启动（Mac/Linux）

```bash
./start-all.sh   # 启动所有服务（Docker + Go + Python + 前端）
./stop-all.sh    # 停止所有服务
```

#### 方式二：一键启动（Windows）

```bash
start-all.bat
```

#### 方式三：手动启动（6 个终端）

```bash
# 终端 1: 启动 Docker 容器（PostgreSQL + Redis）
docker-compose up -d

# 终端 2: Go 后端（端口 8080）
cd backend-go
go run main.go

# 终端 3: Python API 服务（端口 8083）
cd backend-python
conda activate junkfilter    # 如使用 Conda
python api_server.py

# 终端 4: Python Stream 消费者（后台自动评估）
cd backend-python
python main.py

# 终端 5: Telegram Bot（可选，需先在前端配置 bot_token + chat_id）
cd backend-python
python telegram_bot.py

# 终端 6: 前端（端口 5173）
cd frontend-vue
npm install    # 首次运行
npm run dev
```

> VS Code 用户可按 `Cmd+Shift+B`（Mac）一键启动全部 6 个服务，每个独占一个终端 Tab。

### 4. 打开浏览器

访问 http://localhost:5173

## 使用指南

### 添加 RSS 源

1. 进入顶部导航栏 → **配置**
2. 点击「添加订阅源」
3. 填入源名称、RSS Feed URL，选择更新频率
4. 点击添加，系统立即开始抓取

> 系统预置了 26 个常见 RSS 源（微博热搜、GitHub 趋势、掘金热门等）作为 AI 推荐候选池，不会出现在你的源列表中。

### 查看评估结果

- **Timeline** — 所有已评估文章以卡片展示，包含评分、AI 摘要、关键概念标签。点击卡片查看详情和 AI 推理过程
- **Reader** — 三栏阅读器，左侧按源分组浏览，中间选择文章，右侧阅读全文和 AI 分析

### 手动触发抓取

在配置中心点击源右侧的同步按钮，立即触发 RSS 抓取。新内容会通过 Redis Stream 推送给 Python 评估服务自动评估。

### AI 对话

1. 进入顶部导航栏 → **任务**
2. 左侧选择一个任务（RSS 源）
3. 在聊天框与 AI Agent 对话，Agent 了解你的源数据和评估结果
4. 可创建子对话来组织不同话题

### 配置 AI 模型

在配置中心底部设置 LLM 参数：

| 参数 | 说明 | 推荐值 |
|------|------|--------|
| 模型名称 | 如 `gpt-4o`、`deepseek-chat`、`glm-4` | — |
| API 密钥 | OpenAI 兼容的 API Key | — |
| Base URL | API 端点，支持中转站 | — |
| Temperature | 生成随机性 0-1 | 0.7 |
| Max Tokens | 最大输出 token | 2048 |

配置后点击「保存配置」，下次对话和评估时自动使用。

## 数据流

```
RSS 源 → Go 抓取 → 三级去重(Bloom/Redis/PG UNIQUE) → content 表(PENDING)
                                                    ↓
                                              Redis Stream(XADD)
                                                    ↓
                                        Python Consumer(XREADGROUP)
                                                    ↓
                                        LLM 评估(创新度/深度/决策)
                                                    ↓
                                        evaluation 表 + content→EVALUATED
                                                    ↓
                              ┌─────────────────────┼─────────────────────┐
                              ▼                     ▼                     ▼
                          前端展示          通知表+Redis Pub/Sub      Telegram Bot
                          REST API              SSE 推流               push_service
```

1. **Go 后端** 定时抓取已启用的 RSS 源
2. **三级去重**：Bloom Filter（内存）→ Redis Set（URL 去重）→ PostgreSQL UNIQUE 约束（兜底）
3. **短内容过滤**：不足 200 字符的文章直接丢弃；作者黑白名单过滤
4. 新内容写入 `content` 表（`status=PENDING`），同时通过 **Redis Stream** 推送给 Python 消费者
5. **Python Consumer** 从 Stream 批量读取，调用 LLM 评估（创新度 0-10、深度 0-10、决策 INTERESTING/BORING）
6. 评估结果写入 `evaluation` 表，`content.status` 更新为 `EVALUATED`
7. 高分内容触发**通知**：写入 `notifications` 表 → Redis Pub/Sub 广播 → 前端 SSE 实时推流 + Telegram Bot 推送
8. 前端通过 REST API 展示时间轴、阅读器、任务对话等内容

## 项目结构

```
JunkFilter/
├── frontend-vue/              # Vue 3 前端 (:5173)
│   ├── src/
│   │   ├── components/        # 页面组件
│   │   │   ├── AppNavbar.vue           # 顶部导航栏
│   │   │   ├── Home.vue                # 首页
│   │   │   ├── Timeline.vue            # 时间轴（卡片瀑布流）
│   │   │   ├── Reader.vue              # 三栏阅读器
│   │   │   ├── Config.vue              # 配置中心（RSS源 + LLM配置）
│   │   │   ├── TaskDistribution.vue    # 任务页（三栏：任务列表/对话/详情）
│   │   │   ├── Task.vue                # 任务列表
│   │   │   ├── TaskChat.vue            # AI 对话窗口（SSE 流式）
│   │   │   ├── TaskModal.vue           # 任务详情弹窗
│   │   │   ├── TaskSidebar.vue         # 任务侧边栏
│   │   │   ├── ThreadList.vue          # 子对话列表
│   │   │   ├── ChatMessage.vue         # 消息气泡
│   │   │   ├── AIAssistantModal.vue    # AI 助手弹窗
│   │   │   ├── AgentSteps.vue          # Agent 执行步骤展示
│   │   │   ├── NotificationBell.vue    # 通知铃铛
│   │   │   ├── SearchBar.vue           # 搜索栏
│   │   │   ├── EmptyState.vue          # 空状态占位
│   │   │   ├── ErrorCard.vue           # 错误卡片
│   │   │   ├── ExecutionCard.vue       # 执行卡片
│   │   │   ├── ExecutionHistoryModal.vue # 执行历史弹窗
│   │   │   └── SkeletonLoader.vue      # 骨架屏
│   │   ├── composables/       # Vue 3 Composition 复用逻辑
│   │   │   ├── useAPI.js              # HTTP 请求封装
│   │   │   ├── useSSE.js              # SSE 流式连接
│   │   │   ├── useMarkdown.js         # Markdown 渲染
│   │   │   ├── useSearch.js           # 内容搜索
│   │   │   ├── useNotification.js     # 通知管理
│   │   │   ├── useToast.js            # Toast 提示
│   │   │   ├── useFormValidation.js   # 表单校验
│   │   │   ├── useDetailDrawer.js     # 详情抽屉
│   │   │   └── useScrollLock.js       # 滚动锁定
│   │   ├── stores/            # Pinia 状态管理
│   │   │   ├── useTimelineStore.js    # 时间轴状态
│   │   │   ├── useReaderStore.js      # 阅读器状态
│   │   │   ├── useTaskStore.js        # 任务/对话状态
│   │   │   ├── useConfigStore.js      # 配置状态
│   │   │   ├── useThemeStore.js       # 主题状态
│   │   │   └── index.js               # Store 聚合导出
│   │   ├── router/            # Vue Router 路由配置
│   │   ├── App.vue            # 根组件
│   │   └── main.js            # 入口
│   ├── src-tauri/             # Tauri 桌面端封装（可选）
│   ├── package.json           # 依赖
│   └── vite.config.js         # Vite 构建配置
│
├── backend-go/                # Go API 网关 + RSS 抓取 (:8080)
│   ├── main.go                # 入口：配置加载、依赖初始化、服务启动
│   ├── config.yaml            # 基础设施配置（DB/Redis/CORS/抓取参数）
│   ├── handlers/              # Gin 路由处理器（按模块组织）
│   │   ├── source_handler.go         # RSS 源 CRUD
│   │   ├── content_handler.go        # 内容查询、搜索
│   │   ├── evaluation_handler.go     # 评估结果查询
│   │   ├── message_handler.go        # 消息记录
│   │   ├── task_chat_handler.go      # 任务聊天（SSE，转发 Python）
│   │   ├── ai_task_handler.go        # AI 任务分析
│   │   ├── thread_handler.go         # 子对话 CRUD
│   │   ├── config_handler.go         # LLM 配置读写
│   │   ├── notification_handler.go   # 通知 SSE 推流
│   │   ├── search_handler.go         # 全文搜索
│   │   ├── routes.go                 # 路由注册
│   │   └── sse_helpers.go            # SSE 工具函数
│   ├── services/              # 业务逻辑层
│   │   ├── rss_service.go            # RSS 定时抓取（协程池）
│   │   ├── content_service.go        # 内容去重（Bloom/Redis/PG 三级）
│   │   └── dedup_service.go          # 去重策略封装
│   ├── repositories/          # 数据访问层（DAO）
│   │   ├── source_repo.go            # sources 表
│   │   ├── content_repo.go           # content 表
│   │   ├── evaluation_repo.go        # evaluation 表
│   │   ├── message_repository.go     # messages 表
│   │   └── thread_repository.go      # threads 表
│   ├── models/                # 数据模型（struct 定义）
│   │   ├── source.go
│   │   ├── content.go
│   │   ├── evaluation.go
│   │   ├── message.go
│   │   └── junk_filter.go
│   ├── utils/                 # 工具函数
│   │   └── rss_parser.go             # RSS XML 解析
│   └── sql/                   # 初始化 SQL（预置 RSS 源）
│       └── init_sources.sql
│
├── backend-python/            # Python LLM 评估 + ReAct Agent + Bot (:8083)
│   ├── api_server.py          # FastAPI：Agent 对话、评估接口、健康检查
│   ├── main.py                # Redis Stream 消费者（自动评估后台进程）
│   ├── telegram_bot.py        # Telegram Bot（推送+命令+自然语言交互）
│   ├── config.py              # 配置管理（pydantic-settings + DB 热加载）
│   ├── repush_pending.py      # 手动重推 PENDING 文章到 Stream（调试工具）
│   ├── agents/                # Agent 核心
│   │   ├── content_evaluator.py      # 内容评估 Agent（LangGraph）
│   │   ├── react_agent.py            # ReAct Agent 主循环（SSE 流式）
│   │   ├── task_analyzer.py          # 任务分析 Agent
│   │   ├── tools.py                  # Agent 可用工具定义
│   │   └── preference_tools.py       # 用户偏好加载与格式化
│   ├── services/              # 服务层
│   │   ├── stream_consumer.py        # Redis Stream 消费者（评估管道核心）
│   │   ├── db_service.py             # DB 操作封装
│   │   ├── evaluator.py              # 评估器封装
│   │   └── push_service.py           # 推送通道（Telegram/邮件等）
│   ├── models/                # Pydantic 模型
│   │   ├── evaluation.py
│   │   └── ai_task.py
│   └── utils/                 # 工具函数
│       └── llm_client.py             # LLM 客户端封装
│
├── sql/                       # 数据库 Schema 迁移（按序号执行）
│   ├── 01_junk_filter_schema.sql
│   ├── 02_schema.sql
│   ├── 03_migration_chat_v2.sql
│   ├── 04_init_sources.sql
│   ├── 05_add_favicon.sql
│   ├── 06_add_threads.sql
│   ├── 07_add_image_urls.sql
│   ├── 08_add_author_filter.sql
│   ├── 09_add_user_preferences.sql
│   ├── 10_add_notifications.sql
│   ├── 11_add_notification_settings.sql
│   ├── 12_add_push_channels.sql
│   └── 13_add_eval_attempts.sql
│
├── description/               # 项目文档（架构、指南、排错手册）
│   ├── 技术全景手册.md
│   ├── 项目架构说明.md
│   ├── 功能演进计划.md
│   ├── 常见报错手册.md
│   ├── 导航索引.md
│   └── guides/              # 开发指南
│       ├── VSCode开发指南.md
│       ├── LLM配置指南.md
│       └── 代码解剖指南.md
│
├── docker-compose.yml         # PostgreSQL + Redis 容器编排
├── start-all.sh               # Mac/Linux 一键启动
├── stop-all.sh                # Mac/Linux 一键停止
├── problem.md                 # 已知问题与优化计划
├── CLAUDE.md                  # 项目级 Claude Code 指令
└── .env                       # Docker 环境变量（仅容器用）
```

## 常见问题

**Q: Docker 容器启动失败**

检查端口 5432 和 6379 是否被占用：
```bash
netstat -ano | findstr "5432 6379"
```

**Q: Go 后端连不上 Python 服务**

确认环境变量 `PYTHON_API_URL=http://localhost:8083`。如果是系统级环境变量设置了旧值，需要在系统设置中修改或在启动脚本中覆盖。

**Q: AI 对话没有响应**

1. 确认 Python API 服务已启动（`python api_server.py`，端口 8083）
2. 确认前端 **配置中心** 已填写有效的 LLM 配置（API Key、模型、Base URL）
3. 查看 Python API 终端的错误日志（常见原因：API Key 无效、中转站 504）

**Q: RSS 抓取不到内容**

部分境外 RSS 源需要代理。在 `backend-go/config.yaml` 中配置 `proxy_url`，或设置环境变量 `RSS_PROXY_URL`。

**Q: Telegram Bot 无响应**

1. 确认 Bot 进程在运行（`ps aux | grep telegram_bot`）
2. 确认前端 **配置中心 → 消息推送** 已填写有效的 `bot_token` 和 `chat_id`
3. Mac 用户：若息屏后 Bot 死亡，改用 LaunchAgent 管理（见 `description/技术全景手册.md` 6.1.1 节）

**Q: Consumer 卡住不评估**

1. 检查 LLM 配置是否有效（前端配置中心）
2. 检查 `python main.py` 终端是否有错误日志
3. 极端情况下调用管理端点：`curl -X POST http://localhost:8080/api/admin/purge-stream`

## License

MIT
