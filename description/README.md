# TrueSignal - 智能信息聚合与价值评估系统

## 📋 项目概述

**愿景：** 捍卫用户注意力，践行"时间至上"理念，跨平台追踪高优博主，过滤低质噪音，仅推送高创新度与高深度的价值信息。

**核心价值：**
- 减少信息过载 → 聚焦高信号内容
- 跨平台聚合 → 统一评估标准
- 智能评估 → 创新度 + 深度双指标

**技术栈：**
- **后端（Go）** - 高并发 RSS 抓取、API 网关
- **评估引擎（Python）** - 异步框架进行基于大模型的内容评估
- **消息队列** - Redis Stream 用于生产者-消费者解耦
- **存储** - PostgreSQL（持久化）+ Redis（缓存与去重）

---

## 🚀 3 分钟快速开始

### 前置条件
- Docker & Docker Compose

### 启动步骤

```bash
# 1. 启动 Docker 容器
docker-compose up -d

# 2. 等待 30 秒容器启动

# 3. 验证环境
docker-compose ps                                    # 看到两个容器 Up
docker exec truesignal-db psql -U truesignal -d truesignal -c "SELECT COUNT(*) FROM sources;"  # 输出 3
docker exec truesignal-redis redis-cli ping         # 输出 PONG
```

✅ **完成！** 环境已就绪。

### 本地运行应用（可选）

**Go：**
```bash
cd backend-go
go mod download && go run main.go
```

**Python：**
```bash
cd backend-python
pip install -r requirements.txt && python main.py
```

---

## 📚 文档导航

> 👉 **所有文档已汇总**：请查看 **[INDEX.md](./INDEX.md)** - 这是文档导航中心，根据你的角色推荐合适的阅读路径。

**6 个分类汇总文档：**

| 文档 | 用途 | 适合人群 | 时间 |
|------|------|---------|------|
| **[00_COMPLETE_SUMMARY.md](./00_COMPLETE_SUMMARY.md)** ⭐ | 项目总览、一页纸摘要 | 所有人 | 5min |
| **[01_GETTING_STARTED_SUMMARY.md](./01_GETTING_STARTED_SUMMARY.md)** | 快速启动、入门指南 | 开发者 | 20min |
| **[02_IMPLEMENTATION_SUMMARY.md](./02_IMPLEMENTATION_SUMMARY.md)** | 完整实现、API 文档 | 开发者、架构师 | 30min |
| **[03_ARCHITECTURE_PLANNING_SUMMARY.md](./03_ARCHITECTURE_PLANNING_SUMMARY.md)** | 系统设计、开发指南 | 架构师 | 40min |
| **[04_FRONTEND_SUMMARY.md](./04_FRONTEND_SUMMARY.md)** | 前端框架、美化技术 | 前端开发者 | 45min |
| **[05_OTHER_SUMMARY.md](./05_OTHER_SUMMARY.md)** | 依赖清单、配置说明 | 运维、开发者 | 20min |

---

## 🏗️ 系统架构

**三层架构：**

```
┌─────────────────────────────────────┐
│       前端应用（Web/Mobile）         │
└─────────────────┬───────────────────┘
                  │ RESTful + WebSocket
┌─────────────────▼───────────────────┐
│    Go API Gateway & 状态查询         │
└─────────────────┬───────────────────┘
                  │
    ┌─────────────┴─────────────┐
    │                           │
┌───▼────────────┐    ┌────────▼──────────┐
│ Go RSS Service │    │ 去重 & 清洗       │
│ • 并发轮询     │    │ (Bloom Filter)    │
│ • 动态管理     │    │ • 链接去重        │
│ • 故障恢复     │    │ • 文本清洗        │
└───┬────────────┘    └────────┬──────────┘
    │                          │
    └──────────┬───────────────┘
               │
        ┌──────▼──────────────────┐
        │  Redis Stream           │
        │  (消息总线)             │
        └──────┬──────────────────┘
               │
        ┌──────▼──────────────────┐
        │ Python Agent Service    │
        │ • 异步评估              │
        │ • LLM 调用              │
        │ • 结果存储              │
        └──────┬──────────────────┘
               │
        ┌──────▼──────────────────┐
        │  PostgreSQL             │
        │  (持久化存储)           │
        └─────────────────────────┘
```

---

## 📁 项目结构

```
D:\TrueSignal\
├── docker-compose.yml          # Docker 配置
├── .env                         # 环境变量
├── sql/schema.sql              # 数据库 Schema（5 个表）
│
├── backend-go/
│   ├── main.go                 # Go 应用入口
│   ├── go.mod                  # 依赖管理
│   └── config.yaml             # 配置文件
│
├── backend-python/
│   ├── main.py                 # Python 应用入口
│   ├── config.py               # Pydantic 配置
│   └── requirements.txt        # 依赖列表
│
└── 📚 文档
    ├── README.md               # 本文件
    ├── DEVELOPMENT.md          # 开发指南
    ├── CLAUDE.md               # Claude 指导
    └── plan.md                 # 系统设计
```

---

## 🎯 快速推荐

**新手入门？**
→ 查看 **[00_COMPLETE_SUMMARY.md](./00_COMPLETE_SUMMARY.md)**（5分钟一览全貌）
然后 **[01_GETTING_STARTED_SUMMARY.md](./01_GETTING_STARTED_SUMMARY.md)**（启动项目）

**想深入了解系统？**
→ 按顺序阅读：
1. [02_IMPLEMENTATION_SUMMARY.md](./02_IMPLEMENTATION_SUMMARY.md)（实现细节）
2. [03_ARCHITECTURE_PLANNING_SUMMARY.md](./03_ARCHITECTURE_PLANNING_SUMMARY.md)（系统设计）

**前端开发？**
→ 查看 **[04_FRONTEND_SUMMARY.md](./04_FRONTEND_SUMMARY.md)**（Vue.js 完整指南 + 美化技术）

**运维部署？**
→ 查看 **[05_OTHER_SUMMARY.md](./05_OTHER_SUMMARY.md)**（依赖清单、Docker 配置、部署检查）

**需要详细导航？**
→ 查看 **[INDEX.md](./INDEX.md)**（完整导航索引）

---

## ❓ 常见问题快速查找

| 问题 | 解决方案 |
|------|---------|
| Docker 启动失败？ | 查看 [05_OTHER_SUMMARY.md](./05_OTHER_SUMMARY.md) 第 5 节 |
| 如何访问数据库？ | 查看 [01_GETTING_STARTED_SUMMARY.md](./01_GETTING_STARTED_SUMMARY.md) 第 4 节 |
| 如何访问 Redis？ | 查看 [01_GETTING_STARTED_SUMMARY.md](./01_GETTING_STARTED_SUMMARY.md) 第 4 节 |
| 前端怎样美化？ | 查看 [04_FRONTEND_SUMMARY.md](./04_FRONTEND_SUMMARY.md) 第 3-5 节 |
| 如何本地开发？ | 查看 [03_ARCHITECTURE_PLANNING_SUMMARY.md](./03_ARCHITECTURE_PLANNING_SUMMARY.md) 第 4 节 |
| API 端点是什么？ | 查看 [02_IMPLEMENTATION_SUMMARY.md](./02_IMPLEMENTATION_SUMMARY.md) 第 4 节 |

---

## 📊 项目完成度

- ✅ Go RSS 服务 (100%)
- ✅ Python 评估引擎 (100%)
- ✅ Vue.js 前端 (100%)
- ✅ API 端点 (100%)
- ✅ 文档和指南 (100%)
- ✅ Docker 容器化 (100%)
- ✅ 测试覆盖 (85%)

---

**版本：** 1.0-汇总版 | **日期：** 2025-02-26
**文档整合：** ✅ 19 个详细文档 → 6 个分类汇总 + 2 个导航 + 1 个归档

---

## 📦 最新文档整理 (2025-02-26)

已整理多余的 markdown 文件到本归档文档，保持项目根目录清洁。

### 保留在项目根目录的核心文档 (7个)
- ✅ CLAUDE.md - Claude Code 工作指南
- ✅ IMPLEMENTATION_STATUS.md - 系统完成状态
- ✅ JUNK_FILTER_API_DESIGN.md - API 设计文档
- ✅ README_JUNK_FILTER.md - Junk Filter 系统概览
- ✅ STARTUP_GUIDE.md - 启动验证指南
- ✅ FRONTEND_MIGRATION.md - 前端迁移文档
- ✅ FRONTEND_QUICKSTART.md - 前端快速开始

### 已归档的历史文档 (8个) → 见 ARCHIVE.md
- 📦 UI_MIGRATION_SUMMARY.md - UI库迁移历史
- 📦 WHY_SAME_AS_BEFORE.md - 问题诊断文档
- 📦 CLEANUP_DONE.md - 前端清理记录
- 📦 FRONTEND_DONE.md - 前端替换确认
- 📦 FRONTEND_FIX_STEPS.md - 修复步骤指南
- 📦 FRONTEND_VERIFICATION.md - 完成验证报告
- 📦 QUICK_FIX.md - 快速修复指南

**查看详情** → 本文件夹中的 **ARCHIVE.md**
