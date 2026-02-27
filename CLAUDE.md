# CLAUDE.md

本文件为 Claude Code (claude.ai/code) 在本仓库中工作时提供指导。

## 🚨 最重要的事项：保持文件整洁

### 项目根目录文件规范（必须严格执行）

**主目录必须保持极度简洁：**
- ✅ 只允许 1 个 .md 文件：`CLAUDE.md`（本文件）
- ✅ 只允许 2 对脚本文件：`start-all.bat/sh`（启动）、`verify-day1.bat/sh`（验证）
- ✅ 其他配置文件：`docker-compose.yml`、`.env`、`go.mod`、`package.json` 等
- ❌ 禁止在主目录生成其他 .md 文件
- ❌ 禁止在主目录生成多余的 .bat/.sh 脚本

### 文档生成规则（严格遵守）

**如果需要生成新的 .md 文档：**
1. 必须放在 `description/` 文件夹中
2. 按类别分组（创建子文件夹如 `description/guides/`、`description/api/` 等）
3. 不允许在主目录生成任何文档
4. 定期审查和删除过时文档

**现有文档结构（参考）：**
```
D:\TrueSignal\
├── CLAUDE.md                      # 仅有的主目录文档
├── start-all.bat/sh              # 启动脚本对
├── verify-day1.bat/sh            # 验证脚本对
├── docker-compose.yml
└── description/                  # 所有其他文档在这里
    ├── ARCHIVE.md                # 历史文档汇总
    ├── README.md                 # 导航中心
    ├── INDEX.md                  # 完整索引
    └── *_SUMMARY.md              # 分类文档
```

---

## 重要提醒：

### 🚫 绝对禁止：

- 在缺乏证据的情况下做出假设，所有结论都必须援引现有代码或文档
- 在解决用户提出的问题时，禁止添加任何无用文档，直接给出解决方式然后解决问题
- 严厉禁止任何无用文档，直接给出解决方式！！！！
- **在项目主目录生成 .md 文件（除 CLAUDE.md 外）**
- **在项目主目录生成多余的 .bat/.sh 脚本文件**
- **创建功能重复的脚本（如多个启动脚本）**

### ✅ 必须做到：

- 在实现复杂任务前完成详尽规划并记录
- 对跨模块或超过 5 个子任务的工作生成任务分解
- 对复杂任务维护 TODO 清单并及时更新进度
- 在开始开发前校验规划文档得到确认
- 保持小步交付，确保每次提交处于可用状态
- 在执行过程中同步更新计划文档与进度记录
- 主动学习既有实现的优缺点并加以复用或改进
- 连续三次失败后必须暂停操作，重新评估策略
- **定期检查项目根目录，删除冗余的 .md 和脚本文件**
- **新增文档一定要放在 `description/` 文件夹中**

### 📋 内容唯一性规则

- 每一层级必须自洽掌握自身抽象范围，禁止跨层混用内容
- 必须引用其他层的资料而非复制粘贴，保持信息唯一来源
- 每一层级必须站在对应视角描述系统，避免越位细节
- 禁止在高层文档中堆叠实现细节，确保架构与实现边界清晰

### 🗂️ 文档整洁规范（最高优先级）

**主目录文档管理：**
- 将所有可合并的 .md 文档都放在 `description/` 文件夹
- 保持主目录只有 CLAUDE.md 一个文档
- 定期审查主目录，确保没有过时的 .md 和脚本文件
- 任何新增文档都必须创建在 `description/` 中

**脚本文件管理：**
- 主目录只保留必要的脚本：`start-all.bat/sh`、`verify-day1.bat/sh`、`start-phase3.bat/sh`（Phase 3 临时便利脚本）
- 删除所有功能重复的脚本
- 删除所有已过时或极少使用的脚本
- Phase 3 完成后可删除 `start-phase3.bat/sh`

**文档结构清晰：**
- 保持文档结构清晰，层级分明，便于快速定位信息
- 定期审查和重构文档，删除过时信息，保持内容准确
- 使用一致的格式和风格，增强可读性和专业感

## 项目概述

**TrueSignal** 是一个智能信息聚合和价值评估系统，帮助用户从多个 RSS 源中筛选出高价值、高创新度、高深度的内容。

**技术栈：**
- **后端（Go）：** 高并发 RSS 抓取、API 网关
- **评估引擎（Python）：** 异步框架进行基于大模型的内容评估
- **消息队列：** Redis Stream 用于生产者-消费者解耦
- **存储：** PostgreSQL（持久化）+ Redis（缓存与去重）

**架构：** 多服务异步系统，包含三个主要组件：
1. Go RSS 抓取服务 → 获取和去重 RSS 内容
2. Redis Stream → 异步消息总线（XADD/XREADGROUP）
3. Python 评估服务 → 用 LLM 评估内容并将结果写入 PostgreSQL

## 快速开始

### 环境设置

```bash
# 启动 Docker 容器（PostgreSQL + Redis）
docker-compose up -d

# 验证容器运行状态
docker-compose ps

# 检查 PostgreSQL 连接
docker exec truesignal-db psql -U truesignal -d truesignal -c "SELECT version();"

# 检查 Redis 连接
docker exec truesignal-redis redis-cli ping
```

### 本地运行应用

**Go 后端：**
```bash
cd backend-go
go mod download
go run main.go
# 输出显示 DB 和 Redis 连接状态
```

**Python 后端：**
```bash
cd backend-python
pip install -r requirements.txt
python main.py
# 输出显示异步初始化和连接状态
```

### 数据库访问

```bash
# 连接到 PostgreSQL
docker exec -it truesignal-db psql -U truesignal -d truesignal

# psql 中的关键命令：
\dt                    # 列出所有表
SELECT * FROM sources; # 查看 RSS 源
\q                     # 退出
```

```bash
# 连接到 Redis CLI
docker exec -it truesignal-redis redis-cli

# 关键命令：
PING                   # 测试连接
KEYS *                 # 列出所有 key
XLEN ingestion_queue   # 检查 Stream 消息数（Day 2+）
exit                   # 退出
```

## 架构与数据流

### 三级去重机制（Go 服务）

1. **L1：内存 Bloom Filter** - 7 天时间窗口，<0.1% 误触发率
2. **L2：Redis Set** - 精确校验，7 天 TTL
3. **L3：PostgreSQL UNIQUE 约束** - 最后防线，捕获竞态条件

过滤逻辑：`L1（快速拒绝）→ L2（精确检查）→ L3（数据库约束）`

### Redis Stream 数据结构（ingestion_queue）

Go 服务生产的消息包含：
```
task_id (UUID)           - 全局唯一标识符
item_hash (MD5)          - URL 或标题的哈希，用于去重
platform (enum)          - 源类型：blog, twitter, medium
author (string)          - 作者名称
clean_content (string)   - 纯文本，≤5000 字符
published_at (ISO8601)   - 发布时间
```

消费者：Python 评估服务通过 `XREADGROUP` 从 `evaluators` 消费者组读取。

### 数据库 Schema（核心表）

**sources** - RSS 源，包含优先级（1-10）和抓取间隔
**content** - 文章，待评估或已完成评估（状态机：PENDING → PROCESSING → EVALUATED/DISCARDED）
**evaluation** - 评估结果：innovation_score（0-10）、depth_score（0-10）、决策、TLDR、核心概念
**user_subscription** - 用户订阅规则，包含 min_innovation_score 和 min_depth_score 阈值
**status_log** - 状态转换审计日志，用于调试

## 开发工作流

### 配置管理

**Go：**
- 主配置：`backend-go/config.yaml`（YAML 格式）
- 覆盖：环境变量（DB_HOST, DB_PORT, DB_USER, DB_PASSWORD）
- 默认值在 `loadConfig()` 函数中

**Python：**
- 主配置：`backend-python/config.py`（Pydantic Settings）
- 覆盖：环境变量或 .env 文件
- Settings 类管理所有应用配置

### 环境变量（.env）

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=truesignal
DB_PASSWORD=truesignal123
DB_NAME=truesignal

REDIS_URL=redis://localhost:6379/0

LOG_LEVEL=INFO
```

修改这些变量来调整连接字符串或端口。

### 容器生命周期

```bash
# 停止容器（保留数据）
docker-compose stop

# 停止并删除容器（保留数据）
docker-compose down

# 停止并删除所有数据和卷
docker-compose down -v

# 从零开始重建容器
docker-compose down -v
docker-compose up -d
```

## 代码组织与模式

### Go 后端结构（计划中的扩展）

目前是单体 `main.go`，包含：
- `Config` 结构体用于 YAML 配置
- `AppContext` 全局单例，用于 DB/Redis 连接
- `loadConfig()` 用于 YAML + 环境变量解析
- `initDatabase()` 带连接池（MaxOpenConns=20, MaxIdleConns=5）
- `initRedis()` 用于 go-redis 客户端初始化

**Day 2+ 扩展计划：** 拆分为多个包：`services/`、`models/`、`repositories/`、`config/`

### Python 后端结构（计划中的扩展）

目前采用异步优先设计，包含：
- `Database` 类：管理 asyncpg 连接池（min_size=5, max_size=20）
- `Redis` 类：管理 aioredis 异步客户端
- `Settings` 类：Pydantic 配置管理
- `main()` 协程：异步初始化和清理

**Day 2+ 扩展计划：** 添加 `services/evaluator.py`、`models/`、`stream.py` 用于 Redis Stream 消费

## 常见开发任务

### 运行验证脚本

```bash
# Windows
verify-day1.bat

# Linux/Mac
chmod +x verify-day1.sh
./verify-day1.sh
```

这些脚本验证 Docker 容器、PostgreSQL、Redis 和初始数据。

### 查询应用状态

**PostgreSQL：**
```sql
-- 找出所有待处理文章
SELECT id, title, status FROM content WHERE status = 'PENDING' LIMIT 10;

-- 检查评估分布
SELECT decision, COUNT(*) FROM evaluation GROUP BY decision;

-- 查看 RSS 源健康状态
SELECT url, last_fetch_time, priority FROM sources WHERE enabled = TRUE ORDER BY priority DESC;
```

**Redis：**
```bash
# Stream 统计信息
XINFO STREAM ingestion_queue

# 消费者组详情
XINFO GROUPS ingestion_queue

# 去重 key
KEYS dedup:*

# 检查待评估任务
XPENDING ingestion_queue evaluators
```

### 重置状态

```bash
# 重置 PostgreSQL（删除所有数据，保留 schema）
docker exec truesignal-db psql -U truesignal -d truesignal -c "TRUNCATE sources, content, evaluation, user_subscription, status_log CASCADE;"

# 用演示数据重新初始化
docker exec -it truesignal-db psql -U truesignal -d truesignal < sql/schema.sql

# 清空 Redis
docker exec truesignal-redis redis-cli FLUSHDB
```

## 关键依赖

**Go：**
- `gin-gonic/gin` (v1.9.1) - Web 框架（Day 2+）
- `go-redis/redis` (v8.11.5) - Redis 客户端
- `lib/pq` (v1.10.9) - PostgreSQL 驱动
- `google/uuid` (v1.5.0) - UUID 生成
- `yaml.v3` (v3.0.1) - YAML 配置解析

**Python：**
- `aiohttp` (3.9.1) - 异步 HTTP 客户端（Day 2+ 用于 LLM API 调用）
- `asyncpg` (0.29.0) - 异步 PostgreSQL 驱动
- `pydantic` (2.5.0) - 数据验证和设置管理
- `redis` (5.0.1) - Redis 客户端（aioredis 包装器）
- `python-dotenv` (1.0.0) - .env 文件支持

## 测试与验证

**单元测试（计划 Day 2+）：**
- Go：`testing` 包 + `*_test.go` 命名约定
- Python：`pytest` 框架

**集成测试（计划 Day 2+）：**
- 端到端流程：RSS 抓取 → 去重 → 消息队列 → 评估 → 数据库写入

**手动验证：**
- 使用提供的验证脚本或直接查询数据库
- 监控日志：`docker-compose logs <service>`
- 检查 Redis Stream 深度：`docker exec truesignal-redis redis-cli XLEN ingestion_queue`

## 故障排查

**Docker 容器无法启动：**
- 检查端口 5432 或 6379 是否已被占用
- 查看日志：`docker-compose logs`
- 完全重置：`docker-compose down -v && docker-compose up -d`

**数据库连接错误：**
- 验证 `.env` 中的凭证与 docker-compose.yml 环境变量匹配
- 检查容器状态：`docker-compose ps`
- 手动测试：`docker exec truesignal-db psql -U truesignal -d truesignal -c "SELECT 1"`

**Redis 连接错误：**
- 检查 Redis 是否运行：`docker exec truesignal-redis redis-cli ping`
- 验证 .env 中的 REDIS_URL：`redis://localhost:6379/0`
- 查看日志：`docker-compose logs redis`

**Python asyncio 错误：**
- 确保没有阻塞操作（使用 `asyncio.sleep()` 而不是 `time.sleep()`）
- 检查初始化中的事件循环管理
- 验证在使用前已创建 asyncpg 连接池

## 🚫 前端约束说明（2026-02-27 更新）

### 前端修改限制

**当前状态**: 前端已完成 Phase 1-3，功能完整且稳定

**约束原则**:
- ✅ **仅允许优化相关修改**（性能、UX、样式）
- ❌ **禁止大的功能性改动**（除非必要且经过评估）
- ❌ **禁止重构现有组件结构**（TaskDistribution, TaskChat, TaskSidebar 等）
- ❌ **禁止改变 API 层设计**（useAPI, useTaskStore 已经稳定）

**理由**:
1. 前端与后端的适配涉及多个端点（tasks, messages, SSE）
2. 当前 Mock 后端已验证可用，迁移到真实后端时需要最小化改动
3. 维持代码稳定性，确保与 backend-go/backend-python 对接时风险最低

### 对接后端时的改动范围

**允许的改动**:
1. 添加数据转换层（在 useAPI.js 中）
2. 环境变量切换（VITE_API_URL 指向不同后端）
3. 新增可选功能（Config.vue 配置页增强、消息搜索等）
4. 优化相关修改（缓存、去重、智能滚动等）

**不允许的改动**:
1. 重构消息处理流程（TaskChat.vue 的核心逻辑）
2. 改变 Store 设计（useTaskStore 的 state/actions）
3. 修改 SSE 连接机制（useSSE.js 的核心）
4. 更改组件通信方式（Pinia Store → Direct Props）

### 后端迁移计划（不触发前端大改）

**阶段 1 - 现在**（保持现状）:
- 前端继续使用 Mock 后端
- 启动 Go + Python 后端做并行验证

**阶段 2 - 1-2 天**（最小化适配）:
- 在 useAPI.js 添加数据转换函数
- 修改 VITE_API_URL 指向 Go 后端
- Mock 保留用于 SSE 聊天（暂时）

**阶段 3 - 1 周**（完整迁移）:
- Go 后端补充实现消息 API 和 SSE
- 前端 VITE_API_URL 统一指向 Go
- 弃用 Mock 后端

**前端代码改动预计**: <200 行（仅数据转换和配置）

---

## 前端项目进度 (TrueSignal Junk Filter - UI)

### 当前状态 (2026-02-27)

**整体完成度: 100%** 📊

```
Phase 1-2.5 (核心功能):     ████████████ 100% ✅
Phase 3 第一步 (Mock 服务): ████░░░░░░░░  40% ✅
Phase 3 第二步 (前端集成):  ░░░░░░░░░░░░   0% ⏳
Phase 3 第三步 (完整测试):  ░░░░░░░░░░░░   0% ⏳
```

### 前端已完成的工作

#### Phase 1-2.5 完成 (6 个组件 + 5 个 Composables)
- ✅ **UI 组件**: TaskDistribution, TaskSidebar, TaskModal, TaskChat, ChatMessage, ExecutionCard
- ✅ **Composables**: useMarkdown, useScrollLock, useSSE, useToast, useThemeStore
- ✅ **状态管理**: useTaskStore, useConfigStore
- ✅ **功能完整**:
  - 创建任务 (模态框)
  - 选中任务 (侧边栏)
  - 发送消息 (输入框)
  - SSE 流式回复 (逐字显示)
  - Markdown 渲染 + XSS 防护
  - 执行卡片 (4 种状态)
  - 智能滚动锁定
  - 暗黑模式完全支持

#### 前端代码统计
- **组件数**: 6 个 (~1200 行)
- **Composables**: 5 个 (~800 行)
- **Stores**: 2 个 (~300 行)
- **样式**: Tailwind CSS (~500 行)
- **总代码量**: ~2800 行

### Phase 3 进度

#### 第一步: Mock 后端服务器 ✅ 已完成

**创建的文件**:
- `backend-mock/server.js` - 完整的 REST API + SSE 实现 (~400 行)
- `backend-mock/package.json` - 项目配置

**实现的功能**:
- ✅ 8 个 API 端点完整实现
  - GET /api/tasks - 获取任务列表
  - POST /api/tasks - 创建任务
  - GET /api/tasks/:id - 获取任务详情
  - PUT /api/tasks/:id - 更新任务
  - DELETE /api/tasks/:id - 删除任务
  - GET /api/tasks/:id/messages - 获取消息历史
  - POST /api/messages - 保存消息
  - GET /api/chat/stream - SSE 流式端点
- ✅ JSON 文件数据存储 (data/tasks.json, data/messages.json)
- ✅ CORS 跨域配置
- ✅ 错误处理和日志
- ✅ 初始演示数据

#### 第二步: useAPI Composable ✅ 已完成

**创建的文件**:
- `src/composables/useAPI.js` - 统一的 API 调用接口 (~200 行)

**实现的功能**:
- ✅ HTTP 请求通用方法
- ✅ 任务 API 组 (list, get, create, update, delete, execute)
- ✅ 消息 API 组 (list, save)
- ✅ 认证 API 组 (login, register)
- ✅ 错误处理和 Toast 集成
- ✅ 超时处理
- ✅ 重试机制 (指数退避)
- ✅ 加载状态管理

### Phase 3 接下来的工作

#### 第二步: 前端 useTaskStore 集成 (1-2 小时)

**需要修改的文件**:
- `src/stores/useTaskStore.js`

**修改内容**:
1. 导入 useAPI Composable
2. 创建 loadTasks() 方法 (调用 api.tasks.list())
3. 修改 createTask() 使用 API (调用 api.tasks.create())
4. 修改 deleteTask() 使用 API (调用 api.tasks.delete())
5. 在 onMounted() 时自动加载任务列表
6. 添加错误处理和 Toast 提示

#### 第三步: 前端 TaskChat 集成 (1-2 小时)

**需要修改的文件**:
- `src/components/TaskChat.vue`

**修改内容**:
1. 导入 useAPI Composable
2. 创建 loadMessages() 方法 (调用 api.messages.list())
3. 修改 handleSendMessage() 保存消息 (调用 api.messages.save())
4. 修改 handleSSEResponse() 保存 AI 消息
5. 添加任务切换监听 (watch selectedTaskId)
6. 添加消息加载状态和错误处理

#### 第四步: 完整集成测试 (1-2 小时)

**测试项目**:
- [ ] 启动 Mock 服务器: `node backend-mock/server.js`
- [ ] 启动前端: `npm run dev`
- [ ] 创建任务 → 验证数据保存
- [ ] 删除任务 → 验证数据删除
- [ ] 发送消息 → 验证消息保存
- [ ] 加载消息历史 → 验证数据加载
- [ ] SSE 流式回复 → 验证流式工作
- [ ] 刷新页面 → 验证数据持久化

## 项目状态更新

**Phase 3 完成状态**: ✅ **100% 完成并验证**

### 最新进展 (2026-02-27)

**SSE 流式传输问题修复完成**:
- ✅ 识别并修复 SSE 参数缺失问题
- ✅ 实现智能错误判断（有数据→成功）
- ✅ 改进消息添加逻辑（延迟添加防重复）
- ✅ 添加客户端连接监听（后端断开检测）
- ✅ 完整测试验证通过

**现在的系统状态**:
```
前端开发:   ████████████ 100% ✅
后端 Mock:  ████████████ 100% ✅
集成测试:   ████████████ 100% ✅
────────────────────────
Phase 3:    ████████████ 100% ✅
```

### 前端项目完成度

**Phase 1-2.5**: 100% ✅
- UI 组件系统（任务分发、对话交互）
- Markdown 渲染与 XSS 防护
- SSE 流式处理和智能滚动

**Phase 3**: 100% ✅
- Mock 后端服务器（8 个 API 端点）
- 前后端 API 集成
- 数据持久化验证
- **SSE 问题修复** ✅ (2026-02-27 完成)

### 快速启动 Phase 3 (前后端集成)

#### 方式 1: 使用便利脚本 (Windows)
```bash
start-phase3.bat
```

#### 方式 2: 使用便利脚本 (Linux/Mac)
```bash
chmod +x start-phase3.sh
./start-phase3.sh
```

#### 方式 3: 手动启动 (两个终端)

**终端 1 - Mock 服务器**:
```bash
cd backend-mock
node server.js
```

**终端 2 - 前端**:
```bash
cd frontend-vue
npm run dev
```

#### 访问应用
- 前端: http://localhost:5173
- Mock 服务器: http://localhost:3000

## 项目里程碑

**Day 1（已完成）：** 基础设施设置 - Docker、PostgreSQL、Redis、应用框架
**Day 2（已完成）：** 前端 Phase 1-2.5 - 核心 UI 组件、消息交互、SSE 流式
**Day 3（✅ 已完成）：** Phase 3 数据持久化和前后端集成
  - ✅ 第 1 步: Mock 服务器创建完成
  - ✅ 第 2 步: useTaskStore API 集成完成
  - ✅ 第 3 步: TaskChat API 集成完成
  - ✅ 第 4 步: SSE 流式传输问题修复完成 (2026-02-27)
  - ✅ 完整集成测试验证通过

### 修复成果

**问题**: SSE 连接失败导致"流式回复失败：连接错误"

**根本原因**: 四层问题
1. SSE 端点参数缺失（taskId, message）
2. useSSE 错误处理不当（无数据检查）
3. TaskChat 消息添加重复（无状态跟踪）
4. Mock 服务器未检测客户端断开

**修复方案**:
1. ✅ TaskChat.vue - 添加 SSE 参数 (第 228 行)
2. ✅ useSSE.js - 智能错误判断 + 流状态检查 (第 40, 121-146 行)
3. ✅ TaskChat.vue - 延迟添加消息 + 状态跟踪 (第 240-251, 286-299 行)
4. ✅ Mock 服务器 - 客户端连接监听 (第 365-432 行)

**验证结果**: ✅ 已通过完整测试

---

## 快速链接（已更新）

- 📚 **完整修复归档** → `description/PHASE3_FIX_COMPLETE_ARCHIVE.md`
- 🔍 **问题深度分析** → `description/STREAM_STATE_MANAGEMENT_ANALYSIS.md`
- 📖 **现状汇报** → `description/PHASE3_CURRENT_STATUS_REPORT.md`
- 📦 **历史文档** → `description/ARCHIVE.md`
- 🚀 **前端启动** → `start-phase3.bat/sh`
- ✅ **快速测试** → `test-sse.js`

---

## 快速链接

- 📚 **文档导航** → `description/README.md` 或 `description/INDEX.md`
- 📦 **历史文档** → `description/ARCHIVE.md`
- 🚀 **前端启动** → 在 `frontend-vue` 目录执行 `npm run dev`
- 🔧 **Mock 服务器** → 在 `backend-mock` 目录执行 `node server.js`
- ✅ **验证环境** → `verify-day1.bat`（Windows）或 `./verify-day1.sh`（Linux/Mac）
