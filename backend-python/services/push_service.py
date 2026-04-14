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
        f"{decision_emoji} *\\[JunkFilter\\] 高分文章*\n\n"
        f"📰 *标题：* {_escape(title)}\n"
        f"⭐ 创新 {innovation_score}/10 \\| 深度 {depth_score}/10\n"
        f"🏷️ {_escape(decision)}\n\n"
        f"💡 *TL;DR：* {_escape(summary)}\n\n"
    )
    if url:
        text += f"🔗 {_escape(url)}"

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
