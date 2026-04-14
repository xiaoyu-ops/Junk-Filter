# Junk Filter 文档导航中心

**最后更新**: 2026-04-14
**项目进度**: Phase 8-11 完成 + ReAct Agent + Telegram Bot 双向控制，系统完整可用，Mac 开发环境

---

## 快速导航

| 需求 | 文档 |
|------|------|
| **代码结构解读（入门必读）** | [guides/CODE_READING_GUIDE.md](guides/CODE_READING_GUIDE.md) |
| Mac 本地开发（当前）| [guides/VSCODE_DEV_GUIDE.md](guides/VSCODE_DEV_GUIDE.md) |
| 配置 LLM / 中转站 | [guides/LLM_CONFIG_GUIDE.md](guides/LLM_CONFIG_GUIDE.md) |
| RSS 抓取功能分析 | [RSS_FETCHING_CAPABILITY.md](RSS_FETCHING_CAPABILITY.md) |
| Windows 启动参考 | [guides/QUICK_START_WINDOWS.md](guides/QUICK_START_WINDOWS.md) |

---

## 文档结构

```
description/
├── MASTER_INDEX.md              # 本文件：导航中心
├── RSS_FETCHING_CAPABILITY.md   # RSS 功能技术文档
├── plan.md                      # Phase 8-11 功能演进计划
├── project_description.md       # 项目架构与技术细节（含 Telegram Bot）
└── guides/                      # 操作指南
    ├── CODE_READING_GUIDE.md    # 代码结构解读（解剖级）← 新增
    ├── VSCODE_DEV_GUIDE.md      # Mac 本地开发（VS Code Tasks）← 当前主用
    ├── LLM_CONFIG_GUIDE.md      # LLM / 中转站配置
    └── QUICK_START_WINDOWS.md   # Windows 完整启动指南（历史参考）
```

---

## 启动系统（快速参考）

**Mac（VS Code Tasks，当前推荐）**：
```
Cmd + Shift + B  →  🚀 Start All Services
```

**Mac（命令行）**：
```bash
./start-all.sh
```

**Windows**：
```bash
start-all.bat
```

服务地址：
- 前端: http://localhost:5173
- Go API: http://localhost:8080/health
- Python API: http://localhost:8083/health

---

## 系统架构简要

```
RSS 源 → Go 后端抓取 → 三层去重 → Redis Stream → Python 评估(LLM) → PostgreSQL → 前端展示
                                                                           ↓
                                                              Telegram 推送 + Bot 双向控制
```

**核心组件**：
- **Go 后端** (端口 8080) - RSS 抓取、去重、API 网关
- **Python 后端** (端口 8083) - LLM 内容评估、ReAct Agent 聊天
- **Telegram Bot** - 高分推送 + 命令控制 + 自然语言 → ReAct Agent
- **Redis** - 消息队列 (Stream) + 去重缓存
- **PostgreSQL** - 持久化存储
- **Vue 3 前端** (端口 5173) - Timeline、Agent 对话、Config 页面
