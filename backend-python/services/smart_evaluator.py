"""
SmartEvaluator - 智能评估器，LLM 优先 + 自动降级

特点：
1. 优先使用真实 LLM 进行深度评估
2. 失败自动降级到规则评估
3. 性能监控：Token 消耗、耗时统计、吞吐量追踪
4. 生产级别错误处理和日志
"""

import asyncio
import logging
import time
from typing import Optional, Dict, Any
from dataclasses import dataclass
from datetime import datetime

from agents.content_evaluator import ContentEvaluationAgent
from services.rule_evaluator import RuleBasedEvaluator
from models.evaluation import EvaluationResult

logger = logging.getLogger(__name__)


@dataclass
class EvaluationMetrics:
    """评估指标监控"""
    timestamp: datetime
    content_id: int
    source: str  # "llm" | "rule" | "fallback"
    duration_ms: float
    tokens_used: int = 0
    tokens_cost: float = 0.0
    error: Optional[str] = None
    retry_count: int = 0


class PerformanceMonitor:
    """性能监控器 - 追踪吞吐量、成本等"""

    def __init__(self, window_size: int = 100):
        self.metrics: list[EvaluationMetrics] = []
        self.window_size = window_size

    def record(self, metric: EvaluationMetrics):
        """记录一次评估"""
        self.metrics.append(metric)
        # 只保留最近 window_size 条记录
        if len(self.metrics) > self.window_size:
            self.metrics.pop(0)

    def get_stats(self) -> Dict[str, Any]:
        """获取性能统计"""
        if not self.metrics:
            return {}

        recent = self.metrics[-100:]  # 最近 100 条

        # 计算吞吐量 (items/sec)
        if len(recent) > 1:
            time_span = (recent[-1].timestamp - recent[0].timestamp).total_seconds()
            throughput = len(recent) / time_span if time_span > 0 else 0
        else:
            throughput = 0

        # 计算平均耗时
        avg_duration = sum(m.duration_ms for m in recent) / len(recent) if recent else 0

        # Token 成本
        total_tokens = sum(m.tokens_used for m in recent)
        total_cost = sum(m.tokens_cost for m in recent)

        # 成功率
        llm_count = sum(1 for m in recent if m.source == "llm")
        rule_count = sum(1 for m in recent if m.source == "rule")
        fallback_count = sum(1 for m in recent if m.source == "fallback")

        return {
            "throughput_items_per_sec": round(throughput, 2),
            "avg_duration_ms": round(avg_duration, 2),
            "total_tokens": total_tokens,
            "total_cost_usd": round(total_cost, 4),
            "llm_success_rate": round(llm_count / len(recent) * 100, 1) if recent else 0,
            "rule_fallback_rate": round((rule_count + fallback_count) / len(recent) * 100, 1) if recent else 0,
            "window": len(recent)
        }


class SmartEvaluator:
    """
    智能评估器 - LLM 优先，自动降级到规则评估

    使用场景：
    - 正常：LLM API 可用 → 使用 ContentEvaluationAgent
    - 限流：API 返回 429 → 降级到规则评估，同时限流 1s
    - 超时：API 连接超时 → 自动重试 3 次，最后降级
    - 无 Key：API Key 未配置 → 直接用规则评估
    """

    def __init__(
        self,
        llm_evaluator: Optional[ContentEvaluationAgent] = None,
        llm_enabled: bool = True,
        rate_limit_threshold: int = 3,  # 连续 N 个失败后触发降级
    ):
        self.llm_evaluator = llm_evaluator
        self.llm_enabled = llm_enabled and llm_evaluator is not None
        self.rule_evaluator = RuleBasedEvaluator()
        self.performance_monitor = PerformanceMonitor()

        # 限流控制
        self.rate_limit_threshold = rate_limit_threshold
        self.consecutive_failures = 0
        self.rate_limit_until = None

    async def evaluate(
        self,
        content_id: int,
        title: str,
        content: str,
        url: str = "",
        max_retries: int = 3,
    ) -> tuple[EvaluationResult, EvaluationMetrics]:
        """
        智能评估流程

        返回：(评估结果, 性能指标)
        """
        start_time = time.time()
        retry_count = 0
        last_error = None

        # Step 1: 检查是否在限流期间
        if self.rate_limit_until and datetime.now() < self.rate_limit_until:
            logger.warning(f"Rate limit active, using rule evaluator for content {content_id}")
            result = self.rule_evaluator.evaluate(title, content, url)
            duration_ms = (time.time() - start_time) * 1000
            metric = EvaluationMetrics(
                timestamp=datetime.now(),
                content_id=content_id,
                source="fallback",
                duration_ms=duration_ms,
                error="rate_limit"
            )
            self.performance_monitor.record(metric)
            return result, metric

        # Step 2: 尝试 LLM 评估（带重试和降级）
        if self.llm_enabled:
            for attempt in range(max_retries):
                retry_count = attempt + 1
                try:
                    logger.info(
                        f"[LLM Eval] Content {content_id}, "
                        f"Attempt {retry_count}/{max_retries}"
                    )

                    # 调用 LLM
                    result = await self._call_llm_safe(title, content, url)

                    # 成功：重置失败计数
                    self.consecutive_failures = 0

                    duration_ms = (time.time() - start_time) * 1000
                    metric = EvaluationMetrics(
                        timestamp=datetime.now(),
                        content_id=content_id,
                        source="llm",
                        duration_ms=duration_ms,
                        tokens_used=result.get("tokens_used", 0),
                        tokens_cost=result.get("tokens_cost", 0.0),
                        retry_count=retry_count - 1
                    )
                    self.performance_monitor.record(metric)

                    logger.info(
                        f"✅ [LLM Success] Content {content_id}, "
                        f"Duration: {duration_ms:.0f}ms, "
                        f"Tokens: {metric.tokens_used}"
                    )

                    return self._format_result(result), metric

                except Exception as e:
                    last_error = str(e)
                    self.consecutive_failures += 1

                    # 检查是否是限流错误
                    if "429" in last_error or "rate_limit" in last_error.lower():
                        logger.warning(f"Rate limit hit (429), activating fallback")
                        self.rate_limit_until = datetime.now()
                        # 设置 1 秒限流窗口
                        await asyncio.sleep(0.1)  # 不完全阻塞，只是降速
                        break

                    # 检查是否是 API Key 错误
                    if "invalid_api_key" in last_error.lower() or "unauthorized" in last_error.lower():
                        logger.error(f"Invalid API Key, disabling LLM evaluator")
                        self.llm_enabled = False
                        break

                    # 其他错误：继续重试
                    if attempt < max_retries - 1:
                        wait_time = 2 ** attempt  # 指数退避：1s, 2s, 4s
                        logger.warning(
                            f"⚠️  [LLM Retry] Content {content_id}, "
                            f"Error: {last_error}, "
                            f"Waiting {wait_time}s before retry..."
                        )
                        await asyncio.sleep(wait_time)
                    else:
                        logger.error(
                            f"❌ [LLM Failed] Content {content_id}, "
                            f"All retries exhausted: {last_error}"
                        )

        # Step 3: 降级到规则评估
        logger.info(f"[Rule Eval] Falling back to rule-based for content {content_id}")
        result = self.rule_evaluator.evaluate(title, content, url)

        duration_ms = (time.time() - start_time) * 1000
        metric = EvaluationMetrics(
            timestamp=datetime.now(),
            content_id=content_id,
            source="rule",
            duration_ms=duration_ms,
            error=last_error,
            retry_count=retry_count
        )
        self.performance_monitor.record(metric)

        logger.info(
            f"✅ [Rule Fallback] Content {content_id}, "
            f"Duration: {duration_ms:.0f}ms"
        )

        return self._format_result(result), metric

    async def _call_llm_safe(
        self,
        title: str,
        content: str,
        url: str = ""
    ) -> Dict[str, Any]:
        """安全的 LLM 调用，带超时和错误处理"""
        if not self.llm_evaluator:
            raise ValueError("LLM evaluator not initialized")

        try:
            # 设置 30 秒超时
            result = await asyncio.wait_for(
                asyncio.to_thread(
                    self.llm_evaluator.run,
                    title,
                    content,
                    url
                ),
                timeout=30.0
            )
            return result
        except asyncio.TimeoutError:
            raise Exception("LLM evaluation timeout (30s)")
        except Exception as e:
            raise e

    def _format_result(self, result: Dict[str, Any]) -> EvaluationResult:
        """格式化结果为 EvaluationResult"""
        return EvaluationResult(
            innovation_score=result.get("innovation_score", 0),
            depth_score=result.get("depth_score", 0),
            decision=result.get("decision", "SKIP"),
            reasoning=result.get("reasoning", ""),
            tldr=result.get("tldr", ""),
            key_concepts=result.get("key_concepts", [])
        )

    def get_performance_stats(self) -> Dict[str, Any]:
        """获取性能统计（用于监控面板）"""
        return {
            **self.performance_monitor.get_stats(),
            "llm_enabled": self.llm_enabled,
            "consecutive_failures": self.consecutive_failures,
            "rate_limited": self.rate_limit_until is not None and datetime.now() < self.rate_limit_until
        }

    async def health_check(self) -> Dict[str, Any]:
        """健康检查"""
        return {
            "status": "healthy" if self.llm_enabled else "degraded",
            "llm_available": self.llm_enabled,
            "performance": self.get_performance_stats()
        }
