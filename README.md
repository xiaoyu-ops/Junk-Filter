# TrueSignal - AI-Powered RSS Content Intelligence Platform

## 项目概述

**TrueSignal** 是一个智能信息聚合和价值评估系统，帮助用户从多个 RSS 源中筛选出高价值、高创新度、高深度的内容。通过 AI-powered 评估引擎，用户可以更智能地发现和管理内容。

## 项目架构

### 三个核心服务

```
┌─────────────────────────────────────────────────────────┐
│                     TrueSignal Frontend                  │
│              (Vue 3 + TypeScript + Tailwind)             │
│                  http://localhost:3000                   │
└────────────────────────┬────────────────────────────────┘
                         │
        ┌────────────────┼────────────────┐
        ▼                ▼                ▼
┌──────────────┐  ┌─────────────┐  ┌──────────────┐
│  Go Backend  │  │  Redis      │  │ PostgreSQL   │
│  RSS Fetcher │  │  Stream &   │  │  Content DB  │
│  API Gateway │  │  Cache      │  └──────────────┘
│ :8080        │  │  :6379      │
└──────┬───────┘  └─────────────┘
       │
┌──────▼──────────────────────────────┐
│     Python Evaluation Engine        │
│   (Async Framework + LLM API)       │
│   - Innovation Score (0-10)         │
│   - Depth Score (0-10)              │
│   - Pass/Reject Decision            │
└───────────────────────────────────┘
```

## 项目结构

```
TrueSignal/
├── backend-go/                 # Go RSS 抓取服务
│   ├── main.go
│   ├── config.yaml
│   ├── go.mod
│   └── go.sum
│
├── backend-python/             # Python 评估服务
│   ├── main.py
│   ├── config.py
│   ├── requirements.txt
│   └── Dockerfile
│
├── front/                      # Vue 3 前端
│   ├── src/
│   ├── package.json
│   ├── vite.config.ts
│   └── tailwind.config.js
│
├── docker-compose.yml          # Docker 容器编排
├── .env                        # 环境配置
├── .env.example                # 环境示例
├── CLAUDE.md                   # Claude Code 指导文件
└── README.md                   # 本文件
```

## 快速开始

### 前置要求

- Docker & Docker Compose
- Node.js 18+
- Go 1.20+
- Python 3.10+

### 选项 A: Docker Compose (推荐)

```bash
# 启动所有服务
docker-compose up -d

# 验证服务
docker-compose ps

# 查看日志
docker-compose logs -f
```

### 选项 B: 本地开发

#### 1. 启动基础设施 (PostgreSQL + Redis)

```bash
docker-compose up postgres redis -d
```

#### 2. 启动 Go 后端

```bash
cd backend-go
go mod download
go run main.go
# 访问: http://localhost:8080/health
```

#### 3. 启动 Python 评估服务

```bash
cd backend-python
python -m venv venv
source venv/bin/activate  # 或 venv\Scripts\activate (Windows)
pip install -r requirements.txt
python main.py
```

#### 4. 启动前端开发服务器

```bash
cd front
npm install
npm run dev
# 访问: http://localhost:5173
```

## API 端点

### RSS 源管理

| 方法 | 端点 | 功能 |
|------|------|------|
| GET | /api/sources | 获取所有源 |
| POST | /api/sources | 创建新源 |
| PUT | /api/sources/:id | 更新源 |
| DELETE | /api/sources/:id | 删除源 |

### 内容管理

| 方法 | 端点 | 功能 |
|------|------|------|
| GET | /api/content | 获取内容列表（支持分页） |
| GET | /api/content/:id | 获取单个内容 |
| GET | /api/content/source/:sourceId | 按源获取内容 |

### 评估结果

| 方法 | 端点 | 功能 |
|------|------|------|
| GET | /api/evaluations | 获取评估列表 |
| GET | /api/evaluations/content/:contentId | 获取内容评估 |
| GET | /api/evaluations/high-scores | 获取高分内容 |

## 前端功能

### 页面

1. **Home** (`/`)
   - 跨平台搜索（Blog、Twitter、Medium）
   - 快速标签建议
   - 品牌展示

2. **Timeline** (`/timeline`)
   - 双列网格内容展示
   - AI 评分展示（Innovation & Depth）
   - 过滤和分页
   - 博主详情侧滑抽屉

3. **Tasks** (`/tasks`)
   - 自然语言任务创建
   - 对话式 UI
   - 实时任务管理

4. **Config** (`/config`)
   - RSS 源管理（增删改查）
   - AI 参数可视化控制

### 特性

- ✅ 暗黑模式
- ✅ 响应式设计
- ✅ 完整 TypeScript 支持
- ✅ Pinia 状态管理
- ✅ Tailwind CSS 设计系统

## 三级去重机制

Go 服务使用三级去重保证数据质量：

1. **L1: Bloom Filter** (内存)
   - 7天时间窗口
   - <0.1% 误触发率
   - 快速拒绝

2. **L2: Redis Set** (分布式)
   - 精确校验
   - 7天 TTL
   - 原子操作

3. **L3: PostgreSQL** (持久化)
   - UNIQUE 约束
   - 最后防线
   - 捕获竞态条件

## 数据库 Schema

### 核心表

- **sources** - RSS 源配置
- **content** - 文章内容
- **evaluation** - AI 评估结果
- **user_subscription** - 用户订阅规则
- **status_log** - 状态转换审计日志

## 配置文件

### 环境变量 (`.env`)

```
# 数据库
DB_HOST=postgres
DB_PORT=5432
DB_USER=truesignal
DB_PASSWORD=truesignal123
DB_NAME=truesignal

# Redis
REDIS_URL=redis://redis:6379/0

# 日志
LOG_LEVEL=INFO

# 前端
VITE_API_BASE_URL=http://localhost:8080
VITE_APP_TITLE=TrueSignal
```

## Docker Compose 服务

```yaml
services:
  postgres:
    - 数据库: truesignal
    - 用户: truesignal
    - 端口: 5432

  redis:
    - 数据结构存储
    - 消息队列: Stream
    - 缓存层
    - 端口: 6379

  backend-go:
    - RSS 抓取和 API 网关
    - 端口: 8080

  backend-python:
    - 异步评估引擎
    - LLM 集成
```

## 开发工作流

### 数据库访问

```bash
# PostgreSQL
docker exec -it truesignal-db psql -U truesignal -d truesignal

# Redis CLI
docker exec -it truesignal-redis redis-cli
```

### 查看日志

```bash
# 查看所有服务日志
docker-compose logs -f

# 特定服务
docker-compose logs -f backend-go
docker-compose logs -f backend-python
```

### 重置数据

```bash
# 清空数据库（保留 schema）
docker exec truesignal-db psql -U truesignal -d truesignal \
  -c "TRUNCATE sources, content, evaluation, user_subscription, status_log CASCADE;"

# 清空 Redis
docker exec truesignal-redis redis-cli FLUSHDB
```

## 验证脚本

### Windows

```bash
# 验证前端
cd front
verify-frontend.bat
```

### Linux/Mac

```bash
# 验证前端
cd front
chmod +x verify-frontend.sh
./verify-frontend.sh
```

## 文档

- **[CLAUDE.md](./CLAUDE.md)** - Claude Code 项目指南
- **[前端实现总结](./front/IMPLEMENTATION_SUMMARY.md)** - 前端完整文档
- **[前端部署指南](./front/DEPLOYMENT_GUIDE.md)** - 部署和集成指南
- **[description/README.md](./description/README.md)** - 详细技术规范

## 部署

### 生产构建

```bash
# 构建前端
cd front
npm run build
# 输出: dist/ 目录

# 构建 Docker 镜像
docker build -t truesignal:latest .

# 运行容器
docker run -p 80:80 truesignal:latest
```

### 部署选项

- **Vercel** (推荐) - 零配置部署
- **GitHub Pages** - 静态托管
- **AWS** - EC2 + ECS
- **Heroku** - Platform as a Service
- **自管理服务器** - Nginx + Docker

详见 [部署指南](./front/DEPLOYMENT_GUIDE.md)

## 技术栈

### 后端

- **Go** - HTTP 服务器 + RSS 抓取
- **PostgreSQL** - 持久化存储
- **Redis** - 缓存和消息队列
- **Python** - 异步评估引擎

### 前端

- **Vue 3** - UI 框架
- **TypeScript** - 类型安全
- **Tailwind CSS** - 样式框架
- **Pinia** - 状态管理
- **Vite** - 构建工具

## 安全特性

- ✅ 三级去重防护
- ✅ 环境变量加密
- ✅ CORS 配置
- ✅ 输入验证
- ✅ SQL 注入防护
- ✅ HTTPS 支持

## 故障排查

### 常见问题

**Q: 无法连接到数据库**
```bash
# 检查容器状态
docker-compose ps

# 查看日志
docker-compose logs postgres
```

**Q: Redis 连接失败**
```bash
# 测试 Redis 连接
docker exec truesignal-redis redis-cli ping
```

**Q: 前端无法调用 API**
```bash
# 检查 API 基础 URL
echo $VITE_API_BASE_URL

# 检查 CORS 配置
curl -i http://localhost:8080/api/sources
```

## 性能指标

| 指标 | 值 |
|------|-----|
| 前端 JS | 65 kB (gzip) |
| 前端 CSS | 6.2 kB (gzip) |
| 构建时间 | 1.37s |
| 首页加载 | <1s |
| API 响应 | <100ms |

## 学习资源

- [Vue 3 文档](https://vuejs.org/)
- [Go HTTP 服务](https://golang.org/pkg/net/http/)
- [PostgreSQL 手册](https://www.postgresql.org/docs/)
- [Redis 命令](https://redis.io/commands/)
- [Tailwind CSS](https://tailwindcss.com/)

## 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. Push 到分支 (`git push origin feature/amazing-feature`)
5. 开启 Pull Request

## 许可证

本项目采用 MIT 许可证。详见 [LICENSE](./LICENSE)

## 联系方式

- Email: wuzhuoyang252@gmail.com
- GitHub: [TrueSignal](https://github.com/xiaoyu-ops/Junk-Filter)
- 讨论: [GitHub Discussions](https://github.com/xiaoyu-ops/Junk-Filter/discussions)

---

**项目状态**:  Production Ready
**最后更新**: 2026-02-26
**版本**: 1.0.0

**下一步**:
- 👉 [运行快速开始](#-快速开始)
- 👉 [查看前端文档](./front/IMPLEMENTATION_SUMMARY.md)
- 👉 [部署到生产环境](./front/DEPLOYMENT_GUIDE.md)
