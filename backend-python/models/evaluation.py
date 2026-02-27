import asyncio
import json
import logging
from typing import Optional
from datetime import datetime
from pydantic import BaseModel

logger = logging.getLogger(__name__)


class EvaluationRequest(BaseModel):
    """Request model for evaluation"""
    content_id: int
    task_id: str
    title: str
    content: str


class EvaluationResult(BaseModel):
    """Result model for evaluation"""
    innovation_score: int
    depth_score: int
    decision: str  # INTERESTING, SKIP, BOOKMARK
    reasoning: str
    tldr: str
    key_concepts: list[str]
    evaluator_version: str = "mock-v1"


class StreamMessage(BaseModel):
    """Stream message from ingestion queue"""
    content_id: int
    task_id: str
    title: str
    url: str
    content: str
    published_at: str
    platform: str
    author_name: str
    content_hash: str
