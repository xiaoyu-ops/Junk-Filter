"""
ContentEvaluationAgent - 双引擎内容评估智能体

使用 LangGraph 构建有状态的评估流程，支持：
- 主引擎：真实 LLM API (GPT-4/Claude)
- 副引擎：自动降级到 rule_based_evaluator (当 API 失败、超频、无 key 时)
"""

import json
import re
import logging
from typing import TypedDict, Annotated, Optional
from langchain_openai import ChatOpenAI
from langgraph.graph import StateGraph, END
from langgraph.graph.message import add_messages
from models.evaluation import EvaluationResult
from services.rule_evaluator import RuleBasedEvaluator

logger = logging.getLogger(__name__)


class EvaluationState(TypedDict):
    """评估过程的状态"""
    title: str
    content: str
    url: str
    messages: Annotated[list, add_messages]
    innovation_score: int = 0
    depth_score: int = 0
    decision: str = ""
    key_concepts: list = []
    tldr: str = ""
    reasoning: str = ""
    retry_count: int = 0
    error: str = ""
    engine_used: str = "llm"  # llm 或 rule_based
    fallback_triggered: bool = False  # 是否已触发降级


class ContentEvaluationAgent:
    """
    双引擎内容评估 Agent

    流程：
    1. 接收内容 → 初始化状态
    2. 尝试调用真实 LLM API (GPT-4/Claude) → evaluate_node
    3. 解析和验证结果 → parse_node
    4. 如果 API 失败、超频(429)、服务器错误(500)、无 API Key → 自动降级到 rule_based_evaluator
    5. 使用规则引擎进行评估 → fallback_evaluate_node
    6. 返回最终结果

    双引擎策略：
    - 主引擎(LLM)：更准确，支持复杂推理
    - 副引擎(规则)：快速、可靠，在主引擎不可用时自动启用
    """

    def __init__(
        self,
        model: str = "gpt-4",
        api_key: Optional[str] = None,
        api_base: Optional[str] = None,
        max_retries: int = 2,
        temperature: float = 0.7,
        top_p: float = 0.9,
        max_tokens: int = 500,
    ):
        """
        初始化双引擎 Agent

        Args:
            model: LLM 模型名称（gpt-4、gpt-3.5-turbo、claude-3、qwen-max 等）
            api_key: API 密钥（可选，无则自动使用规则引擎）
            api_base: API base URL（支持兼容 OpenAI 的服务）
            max_retries: 最大重试次数
            temperature: LLM 温度参数 (0-2)
            top_p: LLM top_p 参数 (0-1)
            max_tokens: LLM max_tokens 参数
        """
        self.model = model
        self.max_retries = max_retries
        self.temperature = temperature
        self.top_p = top_p
        self.max_tokens = max_tokens
        self.api_key = api_key
        self.api_base = api_base

        # 初始化 LLM（可选，有 API Key 时才初始化）
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
                logger.info(f"[ContentEvaluationAgent] LLM initialized: {model}")
            except Exception as e:
                logger.warning(f"[ContentEvaluationAgent] LLM initialization failed: {e}")
                self.llm = None
        else:
            logger.info("[ContentEvaluationAgent] No API key provided, using rule-based evaluator only")

        # 初始化规则引擎（副引擎）
        self.rule_evaluator = RuleBasedEvaluator()

        self.system_prompt = """你是一个专业的内容评估专家。

你需要评估提供的文章，并生成以下结构化JSON格式的评估：
{
    "innovation_score": <0-10整数>,
    "depth_score": <0-10整数>,
    "decision": "<INTERESTING|BOOKMARK|SKIP>",
    "key_concepts": [<字符串数组，最多5个关键概念>],
    "tldr": "<一句话总结，不超过100字>",
    "reasoning": "<简短的推理过程>"
}

评估维度：
1. innovation_score (0-10)：评估内容的创新度和突破性
   - 8-10：真正突破性的发现，具有革命性影响
   - 6-7：有重要的新见解，能推进领域发展
   - 4-5：有一些新的想法，但不够深入
   - 1-3：主要是既有知识的重述

2. depth_score (0-10)：评估内容的深度和严谨性
   - 8-10：深入的学术级别分析，充分的证据支持
   - 6-7：相当深入的讨论，有逻辑支持
   - 4-5：中等深度，基本的论证
   - 1-3：表面级别的讨论

3. decision：决策标准
   - INTERESTING：innovation_score >= 7 AND depth_score >= 6（高价值内容）
   - BOOKMARK：innovation_score >= 5 OR depth_score >= 5（中等价值）
   - SKIP：其他情况（低价值内容）

请严格按照JSON格式返回，不包含任何其他文本。"""

        # 构建 LangGraph
        self.graph = self._build_graph()

    def _build_graph(self):
        """构建 LangGraph 流程图"""
        workflow = StateGraph(EvaluationState)

        # 添加节点
        workflow.add_node("evaluate", self._evaluate_node)
        workflow.add_node("parse", self._parse_node)
        workflow.add_node("fallback", self._fallback_evaluate_node)
        workflow.add_node("retry", self._retry_node)

        # 添加边
        workflow.add_edge("evaluate", "parse")

        # 条件边：根据解析结果和错误类型决定是否重试或降级
        workflow.add_conditional_edges(
            "parse",
            self._should_retry_or_fallback,
            {
                "retry": "retry",
                "fallback": "fallback",
                "success": END,
                "failed": END,
            }
        )

        # 重试后回到解析
        workflow.add_edge("retry", "parse")

        # 降级后直接结束
        workflow.add_edge("fallback", END)

        # 设置入口
        workflow.set_entry_point("evaluate")

        return workflow.compile()

    def run(self, title: str, content: str, url: str = "") -> EvaluationResult:
        """
        执行评估流程

        Args:
            title: 文章标题
            content: 文章内容
            url: 文章 URL

        Returns:
            EvaluationResult: 评估结果
        """
        # 初始化状态
        initial_state = EvaluationState(
            title=title,
            content=content,
            url=url,
            messages=[],
            retry_count=0,
            engine_used="llm" if self.llm else "rule_based",
            fallback_triggered=False,
        )

        # 执行图
        final_state = self.graph.invoke(initial_state)

        # 返回结果
        return EvaluationResult(
            innovation_score=final_state.get("innovation_score", 5),
            depth_score=final_state.get("depth_score", 5),
            decision=final_state.get("decision", "BOOKMARK"),
            key_concepts=final_state.get("key_concepts", []),
            tldr=final_state.get("tldr", title[:100]),
            reasoning=final_state.get("reasoning", ""),
            evaluator_version=f"dual-engine-{final_state.get('engine_used', 'unknown')}"
        )

    async def evaluate(
        self,
        title: str,
        content: str,
        url: str = "",
        temperature: Optional[float] = None,
        max_tokens: Optional[int] = None,
    ) -> dict:
        """
        异步评估方法（用于 FastAPI）

        Args:
            title: 文章标题
            content: 文章内容
            url: 文章 URL（可选）
            temperature: 温度参数（可选，覆盖初始化值）
            max_tokens: 最大 token 数（可选，覆盖初始化值）

        Returns:
            dict: 评估结果字典
        """
        import asyncio

        # 使用参数值或初始化时的默认值
        if temperature is not None:
            self.temperature = temperature
        if max_tokens is not None:
            self.max_tokens = max_tokens

        # 在线程池中运行同步的 run 方法，避免阻塞事件循环
        loop = asyncio.get_event_loop()
        result = await loop.run_in_executor(
            None,
            self.run,
            title,
            content,
            url
        )

        # 转换 EvaluationResult 对象为字典
        return {
            "innovation_score": result.innovation_score,
            "depth_score": result.depth_score,
            "decision": result.decision,
            "key_concepts": result.key_concepts,
            "tldr": result.tldr,
            "reasoning": result.reasoning,
            "evaluator_version": result.evaluator_version,
        }

    def _evaluate_node(self, state: EvaluationState) -> EvaluationState:
        """LLM 评估节点（主引擎）"""

        # 如果没有 LLM，直接标记需要降级
        if not self.llm:
            state["error"] = "No LLM available, using fallback"
            state["fallback_triggered"] = True
            return state

        user_prompt = f"""
请评估以下内容：

标题：{state['title']}

内容：{state['content'][:3000]}

URL：{state['url']}

请严格返回JSON格式，不添加任何解释文字。
"""

        try:
            from langchain_core.messages import HumanMessage, SystemMessage

            messages = [
                SystemMessage(content=self.system_prompt),
                HumanMessage(content=user_prompt),
            ]

            response = self.llm.invoke(messages)
            state["messages"] = messages + [response]
            state["error"] = ""
            state["engine_used"] = "llm"

        except Exception as e:
            error_msg = str(e)
            state["error"] = f"LLM 调用失败: {error_msg}"
            state["retry_count"] += 1

            # 检查是否为特定错误类型（超频、服务器错误、无 key）
            if "429" in error_msg or "rate_limit" in error_msg.lower():
                logger.warning("[ContentEvaluator] Rate limit (429) detected, triggering fallback")
                state["fallback_triggered"] = True
            elif "500" in error_msg or "server_error" in error_msg.lower():
                logger.warning("[ContentEvaluator] Server error (500) detected, triggering fallback")
                state["fallback_triggered"] = True
            elif "api_key" in error_msg.lower() or "authentication" in error_msg.lower():
                logger.warning("[ContentEvaluator] API key error detected, triggering fallback")
                state["fallback_triggered"] = True

        return state

    def _parse_node(self, state: EvaluationState) -> EvaluationState:
        """解析节点 - 从 LLM 响应中提取 JSON"""
        if state.get("error"):
            # 如果有错误，保持错误状态准备重试或降级
            return state

        try:
            # 检查消息列表是否为空
            if not state.get("messages") or len(state["messages"]) == 0:
                raise ValueError("没有 LLM 响应消息")

            # 获取最后一条消息（LLM 响应）
            last_message = state["messages"][-1]
            response_text = last_message.content

            # 提取 JSON
            json_match = re.search(r"\{.*\}", response_text, re.DOTALL)
            if not json_match:
                raise ValueError("响应中未找到 JSON")

            json_str = json_match.group(0)
            data = json.loads(json_str)

            # 验证和提取字段
            state["innovation_score"] = max(0, min(10, int(data.get("innovation_score", 5))))
            state["depth_score"] = max(0, min(10, int(data.get("depth_score", 5))))
            state["decision"] = data.get("decision", "BOOKMARK").upper()

            # 验证 decision
            if state["decision"] not in ["INTERESTING", "BOOKMARK", "SKIP"]:
                state["decision"] = "BOOKMARK"

            state["key_concepts"] = data.get("key_concepts", [])[:5]
            state["tldr"] = data.get("tldr", state["title"][:100])[:100]
            state["reasoning"] = data.get("reasoning", "")[:200]
            state["error"] = ""
            state["engine_used"] = "llm"

            logger.info(f"[ContentEvaluator] LLM evaluation successful - {state['decision']}")

        except (json.JSONDecodeError, ValueError, TypeError, IndexError) as e:
            state["error"] = f"JSON 解析失败: {str(e)}"
            state["retry_count"] += 1

        return state

    def _fallback_evaluate_node(self, state: EvaluationState) -> EvaluationState:
        """降级节点 - 使用规则引擎进行评估"""
        try:
            logger.info(f"[ContentEvaluator] Fallback triggered - using rule-based evaluator")

            result = self.rule_evaluator.evaluate(
                title=state["title"],
                content=state["content"],
                url=state.get("url", "")
            )

            state["innovation_score"] = result.innovation_score
            state["depth_score"] = result.depth_score
            state["decision"] = result.decision
            state["key_concepts"] = result.key_concepts
            state["tldr"] = result.tldr
            state["reasoning"] = result.reasoning
            state["error"] = ""
            state["engine_used"] = "rule_based"
            state["fallback_triggered"] = True

            logger.info(f"[ContentEvaluator] Rule-based evaluation successful - {state['decision']}")

        except Exception as e:
            logger.error(f"[ContentEvaluator] Fallback evaluation failed: {e}")
            # 最后的降级：返回默认评估
            state["innovation_score"] = 5
            state["depth_score"] = 5
            state["decision"] = "BOOKMARK"
            state["key_concepts"] = []
            state["tldr"] = state["title"][:100]
            state["reasoning"] = f"自动评估失败: {str(e)}"
            state["engine_used"] = "default"

        return state

    def _retry_node(self, state: EvaluationState) -> EvaluationState:
        """重试节点 - 清除错误并重新评估"""
        if state["retry_count"] <= self.max_retries:
            # 清除错误消息和之前的尝试，准备重试
            state["error"] = ""
            if state.get("messages") and len(state["messages"]) > 1:
                state["messages"] = state["messages"][:-1]  # 移除失败的响应
            return state
        else:
            # 超过最大重试次数，触发降级
            logger.info(f"[ContentEvaluator] Max retries ({self.max_retries}) exceeded, triggering fallback")
            state["fallback_triggered"] = True
            return state

    def _should_retry_or_fallback(self, state: EvaluationState) -> str:
        """决定是否重试、降级还是成功"""
        # 如果已标记需要降级，直接降级
        if state.get("fallback_triggered"):
            return "fallback"

        # 如果没有错误，成功
        if not state.get("error"):
            return "success"

        # 如果有错误且未超过最大重试次数，重试
        if state["retry_count"] <= self.max_retries:
            return "retry"

        # 否则降级
        return "fallback"

    def evaluate_batch(self, items: list) -> list:
        """批量评估多个内容"""
        results = []
        for item in items:
            try:
                result = self.run(
                    item.get("title", ""),
                    item.get("content", ""),
                    item.get("url", "")
                )
                results.append(result)
            except Exception as e:
                logger.error(f"[ContentEvaluator] Batch evaluation item failed: {e}")
                # 返回默认结果
                results.append(EvaluationResult(
                    innovation_score=5,
                    depth_score=5,
                    decision="BOOKMARK",
                    reasoning=f"评估失败: {str(e)}",
                    tldr=item.get("title", "")[:100],
                    key_concepts=[],
                ))
        return results

