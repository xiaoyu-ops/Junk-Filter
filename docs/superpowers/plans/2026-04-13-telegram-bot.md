# Telegram Bot Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将推送渠道从 Bark 替换为 Telegram，并新增独立 Telegram Bot 服务支持双向控制（自然语言 + 快捷命令）。

**Architecture:** 新增 `backend-python/telegram_bot.py` 作为独立进程，使用 `python-telegram-bot` 长轮询。`push_service.py` 删除 Bark，改为 Telegram。前端 Config 页替换 Bark UI 为 Telegram 配置（Bot Token + Chat ID）。

**Tech Stack:** python-telegram-bot>=21.0, aiohttp, asyncpg, 现有 ReAct Agent (run_react)

---

## 文件变更清单

| 文件 | 操作 |
|------|------|
| `backend-python/requirements.txt` | 新增 `python-telegram-bot>=21.0` |
| `backend-python/services/push_service.py` | 删除 Bark，新增 `_push_telegram()` |
| `backend-python/telegram_bot.py` | 新建：Bot 主入口 |
| `frontend-vue/src/components/Config.vue` | 替换 Bark UI 为 Telegram 配置 |
| `.vscode/tasks.json` | 新增 Telegram Bot 任务 |

---

## Task 1：安装依赖

**Files:**
- Modify: `backend-python/requirements.txt`

- [ ] **Step 1: 在 requirements.txt 末尾新增依赖**

```
# Telegram Bot
python-telegram-bot>=21.0
```

- [ ] **Step 2: 安装**

```bash
/opt/homebrew/Caskroom/miniconda/base/envs/junkfilter/bin/pip install "python-telegram-bot>=21.0"
```

预期输出：`Successfully installed python-telegram-bot-21.x.x`

- [ ] **Step 3: 验证安装**

```bash
/opt/homebrew/Caskroom/miniconda/base/envs/junkfilter/bin/python -c "import telegram; print(telegram.__version__)"
```

预期输出：`21.x.x`

- [ ] **Step 4: Commit**

```bash
git add backend-python/requirements.txt
git commit -m "deps: add python-telegram-bot"
```

---

## Task 2：改造 push_service.py（Bark → Telegram）

**Files:**
- Modify: `backend-python/services/push_service.py`

- [ ] **Step 1: 完整替换 push_service.py**

用以下内容完整覆盖该文件：

```python
"""
Push Service - Send notifications via Telegram
"""

import logging
import aiohttp

logger = logging.getLogger(__name__)


async def push_to_channels(
    db_pool,
    title: str,
    summary: str,
    innovation_score: int,
    depth_score: int,
    decision: str,
    url: str = "",
):
    """Read push channels from DB and send notification to each enabled Telegram channel"""
    try:
        row = await db_pool.fetchrow(
            "SELECT push_channels, enabled FROM notification_settings WHERE id = 1"
        )
        if not row or not row["enabled"]:
            return

        channels = row["push_channels"]
        if not channels or not isinstance(channels, list):
            return

        for channel in channels:
            if not channel.get("enabled", False):
                continue

            channel_type = channel.get("type", "")
            try:
                if channel_type == "telegram":
                    await _push_telegram(
                        channel, title, summary, innovation_score, depth_score, decision, url
                    )
                else:
                    logger.warning(f"[Push] Unknown channel type: {channel_type}")
            except Exception as e:
                logger.error(f"[Push] Failed to push via {channel_type}: {e}")

    except Exception as e:
        logger.error(f"[Push] Error reading push channels: {e}")


async def _push_telegram(
    channel: dict,
    title: str,
    summary: str,
    innovation_score: int,
    depth_score: int,
    decision: str,
    url: str = "",
):
    """Push notification via Telegram Bot API"""
    bot_token = channel.get("bot_token", "").strip()
    chat_id = channel.get("chat_id", "").strip()
    if not bot_token or not chat_id:
        logger.warning("[Push] Telegram channel missing bot_token or chat_id")
        return

    decision_emoji = {"INTERESTING": "🔥", "BOOKMARK": "📌", "SKIP": "⏭️"}.get(decision, "📄")

    text = (
        f"{decision_emoji} *[JunkFilter] 高分文章*\n\n"
        f"📰 *标题：* {_escape(title)}\n"
        f"⭐ 创新 {innovation_score}/10 \\| 深度 {depth_score}/10\n"
        f"🏷️ {_escape(decision)}\n\n"
        f"💡 *TL;DR：* {_escape(summary)}\n\n"
    )
    if url:
        text += f"🔗 {url}"

    api_url = f"https://api.telegram.org/bot{bot_token}/sendMessage"
    async with aiohttp.ClientSession() as session:
        resp = await session.post(
            api_url,
            json={"chat_id": chat_id, "text": text, "parse_mode": "MarkdownV2"},
            timeout=aiohttp.ClientTimeout(total=10),
        )
        data = await resp.json()
        if data.get("ok"):
            logger.info(f"[Push] Telegram sent: {title[:40]}")
        else:
            logger.warning(f"[Push] Telegram error: {data.get('description')}")


def _escape(text: str) -> str:
    """Escape special characters for Telegram MarkdownV2"""
    special = r"_*[]()~`>#+-=|{}.!"
    return "".join(f"\\{c}" if c in special else c for c in str(text))


async def test_push_channel(channel: dict) -> dict:
    """Test a single push channel with a sample message"""
    channel_type = channel.get("type", "")
    title = "Junk Filter 测试通知"
    summary = "这是一条测试消息，如果你收到了说明推送配置成功。"

    try:
        if channel_type == "telegram":
            await _push_telegram(channel, title, summary, 8, 7, "INTERESTING", "")
        else:
            return {"success": False, "message": f"未知渠道类型: {channel_type}"}
        return {"success": True, "message": "测试消息已发送，请检查 Telegram"}
    except Exception as e:
        return {"success": False, "message": str(e)}
```

- [ ] **Step 2: 手动测试推送（需要真实 Bot Token 和 Chat ID）**

```bash
cd /Users/wuzhuoyang/code/Junk-Filter/backend-python
/opt/homebrew/Caskroom/miniconda/base/envs/junkfilter/bin/python - << 'EOF'
import asyncio
from services.push_service import _push_telegram

channel = {
    "bot_token": "YOUR_BOT_TOKEN",
    "chat_id": "YOUR_CHAT_ID",
    "enabled": True,
    "type": "telegram",
}
asyncio.run(_push_telegram(channel, "测试标题", "测试摘要内容", 8, 7, "INTERESTING", "https://example.com"))
EOF
```

预期：Telegram 收到格式化消息。

- [ ] **Step 3: Commit**

```bash
git add backend-python/services/push_service.py
git commit -m "feat: replace Bark push with Telegram"
```

---

## Task 3：新建 telegram_bot.py

**Files:**
- Create: `backend-python/telegram_bot.py`

- [ ] **Step 1: 创建文件**

```python
"""
Telegram Bot - 双向控制 Junk Filter 系统
支持快捷命令 + 自然语言（接入 ReAct Agent）
"""

import asyncio
import logging
import os

import asyncpg
from telegram import Update
from telegram.ext import Application, CommandHandler, ContextTypes, MessageHandler, filters

from agents.react_agent import run_react
from config import settings

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

GO_API_BASE = os.getenv("GO_API_BASE", "http://localhost:8080")


def get_telegram_config(channels: list) -> tuple[str, str]:
    """从 push_channels 中提取 bot_token 和 chat_id"""
    for ch in channels:
        if ch.get("type") == "telegram" and ch.get("enabled"):
            return ch.get("bot_token", ""), ch.get("chat_id", "")
    return "", ""


async def load_config(pool: asyncpg.pool.Pool) -> tuple[str, str]:
    """从 DB 加载 Telegram bot_token 和 chat_id"""
    row = await pool.fetchrow(
        "SELECT push_channels FROM notification_settings WHERE id = 1"
    )
    if not row or not row["push_channels"]:
        return "", ""
    return get_telegram_config(row["push_channels"])


def auth(func):
    """鉴权装饰器：只响应白名单 chat_id"""
    async def wrapper(update: Update, context: ContextTypes.DEFAULT_TYPE):
        allowed_id = context.bot_data.get("chat_id", "")
        if str(update.effective_chat.id) != str(allowed_id):
            logger.warning(f"[Bot] Unauthorized access from {update.effective_chat.id}")
            return
        return await func(update, context)
    return wrapper


@auth
async def cmd_start(update: Update, context: ContextTypes.DEFAULT_TYPE):
    await update.message.reply_text(
        "👋 Junk Filter Bot\n\n"
        "快捷命令：\n"
        "/status — 管道状态\n"
        "/fetch — 立即抓取所有 RSS 源\n"
        "/recent — 最近 5 篇高分文章\n"
        "/help — 显示此帮助\n\n"
        "直接发送文字可以用自然语言查询和控制系统。"
    )


@auth
async def cmd_help(update: Update, context: ContextTypes.DEFAULT_TYPE):
    await cmd_start(update, context)


@auth
async def cmd_status(update: Update, context: ContextTypes.DEFAULT_TYPE):
    import aiohttp
    async with aiohttp.ClientSession() as session:
        async with session.get(f"{GO_API_BASE}/api/content/stats") as resp:
            data = await resp.json()
    text = (
        f"📊 *管道状态*\n\n"
        f"待评估：{data.get('pending', 0)}\n"
        f"评估中：{data.get('processing', 0)}\n"
        f"已评估：{data.get('evaluated', 0)}\n"
        f"已丢弃：{data.get('discarded', 0)}\n"
        f"总计：{data.get('total', 0)}"
    )
    await update.message.reply_text(text, parse_mode="Markdown")


@auth
async def cmd_fetch(update: Update, context: ContextTypes.DEFAULT_TYPE):
    import aiohttp
    pool: asyncpg.pool.Pool = context.bot_data["db_pool"]
    sources = await pool.fetch("SELECT id FROM sources WHERE enabled = true")
    await update.message.reply_text(f"⏳ 正在触发 {len(sources)} 个 RSS 源抓取...")
    async with aiohttp.ClientSession() as session:
        for row in sources:
            try:
                await session.post(f"{GO_API_BASE}/api/sources/{row['id']}/fetch")
            except Exception:
                pass
    await update.message.reply_text("✅ 抓取任务已触发，请稍后查看结果。")


@auth
async def cmd_recent(update: Update, context: ContextTypes.DEFAULT_TYPE):
    pool: asyncpg.pool.Pool = context.bot_data["db_pool"]
    rows = await pool.fetch("""
        SELECT c.title, c.original_url, e.innovation_score, e.depth_score, e.tldr, e.decision
        FROM content c
        JOIN evaluation e ON e.content_id = c.id
        WHERE e.decision = 'INTERESTING'
        ORDER BY e.id DESC LIMIT 5
    """)
    if not rows:
        await update.message.reply_text("暂无高分文章。")
        return
    for r in rows:
        text = (
            f"🔥 *{r['title']}*\n"
            f"⭐ 创新 {r['innovation_score']}/10 | 深度 {r['depth_score']}/10\n"
            f"💡 {r['tldr']}\n"
            f"🔗 {r['original_url']}"
        )
        await update.message.reply_text(text, parse_mode="Markdown")


@auth
async def handle_message(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """自然语言消息 → ReAct Agent"""
    pool: asyncpg.pool.Pool = context.bot_data["db_pool"]
    llm_config: dict = context.bot_data["llm_config"]
    user_text = update.message.text

    await update.message.reply_text("🤔 思考中...")

    response_text = ""
    try:
        async for chunk in run_react(user_text, [], llm_config, db_pool=pool):
            # 收集 chunk 类型的文字片段
            if '"type": "chunk"' in chunk:
                import json
                data = json.loads(chunk.replace("data: ", ""))
                if data.get("type") == "chunk":
                    response_text += data.get("content", "")
            elif '"type": "done"' in chunk:
                break
    except Exception as e:
        logger.error(f"[Bot] ReAct error: {e}")
        await update.message.reply_text(f"❌ 出错了：{e}")
        return

    if response_text.strip():
        # Telegram 消息上限 4096 字符
        for i in range(0, len(response_text), 4000):
            await update.message.reply_text(response_text[i:i+4000])
    else:
        await update.message.reply_text("（未获得有效回复）")


async def main():
    pool = await asyncpg.create_pool(
        f"postgresql://{settings.db_user}:{settings.db_password}"
        f"@{settings.db_host}:{settings.db_port}/{settings.db_name}"
    )

    bot_token, chat_id = await load_config(pool)
    if not bot_token or not chat_id:
        logger.error("[Bot] No Telegram config found in DB. Add a telegram channel in Config page first.")
        return

    # 加载 LLM 配置供 ReAct Agent 使用
    from config import load_llm_config_from_db
    llm_config = await load_llm_config_from_db(pool) or {}

    app = Application.builder().token(bot_token).build()
    app.bot_data["db_pool"] = pool
    app.bot_data["chat_id"] = chat_id
    app.bot_data["llm_config"] = llm_config

    app.add_handler(CommandHandler("start", cmd_start))
    app.add_handler(CommandHandler("help", cmd_help))
    app.add_handler(CommandHandler("status", cmd_status))
    app.add_handler(CommandHandler("fetch", cmd_fetch))
    app.add_handler(CommandHandler("recent", cmd_recent))
    app.add_handler(MessageHandler(filters.TEXT & ~filters.COMMAND, handle_message))

    logger.info(f"[Bot] Starting, authorized chat_id={chat_id}")
    await app.run_polling(allowed_updates=Update.ALL_TYPES)


if __name__ == "__main__":
    asyncio.run(main())
```

- [ ] **Step 2: 验证语法无误**

```bash
cd /Users/wuzhuoyang/code/Junk-Filter/backend-python
/opt/homebrew/Caskroom/miniconda/base/envs/junkfilter/bin/python -m py_compile telegram_bot.py && echo "OK"
```

预期输出：`OK`

- [ ] **Step 3: Commit**

```bash
git add backend-python/telegram_bot.py
git commit -m "feat: add Telegram Bot service"
```

---

## Task 4：更新前端 Config.vue（Bark → Telegram）

**Files:**
- Modify: `frontend-vue/src/components/Config.vue`

- [ ] **Step 1: 替换渠道标签和字段（模板部分）**

找到以下代码块（约第 463-507 行）：

```html
<!-- 渠道类型 -->
<span class="text-sm font-medium text-[#111827] dark:text-white">Bark (iOS)</span>
<input type="hidden" v-model="ch.type" />
```

替换为：

```html
<!-- 渠道类型 -->
<span class="text-sm font-medium text-[#111827] dark:text-white">Telegram</span>
<input type="hidden" v-model="ch.type" />
```

- [ ] **Step 2: 替换 Bark 配置字段为 Telegram 字段**

找到以下代码块（约第 501-507 行）：

```html
<!-- Bark: server_url -->
<div>
  <label class="block text-xs text-[#6B7280] mb-1">Bark 地址（含 Key）</label>
  <input v-model="ch.server_url" type="text" placeholder="https://api.day.app/YOUR_KEY"
    class="w-full bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 text-[#111827] dark:text-white rounded-lg py-1.5 px-3 text-sm focus:outline-none focus:ring-2 focus:ring-gray-200 dark:focus:ring-gray-700" />
  <p class="text-xs text-[#6B7280] mt-1">打开 Bark App 复制推送地址粘贴到这里</p>
</div>
```

替换为：

```html
<!-- Telegram: bot_token + chat_id -->
<div class="space-y-2">
  <div>
    <label class="block text-xs text-[#6B7280] mb-1">Bot Token</label>
    <input v-model="ch.bot_token" type="text" placeholder="1234567890:ABCDEFGxxxxxxxx"
      class="w-full bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 text-[#111827] dark:text-white rounded-lg py-1.5 px-3 text-sm focus:outline-none focus:ring-2 focus:ring-gray-200 dark:focus:ring-gray-700" />
    <p class="text-xs text-[#6B7280] mt-1">从 @BotFather 获取</p>
  </div>
  <div>
    <label class="block text-xs text-[#6B7280] mb-1">Chat ID</label>
    <input v-model="ch.chat_id" type="text" placeholder="123456789"
      class="w-full bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 text-[#111827] dark:text-white rounded-lg py-1.5 px-3 text-sm focus:outline-none focus:ring-2 focus:ring-gray-200 dark:focus:ring-gray-700" />
    <p class="text-xs text-[#6B7280] mt-1">给 @userinfobot 发消息获取你的 Chat ID</p>
  </div>
</div>
```

- [ ] **Step 3: 替换 JS 中 addPushChannel 的初始对象**

找到（约第 1053-1057 行）：

```js
const addPushChannel = () => {
  if (!notifSettings.value.push_channels) {
    notifSettings.value.push_channels = []
  }
  notifSettings.value.push_channels.push({ type: 'bark', server_url: '', enabled: true })
}
```

替换为：

```js
const addPushChannel = () => {
  if (!notifSettings.value.push_channels) {
    notifSettings.value.push_channels = []
  }
  notifSettings.value.push_channels.push({ type: 'telegram', bot_token: '', chat_id: '', enabled: true })
}
```

- [ ] **Step 4: 验证前端能正常编译**

```bash
cd /Users/wuzhuoyang/code/Junk-Filter/frontend-vue
npm run build 2>&1 | tail -10
```

预期：无错误，输出 `built in Xs`

- [ ] **Step 5: Commit**

```bash
git add frontend-vue/src/components/Config.vue
git commit -m "feat: replace Bark UI with Telegram in Config page"
```

---

## Task 5：更新 VS Code tasks.json

**Files:**
- Modify: `.vscode/tasks.json`

- [ ] **Step 1: 在 `🌐 Frontend :5173` 任务之前新增 Telegram Bot 任务**

在 tasks 数组中，`🌐 Frontend :5173` 任务之前插入：

```json
// ─────────────────────────────────────────────
// Telegram Bot（依赖 DB 就绪）
// ─────────────────────────────────────────────
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
  },
  "problemMatcher": []
},
```

- [ ] **Step 2: 将 `🤖 Telegram Bot` 加入 `🚀 Start All Services` 的 dependsOn**

```json
"dependsOn": [
  "🐹 Go Backend :8080",
  "🐍 Python API :8083",
  "🐍 Python Consumer",
  "🌐 Frontend :5173",
  "🤖 Telegram Bot"
],
```

- [ ] **Step 3: Commit**

```bash
git add .vscode/tasks.json
git commit -m "chore: add Telegram Bot VS Code task"
```

---

## Task 6：端到端验证

- [ ] **Step 1: 在前端 Config 页面配置 Telegram 渠道**

  1. 打开 http://localhost:5173 → Config → 通知设置
  2. 点击"添加渠道"，填入 Bot Token 和 Chat ID
  3. 点击"测试"，Telegram 应收到测试消息
  4. 保存设置

- [ ] **Step 2: 启动 Telegram Bot**

  VS Code 运行 `🤖 Telegram Bot` 任务，日志应显示：
  ```
  [Bot] Starting, authorized chat_id=YOUR_CHAT_ID
  ```

- [ ] **Step 3: 测试快捷命令**

  在 Telegram 发送：
  - `/start` → 收到欢迎消息和命令列表
  - `/status` → 收到管道状态数字
  - `/recent` → 收到最近高分文章

- [ ] **Step 4: 测试自然语言**

  发送"最近有哪些关于 AI 的文章？"，Bot 应调用 ReAct Agent 并返回结果。

- [ ] **Step 5: 触发真实推送**

  手动抓取一个 RSS 源：
  ```bash
  curl -X POST http://localhost:8080/api/sources/1/fetch
  ```
  等待高分文章评估完成，Telegram 应自动收到推送通知。
