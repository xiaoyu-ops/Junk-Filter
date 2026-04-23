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

# Runtime cache: models confirmed to require Responses API (populated on first mismatch error)
_RESPONSES_API_MODELS: set = set()


class _ThoughtFilter:
    """Strips <thought>…</thought> blocks from streaming content (e.g. Gemma 4)."""

    def __init__(self):
        self._buf = ""
        self._in_thought = False

    def feed(self, text: str) -> str:
        self._buf += text
        out = ""
        while True:
            if not self._in_thought:
                idx = self._buf.find("<thought>")
                if idx == -1:
                    # Guard: partial tag might be sitting at the tail
                    tag = "<thought>"
                    safe = len(self._buf)
                    for i in range(1, len(tag)):
                        if self._buf.endswith(tag[:i]):
                            safe = len(self._buf) - i
                            break
                    out += self._buf[:safe]
                    self._buf = self._buf[safe:]
                    break
                out += self._buf[:idx]
                self._buf = self._buf[idx + len("<thought>"):]
                self._in_thought = True
            else:
                idx = self._buf.find("</thought>")
                if idx == -1:
                    self._buf = ""
                    break
                self._buf = self._buf[idx + len("</thought>"):]
                self._in_thought = False
        return out

    def flush(self) -> str:
        if not self._in_thought:
            out = self._buf
            self._buf = ""
            return out
        return ""


def _to_responses_tools(chat_tools: list) -> list:
    """Convert Chat Completions tool format to Responses API format."""
    return [
        {
            "type": "function",
            "name": t["function"]["name"],
            "description": t["function"].get("description", ""),
            "parameters": t["function"].get("parameters", {}),
        }
        for t in chat_tools if t.get("type") == "function"
    ]


async def _responses_api_loop(
    client, model: str, input_items: list, db_pool
) -> AsyncGenerator[str, None]:
    """ReAct loop using OpenAI Responses API."""
    responses_api = getattr(client, "responses", None)
    if responses_api is None:
        yield _sse({"type": "error", "error": "openai 包不支持 Responses API，请升级：pip install -U openai"})
        return

    tools = _to_responses_tools(TOOL_DEFINITIONS)

    for iteration in range(MAX_ITERATIONS):
        try:
            stream = await responses_api.create(
                model=model,
                input=input_items,
                tools=tools,
                stream=True,
            )
        except Exception as e:
            err_str = str(e)
            err_lower = err_str.lower()
            logger.error(f"[ReAct/ResponsesAPI] 调用失败 (iter={iteration}): {err_str[:300]}")
            if "403" in err_str or "forbidden" in err_lower:
                friendly = "LLM 调用被拒绝（403 Forbidden）：API Key 无效、无权限或 IP 被限制，请检查配置"
            elif "<!doctype" in err_lower or "<html" in err_lower or "504" in err_str:
                friendly = "LLM 服务暂时不可用（504 网关超时），请稍后重试"
            else:
                friendly = f"LLM 调用失败: {err_str[:200]}"
            yield _sse({"type": "error", "error": friendly})
            return

        text_content = ""
        tool_calls = []
        thought_filter = _ThoughtFilter()

        async for event in stream:
            etype = getattr(event, "type", "")
            if etype == "response.output_text.delta":
                delta = getattr(event, "delta", "")
                if delta:
                    text_content += delta
                    visible = thought_filter.feed(delta)
                    if visible:
                        yield _sse({"type": "chunk", "content": visible})
            elif etype == "response.output_item.done":
                item = getattr(event, "item", None)
                if item and getattr(item, "type", "") == "function_call":
                    tool_calls.append({
                        "id": getattr(item, "call_id", ""),
                        "name": getattr(item, "name", ""),
                        "args": getattr(item, "arguments", "{}"),
                    })

        if not tool_calls:
            if not text_content:
                yield _sse({"type": "chunk", "content": "（模型未返回内容）"})
            yield _sse({"type": "done"})
            return

        for tc in tool_calls:
            input_items.append({
                "type": "function_call",
                "call_id": tc["id"],
                "name": tc["name"],
                "arguments": tc["args"],
            })

        for tc in tool_calls:
            try:
                tool_args = json.loads(tc["args"] or "{}")
            except Exception:
                tool_args = {}
            logger.info(f"[ReAct] 调用工具 {tc['name']}，参数: {tool_args}")
            result = await execute_tool(tc["name"], tool_args, db_pool=db_pool)
            logger.info(f"[ReAct] 工具 {tc['name']} 结果: {str(result)[:200]}")
            yield _sse({"type": "tool_call", "tool": tc["name"], "args": tool_args, "result": result})
            input_items.append({
                "type": "function_call_output",
                "call_id": tc["id"],
                "output": json.dumps(result, ensure_ascii=False),
            })

    yield _sse({"type": "error", "error": "Agent 达到最大迭代次数，请换个方式提问"})

SYSTEM_PROMPT = """你是 Junk Filter 的 AI 助手，帮用户管理他们的 RSS 内容订阅系统。你可以正常聊天，也可以在需要时调用工具操作系统。

## 可用工具：
1. **query_articles** — 查询已评估的文章，支持按关键词、评分、决策筛选
2. **get_pipeline_status** — 查看评估管道状态（待处理/评估中/已完成数量）
3. **add_source** — 添加新的 RSS 订阅源
4. **remove_source** — 删除已有的 RSS 订阅源
5. **update_preferences** — 更新内容评估偏好（感兴趣话题、过滤条件等）

## 原则：
- 用户问系统数据（文章、状态、偏好）时，调用对应工具获取真实数据，不要编造
- 展示文章时显示 TL;DR 和评估理由
- 执行写操作前先向用户确认
- 普通问题、闲聊直接回答，不需要调工具
- 回复简洁，用中文"""


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

    # 若该模型已被识别为只支持 Responses API，直接走新路径
    if model in _RESPONSES_API_MODELS or llm_config.get("use_responses_api", False):
        async for item in _responses_api_loop(client, model, messages, db_pool):
            yield item
        return

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
            # 首次遇到格式不匹配错误：该模型只支持 Responses API，自动切换
            if "openai_responses" in err_str and iteration == 0:
                logger.info(f"[ReAct] {model} 仅支持 Responses API，自动切换")
                _RESPONSES_API_MODELS.add(model)
                async for item in _responses_api_loop(client, model, messages, db_pool):
                    yield item
                return
            # 中转站返回 HTTP 错误页：提取状态码给出可读提示
            err_lower = err_str.lower()
            if "403" in err_str or "forbidden" in err_lower:
                friendly = "LLM 调用被拒绝（403 Forbidden）：API Key 无效、无权限或 IP 被限制，请检查配置"
            elif "<!doctype" in err_lower or "<html" in err_lower or "504" in err_str:
                friendly = "LLM 服务暂时不可用（504 网关超时），请稍后重试"
            else:
                friendly = f"LLM 调用失败: {err_str[:200]}"
            logger.error(f"[ReAct] LLM 调用失败 (iter={iteration}): {err_str[:300]}")
            yield _sse({"type": "error", "error": friendly})
            return

        # ── 逐 chunk 处理流 ────────────────────────────────────────────────────
        text_content  = ""
        tool_calls_map: dict = {}   # index → {id, name, args}
        thought_filter = _ThoughtFilter()

        async for chunk in stream:
            if not chunk.choices:
                continue
            delta = chunk.choices[0].delta

            # 文字 chunk → 过滤 <thought> 后推送给前端（流式打字效果）
            if delta.content:
                text_content += delta.content
                visible = thought_filter.feed(delta.content)
                if visible:
                    yield _sse({"type": "chunk", "content": visible})

                # Tool call chunks arrive fragmented across multiple stream events;
            # index is stable per tool call, so we accumulate name+args by index
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
            remaining = thought_filter.flush()
            if remaining:
                yield _sse({"type": "chunk", "content": remaining})
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
        # OpenAI API requires content=null (not empty string) when tool_calls are present
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
