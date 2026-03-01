"""
AI Task Creation Models

用于处理 AI 驱动的任务创建请求和响应
"""

from pydantic import BaseModel
from typing import List, Optional


class SourceInfo(BaseModel):
    """RSS 源信息"""
    id: int
    url: str
    author_name: str
    platform: str
    priority: int
    enabled: bool = True


class ConversationMessage(BaseModel):
    """对话消息"""
    role: str  # "user" or "ai"
    content: str


class AITaskCreateRequest(BaseModel):
    """AI 任务创建请求"""
    message: str  # 用户自然语言需求
    sources: List[SourceInfo]  # 可用 RSS 源列表
    conversation_history: Optional[List[ConversationMessage]] = None  # 对话历史
    llm_config: Optional[dict] = None  # LLM 配置 {model_name, api_key, base_url}
    eval_config: Optional[dict] = None  # 评估配置 {temperature, topP, maxTokens}


class PendingTask(BaseModel):
    """待确认的任务信息"""
    id: str  # "source-{source_id}"
    title: str  # 任务标题
    source_name: str  # RSS 源名称
    priority: int  # 优先级 (1-10)
    description: Optional[str] = None  # 任务描述


class AITaskCreateResponse(BaseModel):
    """AI 任务创建响应"""
    reply: str  # AI 对话回复
    pending_task: Optional[PendingTask] = None  # 待确认的任务（可能为 None）
    source_name: Optional[str] = None  # 推荐的源名称
