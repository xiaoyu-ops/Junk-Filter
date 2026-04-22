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

_DB_DSN = (
    f"postgresql://{settings.db_user}:{settings.db_password}"
    f"@{settings.db_host}:{settings.db_port}/{settings.db_name}"
)

_last_message_time: dict[str, float] = {}
RATE_LIMIT_SECONDS = 1.0


def _auth(func):
    """鉴权装饰器：白名单校验 + 限流（每用户每秒最多 1 条）"""
    async def wrapper(update: Update, context: ContextTypes.DEFAULT_TYPE):
        allowed_ids: set = context.bot_data.get("allowed_ids", set())
        incoming_id = str(update.effective_chat.id)
        if incoming_id not in allowed_ids:
            logger.warning(f"[Bot] Unauthorized: {incoming_id}")
            return

        now = asyncio.get_running_loop().time()
        last = _last_message_time.get(incoming_id, 0)
        if now - last < RATE_LIMIT_SECONDS:
            logger.debug(f"[Bot] Rate limited: {incoming_id}")
            await update.message.reply_text("⏳ 请稍等 1 秒再发")
            return
        _last_message_time[incoming_id] = now

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
        async with pool.acquire() as conn:
            sources = await conn.fetch("SELECT id FROM sources WHERE enabled = true")

        await update.message.reply_text(f"⏳ 正在触发 {len(sources)} 个 RSS 源抓取...")

        async def _fire(session, source_id):
            try:
                await session.post(
                    f"{GO_API_BASE}/api/sources/{source_id}/fetch",
                    timeout=aiohttp.ClientTimeout(total=5),
                )
            except Exception:
                pass

        async with aiohttp.ClientSession() as session:
            await asyncio.gather(*[_fire(session, row['id']) for row in sources])

        await update.message.reply_text("✅ 抓取任务已触发，请稍后用 /status 查看结果。")
    except Exception as e:
        logger.error(f"[Bot] /fetch error: {e}", exc_info=True)
        await update.message.reply_text(f"❌ 触发失败：{e}")


@_auth
async def cmd_recent(update: Update, context: ContextTypes.DEFAULT_TYPE):
    pool: asyncpg.pool.Pool = context.bot_data["db_pool"]
    try:
        async with pool.acquire() as conn:
            rows = await conn.fetch("""
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
                f"💡 摘要：{r['tldr']}\n"
                f"🔗 {r['original_url']}"
            )
            await update.message.reply_text(text, parse_mode="Markdown")
    except Exception as e:
        logger.error(f"[Bot] /recent error: {e}", exc_info=True)
        await update.message.reply_text(f"❌ 查询失败：{e}")


async def _reload_llm_config_if_stale(context: ContextTypes.DEFAULT_TYPE):
    """每 60 秒从 DB 热加载一次 LLM 配置，无需重启 Bot"""
    import time
    now = time.monotonic()
    last = context.bot_data.get("llm_config_last_reload", 0)
    if now - last < 60:
        return
    context.bot_data["llm_config_last_reload"] = now
    pool: asyncpg.pool.Pool = context.bot_data["db_pool"]
    try:
        row = await pool.fetchrow(
            "SELECT default_model as model_name, api_key, base_url, temperature, max_tokens "
            "FROM ai_config LIMIT 1"
        )
        if row and row["api_key"] and row["api_key"] != "sk-placeholder":
            context.bot_data["llm_config"] = {
                "model_name": row["model_name"],
                "api_key": row["api_key"],
                "base_url": row["base_url"],
                "temperature": float(row["temperature"] or 0.7),
                "max_tokens": int(row["max_tokens"] or 2000),
            }
            logger.debug("[Bot] LLM config hot-reloaded")
    except Exception as e:
        logger.warning(f"[Bot] Failed to reload LLM config: {e}")


async def _process_message(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """实际处理自然语言消息（在队列 worker 中执行）"""
    await _reload_llm_config_if_stale(context)
    pool: asyncpg.pool.Pool = context.bot_data["db_pool"]
    llm_config: dict = context.bot_data["llm_config"]
    histories: dict = context.bot_data["histories"]
    chat_id = str(update.effective_chat.id)
    user_text = update.message.text
    history = histories.setdefault(chat_id, [])

    thinking_msg = await update.message.reply_text("🤔 思考中...")

    async def _collect_response() -> str:
        collected = ""
        async for chunk in run_react(user_text, history, llm_config, db_pool=pool):
            if not chunk.startswith("data: "):
                continue
            try:
                data = json.loads(chunk[6:])
            except Exception:
                continue
            if data.get("type") == "chunk":
                collected += data.get("content", "")
            elif data.get("type") in ("done", "error"):
                if data.get("type") == "error":
                    collected = f"❌ {data.get('error', '未知错误')}"
                break
        return collected

    try:
        response_text = await asyncio.wait_for(_collect_response(), timeout=120.0)
    except asyncio.TimeoutError:
        logger.warning(f"[Bot] ReAct timeout for: {user_text[:50]}")
        await thinking_msg.delete()
        await update.message.reply_text("❌ 响应超时（120s），请稍后重试")
        return
    except Exception as e:
        logger.error(f"[Bot] ReAct error: {e}")
        await thinking_msg.delete()
        await update.message.reply_text(f"❌ 出错了：{e}")
        return

    await thinking_msg.delete()

    if response_text.strip():
        history.append({"role": "user", "content": user_text})
        history.append({"role": "ai", "content": response_text})
        if len(history) > 20:
            histories[chat_id] = history[-20:]
        for i in range(0, len(response_text), 4000):
            await update.message.reply_text(response_text[i:i + 4000])
    else:
        await update.message.reply_text("（未获得有效回复）")


async def _queue_worker(chat_id: str, queue: asyncio.Queue, context: ContextTypes.DEFAULT_TYPE):
    """每个用户独立的消息处理 worker，串行消费队列"""
    while True:
        update = await queue.get()
        try:
            await _process_message(update, context)
        except Exception as e:
            logger.error(f"[Bot] Worker error for {chat_id}: {e}")
        finally:
            queue.task_done()


@_auth
async def handle_message(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """自然语言消息 → 入队，由 worker 串行处理"""
    chat_id = str(update.effective_chat.id)
    queues: dict = context.bot_data["queues"]
    workers: dict = context.bot_data["workers"]

    if chat_id not in queues:
        queues[chat_id] = asyncio.Queue(maxsize=10)

    queue = queues[chat_id]

    # 启动该用户的 worker（只启动一次）
    if chat_id not in workers or workers[chat_id].done():
        workers[chat_id] = asyncio.create_task(
            _queue_worker(chat_id, queue, context)
        )

    if queue.full():
        await update.message.reply_text("⏳ 队列已满，请稍后再发")
        return

    await queue.put(update)


async def _load_config_from_conn(conn) -> tuple[str, str]:
    """从 notification_settings 读取 bot_token 和 chat_id"""
    row = await conn.fetchrow(
        "SELECT push_channels FROM notification_settings WHERE id = 1"
    )
    if not row:
        return "", ""

    channels = row["push_channels"]
    if isinstance(channels, str):
        try:
            channels = json.loads(channels)
        except Exception:
            return "", ""

    if not channels:
        return "", ""

    for ch in channels:
        if isinstance(ch, str):
            try:
                ch = json.loads(ch)
            except Exception:
                continue
        if ch.get("type") == "telegram" and ch.get("enabled"):
            return ch.get("bot_token", ""), ch.get("chat_id", "")
    return "", ""


async def _get_startup_config() -> tuple[str, str, dict]:
    """用临时连接（非 pool）读取启动所需配置，连接在 asyncio.run() 的 event loop 里创建后立即关闭"""
    conn = await asyncpg.connect(_DB_DSN)
    try:
        bot_token, chat_id = await _load_config_from_conn(conn)
        if not bot_token or not chat_id:
            return "", "", {}

        row = await conn.fetchrow(
            "SELECT default_model as model_name, api_key, base_url, temperature, max_tokens "
            "FROM ai_config LIMIT 1"
        )
        llm_config = {}
        if row and row["api_key"] and row["api_key"] != "sk-placeholder":
            llm_config = {
                "model_name": row["model_name"],
                "api_key": row["api_key"],
                "base_url": row["base_url"],
                "temperature": float(row["temperature"] or 0.7),
                "max_tokens": int(row["max_tokens"] or 2000),
            }
    finally:
        await conn.close()

    return bot_token, chat_id, llm_config


async def _post_init(application: Application) -> None:
    """在 PTB 自己的 event loop 里创建 asyncpg 连接池，避免 event loop 错配"""
    pool = await asyncpg.create_pool(_DB_DSN, min_size=5, max_size=20)
    application.bot_data["db_pool"] = pool
    logger.info("[Bot] DB pool created in PTB event loop")


async def _post_shutdown(application: Application) -> None:
    pool = application.bot_data.get("db_pool")
    if pool:
        await pool.close()
        logger.info("[Bot] DB pool closed")


if __name__ == "__main__":
    # 步骤 1：用 asyncio.run() 读取启动配置（临时连接，用完即关）
    bot_token, chat_id, llm_config = asyncio.run(_get_startup_config())

    if not bot_token or not chat_id:
        logger.error(
            "[Bot] No Telegram config found. "
            "Add a Telegram channel in the Config page first, then restart."
        )
        raise SystemExit(1)

    # 步骤 2：在 PTB 的 event loop 里建 pool（post_init），注册命令
    app = (
        Application.builder()
        .token(bot_token)
        .post_init(_post_init)
        .post_shutdown(_post_shutdown)
        .build()
    )

    # 静态配置写入 bot_data（不依赖 event loop）
    app.bot_data["chat_id"] = chat_id
    app.bot_data["allowed_ids"] = {"8137372066", "6580328406"}
    app.bot_data["llm_config"] = llm_config
    app.bot_data["histories"] = {}
    app.bot_data["queues"] = {}
    app.bot_data["workers"] = {}

    app.add_handler(CommandHandler("start", cmd_start))
    app.add_handler(CommandHandler("help", cmd_help))
    app.add_handler(CommandHandler("status", cmd_status))
    app.add_handler(CommandHandler("fetch", cmd_fetch))
    app.add_handler(CommandHandler("recent", cmd_recent))
    # block=False：允许并发处理多条消息，防止一条消息的 LLM 调用卡住后续消息
    app.add_handler(MessageHandler(filters.TEXT & ~filters.COMMAND, handle_message, block=False))

    logger.info(f"[Bot] Starting, authorized chat_id={chat_id}")

    # 步骤 3：PTB 自己管理 event loop，pool 在 post_init 里创建
    app.run_polling(allowed_updates=Update.ALL_TYPES)
