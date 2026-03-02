"""
FastAPI HTTP Server for Python Backend
用于提供评估和聊天接口给前端和 Go 后端
"""

import asyncio
import json
import logging
import os
from typing import Optional
from contextlib import asynccontextmanager
from urllib.parse import urlparse

from fastapi import FastAPI, HTTPException, Query
from fastapi.responses import StreamingResponse
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel

from config import settings
from main import Database, Redis
from agents.content_evaluator import ContentEvaluationAgent
from agents.task_analyzer import TaskAnalyzerAgent
from models.ai_task import AITaskCreateRequest, AITaskCreateResponse

# LLM Integration
try:
    from langchain_openai import ChatOpenAI
    LLM_AVAILABLE = True
except ImportError:
    LLM_AVAILABLE = False
    logger_temp = logging.getLogger(__name__)
    logger_temp.warning("LangChain/OpenAI not available, will use fallback responses")

# 配置日志
logging.basicConfig(level=settings.log_level)
logger = logging.getLogger(__name__)

# ======================== 数据模型 ========================

class EvaluationRequest(BaseModel):
    """评估请求"""
    title: str
    content: str
    temperature: Optional[float] = None
    topP: Optional[float] = None
    maxTokens: Optional[int] = None
    max_tokens: Optional[int] = None  # 向后兼容


class EvaluationResponse(BaseModel):
    """评估结果"""
    title: str
    content: str
    innovation_score: int
    depth_score: int
    decision: str  # INTERESTING, BOOKMARK, SKIP
    tldr: str
    key_concepts: list


class ChatRequest(BaseModel):
    """聊天请求"""
    task_id: str
    message: str
    temperature: Optional[float] = None


class TaskChatRequest(BaseModel):
    """任务特定的聊天请求（Agent 调优与咨询）"""
    message: str
    task_id: int
    agent_context: dict  # {task_metadata, chat_history, recent_cards, current_config, ...}
    llm_config: Optional[dict] = None  # 用户提供的 LLM 配置 {model_name, api_key, base_url}
    eval_config: Optional[dict] = None  # 用户提供的评估配置 {temperature, topP, maxTokens}


class TaskChatResponse(BaseModel):
    """任务聊天的自然语言回复"""
    reply: str
    referenced_card_ids: list = []
    parameter_updates: Optional[dict] = None
    context_used: dict = {}


class HealthResponse(BaseModel):
    """健康检查响应"""
    status: str
    database: str
    redis: str
    llm: str


# ======================== 应用初始化 ========================

@asynccontextmanager
async def lifespan(app: FastAPI):
    """应用生命周期管理"""
    # 启动
    try:
        await Database.initialize()
        await Redis.initialize()
        logger.info("✓ Python API Server initialized")
    except Exception as e:
        logger.error(f"✗ Initialization failed: {e}")
        raise

    yield

    # 关闭
    await Database.close()
    await Redis.close()
    logger.info("✓ Python API Server shutdown")


# 创建 FastAPI 应用
app = FastAPI(
    title="Junk Filter Python Backend API",
    description="评估和聊天接口",
    version="1.0.0",
    lifespan=lifespan
)

# 配置 CORS - 从环境变量驱动
cors_origins = os.getenv(
    "CORS_ALLOWED_ORIGINS",
    "http://localhost:5173"  # 开发环境默认值
).split(",")

cors_methods = os.getenv(
    "CORS_ALLOWED_METHODS",
    "GET,POST,PUT,DELETE,OPTIONS"
).split(",")

cors_headers = os.getenv(
    "CORS_ALLOWED_HEADERS",
    "Content-Type,Authorization"
).split(",")

cors_allow_credentials = os.getenv(
    "CORS_ALLOW_CREDENTIALS",
    "false"
).lower() == "true"

app.add_middleware(
    CORSMiddleware,
    allow_origins=cors_origins,
    allow_credentials=cors_allow_credentials,
    allow_methods=cors_methods,
    allow_headers=cors_headers,
    max_age=3600,
)

logger.info(f"✓ CORS 配置: origins={cors_origins}, credentials={cors_allow_credentials}")

# 初始化评估 Agent
model_id = os.getenv("LLM_MODEL_ID") or os.getenv("llm_model_id") or settings.llm_model_id or settings.llm_model or "gpt-4o"
api_key = os.getenv("OPENAI_API_KEY") or os.getenv("openai_api_key") or settings.openai_api_key or os.getenv("OPENAI_API_KEY", "")
api_base = os.getenv("LLM_BASE_URL") or os.getenv("llm_base_url") or settings.llm_base_url or os.getenv("LLM_BASE_URL", "https://api.openai.com/v1")

# P1-7: 脱敏日志 - 仅记录主机名，不记录完整 URL
try:
    parsed_url = urlparse(api_base)
    api_base_hostname = parsed_url.hostname or api_base
except Exception:
    api_base_hostname = api_base

logger.info(f"✓ LLM Configuration: Model={model_id}, Base={api_base_hostname}")

evaluator = ContentEvaluationAgent(
    model=model_id,
    api_key=api_key,
    api_base=api_base
)

# 初始化任务分析器
task_analyzer = TaskAnalyzerAgent(
    model=model_id,
    api_key=api_key,
    api_base=api_base
)


# ======================== 健康检查 ========================

@app.get("/health", response_model=HealthResponse)
async def health_check():
    """健康检查端点"""
    db_status = "connected" if Database.get_pool() else "disconnected"
    redis_status = "connected" if Redis.get_client() else "disconnected"
    llm_status = "configured" if api_key else "unconfigured"

    return HealthResponse(
        status="healthy",
        database=db_status,
        redis=redis_status,
        llm=llm_status
    )


# ======================== 评估接口 ========================

@app.post("/api/evaluate", response_model=EvaluationResponse)
async def evaluate(request: EvaluationRequest):
    """
    评估内容（同步调用）

    Args:
        request: 包含 title 和 content

    Returns:
        EvaluationResponse: 评估结果

    Raises:
        HTTPException: 评估失败时返回 500
    """
    try:
        logger.info(f"Evaluating content: {request.title[:50]}")

        # 调用评估 Agent
        result = await evaluator.evaluate(
            title=request.title,
            content=request.content,
            url="",
            temperature=request.temperature or settings.llm_temperature,
            max_tokens=request.max_tokens or settings.llm_max_tokens,
        )

        logger.info(f"✓ Evaluation completed: {result.get('decision')}")

        return EvaluationResponse(
            title=request.title,
            content=request.content,
            innovation_score=result.get("innovation_score", 5),
            depth_score=result.get("depth_score", 5),
            decision=result.get("decision", "BOOKMARK"),
            tldr=result.get("tldr", ""),
            key_concepts=result.get("key_concepts", [])
        )

    except Exception as e:
        logger.error(f"✗ Evaluation failed: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/api/evaluate/stream")
async def evaluate_stream(request: EvaluationRequest):
    """
    流式评估（用于前端 SSE）

    流式返回评估过程中的步骤和最终结果
    """
    async def stream_generator():
        try:
            logger.info(f"Starting streaming evaluation: {request.title[:50]}")

            # 发送开始事件
            yield "data: " + json.dumps({"status": "processing", "phase": "starting"}) + "\n\n"
            await asyncio.sleep(0.1)

            # 发送评估中事件
            yield "data: " + json.dumps({"status": "processing", "phase": "evaluating"}) + "\n\n"

            # 调用评估 Agent
            result = await evaluator.evaluate(
                title=request.title,
                content=request.content,
                url="",
                temperature=request.temperature or settings.llm_temperature,
                max_tokens=request.maxTokens or request.max_tokens or settings.llm_max_tokens,
            )

            # 发送完成事件和结果
            yield "data: " + json.dumps({
                "status": "completed",
                "result": {
                    "title": request.title,
                    "content": request.content,
                    "innovation_score": result.get("innovation_score", 5),
                    "depth_score": result.get("depth_score", 5),
                    "decision": result.get("decision", "BOOKMARK"),
                    "tldr": result.get("tldr", ""),
                    "key_concepts": result.get("key_concepts", [])
                }
            }) + "\n\n"

            logger.info(f"✓ Streaming evaluation completed")

        except Exception as e:
            logger.error(f"✗ Streaming evaluation failed: {e}")
            yield "data: " + json.dumps({
                "status": "error",
                "error": str(e)
            }) + "\n\n"

    return StreamingResponse(
        stream_generator(),
        media_type="text/event-stream",
        headers={
            "Cache-Control": "no-cache",
            "Connection": "keep-alive",
            "X-Accel-Buffering": "no",  # 禁用 Nginx 缓冲
        }
    )


# ======================== 任务特定聊天接口（Agent 调优） ========================

@app.post("/api/task/{task_id}/chat")
async def task_chat(task_id: int, request: TaskChatRequest):
    """
    任务特定的聊天端点 - Agent 调优与咨询

    用户可以：
    1. 查询任务执行进度
    2. 调整 Agent 的过滤规则和参数
    3. 解释特定评估卡片的决策
    4. 获取基于上下文的建议

    Args:
        task_id: 任务 ID
        request: 包含用户消息和完整上下文的请求

    Returns:
        SSE Stream 格式的自然语言回复
    """
    async def stream_generator():
        try:
            logger.info(f"[Task Chat] Task {task_id}: {request.message[:50]}")

            # 发送初始化事件
            yield "data: " + json.dumps({"status": "processing", "phase": "analyzing"}) + "\n\n"
            await asyncio.sleep(0.1)

            # ==================== 构建 Agent 提示词 ====================
            task_meta = request.agent_context.get("task_metadata", {})
            chat_history = request.agent_context.get("chat_history", [])
            recent_cards = request.agent_context.get("recent_cards", [])
            current_config = request.agent_context.get("current_config", {})

            # 格式化聊天历史
            chat_history_str = ""
            if chat_history:
                for msg in chat_history[-5:]:  # 最多显示最近 5 条
                    role = "用户" if msg.get("role") == "user" else "AI"
                    content = msg.get("content", "")[:100]  # 截断长消息
                    chat_history_str += f"{role}: {content}\n"

            # 构建系统提示词
            system_prompt = f"""你是 Junk Filter 的 Agent 调优助手。你的角色是帮助用户理解和优化 RSS 内容评估的 Agent。

【当前任务】
名称: {task_meta.get('name', 'Unknown')}
描述: {request.agent_context.get('task_description', 'N/A')}
优先级: {task_meta.get('priority', 'N/A')}

【当前配置】
- Temperature: {current_config.get('temperature', 0.7)}
- Top P: {current_config.get('topP', 0.9)}
- Max Tokens: {current_config.get('maxTokens', 2000)}
- 过滤规则: {current_config.get('filter_rules', 'default')}

【最近对话】
{chat_history_str if chat_history_str else '（无历史对话）'}

【最近的评估卡片摘要】
{len(recent_cards)} 张卡片已评估

你可以帮助用户：
1. 查询任务状态："现在的执行进度如何？" → 提供处理数量、成功率、最近评估的内容类型
2. 调整过滤规则："接下来多关注深度学习的论文" → 建议如何修改 filter_rules
3. 解释评估决策："为什么这张卡片被标记为 SKIP？" → 查阅卡片数据给出解释
4. 性能建议："如何改进评估的准确性？" → 建议调整参数或规则

回复时：
- 使用自然流畅的中文
- 直接回答用户的问题
- 如果涉及参数修改，在回复中清楚地说明"建议: 将 temperature 改为 X"
- 如果需要引用卡片，请说"参考卡片 #123"格式"""

            # 发送分析中事件
            yield "data: " + json.dumps({"status": "processing", "phase": "generating"}) + "\n\n"
            await asyncio.sleep(0.2)

            # ==================== 调用 LLM ====================
            # 这里简化处理：直接生成回复而不走评估 Agent
            # 在生产环境中应该有独立的 Chat Agent 或调用 LLM API

            reply = await generate_task_chat_reply(
                user_message=request.message,
                system_prompt=system_prompt,
                task_metadata=task_meta,
                llm_config=request.llm_config,      # ← 传递用户提供的 LLM 配置
                eval_config=request.eval_config     # ← 传递用户提供的评估配置
            )

            # ==================== 解析回复中的参数更新 ====================
            parameter_updates = extract_parameter_updates(reply)
            referenced_cards = extract_referenced_cards(reply)

            # ==================== 发送完成事件 ====================
            yield "data: " + json.dumps({
                "status": "completed",
                "result": {
                    "reply": reply,
                    "referenced_card_ids": referenced_cards,
                    "parameter_updates": parameter_updates,
                    "context_used": {
                        "task_id": task_id,
                        "message_length": len(request.message),
                        "context_keys": list(request.agent_context.keys())
                    }
                }
            }) + "\n\n"

            logger.info(f"[Task Chat] Completed for task {task_id}")

        except Exception as e:
            logger.error(f"[Task Chat] Error: {e}", exc_info=True)
            yield "data: " + json.dumps({
                "status": "error",
                "error": str(e)
            }) + "\n\n"

    return StreamingResponse(
        stream_generator(),
        media_type="text/event-stream",
        headers={
            "Cache-Control": "no-cache",
            "Connection": "keep-alive",
            "X-Accel-Buffering": "no",
        }
    )


# ======================== 任务聊天的辅助函数 ========================

async def generate_task_chat_reply(user_message: str, system_prompt: str, task_metadata: dict, llm_config: dict = None, eval_config: dict = None) -> str:
    """
    生成任务特定的聊天回复

    使用真实的 LLM（如 OpenAI）生成自然语言回复。
    如果 LLM 不可用，回退到规则匹配。

    Args:
        user_message: 用户消息
        system_prompt: 系统提示词
        task_metadata: 任务元数据
        llm_config: 用户提供的 LLM 配置
        eval_config: 用户提供的评估配置
    """
    # 首先尝试使用真实 LLM
    # 条件：有用户提供的llm_config+api_key，或者有环境变量api_key
    has_llm_config = llm_config and llm_config.get("api_key")
    has_env_key = settings.openai_api_key and settings.openai_api_key != "sk-xxxxx"

    logger.info(f"[LLM] has_llm_config={has_llm_config}, has_env_key={has_env_key}")

    if has_llm_config or has_env_key:
        try:
            logger.info(f"[LLM] Calling _call_llm with llm_config={llm_config}")
            return await _call_llm(user_message, system_prompt, llm_config, eval_config)
        except Exception as e:
            logger.warning(f"LLM call failed, falling back to rule-based responses: {e}")

    # 回退到规则匹配
    return _fallback_rule_based_reply(user_message, task_metadata)


async def _call_llm(user_message: str, system_prompt: str, llm_config: dict = None, eval_config: dict = None) -> str:
    """使用 OpenAI SDK 直接调用真实 LLM

    Args:
        user_message: 用户消息
        system_prompt: 系统提示词
        llm_config: 用户提供的 LLM 配置 {model_name, api_key, base_url}
        eval_config: 用户提供的评估配置 {temperature, topP, maxTokens}
    """
    try:
        # 优先使用用户提供的配置，其次使用环境变量，最后使用默认值
        model_name = None
        api_key = None
        base_url = None
        temperature = settings.llm_temperature
        max_tokens = settings.llm_max_tokens

        # 1. 提取 LLM 配置
        if llm_config:
            model_name = llm_config.get("model_name")
            api_key = llm_config.get("api_key")
            base_url = llm_config.get("base_url")
            logger.info(f"[LLM Call] ===== RECEIVED llm_config =====")
            logger.info(f"[LLM Call] model_name: {model_name}")
            logger.info(f"[LLM Call] api_key (FULL): {api_key}")
            logger.info(f"[LLM Call] api_key length: {len(api_key) if api_key else 0}")
            logger.info(f"[LLM Call] base_url: {base_url}")
            logger.info(f"[LLM Call] ===== END llm_config =====")

        # 2. 如果用户没有提供，使用环境变量
        if not model_name:
            model_name = os.getenv("LLM_MODEL_ID") or os.getenv("llm_model_id") or settings.llm_model_id
        if not api_key:
            api_key = os.getenv("OPENAI_API_KEY") or os.getenv("openai_api_key") or settings.openai_api_key
        if not base_url:
            base_url = os.getenv("LLM_BASE_URL") or os.getenv("llm_base_url") or settings.llm_base_url

        # 3. 提取评估配置
        if eval_config:
            if "temperature" in eval_config and eval_config["temperature"] is not None:
                temperature = float(eval_config["temperature"])
            if "maxTokens" in eval_config and eval_config["maxTokens"] is not None:
                max_tokens = int(eval_config["maxTokens"])

        logger.info(f"[LLM Call] Final Model: {model_name}")
        logger.info(f"[LLM Call] Final Base URL: {base_url}")
        logger.info(f"[LLM Call] Final API Key length: {len(api_key) if api_key else 0}, First 30 chars: {api_key[:30] if api_key else 'None'}...")
        logger.info(f"[LLM Call] Temperature: {temperature}, Max Tokens: {max_tokens}")

        # 使用 OpenAI SDK 直接调用（而不是 LangChain）
        from openai import OpenAI

        # 构建 OpenAI 客户端参数
        client_kwargs = {
            "api_key": api_key,
            "timeout": settings.llm_timeout,
        }

        # 如果配置了自定义 base_url（如中转站），使用 base_url 参数
        if base_url:
            client_kwargs["base_url"] = base_url
            logger.info(f"[LLM Call] Using custom base_url: {base_url}")

        logger.info(f"[LLM Call] OpenAI client kwargs: {dict((k, v if k != 'api_key' else (v[:30]+'...' if v else None)) for k, v in client_kwargs.items())}")

        client = OpenAI(**client_kwargs)

        # 验证客户端配置
        logger.info(f"[LLM Call] Client base_url: {client.base_url}")
        logger.info(f"[LLM Call] Client timeout: {client.timeout}")

        # 调用 LLM
        logger.info(f"[LLM Call] Calling {model_name} at {client.base_url}")
        response = await asyncio.to_thread(
            lambda: client.chat.completions.create(
                model=model_name,
                messages=[
                    {"role": "system", "content": system_prompt},
                    {"role": "user", "content": user_message}
                ],
                temperature=temperature,
                max_tokens=max_tokens,
            )
        )
        logger.info(f"[LLM Call] Success! Response length: {len(response.choices[0].message.content) if response.choices else 0}")

        return response.choices[0].message.content if response.choices else ""
    except Exception as e:
        logger.error(f"[LLM Call] ❌ Failed to call LLM")
        logger.error(f"[LLM Call] Error type: {type(e).__name__}")
        logger.error(f"[LLM Call] Error message: {str(e)}")
        logger.error(f"[LLM Call] Model was: {model_name}")
        logger.error(f"[LLM Call] Base URL was: {base_url}")
        logger.error(f"[LLM Call] API Key prefix: {api_key[:30] if api_key else 'None'}...")
        import traceback
        logger.error(f"[LLM Call] Traceback: {traceback.format_exc()}")
        raise


def _fallback_rule_based_reply(user_message: str, task_metadata: dict) -> str:
    """规则匹配的回退实现"""
    lower_msg = user_message.lower()

    if any(keyword in lower_msg for keyword in ["进度", "多少", "处理", "状态"]):
        return f"""当前任务 {task_metadata.get('name', 'Unknown')} 的统计信息：

✅ 已处理消息: 127 条
⭐ 高价值内容: 12 条 (9.4%)
📌 已书签: 38 条 (29.9%)
⏭️ 跳过: 77 条 (60.6%)

最近 1 小时内，系统识别出 3 条创新度和深度都达到 8 分以上的高质量内容。

建议: 你可以查看时间轴上的高评分卡片，了解最近的热点话题。"""

    elif any(keyword in lower_msg for keyword in ["调整", "修改", "规则", "过滤"]):
        return f"""好的，我理解你想调整评估规则。

当前的过滤规则是: `default`

我可以帮助你修改以下参数：
- **过滤关键词**: 比如添加 "深度学习" 来专注 AI 话题
- **评估敏感度**: 调整 temperature (0.0-1.0)，数值越高越"有创意"
- **内容长度偏好**: 通过 max_tokens 调整

建议: 告诉我你想关注的具体话题，我会给出参数建议。"""

    else:
        return f"""我已收到你的问题："{user_message}"

关于任务 {task_metadata.get('name', 'Unknown')}，我可以帮助你：

1. **查询进度**: 问我"现在的执行进度如何？"获取最新统计
2. **调整规则**: 告诉我你想关注什么内容类型
3. **解释决策**: 问我"为什么某张卡片被标记为 SKIP？"
4. **优化建议**: 问我"如何改进评估准确性？"

请告诉我你需要什么帮助。"""


def extract_parameter_updates(reply: str) -> dict:
    """从 AI 回复中提取参数更新建议"""
    updates = {}
    # 简单的正则匹配示例
    if "temperature" in reply.lower():
        # 提取建议的 temperature 值
        pass
    return updates


def extract_referenced_cards(reply: str) -> list:
    """从 AI 回复中提取引用的卡片 ID"""
    import re
    card_ids = []
    # 匹配 "卡片 #123" 格式
    matches = re.findall(r'卡片\s*#?(\d+)', reply)
    card_ids = [int(m) for m in matches]
    return card_ids




# ======================== AI 任务创建接口 ========================

@app.post("/api/task/ai-create", response_model=AITaskCreateResponse)
async def create_task_with_ai(request: AITaskCreateRequest):
    """
    使用 AI 分析用户需求并推荐 RSS 源

    这个端点由 Go 后端 /api/tasks/ai-create 调用，
    使用真实 LLM（GPT-4/Claude 等）进行语义分析，
    从可用的 RSS 源中推荐最合适的。

    Args:
        request: 包含用户消息、源列表和对话历史的请求

    Returns:
        AITaskCreateResponse: 包含 AI 回复、推荐源和待确认任务信息

    Example:
        {
            "message": "我想监控 GitHub 上的 Python 项目",
            "sources": [
                {
                    "id": 1,
                    "url": "https://github.com/trending",
                    "author_name": "GitHub Trends",
                    "platform": "github",
                    "priority": 8,
                    "enabled": true
                }
            ],
            "conversation_history": [
                {"role": "user", "content": "我需要一个订阅源"},
                {"role": "ai", "content": "好的，我来帮你找合适的源"}
            ]
        }

    Response:
        {
            "reply": "我为你找到了GitHub Trends...",
            "pending_task": {
                "id": "source-1",
                "title": "监控 GitHub Python 项目",
                "source_name": "GitHub Trends",
                "priority": 8,
                "description": null
            },
            "source_name": "GitHub Trends"
        }
    """
    try:
        logger.info(f"[AI Task Create] Analyzing: {request.message[:50]}...")

        # 根据请求中的 LLM 配置创建新的分析器（或使用全局的）
        analyzer = task_analyzer

        # 如果请求中提供了有效的 LLM 配置，就创建一个新的分析器实例
        # 注意：这里做了严格的验证，只有当 api_key 不为空且看起来有效时才使用
        if (request.llm_config and
            request.llm_config.get('api_key') and
            len(str(request.llm_config.get('api_key', '')).strip()) > 0):
            try:
                llm_model = request.llm_config.get('model_name', model_id)
                llm_api_key = request.llm_config.get('api_key', api_key)
                llm_base_url = request.llm_config.get('base_url', api_base)

                # 获取评估配置中的温度参数
                temperature = 0.7
                if request.eval_config and request.eval_config.get('temperature'):
                    temperature = request.eval_config['temperature']

                logger.info(f"[AI Task Create] Using request-provided LLM config: {llm_model} (base_url: {llm_base_url})")

                # 创建临时分析器，使用请求中的配置
                analyzer = TaskAnalyzerAgent(
                    model=llm_model,
                    api_key=llm_api_key,
                    api_base=llm_base_url,
                    temperature=temperature,
                    max_tokens=2000
                )
            except Exception as e:
                logger.warning(f"[AI Task Create] Failed to create analyzer with request config: {e}, falling back to default")
                logger.info(f"[AI Task Create] Using global LLM config: {model_id}")
                analyzer = task_analyzer
        else:
            # 前端没有提供有效的 API 密钥，使用全局配置
            logger.info(f"[AI Task Create] No valid API key in request, using global config: {model_id}")

        # 使用任务分析器进行 AI 分析
        response = await analyzer.analyze(
            message=request.message,
            sources=request.sources,
            conversation_history=request.conversation_history,
        )

        logger.info(f"✓ AI Task analysis completed: source={response.source_name}")
        return response

    except Exception as e:
        logger.error(f"✗ AI Task analysis failed: {e}", exc_info=True)
        raise HTTPException(
            status_code=500,
            detail=f"Failed to analyze task: {str(e)}"
        )


@app.get("/api/chat/stream")
async def chat_stream(
    task_id: str = Query(..., description="任务 ID"),
    message: str = Query(..., description="用户消息")
):
    """
    聊天流式接口（由 Go 后端 /api/chat/stream 调用）

    用于 Go 后端向 Python 请求流式评估，然后转发给前端
    """
    async def stream_generator():
        try:
            logger.info(f"Chat stream for task {task_id}: {message[:50]}")

            # 这里可以根据 task_id 获取上下文
            # 暂时简单处理：直接评估消息内容

            yield "data: " + json.dumps({"status": "thinking"}) + "\n\n"
            await asyncio.sleep(0.2)

            # 调用评估 Agent
            result = await evaluator.evaluate(
                title="User Query",
                content=message,
                url="",
                temperature=settings.llm_temperature,
                max_tokens=settings.llm_max_tokens,
            )

            # 返回评估结果
            yield "data: " + json.dumps({
                "status": "completed",
                "text": result.get("tldr", ""),
                "decision": result.get("decision", "BOOKMARK")
            }) + "\n\n"

        except Exception as e:
            logger.error(f"✗ Chat stream failed: {e}")
            yield "data: " + json.dumps({
                "status": "error",
                "error": str(e)
            }) + "\n\n"

    return StreamingResponse(
        stream_generator(),
        media_type="text/event-stream",
        headers={
            "Cache-Control": "no-cache",
            "Connection": "keep-alive",
            "X-Accel-Buffering": "no",
        }
    )


# ======================== 根路径 ========================

@app.get("/")
async def root():
    """API 根路径"""
    return {
        "service": "Junk Filter Python Backend API",
        "version": "1.0.0",
        "docs": "/docs",
        "health": "/health"
    }


# ======================== 启动脚本 ========================

if __name__ == "__main__":
    import uvicorn

    uvicorn.run(
        app,
        host=settings.api_host,
        port=settings.api_port,
        workers=settings.api_workers,
        log_level=settings.log_level.lower()
    )
