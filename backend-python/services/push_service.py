"""
Push Service - Send notifications via Bark (iOS)
"""

import logging

import aiohttp

logger = logging.getLogger(__name__)


async def push_to_channels(db_pool, title: str, summary: str, innovation_score: int, depth_score: int, decision: str, url: str = ""):
    """Read push channels from DB and send notification to each enabled Bark channel"""
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
                if channel_type == "bark":
                    await _push_bark(channel, title, summary)
                else:
                    logger.warning(f"[Push] Unknown channel type: {channel_type}")
            except Exception as e:
                logger.error(f"[Push] Failed to push via {channel_type}: {e}")

    except Exception as e:
        logger.error(f"[Push] Error reading push channels: {e}")


async def _push_bark(channel: dict, title: str, summary: str):
    """Push via Bark (iOS)"""
    server_url = channel.get("server_url", "").rstrip("/")
    if not server_url:
        return

    async with aiohttp.ClientSession() as session:
        resp = await session.post(
            server_url,
            json={
                "title": f"[JunkFilter] {title[:50]}",
                "body": summary or title,
                "group": "JunkFilter",
            },
            timeout=aiohttp.ClientTimeout(total=10),
        )
        data = await resp.json()
        if data.get("code") == 200:
            logger.info(f"[Push] Bark sent: {title[:30]}")
        else:
            logger.warning(f"[Push] Bark error: {data}")


async def test_push_channel(channel: dict) -> dict:
    """Test a single push channel with a sample message"""
    channel_type = channel.get("type", "")
    title = "Junk Filter 测试通知"
    summary = "这是一条测试消息，如果你收到了说明推送配置成功。"

    try:
        if channel_type == "bark":
            await _push_bark(channel, title, summary)
        else:
            return {"success": False, "message": f"未知渠道类型: {channel_type}"}

        return {"success": True, "message": "测试消息已发送，请检查手机"}
    except Exception as e:
        return {"success": False, "message": str(e)}
