import random
import asyncio
import logging
from typing import Optional
from models.evaluation import EvaluationResult

logger = logging.getLogger(__name__)


class LLMClient:
    """Mock LLM client for evaluation. Can be extended with real API calls."""

    def __init__(self, model_name: str = "mock-evaluator"):
        self.model_name = model_name
        self.version = "mock-v1"

    async def evaluate(
        self,
        title: str,
        content: str,
        url: Optional[str] = None,
    ) -> EvaluationResult:
        """
        Evaluate content and generate scores and decision.

        In production, this would call an LLM API like OpenAI, Claude, etc.
        Currently returns mock evaluation for testing.
        """
        # Mock evaluation logic
        innovation_score = random.randint(3, 9)
        depth_score = random.randint(2, 9)

        # Decision based on scores
        combined_score = innovation_score + depth_score
        if combined_score >= 15:
            decision = "INTERESTING"
            reasoning = f"High innovation ({innovation_score}) and depth ({depth_score}) scores indicate valuable content."
        elif combined_score >= 10:
            decision = "BOOKMARK"
            reasoning = f"Moderate scores ({combined_score}/18) suggest potentially useful reference material."
        else:
            decision = "SKIP"
            reasoning = f"Low combined score ({combined_score}/18) suggests limited value."

        # Extract key concepts from title (mock)
        key_concepts = self._extract_concepts(title)

        # Generate TLDR (mock)
        tldr = self._generate_tldr(title, content)

        return EvaluationResult(
            innovation_score=innovation_score,
            depth_score=depth_score,
            decision=decision,
            reasoning=reasoning,
            tldr=tldr,
            key_concepts=key_concepts,
            evaluator_version=self.version,
        )

    @staticmethod
    def _extract_concepts(text: str) -> list[str]:
        """Extract key concepts from text (mock implementation)"""
        # Simple extraction: split by common delimiters
        words = text.split()
        # Return significant words (length > 4)
        concepts = [w.strip(".,!?;:") for w in words if len(w) > 4][:5]
        return concepts

    @staticmethod
    def _generate_tldr(title: str, content: str) -> str:
        """Generate TLDR summary (mock implementation)"""
        # Simple mock: use first sentence + key title words
        sentences = content.split(".")
        first_sentence = sentences[0][:100] if sentences else content[:100]
        return f"{first_sentence}... ({title[:50]})"

    async def evaluate_batch(
        self,
        items: list[dict],
    ) -> list[EvaluationResult]:
        """Evaluate multiple items concurrently"""
        tasks = [
            self.evaluate(
                title=item.get("title", ""),
                content=item.get("content", ""),
                url=item.get("url"),
            )
            for item in items
        ]
        return await asyncio.gather(*tasks)


# Real API implementation placeholder
class RealLLMClient:
    """
    Real LLM client using OpenAI/Claude API.

    To use:
    1. Set OPENAI_API_KEY or ANTHROPIC_API_KEY env var
    2. Replace LLMClient with RealLLMClient in stream_consumer.py
    """

    def __init__(
        self,
        api_key: Optional[str] = None,
        model_name: str = "gpt-4",
        api_provider: str = "openai",
    ):
        self.api_key = api_key
        self.model_name = model_name
        self.api_provider = api_provider

    async def evaluate(
        self,
        title: str,
        content: str,
        url: Optional[str] = None,
    ) -> EvaluationResult:
        """Real evaluation using LLM API"""
        # Implementation would go here
        raise NotImplementedError("Real LLM client not yet implemented")
