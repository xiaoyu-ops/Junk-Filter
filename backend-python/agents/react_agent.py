"""
ReAct Agent 主循环 —— 推理-行动（Reasoning-Acting）循环

架构：
  1. 调用 LLM（带 tools 定义），流式接收响应
  2. 若 LLM 返回 tool_calls → 执行对应工具 → 结果回传给 LLM → 再次调用 LLM
  3. 若 LLM 只返回文字 → 流式推送给用户，结束

输出：SSE（Server-Sent Events）流，前端用 EventSource 接收
  {"type": "chunk",     "content": "流式文字片段"}   ← 正常对话文字
  {"type": "tool_call", "tool": "...", "args": {...}, "result": {...}}  ← 工具执行
  {"type": "done"}                                          ← 一轮对话结束
  {"type": "error",     "error": "错误信息"}            ← 异常终止

兼容层：
  - Chat Completions API（标准 OpenAI 格式）
  - Responses API（OpenAI 新格式，如 o1/o3 系列）
  - 运行时自动检测：首次调用失败后缓存到 _RESPONSES_API_MODELS
"""

import json
import logging
from typing import AsyncGenerator, List

from agents.tools import TOOL_DEFINITIONS, execute_tool

logger = logging.getLogger(__name__)

MAX_ITERATIONS = 8  # ReAct 循环最大迭代次数，防止无限递归（如工具链死循环）

# 运行时缓存：已确认只支持 Responses API 的模型（首次调用失败后自动记录）
# 避免每次重启都重复探测模型类型
_RESPONSES_API_MODELS: set = set()


class _ThoughtFilter:
    """
    过滤流式输出中的 <thought>…</thought> 标签（如 Gemma 4 的思考链）。

    为什么需要：某些模型在流式输出时会夹杂推理过程，如
      <thought>我需要查询文章列表...</thought>用户想要...
    这些思考内容不应展示给用户，需要实时过滤。

    实现为状态机：维护一个缓冲区，遇到 <thought> 进入跳过态，
    遇到 </thought> 退出。注意处理标签跨 chunk 的边界情况。
    """

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
    """
    将 Chat Completions 的 tool 格式转换为 Responses API 格式。

    Chat Completions: {"type": "function", "function": {"name": "...", "parameters": {...}}}
    Responses API:    {"type": "function", "name": "...", "parameters": {...}}
    区别：后者少了嵌套的 "function" 键，字段平铺。
    """
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
    """
    Responses API 路径的 ReAct 循环。

    与 Chat Completions 的区别：
      - input 参数格式不同（item list vs messages list）
      - 工具定义格式不同（name/params vs function.name/function.parameters）
      - tool_call 结果回传格式不同（function_call_output vs tool role message）
    """
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
    ReAct 主入口。异步生成器，yield SSE 文本行。

    执行流程：
      1. 用 llm_config 初始化 AsyncOpenAI 客户端
      2. 构建消息列表（system prompt + 历史消息 + 当前用户消息）
      3. 若模型已知只支持 Responses API → 走 _responses_api_loop
      4. 否则走标准 Chat Completions 路径
      5. 在循环中：LLM 调用 → 检查 tool_calls → 有则执行工具并回传 → 无则输出文字并结束

    关键设计：流式输出不等待完整响应，收到 chunk 立即 yield，
    前端可以实时看到"打字效果"。
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

    # 构建消息列表：system prompt + 最近 10 条历史 + 当前用户消息
    # 只取最近 10 条是为了控制 token 消耗，避免长历史导致超出模型上下文窗口
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
    # 每次迭代 = 一次 LLM 调用。若 LLM 决定调用工具，执行后回传结果，进入下一轮迭代。
    # 最大迭代次数限制防止无限循环（如工具链调用形成闭环）。
    for iteration in range(MAX_ITERATIONS):
        try:
            stream = await client.chat.completions.create(
                model=model,
                messages=messages,
                tools=TOOL_DEFINITIONS,
                tool_choice="auto",  # "auto" = LLM 自行决定是否调用工具；"none" = 禁止调工具
                temperature=0.7,       # 控制创造性：0 = 确定性输出，1 = 高度随机
                max_tokens=1500,       # 限制单轮输出长度，防止超长回复消耗过多 token
                stream=True,  # 流式响应，chunk 到达即处理
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

            # 工具调用在流式输出中是分片到达的：
            # chunk 1: id="call_xxx", function.name="query_articl"
            # chunk 2: function.name="es", function.arguments="{\"keywor"
            # chunk 3: function.arguments="d\": \"AI\"}"
            # 同一个 tool call 的 index 稳定，按 index 聚合 name + args
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
        # 把 assistant 的工具调用意图加入消息历史（content 必须为 null 而非空字符串）
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
            "content": text_content or None,  # OpenAI 要求 tool_calls 时 content=null
            "tool_calls": tool_calls_list,
        })

        # 逐个执行工具，结果回传给 LLM
        for tc in tool_calls_map.values():
            tool_name = tc["name"]
            try:
                tool_args = json.loads(tc["args"] or "{}")
            except Exception:
                tool_args = {}

            logger.info(f"[ReAct] 调用工具 {tool_name}，参数: {tool_args}")
            result = await execute_tool(tool_name, tool_args, db_pool=db_pool)
            logger.info(f"[ReAct] 工具 {tool_name} 结果: {str(result)[:200]}")

            # 通知前端：工具已执行及结果（用于 UI 展示工具调用过程）
            yield _sse({"type": "tool_call", "tool": tool_name, "args": tool_args, "result": result})

            # 工具结果以 "tool" role 消息回传，LLM 下一轮据此生成回复
            messages.append({
                "role": "tool",
                "tool_call_id": tc["id"],
                "content": json.dumps(result, ensure_ascii=False),
            })

    yield _sse({"type": "error", "error": "Agent 达到最大迭代次数，请换个方式提问"})


def _sse(data: dict) -> str:
    """
    包装 SSE（Server-Sent Events）格式的一行数据。

    SSE 协议要求每行以 "data: " 开头，以双换行符结束。
    前端 EventSource 会自动解析 data 字段中的 JSON。
    """
    return "data: " + json.dumps(data, ensure_ascii=False) + "\n\n"
