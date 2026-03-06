# Junk Filter

智能信息聚合与价值评估系统。从多个 RSS 源自动抓取内容，利用 LLM 评估文章的创新度和深度，帮助你从信息洪流中筛选出真正有价值的内容。

## 功能概览

- **RSS 源管理** — 添加、删除、手动同步 RSS 订阅源，支持自定义抓取频率
- **AI 内容评估** — 基于 LLM 自动评分（创新度 / 深度 0-10），生成 TLDR 摘要和关键概念
- **Timeline 时间轴** — 卡片式瀑布流展示已评估内容，支持查看评估详情和 AI 推理过程
- **Reader 阅读器** — 三栏式阅读界面（源列表 → 文章列表 → 阅读面板），按源/作者筛选
- **Task 任务对话** — 与 AI Agent 对话咨询，支持子对话（Threads）组织不同话题
- **配置中心** — 管理 RSS 源和 AI 模型参数（模型、API Key、Base URL、Temperature 等）
- **暗黑模式** — 全局亮色/暗色主题切换

## 技术架构

```
┌────────────┐     ┌────────────┐     ┌─────────────────┐
│  Vue 3     │────▶│  Go 后端    │────▶│  PostgreSQL     │
│  前端       │     │  API 网关   │     │  持久化存储      │
│  :5173     │     │  :8080     │     │  :5432          │
└────────────┘     └─────┬──────┘     └─────────────────┘
                         │
                         │ Redis Stream
                         ▼
                   ┌─────────────┐     ┌─────────────────┐
                   │  Redis      │────▶│  Python 评估服务  │
                   │  消息队列    │     │  LLM Agent      │
                   │  :6379      │     │  :8083          │
                   └─────────────┘     └─────────────────┘
```

| 组件 | 技术栈 | 职责 |
|------|--------|------|
| 前端 | Vue 3 + Pinia + Tailwind CSS | UI、状态管理、SSE 流式交互 |
| Go 后端 | Gin + go-redis + lib/pq | API 网关、RSS 抓取、三级去重、消息分发 |
| Python 服务 | FastAPI + LangChain + asyncpg | LLM 内容评估、AI Agent 对话 |
| 数据库 | PostgreSQL 15 | 源、内容、评估结果、消息持久化 |
| 消息队列 | Redis 7 (Stream) | Go → Python 异步消息传递 |

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

在项目根目录创建 `.env` 文件，必须配置以下项：

```env
# 数据库（默认值即可）
DB_USER=junkfilter
DB_PASSWORD=junkfilter123
DB_NAME=junkfilter

# LLM 配置（必填，任选一种）
OPENAI_API_KEY=sk-your-api-key
LLM_BASE_URL=https://api.openai.com/v1    # 或中转站地址
LLM_MODEL_ID=gpt-4o                        # 模型名称

# Go → Python 通信
PYTHON_API_URL=http://localhost:8083
```

### 3. 启动服务

#### 方式一：一键启动（Windows）

```bash
start-all.bat
```

脚本会依次启动 Docker 容器 → Go 后端 → Python API → Python 消费者 → 前端。

#### 方式二：手动启动（5 个终端）

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

# 终端 5: 前端（端口 5173）
cd frontend-vue
npm install    # 首次运行
npm run dev
```

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
RSS 源 → Go 抓取 → 三级去重(Bloom/Redis/PG) → Redis Stream → Python 消费者 → LLM 评估 → PostgreSQL → 前端展示
```

1. Go 后端定时抓取已启用的 RSS 源
2. 新内容经过三级去重（Bloom Filter → Redis Set → PostgreSQL UNIQUE 约束）
3. 通过 Redis Stream 推送给 Python 消费者
4. Python 调用 LLM 评估创新度和深度，生成摘要和关键概念
5. 评估结果写入 PostgreSQL
6. 前端通过 REST API 展示

## 项目结构

```
JunkFilter/
├── frontend-vue/           # Vue 3 前端
│   └── src/
│       ├── components/     # 页面组件
│       │   ├── Timeline.vue        # 时间轴
│       │   ├── Reader.vue          # 阅读器
│       │   ├── Config.vue          # 配置中心
│       │   ├── TaskDistribution.vue # 任务页面（三栏布局）
│       │   ├── TaskChat.vue        # 对话窗口
│       │   ├── ThreadList.vue      # 子对话列表
│       │   └── ChatMessage.vue     # 消息气泡
│       ├── composables/    # 可复用逻辑（useAPI, useSSE, useMarkdown）
│       ├── stores/         # Pinia 状态管理
│       └── router/         # 路由
├── backend-go/             # Go 后端
│   ├── main.go             # 入口 + 配置
│   ├── config.yaml         # YAML 配置
│   ├── handlers/           # 路由处理器
│   ├── repositories/       # 数据库访问
│   ├── models/             # 数据模型
│   └── services/           # RSS 抓取 + 去重
├── backend-python/         # Python 评估服务
│   ├── api_server.py       # FastAPI 服务（Agent 对话、评估接口）
│   ├── main.py             # Redis Stream 消费者（自动评估）
│   ├── config.py           # 配置管理
│   └── agents/             # LLM Agent
│       ├── content_evaluator.py  # 内容评估 Agent
│       └── task_analyzer.py      # 任务分析 Agent
├── sql/                    # 数据库 Schema 和迁移
├── docker-compose.yml      # PostgreSQL + Redis
├── start-all.bat/sh        # 一键启动脚本
└── .env                    # 环境变量
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
2. 确认 `.env` 中的 `OPENAI_API_KEY` 有效
3. 查看 Go 后端终端的错误日志

**Q: RSS 抓取不到内容**

部分境外 RSS 源需要代理。在 `backend-go/config.yaml` 中配置 `proxy_url`，或设置环境变量 `RSS_PROXY_URL`。

**Q: 暗黑模式下文字看不清**

已在 MVP 中修复。如果仍有问题，确保前端代码是最新版本（`cd frontend-vue && npm run dev`）。

## License

MIT
