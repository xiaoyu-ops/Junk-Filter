"""
ContentEvaluationAgent - 基于 LangGraph 的内容评估智能体

使用 LangGraph 构建有状态的评估流程，支持多步推理和自动重试
"""

import json
import re
from typing import TypedDict, Annotated
from langchain_openai import ChatOpenAI
from langgraph.graph import StateGraph, END
from langgraph.graph.message import add_messages
from models.evaluation import EvaluationResult


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


class ContentEvaluationAgent:
    """
    基于 LangGraph 的内容评估 Agent

    流程：
    1. 接收内容 → 初始化状态
    2. 调用 LLM 进行评估 → evaluate_node
    3. 解析和验证结果 → parse_node
    4. 如果失败且重试次数未超限 → retry（自动重试）
    5. 返回最终结果
    """

    def __init__(
        self,
        model: str = "gpt-4",
        api_key: str = None,
        max_retries: int = 2,
    ):
        """
        初始化 LangGraph Agent

        Args:
            model: LLM 模型名称（gpt-4、gpt-3.5-turbo、claude-3、qwen-max 等）
            api_key: API 密钥
            max_retries: 最大重试次数
        """
        self.model = model
        self.max_retries = max_retries
        self.llm = ChatOpenAI(model=model, api_key=api_key, temperature=0.7)

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
        workflow.add_node("retry", self._retry_node)

        # 添加边
        workflow.add_edge("evaluate", "parse")

        # 条件边：根据解析结果决定是否重试
        workflow.add_conditional_edges(
            "parse",
            self._should_retry,
            {
                "retry": "retry",
                "success": END,
                "failed": END,
            }
        )

        # 重试后回到解析
        workflow.add_edge("retry", "parse")

        # 设置入口
        workflow.set_entry_point("evaluate")

        return workflow.compile()

    def run(self, title: str, content: str, url: str) -> EvaluationResult:
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
        )

    def _evaluate_node(self, state: EvaluationState) -> EvaluationState:
        """LLM 评估节点"""
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

        except Exception as e:
            state["error"] = f"LLM 调用失败: {str(e)}"
            state["retry_count"] += 1

        return state

    def _parse_node(self, state: EvaluationState) -> EvaluationState:
        """解析节点 - 从 LLM 响应中提取 JSON"""
        if state.get("error"):
            # 如果有错误，保持错误状态准备重试
            return state

        try:
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

        except (json.JSONDecodeError, ValueError, TypeError) as e:
            state["error"] = f"JSON 解析失败: {str(e)}"
            state["retry_count"] += 1

        return state

    def _retry_node(self, state: EvaluationState) -> EvaluationState:
        """重试节点 - 清除错误并重新评估"""
        if state["retry_count"] <= self.max_retries:
            # 清除错误消息和之前的尝试，准备重试
            state["error"] = ""
            state["messages"] = state["messages"][:-1]  # 移除失败的响应
            return state
        else:
            # 超过最大重试次数，返回默认值
            state["decision"] = "BOOKMARK"
            state["innovation_score"] = 5
            state["depth_score"] = 5
            state["reasoning"] = f"评估失败，已重试 {state['retry_count']} 次"
            return state

    def _should_retry(self, state: EvaluationState) -> str:
        """决定是否重试"""
        if state.get("error"):
            if state["retry_count"] <= self.max_retries:
                return "retry"
            else:
                return "failed"
        return "success"

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
                print(f"批量评估中的项目失败：{e}")
                continue
        return results
