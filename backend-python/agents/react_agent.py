"""
ReAct Agent 主循环
基于 OpenAI function calling（兼容 GLM-5 等 OpenAI 兼容接口）

SSE 事件格式：
  {"type": "tool_call", "tool": "...", "args": {...}, "result": {...}}
  {"type": "chunk",     "content": "流式文字片段"}
  {"type": "done"}
  {"type": "error",     "error": "错误信息"}
"""

import json
import logging
from typing import AsyncGenerator, List

from agents.tools import TOOL_DEFINITIONS, execute_tool

logger = logging.getLogger(__name__)

MAX_ITERATIONS = 8

SYSTEM_PROMPT = """你是 Junk Filter 系统的专属 AI 助手，不是通用聊天机器人。

## 你拥有以下 5 个工具（必须优先使用，不要凭空回答）：

1. **query_articles** — 查询已评估的文章列表，支持按关键词、评分、状态筛选
2. **get_pipeline_status** — 获取评估管道当前状态（队列深度、待处理数量等）
3. **add_source** — 添加新的 RSS 订阅源（需要名称和 URL）
4. **remove_source** — 删除已有的 RSS 订阅源
5. **update_preferences** — 更新用户的内容评估偏好（感兴趣的话题、过滤条件等）

## 工作原则：
1. 用户询问数据时，**必须先调用工具**获取真实信息，再作答，不要编造数据
2. 展示文章时，有 tldr 就显示 TL;DR，有 reasoning 就显示评估理由，不要省略
3. 用户询问"你有什么工具/功能"时，列出上面 5 个工具并简要说明
4. 执行写操作（添加/删除源、更新偏好）时，先向用户确认再执行
5. 回复简洁，用中文
6. 你不是通用 AI，只负责 Junk Filter 系统相关的任务"""


async def run_react(
    message: str,
    history: List[dict],
    llm_config: dict,
    db_pool=None,
) -> AsyncGenerator[str, None]:
    """
    ReAct 主循环，异步生成器，yield SSE 文本行。
    使用 AsyncOpenAI 实现真实流式输出。
    """
    try:
        from openai import AsyncOpenAI
    except ImportError:
        yield _sse({"type": "error", "error": "openai 包未安装"})
        return

    model    = llm_config.get("model_name") or "gpt-5.4"
    api_key  = llm_config.get("api_key") or ""
    base_url = llm_config.get("base_url") or None

    if not api_key:
        yield _sse({"type": "error", "error": "未配置 API Key，请在设置页面填写 LLM 配置"})
        return

    client_kwargs = {"api_key": api_key}
    if base_url:
        client_kwargs["base_url"] = base_url
    try:
        client = AsyncOpenAI(**client_kwargs)
    except Exception as e:
        yield _sse({"type": "error", "error": f"LLM 客户端初始化失败: {e}"})
        return

    # 构建消息列表
    messages = [{"role": "system", "content": SYSTEM_PROMPT}]
    for msg in history[-10:]:
        role = "assistant" if msg.get("role") == "ai" else msg.get("role", "user")
        content = msg.get("content", "")
        if content:
            messages.append({"role": role, "content": content})
    messages.append({"role": "user", "content": message})

    # ── ReAct 循环 ──────────────────────────────────────────────────────────────
    for iteration in range(MAX_ITERATIONS):
        try:
            stream = await client.chat.completions.create(
                model=model,
                messages=messages,
                tools=TOOL_DEFINITIONS,
                tool_choice="auto",
                temperature=0.7,
                max_tokens=1500,
                stream=True,
            )
        except Exception as e:
            err_str = str(e)
            # 504/HTML 响应：精简为可读提示
            if "<!DOCTYPE" in err_str or "<html" in err_str or "504" in err_str:
                friendly = "LLM 服务暂时不可用（504 网关超时），请稍后重试"
            else:
                friendly = f"LLM 调用失败: {err_str[:200]}"
            logger.error(f"[ReAct] LLM 调用失败 (iter={iteration}): {err_str[:300]}")
            yield _sse({"type": "error", "error": friendly})
            return

        # ── 逐 chunk 处理流 ────────────────────────────────────────────────────
        text_content  = ""
        tool_calls_map: dict = {}   # index → {id, name, args}

        async for chunk in stream:
            if not chunk.choices:
                continue
            delta = chunk.choices[0].delta

            # 文字 chunk → 立即推送给前端（流式打字效果）
            if delta.content:
                text_content += delta.content
                yield _sse({"type": "chunk", "content": delta.content})

            # 工具调用 chunk → 只收集，不推送
            if delta.tool_calls:
                for tc in delta.tool_calls:
                    idx = tc.index
                    if idx not in tool_calls_map:
                        tool_calls_map[idx] = {"id": "", "name": "", "args": ""}
                    if tc.id:
                        tool_calls_map[idx]["id"] = tc.id
                    if tc.function and tc.function.name:
                        tool_calls_map[idx]["name"] += tc.function.name
                    if tc.function and tc.function.arguments:
                        tool_calls_map[idx]["args"] += tc.function.arguments

        # ── 无工具调用 → 流式文字已全部推送，结束 ─────────────────────────────
        if not tool_calls_map:
            # 兼容推理模型（content 为空但有 reasoning 的情况）
            if not text_content:
                yield _sse({"type": "chunk", "content": "（模型未返回内容）"})
            yield _sse({"type": "done"})
            return

        # ── 有工具调用 → 执行工具，再循环 ─────────────────────────────────────
        tool_calls_list = [
            {
                "id": tc["id"],
                "type": "function",
                "function": {"name": tc["name"], "arguments": tc["args"]},
            }
            for tc in tool_calls_map.values()
        ]
        messages.append({
            "role": "assistant",
            "content": text_content or None,
            "tool_calls": tool_calls_list,
        })

        for tc in tool_calls_map.values():
            tool_name = tc["name"]
            try:
                tool_args = json.loads(tc["args"] or "{}")
            except Exception:
                tool_args = {}

            logger.info(f"[ReAct] 调用工具 {tool_name}，参数: {tool_args}")
            result = await execute_tool(tool_name, tool_args, db_pool=db_pool)
            logger.info(f"[ReAct] 工具 {tool_name} 结果: {str(result)[:200]}")

            yield _sse({"type": "tool_call", "tool": tool_name, "args": tool_args, "result": result})

            messages.append({
                "role": "tool",
                "tool_call_id": tc["id"],
                "content": json.dumps(result, ensure_ascii=False),
            })

    yield _sse({"type": "error", "error": "Agent 达到最大迭代次数，请换个方式提问"})


def _sse(data: dict) -> str:
    return "data: " + json.dumps(data, ensure_ascii=False) + "\n\n"
