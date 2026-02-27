import asyncio
import json
import logging
import os
from typing import Optional
import redis.asyncio as aioredis
import asyncpg
from models.evaluation import StreamMessage
from services.evaluator import EvaluatorService
from services.db_service import DBService
from agents.content_evaluator import ContentEvaluationAgent
from config import settings
from config import settings
import os

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
        self.consumer_name = consumer_name
        self.stream_name = "ingestion_queue"

        # Services
        self.db_service = DBService(db_pool)

        # ContentEvaluationAgent - 从 settings 读取配置
        model_id = settings.llm_model_id or settings.llm_model
        api_key = settings.openai_api_key or os.getenv("OPENAI_API_KEY","")
        api_base = settings.llm_base_url or os.getenv("LLM_BASE_URL","https://elysiver.h-e.top/v1")
        logger.info(f"Initializing ContentEvaluationAgent with model:{model_id}")

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
            await self.redis.xgroup_create(
                self.stream_name,
                self.consumer_group,
                id="$",
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
                    continue

                # Process messages
                for stream_name, stream_messages in messages:
                    batch = []
                    message_ids = []

                    for msg_id, msg_data in stream_messages:
                        try:
                            # Parse message
                            data_str = msg_data.get(b"data", b"{}").decode("utf-8")
                            data = json.loads(data_str)

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

            # Run agent evaluation (synchronously within async context)
            # Note: ContentEvaluationAgent.run is synchronous, we run it in executor
            loop = asyncio.get_event_loop()
            result = await loop.run_in_executor(
                None,
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

