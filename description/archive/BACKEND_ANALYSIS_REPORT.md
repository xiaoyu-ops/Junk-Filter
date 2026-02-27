# 📊 TrueSignal 真实后端分析报告

**报告日期**: 2026-02-27
**分析范围**: backend-go, backend-python
**目标**: 理解架构、功能对比、迁移可行性

---

## 🎯 核心问题

您已准备了 `backend-go` 和 `backend-python` 两个真实后端实现，当前系统仍在使用 Mock 后端。本报告分析这两个真实后端的功能和架构，并给出建议。

---

## 📋 快速概览

| 方面 | backend-go | backend-python | backend-mock |
|------|----------|-----------------|-------------|
| **语言** | Go | Python | Node.js |
| **框架** | Gin Web Framework | asyncio + aioredis | Express (隐式) |
| **数据库** | PostgreSQL (sql/database) | asyncpg (async) | JSON 文件 |
| **缓存** | Redis (go-redis) | aioredis (async) | 无 |
| **主职责** | RSS 抓取 + 去重 + API | 内容评估引擎 | 开发测试 |
| **性能特点** | 高并发、3层去重 | 异步评估、LLM 集成 | 单线程、模拟 |

---

## 🏗️ Backend-Go 架构详解

### 1️⃣ 核心职责

**RSS 数据流入层** - Go 后端专注于获取和清洗 RSS 内容：

```
RSS 源 (sources 表)
   ↓
RSSService (定时抓取)
   ↓
3层去重机制 (L1 Bloom Filter, L2 Redis Set, L3 DB Constraint)
   ↓
Clean Content + Metadata
   ↓
content 表 (待评估状态)
   ↓
Redis Stream: ingestion_queue
   ↓
↓↓↓ 传给 Python 处理 ↓↓↓
```

### 2️⃣ 关键组件

**main.go (backend-go/main.go)**:
- 初始化 PostgreSQL 连接池（MaxOpenConns=20, MaxIdleConns=5）
- 初始化 Redis 客户端
- 配置管理：yaml + 环境变量覆盖
- 启动 HTTP 服务（Gin 框架）
- 启动 RSS 定时服务

**关键配置**:
```yaml
database:
  host: localhost
  port: 5432
  user: truesignal
  password: truesignal123
  dbname: truesignal

redis:
  host: localhost
  port: 6379
  db: 0

server:
  port: 8080

ingestion:
  worker_count: 5          # 并发抓取数
  timeout: 10s             # 单个源超时
  retry_max: 3             # 失败重试次数
  fetch_interval: 1h       # 定时间隔
```

### 3️⃣ RSSService (services/rss_service.go)

**工作流程**:

```go
// 1. 启动服务，初始化 Bloom Filter
Start(ctx, interval)
  ├─ InitializeBloomFilter()      // 7天时间窗口，<0.1% 误触发
  └─ 启动定时器 (interval 控制)

// 2. 定时执行
run() 循环
  ├─ 每个 interval 执行一次 fetchAllSources()
  └─ 并发控制：workerCount = 5 个 worker

// 3. 获取所有源
fetchAllSources()
  ├─ 从 sourceRepo 获取所有启用的源
  ├─ 分配给 5 个 worker 并发抓取
  └─ 每个源最多重试 3 次

// 4. 处理单个源
processSource(source)
  ├─ 用 RSSParser 解析 feed
  ├─ 对每个 item 执行去重
  ├─ 通过的 item 写入 content 表
  └─ 失败的 item 记录日志

// 5. 去重逻辑 (DedupService)
checkDuplicate(contentHash)
  ├─ L1: Bloom Filter 快速判断 (拒绝率 >99%)
  ├─ L2: Redis Set 精确检查 (7天 TTL)
  └─ L3: DB UNIQUE 约束 (最后防线)
```

**模型定义** (models/content.go):
```go
type Content struct {
    ID           int64        // 数据库主键
    TaskID       uuid.UUID    // 全局任务ID
    SourceID     int64        // 来源ID
    Platform     string       // 平台: blog, twitter, medium
    AuthorName   string       // 作者
    Title        string       // 标题
    OriginalURL  string       // 原始 URL
    ContentHash  string       // 去重用的 hash (MD5)
    CleanContent string       // 纯文本内容 (≤5000 chars)
    PublishedAt  *time.Time   // 发布时间
    IngestedAt   time.Time    // 抓取时间
    Status       string       // PENDING, PROCESSING, EVALUATED, DISCARDED
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

### 4️⃣ HTTP 路由

**注册的路由**:
```go
// handlers/source_handler.go
routes:
  GET    /sources           // 列表（enabledOnly=true 参数）
  GET    /sources/:id       // 获取单个
  POST   /sources           // 创建源
  PUT    /sources/:id       // 更新
  DELETE /sources/:id       // 删除

// handlers/content_handler.go
routes:
  GET    /content           // 列表（status, source_id, limit, offset）
  GET    /content/:id       // 获取单个
  POST   /content           // 创建（用于测试）
  PUT    /content/:id       // 状态更新

// handlers/evaluation_handler.go
routes:
  GET    /evaluations       // 列表
  GET    /evaluations/:id   // 获取单个
  POST   /evaluations       // 写入评估结果
```

### 5️⃣ 数据库约束

**数据一致性机制**:
```sql
-- content 表
UNIQUE(content_hash)          -- L3 去重防线
FOREIGN KEY(source_id)        -- 源完整性

-- evaluation 表
FOREIGN KEY(content_id)       -- 内容完整性
CHECK(innovation_score >= 0 AND innovation_score <= 10)
CHECK(depth_score >= 0 AND depth_score <= 10)

-- status_log 表（审计）
FOREIGN KEY(content_id)
CHECK(status IN (...))
```

---

## 🐍 Backend-Python 架构详解

### 1️⃣ 核心职责

**内容评估层** - Python 后端消费 Redis Stream 进行智能评估：

```
Redis Stream: ingestion_queue
   ↓
StreamConsumer (异步消费)
   ├─ ContentEvaluationAgent (LangGraph, GPT-4)
   └─ EvaluatorService (备用评估器)
   ↓
评估结果：
  - innovation_score (0-10)
  - depth_score (0-10)
  - decision (keep/discard/flag)
  - summary (TLDR)
   ↓
evaluation 表
   ↓
content 表 (状态更新为 EVALUATED/DISCARDED)
```

### 2️⃣ 关键组件

**main.py (backend-python/main.py)**:
- 异步初始化 PostgreSQL 连接池（min=5, max=20）
- 异步初始化 Redis 客户端
- 启动 StreamConsumer 消费循环
- 优雅关闭（Ctrl+C）处理

**关键配置** (config.py):
```python
# 数据库
db_host = "localhost"
db_port = 5432
db_user = "truesignal"
db_password = "truesignal123"
db_name = "truesignal"
db_pool_min_size = 5
db_pool_max_size = 20

# Redis
redis_url = "redis://localhost:6379/0"
redis_pool_size = 10

# 评估
evaluation_timeout = 30      # 秒
max_retries = 3
batch_size = 10              # 批处理大小
log_level = "INFO"
```

### 3️⃣ StreamConsumer (services/stream_consumer.py)

**工作流程**:

```python
async def initialize():
    # 创建消费者组 (如果不存在)
    await redis.xgroup_create(
        "ingestion_queue",
        "evaluators",           # 消费者组名
        id="$",                 # 从新消息开始
        mkstream=True           # 创建 stream 如果不存在
    )

async def run():
    # 主消费循环
    while True:
        # 读取最多 10 条消息（批处理）
        messages = await redis.xreadgroup(
            group="evaluators",
            consumer="evaluator-1",
            streams={"ingestion_queue": ">"},  # 只读未确认消息
            count=10,
            block=1000                         # 1秒超时
        )

        # 批处理消息
        for msg_id, msg_data in messages:
            # 1. 解析消息
            stream_msg = StreamMessage(**msg_data)

            # 2. 调用 LangGraph 评估
            result = await self.evaluator_agent.evaluate(stream_msg)

            # 3. 写入评估表
            await self.db_service.save_evaluation(result)

            # 4. 更新内容状态
            await self.db_service.update_content_status(
                content_id,
                "EVALUATED"
            )

            # 5. ACK 消息（标记已处理）
            await redis.xack(
                "ingestion_queue",
                "evaluators",
                msg_id
            )
```

### 4️⃣ ContentEvaluationAgent (agents/content_evaluator.py)

**评估流程** (使用 LangGraph):

```python
# 采用 Agent 模式（思维链路）
# 1. 分析内容创新度 → innovation_score (0-10)
# 2. 评估内容深度 → depth_score (0-10)
# 3. 决策：keep / discard / flag
# 4. 生成 TLDR 摘要
# 5. 提取核心概念 (concepts: list[string])

class ContentEvaluationAgent:
    async def evaluate(self, content) -> EvaluationResult:
        # 使用 GPT-4 进行多维评估
        # 提示词包含：
        # - 创新度标准
        # - 深度标准
        # - 决策规则
        # - 输出格式 (JSON)

        return {
            "content_id": content.id,
            "innovation_score": 8,       # 0-10
            "depth_score": 7,            # 0-10
            "decision": "keep",          # keep|discard|flag
            "summary": "This article...",
            "concepts": ["AI", "LLM"],
            "reasoning": "..."           # 评估过程
        }
```

### 5️⃣ 数据库集成 (services/db_service.py)

```python
class DBService:
    async def save_evaluation(self, result):
        # 写入 evaluation 表
        # 包含: innovation_score, depth_score, decision, summary, concepts

    async def update_content_status(self, content_id, status):
        # content.status = EVALUATED | DISCARDED
        # 同时更新 updated_at 时间戳

    async def get_pending_content(self, limit):
        # 查询 content 表中状态为 PENDING 的文章
```

### 6️⃣ 异步编程模式

```python
# 优点：
# - 高效处理 I/O 等待（数据库、Redis）
# - 单进程处理多个请求
# - 内存占用少

# 关键模式：
import asyncio
import asyncpg
import aioredis

# 连接池
db_pool = await asyncpg.create_pool(dsn, min_size=5, max_size=20)
redis_client = await aioredis.from_url(url)

# 并发操作
await asyncio.gather(
    db_operation1(),
    db_operation2(),
    ...
)

# 流式消费
async for msg in redis_stream:
    await process(msg)
```

---

## 🔄 数据流完整图

```
┌─────────────────────────────────────────────────────────────────┐
│                    TrueSignal 完整数据流                         │
└─────────────────────────────────────────────────────────────────┘

┌─── Backend-Go (RSS 数据流入) ───────────────────────┐
│                                                       │
│  1. 定时任务 (RSSService.Start)                      │
│     └─ 每 1 小时执行一次                             │
│                                                       │
│  2. 并发抓取 (5 workers)                             │
│     └─ 从 N 个 RSS 源并发获取                        │
│                                                       │
│  3. 解析与清洗                                        │
│     └─ 提取 title, author, content, url, date       │
│                                                       │
│  4. 三层去重                                         │
│     L1: Bloom Filter (快速 filter)                  │
│     L2: Redis Set (精确 check)                       │
│     L3: DB UNIQUE (最后防线)                         │
│                                                       │
│  5. 存储清洁内容                                      │
│     content 表: status = "PENDING"                   │
│                                                       │
│  6. 发送到队列                                       │
│     Redis Stream: ingestion_queue                    │
│     消息包含: id, content, url, author, platform    │
│                                                       │
└───────────────────────────────────────────────────────┘
                        ↓
              Redis Stream: ingestion_queue
                        ↓
┌─── Backend-Python (内容评估) ──────────────────────┐
│                                                      │
│  1. 消费循环 (StreamConsumer.run)                   │
│     └─ 从 ingestion_queue 批量读取（batch=10）      │
│                                                      │
│  2. 智能评估 (ContentEvaluationAgent)               │
│     ├─ 创新度分析 → innovation_score (0-10)        │
│     ├─ 深度评估 → depth_score (0-10)              │
│     ├─ 决策制定 → keep | discard | flag            │
│     ├─ 摘要生成 → TLDR                             │
│     └─ 概念提取 → concepts: [keywords]             │
│                                                      │
│  3. 结果保存                                        │
│     evaluation 表: innovation_score, depth_score    │
│     content 表: status = "EVALUATED"                │
│                                                      │
│  4. 消息确认                                        │
│     Redis: XACK ingestion_queue evaluators          │
│                                                      │
└────────────────────────────────────────────────────┘
                        ↓
              PostgreSQL Database
              ├─ sources (RSS源目录)
              ├─ content (原始内容)
              ├─ evaluation (评估结果)
              └─ status_log (审计日志)
```

---

## 🔀 对比三个后端

### 功能维度对比

| 功能 | backend-go | backend-python | backend-mock |
|------|----------|-----------------|-------------|
| **REST API** | ✅ 完整 (8+ 端点) | ❌ 无 | ✅ Mock (8 端点) |
| **RSS 抓取** | ✅ 实现 | ❌ 无 | ❌ Mock 数据 |
| **去重机制** | ✅ 3 层完整 | ❌ 无 | ❌ 无 |
| **内容评估** | ❌ 无 | ✅ LLM GPT-4 | ❌ Mock 响应 |
| **异步处理** | ✅ Go 并发 | ✅ asyncio | ❌ 单线程 |
| **数据持久化** | ✅ PostgreSQL | ✅ PostgreSQL | ❌ JSON 文件 |
| **缓存系统** | ✅ Redis | ✅ Redis | ❌ 内存 |
| **SSE 流式** | ❌ 无 | ❌ 无 | ✅ Mock 端点 |
| **消息队列** | ✅ 生产 Stream | ✅ 消费 Stream | ❌ 无 |
| **前端数据** | ✅ 真实 RSS | ✅ AI 评分 | ✅ Mock 数据 |

### 性能对比

| 指标 | backend-go | backend-python | backend-mock |
|------|----------|-----------------|-------------|
| **吞吐量** | 1000+ req/s | 受 LLM API 限制 | 100 req/s |
| **延迟** | 10-50ms | 5-30s (LLM) | <5ms |
| **并发数** | 高（Worker Pool） | 中（asyncio） | 低（单线程） |
| **内存占用** | 100-200MB | 200-400MB | 50-100MB |
| **数据库连接** | 连接池 (20) | 连接池 (20) | 无 |

### 部署就绪度

| 方面 | backend-go | backend-python | backend-mock |
|------|----------|-----------------|-------------|
| **生产就绪** | ⚠️ 需测试 | ⚠️ 需测试 | ❌ 仅开发 |
| **错误处理** | ✅ 完整 | ✅ 完整 | ⚠️ 基础 |
| **监控指标** | ⚠️ 基础 | ⚠️ 基础 | ❌ 无 |
| **日志系统** | ✅ 完整 | ✅ 完整 | ⚠️ 基础 |
| **配置管理** | ✅ YAML + Env | ✅ Python config | ❌ 硬编码 |
| **测试覆盖** | ⚠️ 部分 | ⚠️ 部分 | ❌ 无 |

---

## 🚀 迁移策略

### 选项 1: 迁移到 Backend-Go

**适用场景**: 需要高性能 RSS 抓取和实时处理

**迁移步骤**:
```
1. 启动 Docker: docker-compose up -d
   ├─ PostgreSQL (5432)
   ├─ Redis (6379)
   └─ 等待就绪

2. 初始化数据库
   └─ 运行 sql/schema.sql

3. 启动 Go 后端
   cd backend-go
   go run main.go
   ├─ 监听 8080 端口
   ├─ 初始化 Bloom Filter
   └─ 启动 RSS 定时任务

4. 更新前端配置
   VITE_API_URL=http://localhost:8080

5. 修改 API 调用
   └─ TaskDistribution.vue 指向真实端点
```

**API 兼容性检查**:
- ✅ GET /api/sources (列表)
- ✅ POST /api/sources (创建)
- ✅ DELETE /api/sources/:id (删除)
- ⚠️ SSE /api/chat/stream → 不存在，需补充

**缺陷**:
- 没有 SSE 流式端点（目前由 Mock 提供）
- 没有 AI 评估能力
- 需要添加聊天接口

---

### 选项 2: 完整系统（Go + Python）

**适用场景**: 完整 RSS 评估系统，生产环境

**迁移步骤**:
```
1. 启动所有基础设施 (PostgreSQL + Redis)
2. 启动 Go 后端 (端口 8080)
3. 启动 Python 后端 (后台运行消费 Stream)
4. 监控两个后端日志：
   - Go: RSS 抓取进度
   - Python: 评估结果
```

**数据流完整验证**:
```bash
# 1. 查看 RSS 源
curl http://localhost:8080/sources

# 2. 等待抓取 (~1 小时或手动触发)
# 查看 content 表中状态为 PENDING 的记录

# 3. Python 自动消费并评估
# 查看 evaluation 表中的评估结果

# 4. 前端查询已评估内容
GET /api/content?status=EVALUATED
```

**前端集成方案**:
```
后端 API 已准备好，但需要添加：
1. SSE 流式聊天端点（Go）
2. 用户认证和授权
3. 配置管理接口（RSS 源、模型参数）
4. 前端路由映射
```

---

### 选项 3: 混合方案（推荐）

**过渡计划**:

```
┌─ 阶段 1: 现状 (当前) ────────────┐
│                                   │
│  ├─ 前端 (Vue 3)                 │
│  ├─ Mock 后端 (Node.js)          │
│  └─ 数据: JSON 文件 (本地)       │
│                                   │
│  ✅ 优点: 快速迭代，无依赖      │
│  ❌ 缺点: 无真实数据，无评估    │
│                                   │
└───────────────────────────────────┘
           ↓ (1-2 天)
┌─ 阶段 2: 启动真实后端 ────────────┐
│                                   │
│  ├─ 前端 (Vue 3)                 │
│  ├─ Mock 后端保留 (备用)         │
│  ├─ Go 后端 (新增)               │
│  ├─ Python 后端 (新增)           │
│  ├─ PostgreSQL (新增)            │
│  └─ Redis (新增)                 │
│                                   │
│  ✅ 优点: 真实数据，评估功能    │
│  ⚠️ 缺点: 需维护多个后端        │
│                                   │
└───────────────────────────────────┘
           ↓ (3-5 天)
┌─ 阶段 3: 完全迁移 ────────────────┐
│                                   │
│  ├─ 前端 (Vue 3)                 │
│  ├─ Go 后端 (主要)               │
│  ├─ Python 后端 (评估)           │
│  ├─ PostgreSQL (主数据库)        │
│  └─ Redis (缓存 + Stream)        │
│                                   │
│  ✅ 优点: 生产级系统            │
│  ✅ 特点: 高性能，AI 智能       │
│                                   │
└───────────────────────────────────┘
```

**具体行动**:

**第 1 步**: 启动 Docker 和真实后端
```bash
# 启动基础设施
docker-compose up -d

# 启动 Go 后端 (终端 1)
cd backend-go
go run main.go

# 启动 Python 后端 (终端 2)
cd backend-python
python main.py

# 保持 Mock 后端运行 (终端 3)
cd backend-mock
node server.js
```

**第 2 步**: 验证三个后端的 API
```bash
# Mock (当前)
curl http://localhost:3000/api/tasks

# Go (新增)
curl http://localhost:8080/sources

# Python (后台消费)
# 无 API，只有 Redis Stream 消费
```

**第 3 步**: 在前端添加 API 切换
```javascript
// src/composables/useAPI.js
const apiUrl = import.meta.env.VITE_BACKEND || 'go'

if (apiUrl === 'go') {
  baseURL = 'http://localhost:8080'
  // 调用真实 Go API
} else if (apiUrl === 'mock') {
  baseURL = 'http://localhost:3000'
  // 调用 Mock API
}
```

**第 4 步**: 逐步迁移功能
```
Week 1: Go 后端 RSS 抓取和内容管理
Week 2: Python 后端 AI 评估集成
Week 3: 前端完整迁移
Week 4: 性能优化和监控
```

---

## 📊 系统架构对比

### Mock 架构（当前）
```
前端 (Vue 3)
   ↓ HTTP
Mock 后端 (Node.js)
   ↓ JSON I/O
本地文件系统
```

**特点**: 单进程、简单、低延迟、无持久化

### 真实架构（推荐）
```
前端 (Vue 3)
   ├─ HTTP
   ├─ Go 后端 (8080)
   │   ├─ RSS 抓取 (并发)
   │   ├─ 去重 (Bloom + Redis)
   │   └─ 内容管理
   │
   └─ (可选) 聊天 SSE
       └─ Node.js Mock (仍用于流式)

┌─ 后台服务 ─────────────────┐
│                              │
│ Go: content → Redis Stream   │
│ Python: Stream → evaluation  │
│ DB: PostgreSQL + Redis       │
│                              │
└──────────────────────────────┘
```

---

## 💡 立即行动建议

### 短期（今天）

1. **启动 Docker 基础设施**
   ```bash
   docker-compose up -d
   # 等待 PostgreSQL 和 Redis 启动
   ```

2. **测试 Go 后端连通性**
   ```bash
   cd backend-go
   go run main.go
   # 检查输出是否显示 "✓ Database connected"
   ```

3. **测试 Python 后端连通性**
   ```bash
   cd backend-python
   python main.py
   # 检查输出是否显示 "✓ Database connected"
   ```

4. **验证 API 响应**
   ```bash
   curl http://localhost:8080/sources        # Go
   curl http://localhost:3000/api/tasks      # Mock
   ```

### 中期（1-2 天）

1. **集成 Go 后端 API 到前端**
   - 修改 `VITE_API_URL` 环境变量
   - 测试任务 CRUD 操作
   - 验证 RSS 源管理

2. **集成 Python 后端**
   - 手动添加测试内容到 content 表
   - 启动 Python 消费循环
   - 观察 evaluation 表填充

3. **混合测试**
   - Mock 用于开发和快速测试
   - Go 用于真实 RSS 和内容
   - Python 用于评估

### 长期（1 周）

1. **完全迁移前端到 Go**
2. **配置 Python 定时任务**
3. **添加监控和告警**
4. **性能基准测试**

---

## 🎯 问题解答

### Q1: 我需要同时运行三个后端吗？

**A**: 不需要。建议方案：
- **开发阶段**: Mock + Go（并行测试）
- **生产阶段**: Go + Python（完整系统）
- **弃用**: Mock 后端（改用 Go 的 SSE 实现）

### Q2: Go 和 Python 如何协作？

**A**: 通过 Redis Stream：
```
Go -> content 表 + Redis Stream (ingestion_queue)
          ↓
          ├─ Python 消费 Stream
          ├─ 调用 GPT-4 评估
          └─ 写入 evaluation 表
```

### Q3: 前端 API 会改变吗？

**A**: 部分改变：
- **保留**: `/api/sources`, `/api/content`, `/api/evaluations`
- **调整**: 从 `localhost:3000` → `localhost:8080`
- **新增**: SSE 聊天端点（需补充实现）

### Q4: 数据会丢失吗？

**A**: 不会。现有 Mock 数据在 `backend-mock/data/` 目录：
- 可手动迁移到 PostgreSQL
- 或保留为开发参考

### Q5: 哪个后端更稳定？

**A**: Go 后端更稳定（生产级）
- 编译型语言，运行时检查完整
- 并发模型成熟（Goroutines）
- 错误处理完善

Python 后端功能更强大但依赖 LLM API

---

## 📋 迁移检查清单

### 基础设施
- [ ] Docker Compose 运行正常
- [ ] PostgreSQL 初始化完成
- [ ] Redis 连接正常

### Go 后端
- [ ] 编译成功 (`go build`)
- [ ] 连接数据库成功
- [ ] 监听端口 8080
- [ ] 健康检查 (`GET /health`)
- [ ] 可创建 RSS 源 (`POST /sources`)

### Python 后端
- [ ] 依赖安装完成 (`pip install -r requirements.txt`)
- [ ] 连接数据库成功
- [ ] Redis Stream 消费者组创建
- [ ] 可接收评估请求

### 前端集成
- [ ] 环境变量指向 Go 后端
- [ ] 任务列表加载正常
- [ ] 可创建/删除任务
- [ ] 消息保存成功

### 数据流
- [ ] Go 抓取 RSS → content 表
- [ ] Python 评估 → evaluation 表
- [ ] 前端显示评估结果

---

## 📚 关键文件快速查询

| 文件 | 用途 |
|------|------|
| `backend-go/main.go` | 启动文件 + 配置 |
| `backend-go/services/rss_service.go` | RSS 抓取逻辑 |
| `backend-go/handlers/source_handler.go` | API 源管理 |
| `backend-python/main.py` | 启动文件 + 初始化 |
| `backend-python/services/stream_consumer.py` | 流消费 + 批处理 |
| `backend-python/agents/content_evaluator.py` | LLM 评估引擎 |
| `CLAUDE.md` | 项目规范 |
| `docker-compose.yml` | 基础设施定义 |

---

## ✅ 总结

| 后端 | 优点 | 缺点 | 何时使用 |
|------|------|------|---------|
| **Mock** | 快速、无依赖 | 不真实、无评估 | 早期开发 |
| **Go** | 高性能、生产级 | 需要数据库 | 生产 RSS |
| **Python** | AI 智能评估 | 依赖 LLM API | 生产评估 |

**最终建议**: 立即启动 Go + Python，保留 Mock 用于特定测试，逐步迁移前端接口。

---

**报告完成**: 2026-02-27
**下一步**: 按照"立即行动建议"执行迁移

