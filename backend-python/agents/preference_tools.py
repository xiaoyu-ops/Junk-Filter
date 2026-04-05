"""
Preference Tools - LangChain Tools for managing user evaluation preferences

These tools are used by the chat Agent (ReAct) to silently capture and update
user preferences during natural conversation. The user does not need to know
these tools exist.
"""

import json
import logging
from typing import Optional

from langchain_core.tools import tool

logger = logging.getLogger(__name__)

# Global reference to DB pool, set during app startup
_db_pool = None


def set_db_pool(pool):
    """Set the database pool for preference tools"""
    global _db_pool
    _db_pool = pool


async def _get_preferences_from_db(source_id: Optional[int] = None) -> dict:
    """Internal: fetch preferences from database"""
    if _db_pool is None:
        return {}
    try:
        if source_id is not None:
            row = await _db_pool.fetchrow(
                "SELECT preferences FROM user_preferences WHERE source_id = $1",
                source_id,
            )
        else:
            row = await _db_pool.fetchrow(
                "SELECT preferences FROM user_preferences WHERE source_id IS NULL",
            )
        if row and row["preferences"]:
            return json.loads(row["preferences"]) if isinstance(row["preferences"], str) else row["preferences"]
    except Exception as e:
        logger.error(f"[Preferences] Error reading preferences: {e}")
    return {}


async def _save_preferences_to_db(preferences: dict, source_id: Optional[int] = None):
    """Internal: save preferences to database"""
    if _db_pool is None:
        return
    try:
        prefs_json = json.dumps(preferences, ensure_ascii=False)
        if source_id is not None:
            await _db_pool.execute(
                """INSERT INTO user_preferences (source_id, preferences, updated_at)
                   VALUES ($1, $2, NOW())
                   ON CONFLICT (source_id) DO UPDATE SET preferences = $2, updated_at = NOW()""",
                source_id, prefs_json,
            )
        else:
            # Global preferences (source_id IS NULL)
            existing = await _db_pool.fetchrow(
                "SELECT id FROM user_preferences WHERE source_id IS NULL"
            )
            if existing:
                await _db_pool.execute(
                    "UPDATE user_preferences SET preferences = $1, updated_at = NOW() WHERE source_id IS NULL",
                    prefs_json,
                )
            else:
                await _db_pool.execute(
                    "INSERT INTO user_preferences (source_id, preferences) VALUES (NULL, $1)",
                    prefs_json,
                )
        logger.info(f"[Preferences] Saved preferences for source_id={source_id}")
    except Exception as e:
        logger.error(f"[Preferences] Error saving preferences: {e}")


@tool
def get_user_preferences(source_id: Optional[int] = None) -> str:
    """获取用户的评估偏好画像。source_id 为空时获取全局偏好，指定时获取特定 RSS 源的偏好。
    返回 JSON 格式的偏好信息，包含 liked_topics, disliked_topics, quality_bias, style_preference, score_threshold 等字段。"""
    import asyncio
    try:
        loop = asyncio.get_event_loop()
        if loop.is_running():
            import concurrent.futures
            with concurrent.futures.ThreadPoolExecutor() as executor:
                future = executor.submit(asyncio.run, _get_preferences_from_db(source_id))
                prefs = future.result()
        else:
            prefs = asyncio.run(_get_preferences_from_db(source_id))
    except Exception:
        prefs = {}

    if not prefs:
        return json.dumps({"message": "暂无偏好记录", "preferences": {}}, ensure_ascii=False)
    return json.dumps(prefs, ensure_ascii=False)


@tool
def update_user_preferences(updates: str, source_id: Optional[int] = None) -> str:
    """更新用户偏好画像（增量合并，不覆盖已有内容）。updates 是 JSON 字符串，包含要更新的偏好字段。
    支持的字段：liked_topics(数组), disliked_topics(数组), quality_bias(字符串), style_preference(字符串), score_threshold(数字), custom_notes(字符串)。
    数组类型的字段会追加而非替换。"""
    import asyncio

    try:
        new_prefs = json.loads(updates)
    except json.JSONDecodeError:
        return "更新失败：无效的 JSON 格式"

    try:
        loop = asyncio.get_event_loop()
        if loop.is_running():
            import concurrent.futures
            with concurrent.futures.ThreadPoolExecutor() as executor:
                future = executor.submit(asyncio.run, _merge_and_save(new_prefs, source_id))
                return future.result()
        else:
            return asyncio.run(_merge_and_save(new_prefs, source_id))
    except Exception as e:
        logger.error(f"[Preferences] Update error: {e}")
        return f"更新失败：{str(e)}"


async def _merge_and_save(new_prefs: dict, source_id: Optional[int]) -> str:
    """Merge new preferences with existing ones and save"""
    existing = await _get_preferences_from_db(source_id)

    # Merge: arrays append, scalars overwrite
    array_fields = ["liked_topics", "disliked_topics"]
    for key, value in new_prefs.items():
        if key in array_fields and isinstance(value, list):
            existing_list = existing.get(key, [])
            # Append without duplicates
            for item in value:
                if item not in existing_list:
                    existing_list.append(item)
            existing[key] = existing_list
        else:
            existing[key] = value

    await _save_preferences_to_db(existing, source_id)
    return json.dumps({"message": "偏好已更新", "preferences": existing}, ensure_ascii=False)


def format_preferences_for_prompt(preferences: dict) -> str:
    """Format preference dict into natural language for injection into evaluation prompt"""
    if not preferences:
        return ""

    lines = ["\n[用户偏好提示]"]

    liked = preferences.get("liked_topics", [])
    if liked:
        lines.append(f"- 用户关注的领域：{', '.join(liked)}，相关主题可适当加分")

    disliked = preferences.get("disliked_topics", [])
    if disliked:
        lines.append(f"- 用户不感兴趣的领域：{', '.join(disliked)}，相关内容应降低评分")

    quality = preferences.get("quality_bias", "")
    if quality:
        lines.append(f"- 质量偏好：{quality}")

    style = preferences.get("style_preference", "")
    if style:
        lines.append(f"- 风格偏好：{style}")

    threshold = preferences.get("score_threshold")
    if threshold:
        lines.append(f"- 用户期望的最低总分门槛为 {threshold} 分，请严格把控质量")

    custom = preferences.get("custom_notes", "")
    if custom:
        lines.append(f"- 额外说明：{custom}")

    if len(lines) <= 1:
        return ""

    return "\n".join(lines)
