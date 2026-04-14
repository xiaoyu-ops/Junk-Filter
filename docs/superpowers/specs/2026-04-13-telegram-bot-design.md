# Telegram Bot 设计文档

**日期**：2026-04-13  
**状态**：已批准

---

## 背景

当前项目使用 Bark（iOS）推送高分文章通知。用户希望：
1. 将推送渠道从 Bark 改为 Telegram
2. 通过 Telegram Bot 双向控制项目（查询文章、触发抓取、管理 RSS 源、更新偏好等）

---

## 目标

- 高分文章评估完成后自动推送到 Telegram
- 支持自然语言对话（接入现有 ReAct Agent）
- 支持固定快捷命令（`/status`、`/fetch` 等）
- 仅响应白名单 `chat_id`（私人 Bot，安全隔离）

---

## 架构

### 新增文件

```
backend-python/
├── telegram_bot.py        # 新增：Bot 主入口，独立进程
└── services/
    └── push_service.py    # 改动：新增 _push_telegram()，保留 Bark 结构
```

### 运行方式

独立进程，VS Code 新增 `🤖 Telegram Bot` 任务。使用 `python-telegram-bot` 库长轮询模式，无需公网域名。

### 数据流

```
Telegram 用户发消息
  → telegram_bot.py 接收
  → 命令路由：/xxx → 直接执行工具 | 自然语言 → run_react()
  → 返回结果给 Telegram

高分文章评估完成
  → stream_consumer.py 调用 push_to_channels()
  → push_service._push_telegram()
  → 发送格式化消息到 Telegram
```

---

## 功能设计

### 快捷命令

| 命令 | 说明 |
|------|------|
| `/start` | 欢迎消息，列出可用命令 |
| `/status` | 管道状态（各状态文章数） |
| `/fetch` | 触发所有 RSS 源立即抓取 |
| `/recent` | 最近 5 篇高分文章（INTERESTING） |
| `/help` | 命令列表 |

其余任意文本 → 转发给 ReAct Agent 处理。

### 推送消息格式

```
🔥 [JunkFilter] 高分文章

📰 标题：<title>
⭐ 创新 <innovation>/10 | 深度 <depth>/10
🏷️ <decision>

💡 TL;DR：<tldr>

🧠 评估理由：<reasoning>

🔗 <url>
```

### 安全鉴权

- 从 DB `ai_config` 表或环境变量读取 `TELEGRAM_BOT_TOKEN` 和 `TELEGRAM_CHAT_ID`
- 所有入站消息检查 `message.chat.id == TELEGRAM_CHAT_ID`，不匹配直接忽略

---

## 配置

在 `notification_settings.push_channels` 中新增 telegram 类型：

```json
{
  "type": "telegram",
  "bot_token": "xxx:yyy",
  "chat_id": "123456789",
  "enabled": true
}
```

`push_service.py` 读取该配置，`telegram_bot.py` 启动时也从同一来源读取 token 和 chat_id。

---

## 前端改动（Config 页面）

- 将渠道类型从 Bark 替换为 Telegram
- Telegram 渠道显示两个输入框：Bot Token、Chat ID
- 删除所有 Bark 相关代码和 UI

---

## 依赖

```
python-telegram-bot>=21.0
```

加入 `backend-python/requirements.txt`。

---

## VS Code 任务

```json
{
  "label": "🤖 Telegram Bot",
  "type": "shell",
  "command": "cd backend-python && /opt/homebrew/Caskroom/miniconda/base/envs/junkfilter/bin/python telegram_bot.py",
  "dependsOn": ["🐘 DB + Redis (Docker)"],
  "presentation": {
    "label": "Telegram Bot",
    "group": "junkfilter",
    "panel": "dedicated",
    "reveal": "always",
    "close": false
  }
}
```

---

## 不在本次范围内

- Webhook 模式（长轮询已够用）
- 多用户支持
- Telegram 文件/图片发送
- 内联键盘按钮（后续可加）
