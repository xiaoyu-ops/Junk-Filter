"""
Telegram Bot - 双向控制 Junk Filter 系统
支持快捷命令 + 自然语言（接入 ReAct Agent）
仅响应白名单 chat_id（私人 Bot）
"""

import asyncio
import json
import logging
import os

import asyncpg
from telegram import Update
from telegram.ext import Application, CommandHandler, ContextTypes, MessageHandler, filters

from agents.react_agent import run_react
from config import settings

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
)
logger = logging.getLogger(__name__)

GO_API_BASE = os.getenv("GO_API_BASE", "http://localhost:8080")


def _auth(func):
    """鉴权装饰器：只响应白名单 chat_id"""
    async def wrapper(update: Update, context: ContextTypes.DEFAULT_TYPE):
        allowed_id = str(context.bot_data.get("chat_id", ""))
        incoming_id = str(update.effective_chat.id)
        if incoming_id != allowed_id:
            logger.warning(f"[Bot] Unauthorized: {incoming_id}")
            return
        return await func(update, context)
    return wrapper


@_auth
async def cmd_start(update: Update, context: ContextTypes.DEFAULT_TYPE):
    await update.message.reply_text(
        "👋 *Junk Filter Bot*\n\n"
        "快捷命令：\n"
        "/status — 管道状态\n"
        "/fetch — 立即抓取所有 RSS 源\n"
        "/recent — 最近 5 篇高分文章\n"
        "/help — 显示此帮助\n\n"
        "直接发文字可用自然语言查询和控制系统。",
        parse_mode="Markdown",
    )


@_auth
async def cmd_help(update: Update, context: ContextTypes.DEFAULT_TYPE):
    await cmd_start(update, context)


@_auth
async def cmd_status(update: Update, context: ContextTypes.DEFAULT_TYPE):
    import aiohttp
    try:
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
    except Exception as e:
        text = f"❌ 获取状态失败：{e}"
    await update.message.reply_text(text, parse_mode="Markdown")


@_auth
async def cmd_fetch(update: Update, context: ContextTypes.DEFAULT_TYPE):
    import aiohttp
    pool: asyncpg.pool.Pool = context.bot_data["db_pool"]
    try:
        sources = await pool.fetch("SELECT id FROM sources WHERE enabled = true")
        await update.message.reply_text(f"⏳ 正在触发 {len(sources)} 个 RSS 源抓取...")
        async with aiohttp.ClientSession() as session:
            for row in sources:
                try:
                    await session.post(
                        f"{GO_API_BASE}/api/sources/{row['id']}/fetch",
                        timeout=aiohttp.ClientTimeout(total=5),
                    )
                except Exception:
                    pass
        await update.message.reply_text("✅ 抓取任务已触发，请稍后用 /status 查看结果。")
    except Exception as e:
        await update.message.reply_text(f"❌ 触发失败：{e}")


@_auth
async def cmd_recent(update: Update, context: ContextTypes.DEFAULT_TYPE):
    pool: asyncpg.pool.Pool = context.bot_data["db_pool"]
    try:
        rows = await pool.fetch("""
            SELECT c.title, c.original_url, e.innovation_score, e.depth_score, e.tldr
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
    except Exception as e:
        await update.message.reply_text(f"❌ 查询失败：{e}")


@_auth
async def handle_message(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """自然语言消息 → ReAct Agent"""
    pool: asyncpg.pool.Pool = context.bot_data["db_pool"]
    llm_config: dict = context.bot_data["llm_config"]
    user_text = update.message.text

    thinking_msg = await update.message.reply_text("🤔 思考中...")

    response_text = ""
    try:
        async for chunk in run_react(user_text, [], llm_config, db_pool=pool):
            if not chunk.startswith("data: "):
                continue
            try:
                data = json.loads(chunk[6:])
            except Exception:
                continue
            if data.get("type") == "chunk":
                response_text += data.get("content", "")
            elif data.get("type") in ("done", "error"):
                if data.get("type") == "error":
                    response_text = f"❌ {data.get('error', '未知错误')}"
                break
    except Exception as e:
        logger.error(f"[Bot] ReAct error: {e}")
        await thinking_msg.delete()
        await update.message.reply_text(f"❌ 出错了：{e}")
        return

    await thinking_msg.delete()

    if response_text.strip():
        # Telegram 单条消息上限 4096 字符
        for i in range(0, len(response_text), 4000):
            await update.message.reply_text(response_text[i:i + 4000])
    else:
        await update.message.reply_text("（未获得有效回复）")


async def _load_config(pool: asyncpg.pool.Pool) -> tuple[str, str]:
    """从 notification_settings 读取 bot_token 和 chat_id"""
    row = await pool.fetchrow(
        "SELECT push_channels FROM notification_settings WHERE id = 1"
    )
    if not row or not row["push_channels"]:
        return "", ""
    for ch in row["push_channels"]:
        if ch.get("type") == "telegram" and ch.get("enabled"):
            return ch.get("bot_token", ""), ch.get("chat_id", "")
    return "", ""


async def main():
    pool = await asyncpg.create_pool(
        f"postgresql://{settings.db_user}:{settings.db_password}"
        f"@{settings.db_host}:{settings.db_port}/{settings.db_name}"
    )

    bot_token, chat_id = await _load_config(pool)
    if not bot_token or not chat_id:
        logger.error(
            "[Bot] No Telegram config found. "
            "Add a Telegram channel in the Config page first, then restart."
        )
        await pool.close()
        return

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
