import asyncio
import json
import logging
import os
from typing import Optional
import redis.asyncio as aioredis
import asyncpg
from concurrent.futures import ThreadPoolExecutor
from models.evaluation import StreamMessage
from services.evaluator import EvaluatorService
from services.db_service import DBService
from agents.content_evaluator import ContentEvaluationAgent
from config import settings
from config import settings

logger = logging.getLogger(__name__)


class StreamConsumer:
    """Consumer for Redis Stream ingestion_queue"""

    def __init__(
        self,
        redis_client: aioredis.Redis,
        db_pool: asyncpg.pool.Pool,
        batch_size: int = 10,
        consumer_group: str = "evaluators",
        consumer_name: str = "evaluator-1",
    ):
        self.redis = redis_client
        self.db_pool = db_pool
        self.batch_size = batch_size
        self.consumer_group = consumer_group
        # P0: 支持多个消费者，从环境变量读取 WORKER_ID
        default_worker_id = os.getenv('WORKER_ID', '1')
        self.consumer_name = f"evaluator-{default_worker_id}"
        self.stream_name = "ingestion_queue"

        # Services
        self.db_service = DBService(db_pool)

        # P0: 创建自定义 ThreadPoolExecutor，而不是用默认的 8 个线程
        self.executor = ThreadPoolExecutor(
            max_workers=settings.llm_max_workers,
            thread_name_prefix="llm-worker-"
        )

        # ContentEvaluationAgent - 从 settings 读取配置
        model_id = settings.llm_model_id or settings.llm_model
        api_key = settings.openai_api_key or os.getenv("OPENAI_API_KEY","")
        api_base = settings.llm_base_url or os.getenv("LLM_BASE_URL","https://elysiver.h-e.top/v1")
        logger.info(f"Initializing ContentEvaluationAgent with model:{model_id} (Consumer: {self.consumer_name})")

        self.evaluator_agent = ContentEvaluationAgent(
            model=model_id,
            api_key=api_key,
            api_base=api_base,
        )

        # Legacy evaluator service 保留（用于备份）
        self.legacy_evaluator = EvaluatorService(None, self.db_service)

    async def initialize(self):
        """Initialize consumer group"""
        try:
            # Create consumer group if it doesn't exist
            # Using id="0-0" to read all messages from the beginning (for demo)
            # In production, might want to use id="$" to only read new messages
            await self.redis.xgroup_create(
                self.stream_name,
                self.consumer_group,
                id="0-0",  # Read from beginning
                mkstream=True,
            )
            logger.info(f"Created consumer group: {self.consumer_group}")
        except Exception as e:
            if "BUSYGROUP" in str(e):
                logger.info(f"Consumer group already exists: {self.consumer_group}")
            else:
                logger.error(f"Error creating consumer group: {e}")
                raise

    async def run(self):
        """Main consumer loop"""
        logger.info(
            f"Starting consumer: {self.consumer_name} "
            f"(group: {self.consumer_group}, batch_size: {self.batch_size})"
        )

        while True:
            try:
                # Read from stream
                messages = await self.redis.xreadgroup(
                    self.consumer_group,
                    self.consumer_name,
                    {self.stream_name: ">"},
                    count=self.batch_size,
                    block=1000,  # 1 second timeout
                )

                if not messages:
                    logger.debug(f"No messages available for {self.consumer_name}")
                    continue

                logger.info(f"Received {len(messages)} streams with messages for {self.consumer_name}")

                # Process messages
                for stream_name, stream_messages in messages:
                    batch = []
                    message_ids = []

                    for msg_id, msg_data in stream_messages:
                        try:
                            # Parse message - msg_data is a dict with keys/values
                            # Since decode_responses=True in Redis init, keys and values are already strings
                            data = {}
                            for key, value in msg_data.items():
                                # Handle both bytes and string keys/values
                                str_key = key.decode("utf-8") if isinstance(key, bytes) else key
                                str_value = value.decode("utf-8") if isinstance(value, bytes) else value
                                data[str_key] = str_value

                            logger.debug(f"Parsed message keys: {list(data.keys())}")

                            # Convert content_id to int if it exists
                            if "content_id" in data:
                                data["content_id"] = int(data["content_id"])

                            # Convert to StreamMessage
                            stream_msg = StreamMessage(**data)
                            batch.append(stream_msg)
                            message_ids.append(msg_id)
                        except Exception as e:
                            logger.error(f"Error parsing message {msg_id}: {e}")
                            # ACK failed message anyway
                            await self.redis.xack(
                                self.stream_name,
                                self.consumer_group,
                                msg_id,
                            )

                    if batch:
                        # Evaluate batch using ContentEvaluationAgent
                        success, failure = await self.evaluate_batch(batch)

                        # ACK processed messages
                        for msg_id in message_ids:
                            await self.redis.xack(
                                self.stream_name,
                                self.consumer_group,
                                msg_id,
                            )

                        logger.info(
                            f"Processed batch: {success} success, {failure} failures"
                        )

            except asyncio.CancelledError:
                logger.info("Consumer cancelled, shutting down gracefully...")
                break
            except Exception as e:
                logger.error(f"Error in consumer loop: {e}", exc_info=True)
                # Brief pause before retry
                await asyncio.sleep(5)

    async def evaluate_batch(self, messages: list[StreamMessage]) -> tuple[int, int]:
        """
        Evaluate multiple items using ContentEvaluationAgent

        Args:
            messages: List of StreamMessage objects

        Returns:
            Tuple of (success_count, failure_count)
        """
        success_count = 0
        failure_count = 0

        # ✨ 每次处理批次前，重新加载数据库配置（支持动态更新）
        await self._reload_llm_config()

        for message in messages:
            try:
                # Use ContentEvaluationAgent to evaluate
                result = await self._evaluate_with_agent(message)

                if result is None:
                    failure_count += 1
                    continue

                # Store evaluation result
                eval_id = await self.db_service.create_evaluation(
                    content_id=message.content_id,
                    task_id=message.task_id,
                    result=result,
                )

                if eval_id is None:
                    logger.warning(f"Could not create evaluation for {message.content_id}")
                    failure_count += 1
                    continue

                # Update content status to EVALUATED
                await self.db_service.update_content_status(
                    message.content_id,
                    "EVALUATED",
                )

                # Log status change
                await self.db_service.log_status_change(
                    content_id=message.content_id,
                    task_id=message.task_id,
                    from_status="PENDING",
                    to_status="EVALUATED",
                    reason=f"Evaluated with decision: {result.decision}",
                )

                logger.info(
                    f"Evaluated content {message.content_id}: "
                    f"Innovation={result.innovation_score}, "
                    f"Depth={result.depth_score}, "
                    f"Decision={result.decision}"
                )
                success_count += 1

            except Exception as e:
                logger.error(f"Error evaluating content {message.content_id}: {e}")
                # Mark as DISCARDED on error
                try:
                    await self.db_service.update_content_status(
                        message.content_id,
                        "DISCARDED",
                    )
                    await self.db_service.log_status_change(
                        content_id=message.content_id,
                        task_id=message.task_id,
                        from_status="PENDING",
                        to_status="DISCARDED",
                        reason=f"Evaluation error: {str(e)}",
                    )
                except Exception as log_e:
                    logger.error(f"Error logging status change: {log_e}")
                failure_count += 1

        return success_count, failure_count

    async def _reload_llm_config(self):
        """
        ✨ 重新加载数据库中的 LLM 配置（支持动态更新）

        流程：
        1. 查询 model_config 表，获取最新启用的配置
        2. 如果找到配置且 API Key 有效，更新 evaluator_agent
        3. 如果没有配置，保持现有设置（回退到启动时的配置）
        """
        try:
            from config import load_llm_config_from_db

            db_config = await load_llm_config_from_db(self.db_pool)

            if db_config and db_config.get('api_key'):
                # 配置有效，重新创建 evaluator_agent
                new_model = db_config.get('model_name', self.evaluator_agent.model)
                new_api_key = db_config['api_key']
                new_api_base = db_config.get('base_url') or os.getenv("LLM_BASE_URL", "https://elysiver.h-e.top/v1")

                # 只在配置变化时重新初始化（避免频繁重建）
                if (new_model != self.evaluator_agent.model or
                    new_api_key != getattr(self.evaluator_agent, 'api_key', '')):

                    self.evaluator_agent = ContentEvaluationAgent(
                        model=new_model,
                        api_key=new_api_key,
                        api_base=new_api_base,
                    )
                    logger.info(f"[Config] Reloaded LLM config from database: {new_model}")
            else:
                logger.debug("[Config] No enabled LLM config in database, using current settings")

        except Exception as e:
            logger.warning(f"[Config] Error reloading LLM config: {e}")
            # 配置重载失败，继续使用现有配置（不中断评估）

    async def _evaluate_with_agent(self, message: StreamMessage):
        """
        Use ContentEvaluationAgent to evaluate a single message

        Args:
            message: StreamMessage to evaluate

        Returns:
            EvaluationResult or None if evaluation fails
        """
        try:
            # Update content status to PROCESSING
            await self.db_service.update_content_status(
                message.content_id,
                "PROCESSING",
            )

            # P0: 使用自定义 ThreadPoolExecutor（50 个线程）而不是默认的 8 个
            # 这样可以支持 25 items/sec，而不是 4 items/sec
            loop = asyncio.get_event_loop()
            result = await loop.run_in_executor(
                self.executor,  # ← P0: 自定义线程池，不用 None (默认)
                self.evaluator_agent.run,
                message.title,
                message.content,
                message.url,
            )

            return result

        except Exception as e:
            logger.error(f"Error in agent evaluation: {e}")
            return None

    async def get_pending_count(self) -> int:
        """Get count of pending messages"""
        try:
            count = await self.redis.xlen(self.stream_name)
            return count
        except Exception as e:
            logger.error(f"Error getting pending count: {e}")
            return 0

    async def get_consumer_info(self) -> dict:
        """Get consumer group info"""
        try:
            info = await self.redis.xinfo_groups(self.stream_name)
            return info
        except Exception as e:
            logger.error(f"Error getting consumer info: {e}")
            return {}

    async def stop(self):
        """Graceful shutdown"""
        logger.info("Stopping consumer...")
        # Consumer cleanup handled by event loop cancellation

