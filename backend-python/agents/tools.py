"""
Agent Tools — 5 个工具实现
调用 Go API 或直接操作 DB
"""

import json
import logging
import os
import aiohttp

logger = logging.getLogger(__name__)

GO_API_BASE = os.getenv("GO_API_BASE", "http://localhost:8080")

# ── OpenAI function calling 格式的工具定义 ─────────────────────────────────────

TOOL_DEFINITIONS = [
    {
        "type": "function",
        "function": {
            "name": "add_source",
            "description": "添加新的 RSS 订阅源到系统",
            "parameters": {
                "type": "object",
                "properties": {
                    "name": {"type": "string", "description": "订阅源显示名称"},
                    "url":  {"type": "string", "description": "RSS 订阅源 URL"},
                    "priority": {"type": "integer", "description": "优先级 1-10，默认 5"},
                },
                "required": ["name", "url"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "remove_source",
            "description": "删除指定 ID 的 RSS 订阅源",
            "parameters": {
                "type": "object",
                "properties": {
                    "source_id": {"type": "integer", "description": "订阅源的数字 ID"},
                },
                "required": ["source_id"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "query_articles",
            "description": "查询文章列表，支持按关键词、状态、最低评分过滤",
            "parameters": {
                "type": "object",
                "properties": {
                    "keyword":   {"type": "string",  "description": "在标题和 TLDR 中搜索的关键词"},
                    "status":    {"type": "string",  "enum": ["PENDING", "PROCESSING", "EVALUATED", "DISCARDED"], "description": "按状态过滤"},
                    "min_score": {"type": "number",  "description": "最低综合评分（innovation+depth 平均，0-10）"},
                    "limit":     {"type": "integer", "description": "返回条数上限，默认 10，最大 20"},
                },
                "required": [],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "get_pipeline_status",
            "description": "查询当前评估管道状态：各状态文章数量（PENDING/PROCESSING/EVALUATED/DISCARDED）",
            "parameters": {
                "type": "object",
                "properties": {},
                "required": [],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "update_preferences",
            "description": "更新用户的内容评估偏好，影响后续 LLM 评估结果",
            "parameters": {
                "type": "object",
                "properties": {
                    "liked_topics":          {"type": "array",   "items": {"type": "string"}, "description": "感兴趣的主题列表"},
                    "disliked_topics":       {"type": "array",   "items": {"type": "string"}, "description": "不感兴趣的主题列表"},
                    "min_innovation_score":  {"type": "integer", "description": "最低创新分要求（0-10）"},
                    "min_depth_score":       {"type": "integer", "description": "最低深度分要求（0-10）"},
                },
                "required": [],
            },
        },
    },
]

# ── 调度入口 ───────────────────────────────────────────────────────────────────

async def execute_tool(name: str, args: dict, db_pool=None) -> dict:
    """根据工具名分发执行，统一捕获异常"""
    try:
        if name == "add_source":
            return await _add_source(**args)
        elif name == "remove_source":
            return await _remove_source(**args)
        elif name == "query_articles":
            return await _query_articles(db_pool=db_pool, **args)
        elif name == "get_pipeline_status":
            return await _get_pipeline_status()
        elif name == "update_preferences":
            return await _update_preferences(db_pool=db_pool, **args)
        else:
            return {"error": f"未知工具: {name}"}
    except Exception as e:
        logger.error(f"[Tool] {name} 执行失败: {e}")
        return {"error": str(e)}

# ── 工具实现 ───────────────────────────────────────────────────────────────────

async def _add_source(name: str, url: str, priority: int = 5) -> dict:
    payload = {"name": name, "url": url, "priority": priority, "enabled": True}
    async with aiohttp.ClientSession() as session:
        async with session.post(f"{GO_API_BASE}/api/sources", json=payload) as resp:
            if resp.status in (200, 201):
                data = await resp.json()
                return {"success": True, "source_id": data.get("id"), "message": f"已添加订阅源「{name}」"}
            text = await resp.text()
            return {"success": False, "error": text}


async def _remove_source(source_id: int) -> dict:
    async with aiohttp.ClientSession() as session:
        async with session.delete(f"{GO_API_BASE}/api/sources/{source_id}") as resp:
            if resp.status in (200, 204):
                return {"success": True, "message": f"已删除订阅源 ID={source_id}"}
            text = await resp.text()
            return {"success": False, "error": text}


async def _query_articles(
    keyword: str = None,
    status: str = None,
    min_score: float = None,
    limit: int = 10,
    db_pool=None,
) -> dict:
    if not db_pool:
        return {"error": "数据库不可用"}

    limit = min(int(limit), 20)
    conditions = []
    params = []
    idx = 1

    if status:
        conditions.append(f"c.status = ${idx}")
        params.append(status)
        idx += 1

    if keyword:
        conditions.append(f"(c.title ILIKE ${idx} OR e.tldr ILIKE ${idx})")
        params.append(f"%{keyword}%")
        idx += 1

    where = ("WHERE " + " AND ".join(conditions)) if conditions else ""

    score_filter = ""
    if min_score is not None:
        params.append(float(min_score))
        score_filter = f"HAVING (COALESCE(e.innovation_score,0) + COALESCE(e.depth_score,0)) / 2.0 >= ${idx}"
        idx += 1

    query = f"""
        SELECT c.id, c.title, c.status, c.original_url,
               e.innovation_score, e.depth_score, e.decision, e.tldr, e.reasoning
        FROM content c
        LEFT JOIN evaluation e ON e.content_id = c.id
        {where}
        GROUP BY c.id, c.title, c.status, c.original_url,
                 e.innovation_score, e.depth_score, e.decision, e.tldr, e.reasoning
        {score_filter}
        ORDER BY c.created_at DESC
        LIMIT {limit}
    """

    rows = await db_pool.fetch(query, *params)
    articles = [
        {
            "id": r["id"],
            "title": r["title"],
            "status": r["status"],
            "url": r["original_url"],
            "innovation_score": r["innovation_score"],
            "depth_score": r["depth_score"],
            "decision": r["decision"],
            "tldr": (r["tldr"] or "")[:120],
            "reasoning": (r["reasoning"] or "")[:200],
        }
        for r in rows
    ]
    return {"articles": articles, "count": len(articles)}


async def _get_pipeline_status() -> dict:
    async with aiohttp.ClientSession() as session:
        async with session.get(f"{GO_API_BASE}/api/content/stats") as resp:
            if resp.status == 200:
                return await resp.json()
            return {"error": "无法获取管道状态"}


async def _update_preferences(
    liked_topics: list = None,
    disliked_topics: list = None,
    min_innovation_score: int = None,
    min_depth_score: int = None,
    db_pool=None,
) -> dict:
    if not db_pool:
        return {"error": "数据库不可用"}

    prefs = {}
    if liked_topics is not None:
        prefs["liked_topics"] = liked_topics
    if disliked_topics is not None:
        prefs["disliked_topics"] = disliked_topics
    if min_innovation_score is not None:
        prefs["min_innovation_score"] = min_innovation_score
    if min_depth_score is not None:
        prefs["min_depth_score"] = min_depth_score

    if not prefs:
        return {"error": "未提供任何偏好参数"}

    try:
        from agents.preference_tools import _merge_and_save
        await _merge_and_save(prefs, source_id=None)
        return {"success": True, "updated": prefs, "message": "偏好已更新"}
    except Exception as e:
        return {"error": str(e)}
