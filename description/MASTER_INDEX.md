# Junk Filter 文档导航中心

**最后更新**: 2026-03-07
**项目进度**: Phase 5.3 完成，系统完整可用

---

## 快速导航

| 需求 | 文档 |
|------|------|
| 启动系统 | [guides/QUICK_START_WINDOWS.md](guides/QUICK_START_WINDOWS.md) |
| 配置 LLM / 中转站 | [guides/LLM_CONFIG_GUIDE.md](guides/LLM_CONFIG_GUIDE.md) |
| RSS 抓取功能分析 | [RSS_FETCHING_CAPABILITY.md](RSS_FETCHING_CAPABILITY.md) |

---

## 文档结构

```
description/
├── MASTER_INDEX.md              # 本文件：导航中心
├── RSS_FETCHING_CAPABILITY.md   # RSS 功能技术文档
└── guides/                      # 操作指南
    ├── LLM_CONFIG_GUIDE.md      # LLM / 中转站配置
    └── QUICK_START_WINDOWS.md   # Windows 完整启动指南
```

---

## 启动系统（快速参考）

```bash
# Windows
start-all.bat

# Linux/Mac
./start-all.sh
```

服务地址：
- 前端: http://localhost:5173
- Go API: http://localhost:8080/health
- Python API: http://localhost:8083/health

---

## 系统架构简要

```
RSS 源 → Go 后端抓取 → 三层去重 → Redis Stream → Python 评估(LLM) → PostgreSQL → 前端展示
```

**核心组件**：
- **Go 后端** (端口 8080) - RSS 抓取、去重、API 网关
- **Python 后端** (端口 8083) - LLM 内容评估、Agent 聊天
- **Redis** - 消息队列 (Stream) + 去重缓存
- **PostgreSQL** - 持久化存储
- **Vue 3 前端** (端口 5173) - 任务管理 UI
