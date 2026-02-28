"""
改进的 StreamConsumer - 集成 SmartEvaluator 和性能监控

特点：
1. 使用 SmartEvaluator (LLM 优先 + 自动降级)
2. 性能监控和实时统计
3. 异步批处理，支持 25 items/sec 吞吐量
4. 优雅退出和错误恢复
"""

import asyncio
import json
import logging
import os
import time
from typing import Optional
import redis.asyncio as aioredis
import asyncpg
from concurrent.futures import ThreadPoolExecutor

from models.evaluation import StreamMessage
from services.db_service import DBService
from services.smart_evaluator import SmartEvaluator
from agents.content_evaluator import ContentEvaluationAgent
from config import settings

logger = logging.getLogger(__name__)


class StreamConsumerV2:
    """
    改进的 Redis Stream 消费者

    使用 SmartEvaluator 替代原来的 ContentEvaluationAgent，
    实现 LLM 优先 + 自动降级的评估策略。
    """

    def __init__(
        self,
        redis_client: aioredis.Redis,
        db_pool: asyncpg.pool.Pool,
        batch_size: int = 50,  # P0: 50 items/batch
        consumer_group: str = "evaluators",
        consumer_name: str = "evaluator-1",
    ):
        self.redis = redis_client
        self.db_pool = db_pool
        self.batch_size = batch_size
        self.consumer_group = consumer_group

        # 从环境变量读取 WORKER_ID（支持多消费者）
        default_worker_id = os.getenv('WORKER_ID', '1')
        self.consumer_name = f"evaluator-{default_worker_id}"
        self.stream_name = "ingestion_queue"

        # 数据库服务
        self.db_service = DBService(db_pool)

        # 性能监控
        self.processed_count = 0
        self.failed_count = 0
        self.start_time = time.time()
        self.batch_times = []  # 记录每个 batch 的处理时间

        # 初始化 ContentEvaluationAgent
        model_id = settings.llm_model_id or settings.llm_model or "gpt-4-turbo"
        api_key = settings.openai_api_key or os.getenv("OPENAI_API_KEY", "")
        api_base = settings.llm_base_url or os.getenv("LLM_BASE_URL", "https://api.openai.com/v1")

        # 关键日志：输出实际使用的配置
        logger.info(f"🔧 LLM Configuration:")
        logger.info(f"  - Model: {model_id}")
        logger.info(f"  - API Base: {api_base}")
        logger.info(f"  - API Key: {api_key[:10]}..." if api_key else "  - API Key: NOT SET")

        try:
            self.llm_agent = ContentEvaluationAgent(
                model=model_id,
                api_key=api_key,
                api_base=api_base,
            )
            llm_enabled = bool(api_key)
        except Exception as e:
            logger.warning(f"⚠️  Failed to initialize LLM agent: {e}, will use rule-based only")
            self.llm_agent = None
            llm_enabled = False

        # 初始化 SmartEvaluator（LLM 优先 + 自动降级）
        self.evaluator = SmartEvaluator(
            llm_evaluator=self.llm_agent,
            llm_enabled=llm_enabled,
            rate_limit_threshold=3
        )

        logger.info(
            f"📊 StreamConsumer initialized: "
            f"Consumer={self.consumer_name}, "
            f"Batch={self.batch_size}, "
            f"LLM={'enabled' if llm_enabled else 'disabled'}"
        )

    async def initialize(self):
        """初始化消费者组"""
        try:
            await self.redis.xgroup_create(
                self.stream_name,
                self.consumer_group,
                id="0-0",  # 从头读取所有消息（演示模式）
                mkstream=True,
            )
            logger.info(f"✅ Consumer group created: {self.consumer_group}")
        except Exception as e:
            if "BUSYGROUP" in str(e):
                logger.info(f"✅ Consumer group already exists: {self.consumer_group}")
            else:
                logger.error(f"❌ Error creating consumer group: {e}")
                raise

    async def run(self):
        """主消费循环"""
        logger.info(
            f"🚀 Starting consumer: {self.consumer_name} "
            f"(group: {self.consumer_group}, batch_size: {self.batch_size})"
        )

        while True:
            try:
                # 从 Stream 读取消息
                messages = await self.redis.xreadgroup(
                    self.consumer_group,
                    self.consumer_name,
                    {self.stream_name: ">"},
                    count=self.batch_size,
                    block=1000,  # 1 秒超时
                )

                if not messages:
                    logger.debug(f"No messages for {self.consumer_name}")
                    continue

                # 处理消息
                for stream_name, stream_messages in messages:
                    batch = []
                    message_ids = []

                    # 解析消息
                    for msg_id, msg_data in stream_messages:
                        try:
                            # 转换为字典
                            data = {}
                            for key, value in msg_data.items():
                                str_key = key.decode("utf-8") if isinstance(key, bytes) else key
                                str_value = value.decode("utf-8") if isinstance(value, bytes) else value
                                data[str_key] = str_value

                            if "content_id" in data:
                                data["content_id"] = int(data["content_id"])

                            stream_msg = StreamMessage(**data)
                            batch.append(stream_msg)
                            message_ids.append(msg_id)
                        except Exception as e:
                            logger.error(f"❌ Error parsing message {msg_id}: {e}")
                            # 错误消息也要 ACK（否则会一直重处理）
                            await self.redis.xack(
                                self.stream_name,
                                self.consumer_group,
                                msg_id,
                            )

                    if batch:
                        # 评估批次
                        success, failure = await self.evaluate_batch(batch)

                        # ACK 已处理的消息
                        for msg_id in message_ids:
                            await self.redis.xack(
                                self.stream_name,
                                self.consumer_group,
                                msg_id,
                            )

                        # 输出性能统计
                        logger.info(
                            f"📈 Batch processed: "
                            f"Success={success}, Failed={failure}, "
                            f"Throughput={self._get_throughput():.1f} items/sec"
                        )

            except asyncio.CancelledError:
                logger.info("🛑 Consumer cancelled, shutting down...")
                break
            except Exception as e:
                logger.error(f"❌ Error in consumer loop: {e}", exc_info=True)
                await asyncio.sleep(5)

        # 输出最终统计
        self._log_final_stats()

    async def evaluate_batch(self, messages: list[StreamMessage]) -> tuple[int, int]:
        """
        评估一批消息

        使用 SmartEvaluator，LLM 优先 + 自动降级

        返回：(成功数, 失败数)
        """
        batch_start = time.time()
        success_count = 0
        failure_count = 0

        # 并发评估（使用 asyncio.gather）
        tasks = [
            self._evaluate_single(message)
            for message in messages
        ]

        results = await asyncio.gather(*tasks, return_exceptions=True)

        for i, result in enumerate(results):
            if isinstance(result, Exception):
                logger.error(f"❌ Evaluation exception: {result}")
                failure_count += 1
                continue

            if result is None:
                failure_count += 1
                continue

            success_count += 1

        # 记录批次耗时
        batch_duration = time.time() - batch_start
        self.batch_times.append(batch_duration)
        if len(self.batch_times) > 100:  # 只保留最近 100 个
            self.batch_times.pop(0)

        self.processed_count += success_count
        self.failed_count += failure_count

        return success_count, failure_count

    async def _evaluate_single(self, message: StreamMessage) -> Optional[int]:
        """
        评估单个内容

        返回：evaluation_id 或 None（失败）
        """
        try:
            # 调用 SmartEvaluator
            result, metrics = await self.evaluator.evaluate(
                content_id=message.content_id,
                title=message.title or "Untitled",
                content=message.content or "",
                url=message.url or "",
            )

            # 保存评估结果
            eval_id = await self.db_service.create_evaluation(
                content_id=message.content_id,
                task_id=message.task_id,
                result=result,
            )

            if eval_id is None:
                logger.warning(f"⚠️  Failed to create evaluation for content {message.content_id}")
                return None

            # 更新内容状态
            await self.db_service.update_content_status(
                content_id=message.content_id,
                status="EVALUATED"
            )

            # 日志（展示评估来源和性能）
            logger.info(
                f"✅ Evaluated content {message.content_id}: "
                f"Source={metrics.source}, "
                f"Score=({result.innovation_score}, {result.depth_score}), "
                f"Duration={metrics.duration_ms:.0f}ms, "
                f"Decision={result.decision}"
            )

            return eval_id

        except Exception as e:
            logger.error(f"❌ Error evaluating content {message.content_id}: {e}", exc_info=True)
            return None

    def _get_throughput(self) -> float:
        """计算实时吞吐量 (items/sec)"""
        elapsed = time.time() - self.start_time
        if elapsed < 1:
            return 0
        return self.processed_count / elapsed

    def _log_final_stats(self):
        """输出最终统计"""
        elapsed = time.time() - self.start_time
        logger.info(
            f"\n📊 === Final Statistics === \n"
            f"  Total Processed: {self.processed_count}\n"
            f"  Total Failed: {self.failed_count}\n"
            f"  Elapsed Time: {elapsed:.1f}s\n"
            f"  Average Throughput: {self._get_throughput():.1f} items/sec\n"
            f"  Success Rate: {self.processed_count / (self.processed_count + self.failed_count) * 100:.1f}%\n"
        )

        if self.batch_times:
            avg_batch_time = sum(self.batch_times) / len(self.batch_times)
            logger.info(f"  Average Batch Time: {avg_batch_time:.2f}s\n")

        # 性能统计
        perf_stats = self.evaluator.get_performance_stats()
        logger.info(f"  Performance Stats:\n")
        for key, value in perf_stats.items():
            logger.info(f"    - {key}: {value}\n")

    async def get_health(self) -> dict:
        """健康检查端点"""
        return {
            "status": "healthy",
            "consumer": self.consumer_name,
            "processed": self.processed_count,
            "failed": self.failed_count,
            "throughput": self._get_throughput(),
            "evaluator_health": await self.evaluator.health_check(),
        }
