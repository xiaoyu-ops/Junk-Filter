"""
ContentEvaluationAgent - LLM 内容评估智能体

使用 LangGraph 构建有状态的评估流程。
LLM 调用失败时直接抛出异常，由调用方决定重试或放弃，不使用规则降级。
"""

import json
import re
import logging
from typing import TypedDict, Annotated, Optional
from langchain_openai import ChatOpenAI
from langgraph.graph import StateGraph, END
from langgraph.graph.message import add_messages
from models.evaluation import EvaluationResult

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
    engine_used: str = "llm"
    fallback_triggered: bool = False


class ContentEvaluationAgent:
    """
    LLM 内容评估 Agent

    流程：
    1. 接收内容 → 初始化状态
    2. 调用 LLM API → evaluate_node
    3. 解析和验证结果 → parse_node
    4. 失败时重试（最多 max_retries 次）
    5. 超过重试次数 → 抛出异常，由调用方处理

    不使用规则降级，失败就是失败。
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

        if not api_key:
            raise ValueError("[ContentEvaluationAgent] No API key provided")

        llm_kwargs = {
            "model": model,
            "temperature": temperature,
            "top_p": top_p,
            "max_tokens": max_tokens,
            "api_key": api_key,
            "streaming": True,  # some relays only return content in stream mode
        }
        if api_base:
            llm_kwargs["base_url"] = api_base

        self._llm_kwargs = llm_kwargs.copy()
        self._use_responses_api = False
        self.llm = ChatOpenAI(**llm_kwargs)
        logger.info(f"[ContentEvaluationAgent] LLM initialized: {model}")

        self.system_prompt = """/no_think
你是一个专业的内容评估专家。

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
        """构建 LangGraph 流程图

        Graph topology:
          evaluate → parse → [success→END, failed→END, retry→evaluate→parse→...]

        The retry edge loops back to evaluate (re-invokes LLM), not to parse,
        so each retry gets a fresh LLM response rather than re-parsing the same bad output.
        """
        workflow = StateGraph(EvaluationState)

        workflow.add_node("evaluate", self._evaluate_node)
        workflow.add_node("parse", self._parse_node)
        workflow.add_node("retry", self._retry_node)

        workflow.add_edge("evaluate", "parse")

        workflow.add_conditional_edges(
            "parse",
            self._should_retry_or_fail,
            {
                "retry": "retry",
                "success": END,
                "failed": END,
            }
        )

        # retry clears error state then feeds back into evaluate for a fresh LLM call
        workflow.add_edge("retry", "evaluate")

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
        initial_state = EvaluationState(
            title=title,
            content=content,
            url=url,
            messages=[],
            retry_count=0,
            engine_used="llm",
            fallback_triggered=False,
        )

        final_state = self.graph.invoke(initial_state)

        if final_state.get("error") or not final_state.get("decision"):
            raise RuntimeError(f"LLM evaluation failed: {final_state.get('error', 'no decision returned')}")

        return EvaluationResult(
            innovation_score=final_state.get("innovation_score", 5),
            depth_score=final_state.get("depth_score", 5),
            decision=final_state.get("decision"),
            key_concepts=final_state.get("key_concepts", []),
            tldr=final_state.get("tldr", title[:100]),
            reasoning=final_state.get("reasoning", ""),
            evaluator_version="llm"
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
        """LLM 评估节点"""
        user_prompt = f"""
请评估以下内容：

标题：{state['title']}

内容：{state['content'][:3000]}

URL：{state['url']}

请严格返回JSON格式，不添加任何解释文字。
"""
        from langchain_core.messages import HumanMessage, SystemMessage

        messages = [
            SystemMessage(content=self.system_prompt),
            HumanMessage(content=user_prompt),
        ]

        def _extract_content(response):
            # Some reasoning models put actual output in reasoning_content instead of content
            if not response.content:
                reasoning = response.additional_kwargs.get("reasoning_content") or \
                            response.additional_kwargs.get("reasoning") or ""
                if reasoning:
                    response.content = reasoning
                    logger.debug("[ContentEvaluator] Using reasoning_content as response")
            return response

        try:
            response = _extract_content(self.llm.invoke(messages))
            state["messages"] = messages + [response]
            state["error"] = ""
            state["engine_used"] = "llm"

        except Exception as e:
            err_str = str(e)
            # Auto-detect: model only supports Responses API → switch transparently
            if "openai_responses" in err_str and not self._use_responses_api:
                logger.info(f"[ContentEvaluator] {self.model} 仅支持 Responses API，自动切换")
                try:
                    self.llm = ChatOpenAI(**{**self._llm_kwargs, "use_responses_api": True})
                    self._use_responses_api = True
                    response = _extract_content(self.llm.invoke(messages))
                    state["messages"] = messages + [response]
                    state["error"] = ""
                    state["engine_used"] = "llm"
                except Exception as e2:
                    state["error"] = f"LLM 调用失败: {str(e2)}"
                    state["retry_count"] += 1
            else:
                state["error"] = f"LLM 调用失败: {err_str}"
                state["retry_count"] += 1

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
            state["tldr"] = data.get("tldr", "")[:200]
            state["reasoning"] = data.get("reasoning", "")[:500]
            state["error"] = ""
            state["engine_used"] = "llm"

            logger.info(
                f"[ContentEvaluator] LLM evaluation successful - {state['decision']}\n"
                f"  标题: {state['title'][:60]}\n"
                f"  创新: {state['innovation_score']} | 深度: {state['depth_score']} | 决策: {state['decision']}\n"
                f"  TLDR: {state['tldr']}\n"
                f"  推理: {state['reasoning']}"
            )

        except (json.JSONDecodeError, ValueError, TypeError, IndexError) as e:
            state["error"] = f"JSON 解析失败: {str(e)}"
            state["retry_count"] += 1

        return state

    def _retry_node(self, state: EvaluationState) -> EvaluationState:
        """重试节点 - 清除错误准备重新评估"""
        state["error"] = ""
        if state.get("messages") and len(state["messages"]) > 1:
            state["messages"] = state["messages"][:-1]
        return state

    def _should_retry_or_fail(self, state: EvaluationState) -> str:
        """决定是否重试或失败"""
        if not state.get("error"):
            return "success"
        if state["retry_count"] <= self.max_retries:
            return "retry"
        logger.warning(f"[ContentEvaluator] Max retries ({self.max_retries}) exceeded, giving up")
        return "failed"

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

