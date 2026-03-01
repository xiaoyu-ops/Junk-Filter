"""
Task Analyzer Agent - AI 任务创建助手

使用 LLM 分析用户自然语言需求，从可用的 RSS 源中推荐最合适的源，
并生成任务标题和优先级建议。

与 ContentEvaluationAgent 采用相同的双引擎模式：
- 主引擎：真实 LLM API (GPT-4/Claude)
- 副引擎：规则匹配（当 LLM 不可用时）
"""

import json
import logging
from typing import Optional, Dict, List
from models.ai_task import SourceInfo, ConversationMessage, AITaskCreateResponse, PendingTask
from langchain_openai import ChatOpenAI
from langchain_core.messages import HumanMessage, SystemMessage

logger = logging.getLogger(__name__)


class TaskAnalyzerAgent:
    """AI 任务创建分析器"""

    def __init__(
        self,
        model: str = "gpt-4",
        api_key: Optional[str] = None,
        api_base: Optional[str] = None,
        temperature: float = 0.7,
        max_tokens: int = 1000,
    ):
        """
        初始化任务分析器

        Args:
            model: LLM 模型名称
            api_key: API 密钥（可选，无则使用规则匹配）
            api_base: API base URL（支持中转站）
            temperature: LLM 温度参数
            max_tokens: LLM 最大 token 数
        """
        self.model = model
        self.api_key = api_key
        self.api_base = api_base
        self.temperature = temperature
        self.max_tokens = max_tokens

        # 初始化 LLM（可选）
        self.llm = None
        if api_key:
            try:
                llm_kwargs = {
                    "model": model,
                    "temperature": temperature,
                    "api_key": api_key,
                }
                if api_base:
                    llm_kwargs["base_url"] = api_base

                self.llm = ChatOpenAI(**llm_kwargs)
                logger.info(f"[TaskAnalyzerAgent] LLM initialized: {model}")
            except Exception as e:
                logger.warning(f"[TaskAnalyzerAgent] LLM initialization failed: {e}")
                self.llm = None
        else:
            logger.info("[TaskAnalyzerAgent] No API key provided, using rule-based matching")

    async def analyze(
        self,
        message: str,
        sources: List[SourceInfo],
        conversation_history: Optional[List[ConversationMessage]] = None,
    ) -> AITaskCreateResponse:
        """
        使用 AI 分析用户需求并推荐 RSS 源

        Args:
            message: 用户自然语言需求
            sources: 可用的 RSS 源列表
            conversation_history: 对话历史（用于上下文理解）

        Returns:
            AITaskCreateResponse: 包含 AI 回复、推荐源和任务信息
        """
        logger.info(f"[TaskAnalyzer] Analyzing message: {message[:50]}...")

        # 如果有 LLM，尝试使用真实 AI 分析
        if self.llm:
            try:
                return await self._analyze_with_llm(message, sources, conversation_history)
            except Exception as e:
                logger.warning(
                    f"[TaskAnalyzer] LLM analysis failed: {e}, falling back to rule-based"
                )

        # 降级到规则匹配
        return self._analyze_with_rules(message, sources, conversation_history)

    async def _analyze_with_llm(
        self,
        message: str,
        sources: List[SourceInfo],
        conversation_history: Optional[List[ConversationMessage]] = None,
    ) -> AITaskCreateResponse:
        """使用真实 LLM 进行分析"""
        import asyncio

        # 构建源列表格式化字符串
        sources_str = self._format_sources(sources)

        # 构建对话历史字符串
        history_str = ""
        if conversation_history:
            for msg in conversation_history[-3:]:  # 只显示最近 3 条消息
                role = "用户" if msg.role == "user" else "AI"
                history_str += f"{role}: {msg.content}\n"

        # 构建系统提示词
        system_prompt = f"""你是 Junk Filter 的智能任务创建助手。你的职责是帮助用户理解他们的需求，
并从可用的 RSS 源中推荐最合适的来源。

【可用的 RSS 源】
{sources_str}

【用户任务】
用户描述了他们的监控需求。你需要：
1. 理解用户的真实意图
2. 从可用源中找到最匹配的
3. 生成一个清晰、友好的回复
4. 如果找到匹配的源，提供源 ID 和建议的任务标题

【回复格式】
你必须返回以下 JSON 格式（不包含任何其他文本）：
{{
    "source_id": <source_id 或 -1 如果没有匹配>,
    "source_name": "<源名称>",
    "task_title": "<建议的任务标题>",
    "priority": <1-10的优先级>,
    "reasoning": "<为什么推荐这个源>",
    "reply": "<对用户的友好回复，使用中文>"
}}

【重要规则】
- 如果没有完全匹配的源，返回 source_id: -1
- task_title 应该清晰、简洁、可执行
- priority 应该基于源的优先级和用户的需求
- reply 应该是自然、友好的对话语气"""

        # 构建用户消息
        history_part = f"对话历史：\n{history_str}" if history_str else ""
        user_prompt = f"""用户说："{message}"

{history_part}

请分析这个需求，从可用的源中推荐最适合的。"""

        try:
            messages = [
                SystemMessage(content=system_prompt),
                HumanMessage(content=user_prompt),
            ]

            # 在线程池中调用 LLM（避免阻塞）
            loop = asyncio.get_event_loop()
            response = await loop.run_in_executor(
                None,
                lambda: self.llm.invoke(messages),
            )

            response_text = response.content if hasattr(response, "content") else str(response)
            logger.info(f"[TaskAnalyzer] LLM response: {response_text[:100]}...")

            # 解析 JSON 响应
            return self._parse_llm_response(response_text, sources)

        except Exception as e:
            logger.error(f"[TaskAnalyzer] LLM call failed: {e}", exc_info=True)
            raise

    def _analyze_with_rules(
        self,
        message: str,
        sources: List[SourceInfo],
        conversation_history: Optional[List[ConversationMessage]] = None,
    ) -> AITaskCreateResponse:
        """使用规则匹配进行降级分析"""
        logger.info(f"[TaskAnalyzer] Using rule-based analysis for: {message[:50]}")

        # 关键词匹配
        message_lower = message.lower()
        matched_source = None
        match_score = 0

        for source in sources:
            if not source.enabled:
                continue

            # 构建搜索关键词
            source_name_lower = source.author_name.lower()
            platform_lower = source.platform.lower()
            url_keywords = source.url.lower()

            # 计算匹配分数
            score = 0
            if source_name_lower in message_lower:
                score += 10  # 完全匹配源名称
            elif any(
                keyword in message_lower for keyword in source_name_lower.split()
            ):
                score += 5  # 部分匹配源名称

            if platform_lower in message_lower:
                score += 5  # 平台匹配

            if score > match_score:
                match_score = score
                matched_source = source

        # 如果找到匹配源
        if matched_source and match_score > 0:
            task_title = f"监控 {matched_source.author_name}"
            reply = f"""我理解你的需求了。我发现 "{matched_source.author_name}" 这个源非常适合你！

这个源专注于 {matched_source.platform} 领域，优先级为 {matched_source.priority}/10。

我建议创建以下任务：
- **任务名称**：{task_title}
- **优先级**：{matched_source.priority}

这样能够帮助你及时获取最新的相关内容。需要确认创建这个任务吗？"""

            return AITaskCreateResponse(
                reply=reply,
                pending_task=PendingTask(
                    id=f"source-{matched_source.id}",
                    title=task_title,
                    source_name=matched_source.author_name,
                    priority=matched_source.priority,
                ),
                source_name=matched_source.author_name,
            )

        # 如果没有找到匹配
        reply = """我理解你想要创建一个新的监控任务。不过在我们现有的源中没有完全匹配你的需求。

你可以：
1. 提供一个 RSS 源的 URL，我可以帮助添加它
2. 告诉我更多细节，看看现有的源是否能满足需求
3. 查看所有可用的源，选择一个相近的

你想怎么做呢？"""

        return AITaskCreateResponse(reply=reply)

    def _parse_llm_response(
        self, response_text: str, sources: List[SourceInfo]
    ) -> AITaskCreateResponse:
        """解析 LLM JSON 响应"""
        try:
            # 提取 JSON 部分（可能被其他文本包裹）
            import json
            import re

            # 尝试直接解析
            try:
                data = json.loads(response_text)
            except json.JSONDecodeError:
                # 尝试从文本中提取 JSON
                json_match = re.search(r"\{.*\}", response_text, re.DOTALL)
                if not json_match:
                    raise ValueError("No JSON found in response")
                data = json.loads(json_match.group())

            source_id = data.get("source_id", -1)
            source_name = data.get("source_name", "")
            task_title = data.get("task_title", "")
            priority = data.get("priority", 5)
            reply = data.get("reply", "")

            logger.info(
                f"[TaskAnalyzer] Parsed response: source_id={source_id}, title={task_title}"
            )

            # 验证 source_id 是否有效
            if source_id > 0:
                matched_source = next(
                    (s for s in sources if s.id == source_id), None
                )
                if matched_source:
                    return AITaskCreateResponse(
                        reply=reply,
                        pending_task=PendingTask(
                            id=f"source-{source_id}",
                            title=task_title,
                            source_name=source_name,
                            priority=priority,
                        ),
                        source_name=source_name,
                    )

            # source_id 无效，返回仅有回复的响应
            return AITaskCreateResponse(reply=reply)

        except Exception as e:
            logger.error(f"[TaskAnalyzer] Failed to parse LLM response: {e}")
            # 返回原始回复
            return AITaskCreateResponse(reply=response_text)

    @staticmethod
    def _format_sources(sources: List[SourceInfo]) -> str:
        """格式化源列表为易读的字符串"""
        if not sources:
            return "（无可用源）"

        lines = []
        for source in sources:
            status = "启用" if source.enabled else "禁用"
            lines.append(
                f"- ID: {source.id}, 名称: {source.author_name}, 平台: {source.platform}, "
                f"优先级: {source.priority}/10, 状态: {status}"
            )

        return "\n".join(lines)
