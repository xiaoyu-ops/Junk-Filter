# JunkFilter - 智能 RSS 内容评估系统

> 从海量 RSS 源中智能筛选高价值、高创新度、高深度的内容

## 📖 项目简介

**JunkFilter** 是一个全栈智能信息聚合和价值评估系统。它能帮助你：

- 🔗 **管理多个 RSS 源** - 集中聚合各类内容源
- 🤖 **智能内容评估** - 用大模型（LLM）自动评估内容的创新度和深度
- 💬 **Agent 聊天咨询** - 与 AI Agent 对话，优化评估规则
- 💾 **持久化存储** - 所有内容和评估结果保存到数据库
- 📊 **实时流式展示** - Server-Sent Events (SSE) 实现流式响应

### 核心工作流

```
RSS 源 → Go 后端（抓取 + 去重）→ Redis Stream（消息队列）
                                    ↓
                        Python 评估引擎（LLM 评估）
                                    ↓
                            评估结果写入数据库
                                    ↓
前端展示 ← Go API ← 数据库 / 缓存
                ↑
           用户配置和聊天
```

---

## 🚀 快速开始

### 系统要求

- **Docker & Docker Compose**
- **Go 1.18+**（如本地运行后端）
- **Python 3.10+**（如本地运行评估引擎）
- **Node.js 16+**（前端开发）
- **现代浏览器**（支持 SSE、ES6）

### 1️⃣ 启动全栈系统

#### 方式一：使用启动脚本（推荐）

**Windows**：
```bash
start-all.bat
```

**Linux/Mac**：
```bash
chmod +x start-all.sh
./start-all.sh
```

#### 方式二：手动启动

**终端 1 - Docker 基础设施**：
```bash
docker-compose up -d
```

**终端 2 - Go 后端**：
```bash
cd backend-go
go run main.go
```

**终端 3 - Python 评估引擎**：
```bash
cd backend-python
python main.py
```

**终端 4 - Vue 前端**：
```bash
cd frontend-vue
npm install
npm run dev
```

### 2️⃣ 访问应用

打开浏览器访问：

```
前端应用:   http://localhost:5173
Go 后端:    http://localhost:8080
Python 后端: http://localhost:8081
```

---

## 💡 如何使用

### 场景 1: 添加 RSS 源并查看配置

#### 步骤 1: 打开配置面板

点击导航栏中的 **"配置"** 按钮，进入 `/config` 页面。

#### 步骤 2: 添加 RSS 源

在 **"订阅源管理"** 部分：
1. 点击 **"添加订阅源"** 按钮
2. 填入信息：
   - **源名称**：例如 "TechCrunch"
   - **RSS Feed URL**：例如 `https://techcrunch.com/feed/`
   - **更新频率**：每小时 / 每30分钟 / 每2小时 / 每天
   - **过滤规则**（可选）：例如 "优先级 >= 7"

3. 点击 **"添加"** 按钮

系统将立即开始从该 RSS 源抓取内容。

#### 步骤 3: 查看源状态和同步日志

在订阅源列表中：
- **状态列** 显示源是否活跃
- 点击左侧 **展开按钮** 查看同步日志
- **手动同步** 按钮：立即触发一次抓取
- **删除** 按钮：移除该源

---

### 场景 2: 配置 AI 模型进行智能评估

#### 步骤 1: 进入 AI 模型配置

在配置面板的 **"AI 模型配置"** 部分填入你的 LLM 参数。

#### 步骤 2: 填入 LLM 信息

**支持三种方案**：

**方案 A: 使用 OpenAI 官方 API**
```
模型名称:  gpt-4-turbo
API 密钥:  sk-xxxx（从 https://platform.openai.com/api-keys 获取）
Base URL:  https://api.openai.com/v1
```

**方案 B: 使用中转站（推荐测试）**
```
模型名称:  gpt-4o 或 gemini-2.5-pro
API 密钥:  sk-xxxx（中转站提供的密钥）
Base URL:  https://elysiver.h-e.top/v1
```

**方案 C: 使用 DeepSeek API**
```
模型名称:  deepseek-chat
API 密钥:  sk-xxxx（从 https://platform.deepseek.com/ 获取）
Base URL:  https://api.deepseek.com
```

#### 步骤 3: 调整评估参数

- **温度 (Temperature)**：0.0-1.0
  - 低值（0.0-0.3）：更保守、确定性强
  - 高值（0.7-1.0）：更有创意、多样性强
  - 推荐：0.7

- **Top P (核采样)**：0.0-1.0
  - 控制词汇多样性
  - 推荐：0.9

- **最大 Token 数**：
  - 限制 AI 回复长度
  - 推荐：2048

#### 步骤 4: 保存配置

点击 **"保存配置"** 按钮，配置将保存到浏览器 localStorage。

✅ **配置已保存！下次聊天时自动使用**

---

### 场景 3: 与 Agent 进行对话，优化评估规则

#### 步骤 1: 选择一个任务

从左侧任务列表中点击选择一个 RSS 源（任务）。

#### 步骤 2: 打开聊天界面

右侧显示该任务的 **聊天历史** 和 **消息输入框**。

#### 步骤 3: 与 Agent 对话

输入你的问题或指令，按 **Enter** 发送：

**示例对话**：

```
用户: "现在的执行进度如何？"

Agent: "当前任务已处理消息: 127 条
⭐ 高价值内容: 12 条 (9.4%)
📌 已书签: 38 条 (29.9%)
⏭️ 跳过: 77 条 (60.6%)
最近 1 小时内，系统识别出 3 条创新度和深度都达到 8 分以上的高质量内容。"

---

用户: "接下来多关注深度学习的论文"

Agent: "好的，我理解你想调整评估规则。
当前的过滤规则是: default
我可以帮助你修改以下参数：
- **过滤关键词**: 比如添加 "深度学习" 来专注 AI 话题
- **评估敏感度**: 调整 temperature (0.0-1.0)，数值越高越"有创意"
- **内容长度偏好**: 通过 max_tokens 调整
建议: 告诉我你想关注的具体话题，我会给出参数建议。"

---

用户: "为什么这张卡片被标记为 SKIP？"

Agent: "参考卡片 #123：这篇内容的创新度为 4 分，深度为 3 分，
均未达到你设定的阈值。内容主要重述已知信息，缺乏新的见解。
建议: 如果你认为这个话题有价值，可以调整 min_innovation_score 阈值。"
```

#### 步骤 4: Agent 提供参数建议

Agent 可能建议调整：
- `temperature` - 调整创意度
- `topP` - 调整词汇多样性
- `maxTokens` - 调整响应长度
- `filter_rules` - 调整过滤规则

根据建议回到 **配置面板** 修改参数并重新保存。

---

### 场景 4: 查看和搜索消息历史

#### 搜索功能

在聊天窗口顶部搜索框输入关键词，实时搜索该任务的所有消息。

#### 过滤功能

使用过滤器快速定位：
- **日期范围**：全部时间 / 今天 / 本周 / 本月
- **消息状态**：全部 / 未读 / 已读
- **导出格式**：Markdown / JSON / CSV

#### 导出消息

点击 **"导出"** 按钮，将消息导出为选定格式供本地保存和分析。

---

## 🎯 核心功能

### 🔧 后端（Go）

| 功能 | 说明 |
|------|------|
| **RSS 源管理** | 添加、删除、启用/禁用 RSS 源 |
| **内容抓取** | 定期从 RSS 源抓取新内容 |
| **去重系统** | 三层去重（Bloom Filter → Redis → DB 约束） |
| **消息队列** | Redis Stream 解耦生产者和消费者 |
| **API 服务** | RESTful API 提供所有操作接口 |
| **聊天接口** | Task Chat API 支持 Agent 对话 |

### 🧠 评估引擎（Python）

| 功能 | 说明 |
|------|------|
| **内容评估** | 计算创新度和深度分数（0-10） |
| **决策制定** | 自动标记为 INTERESTING / BOOKMARK / SKIP |
| **LLM 集成** | 支持 OpenAI、DeepSeek、中转站等 |
| **Agent 聊天** | 自然语言回复，提供参数优化建议 |
| **流式响应** | SSE 实现实时响应流 |

### 💻 前端（Vue 3）

| 功能 | 说明 |
|------|------|
| **任务管理** | 创建、查看、删除任务 |
| **配置面板** | 直观的 RSS 源和 LLM 参数配置 |
| **实时聊天** | 与 Agent 进行实时对话 |
| **消息管理** | 搜索、过滤、导出消息历史 |
| **暗黑模式** | 完整的深色主题支持 |
| **响应式设计** | 完美适配各种屏幕尺寸 |

---

## 📊 API 参考

### Go 后端 API

#### 获取所有 RSS 源
```bash
GET http://localhost:8080/api/sources
```

**响应示例**：
```json
[
  {
    "id": 1,
    "url": "https://techcrunch.com/feed/",
    "author_name": "TechCrunch",
    "priority": 5,
    "enabled": true,
    "last_fetch_time": "2026-03-01T15:30:00Z",
    "created_at": "2026-03-01T10:00:00Z"
  }
]
```

#### 创建新 RSS 源
```bash
POST http://localhost:8080/api/sources
Content-Type: application/json

{
  "url": "https://techcrunch.com/feed/",
  "author_name": "TechCrunch",
  "priority": 5,
  "enabled": true
}
```

#### 任务聊天（Agent 对话）
```bash
POST http://localhost:8080/api/tasks/{id}/chat
Content-Type: application/json

{
  "message": "现在的执行进度如何？"
}
```

**响应**：Server-Sent Events (SSE) 流式响应
```
data: {"status":"processing","phase":"analyzing"}

data: {"status":"completed","result":{"reply":"当前任务已处理消息: 127 条..."}}
```

### Python 后端 API

#### 健康检查
```bash
GET http://localhost:8081/health
```

**响应**：
```json
{
  "status": "healthy",
  "database": "connected",
  "redis": "connected",
  "llm": "configured"
}
```

#### 评估内容（同步）
```bash
POST http://localhost:8081/api/evaluate
Content-Type: application/json

{
  "title": "深度学习新突破",
  "content": "最近研究表明...",
  "temperature": 0.7,
  "maxTokens": 2048
}
```

**响应**：
```json
{
  "title": "深度学习新突破",
  "content": "最近研究表明...",
  "innovation_score": 8,
  "depth_score": 9,
  "decision": "INTERESTING",
  "tldr": "这项研究在深度学习优化方面取得重要进展...",
  "key_concepts": ["深度学习", "优化算法", "神经网络"]
}
```

---

## 🔐 安全特性

### ✅ 已实现的安全措施

| 措施 | 说明 |
|------|------|
| **CORS 严格模式** | 仅允许 localhost:5173 访问 API |
| **环境变量驱动** | 所有敏感配置从环境变量读取 |
| **Redis 密码认证** | Redis 启用强密码保护 |
| **DB SSL 支持** | 支持生产环境 SSL/TLS 连接 |
| **日志脱敏** | 敏感信息（如 API URL）仅记录主机名 |
| **参数化配置** | 数据库凭证不硬编码在代码中 |

### 📋 环境变量配置

创建 `.env` 文件（参考 `.env.example`）：

```bash
# 数据库
DB_HOST=localhost
DB_PORT=5432
DB_USER=truesignal
DB_PASSWORD=your_secure_password
DB_NAME=truesignal
DB_SSL_MODE=disable  # 开发；生产改为 require

# Redis
REDIS_URL=redis://localhost:6379/0
REDIS_PASSWORD=your_secure_redis_password

# CORS（严格白名单）
CORS_ALLOWED_ORIGINS=http://localhost:5173
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization
CORS_ALLOW_CREDENTIALS=false

# LLM 配置
OPENAI_API_KEY=sk-your-real-api-key
LLM_BASE_URL=https://api.openai.com/v1
LLM_MODEL_ID=gpt-4-turbo

# 日志
LOG_LEVEL=INFO

# 服务端口
GO_SERVER_PORT=8080
PYTHON_API_URL=http://localhost:8081
```

---

## 📁 项目结构

```
JunkFilter/
├── backend-go/              # Go 后端（RSS 抓取 + API）
│   ├── main.go             # 主程序、配置管理
│   ├── handlers/           # API 处理器
│   ├── repositories/       # 数据库层
│   ├── services/           # 业务逻辑
│   └── config.yaml         # YAML 配置文件
│
├── backend-python/          # Python 评估引擎
│   ├── api_server.py       # FastAPI 应用
│   ├── config.py           # Pydantic 配置
│   ├── main.py             # 启动脚本
│   ├── agents/             # AI Agent 实现
│   └── requirements.txt     # 依赖列表
│
├── frontend-vue/            # Vue 3 前端应用
│   ├── src/
│   │   ├── components/     # Vue 组件
│   │   │   ├── Config.vue       # 配置面板
│   │   │   └── TaskChat.vue     # 聊天界面
│   │   ├── stores/         # Pinia 状态管理
│   │   ├── composables/    # Vue Composables
│   │   │   └── useAPI.js   # API 适配层
│   │   └── App.vue         # 根组件
│   └── package.json        # 依赖配置
│
├── .env                     # 环境变量（本地）
├── .env.example             # 环境变量模板（安全）
├── docker-compose.yml       # Docker 容器编排
├── start-all.bat/sh         # 启动脚本
├── verify-day1.bat/sh       # 验证脚本
└── README.md                # 项目文档（本文件）
```

---

## 🐳 Docker 操作

### 启动容器
```bash
docker-compose up -d
```

### 查看容器状态
```bash
docker-compose ps
```

### 查看日志
```bash
# 所有容器
docker-compose logs -f

# 特定容器
docker-compose logs -f junkfilter-db
docker-compose logs -f junkfilter-redis
```

### 进入容器 Shell
```bash
# PostgreSQL
docker-compose exec junkfilter-db psql -U truesignal -d truesignal

# Redis
docker-compose exec junkfilter-redis redis-cli
```

### 停止容器
```bash
docker-compose down
```

### 完全重置（删除数据）
```bash
docker-compose down -v
docker-compose up -d
```

---

## 🔍 故障排查

### 问题 1: Docker 容器无法启动

**症状**：`docker-compose up` 失败

**解决方案**：
```bash
# 1. 检查端口是否被占用
netstat -an | grep 5432  # PostgreSQL
netstat -an | grep 6379  # Redis

# 2. 完全重置
docker-compose down -v
docker-compose up -d

# 3. 查看详细日志
docker-compose logs
```

### 问题 2: 前端无法连接后端

**症状**：浏览器控制台出现 CORS 错误

**解决方案**：
```bash
# 1. 检查 CORS 配置
cat .env | grep CORS_ALLOWED_ORIGINS

# 2. 确保在 http://localhost:5173 访问
# 3. Go 后端已启动
curl http://localhost:8080/health
```

### 问题 3: LLM API 调用失败

**症状**：Agent 回复出错，日志显示 API 调用失败

**解决方案**：
```bash
# 1. 检查 API 密钥是否正确
cat .env | grep OPENAI_API_KEY

# 2. 检查 Base URL 是否可访问
curl https://api.openai.com/v1/models  # OpenAI
curl https://elysiver.h-e.top/v1/models  # 中转站

# 3. 检查网络连接和防火墙
```

### 问题 4: 数据库连接错误

**症状**：后端日志显示 "Failed to connect to database"

**解决方案**：
```bash
# 1. 检查容器是否运行
docker-compose ps junkfilter-db

# 2. 验证数据库凭证
cat .env | grep DB_

# 3. 手动测试连接
docker-compose exec junkfilter-db psql -U truesignal -d truesignal -c "SELECT 1"
```

---

## 📞 获取帮助

### 查看日志
```bash
# 查看所有日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f backend-go
docker-compose logs -f backend-python
```

### 验证系统健康状况
```bash
# 前端应用
curl http://localhost:5173

# Go 后端
curl http://localhost:8080/health

# Python 后端
curl http://localhost:8081/health
```

### 报告问题

如遇到问题，请提供：
1. 错误日志（`docker-compose logs`）
2. .env 文件配置（脱敏后）
3. 系统信息（OS、Docker 版本）
4. 重现步骤

---

## 📚 进阶配置

### 生产环境部署

1. **更新 .env 文件**：
```bash
# 数据库
DB_HOST=your-db-host
DB_USER=production-user
DB_PASSWORD=strong-secure-password
DB_SSL_MODE=require  # ← 生产必须启用

# Redis
REDIS_PASSWORD=production-secure-password

# CORS
CORS_ALLOWED_ORIGINS=https://your-domain.com

# LLM
OPENAI_API_KEY=sk-prod-key
```

2. **启用 HTTPS**：
   在反向代理（Nginx/Traefik）中配置 SSL 证书

3. **数据备份**：
```bash
# PostgreSQL 备份
docker-compose exec junkfilter-db pg_dump -U truesignal truesignal > backup.sql

# 恢复
docker-compose exec -T junkfilter-db psql -U truesignal truesignal < backup.sql
```

### 性能优化

- 调整 Go 连接池：`DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS`
- 优化 Redis 缓存：增加 TTL，使用更多分片
- 启用数据库索引：现有索引已覆盖关键查询

---

## 📄 许可证

MIT License - 详见 LICENSE 文件

---

## 🙏 致谢

感谢所有开源项目的贡献：
- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [FastAPI](https://github.com/tiangolo/fastapi)
- [Vue 3](https://github.com/vuejs/core)
- [LangChain](https://github.com/langchain-ai/langchain)
- [PostgreSQL](https://www.postgresql.org/)
- [Redis](https://redis.io/)

---

**最后更新**: 2026-03-01
**版本**: 1.0.0 (Phase 5.3)
**状态**: ✅ 生产就绪
