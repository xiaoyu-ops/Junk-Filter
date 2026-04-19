import asyncio
import json
import logging
import os
import time
from typing import Optional
import redis.asyncio as aioredis
import asyncpg
from models.evaluation import StreamMessage
from services.db_service import DBService
from agents.content_evaluator import ContentEvaluationAgent
from config import settings
from agents.preference_tools import _get_preferences_from_db, format_preferences_for_prompt
from services.push_service import push_to_channels

logger = logging.getLogger(__name__)

# LLM config cache TTL (seconds)
_LLM_CONFIG_CACHE_TTL = 60


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

        # ContentEvaluationAgent - 从 settings 读取配置（启动时初始化，之后热加载）
        # 优先用 DB 热加载的 llm_model，再 fallback 到环境变量 llm_model_id
        model_id = settings.llm_model or settings.llm_model_id
        api_key = settings.openai_api_key or os.getenv("OPENAI_API_KEY", "")
        api_base = settings.llm_base_url or os.getenv("LLM_BASE_URL", "")
        logger.info(f"Initializing ContentEvaluationAgent with model:{model_id} (Consumer: {self.consumer_name})")

        self.evaluator_agent = ContentEvaluationAgent(
            model=model_id,
            api_key=api_key,
            api_base=api_base,
            max_tokens=settings.llm_max_tokens,
        )
        self._base_system_prompt = self.evaluator_agent.system_prompt

        # LLM config hot-reload cache
        self._llm_config_last_check = 0


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

                            # Go 发布时将整个 JSON 放在 "data" 字段里，需要先解包
                            if "data" in data and len(data) == 1:
                                data = json.loads(data["data"])

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
                        # Drain previously-failed PENDING items first so they don't starve
                        # behind a flood of new messages after a consumer restart
                        await self._requeue_pending_content()

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
                # Defensive check: skip stale messages whose content_id no longer exists in DB
                exists = await self.db_pool.fetchval(
                    "SELECT EXISTS(SELECT 1 FROM content WHERE id = $1)", message.content_id
                )
                if not exists:
                    logger.warning(f"Skipping content {message.content_id}: not found in database (stale message)")
                    continue

                # Rate limit: pause between LLM calls (serial)
                if success_count + failure_count > 0:
                    await asyncio.sleep(settings.llm_request_interval)

                # Use ContentEvaluationAgent to evaluate
                result = await self._evaluate_with_agent(message)

                if result is None:
                    # LLM failed - increment eval_attempts, decide PENDING or DISCARDED
                    await self._handle_eval_failure(message)
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

                # Check if content meets notification threshold (user-configurable)
                if await self._should_notify(message, result):
                    await self._create_notification(message, result)

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
        Hot-reload LLM config from database with TTL cache.

        - Checks DB at most once per _LLM_CONFIG_CACHE_TTL seconds
        - Only recreates evaluator_agent when model/api_key/base_url actually changed
        """
        now = time.time()
        if now - self._llm_config_last_check < _LLM_CONFIG_CACHE_TTL:
            return  # cache still fresh, skip DB query

        self._llm_config_last_check = now

        try:
            from config import load_llm_config_from_db

            db_config = await load_llm_config_from_db(self.db_pool)

            if db_config and db_config.get('api_key'):
                new_model = db_config.get('model_name', self.evaluator_agent.model)
                new_api_key = db_config['api_key']
                new_api_base = db_config.get('base_url') or os.getenv("LLM_BASE_URL", "")

                # Only reinitialize when config actually changed
                if (new_model != self.evaluator_agent.model or
                    new_api_key != getattr(self.evaluator_agent, 'api_key', '') or
                    new_api_base != getattr(self.evaluator_agent, 'api_base', '')):

                    self.evaluator_agent = ContentEvaluationAgent(
                        model=new_model,
                        api_key=new_api_key,
                        api_base=new_api_base,
                        max_tokens=settings.llm_max_tokens,
                    )
                    self._base_system_prompt = self.evaluator_agent.system_prompt
                    logger.info(f"[Config] Hot-reloaded LLM config: model={new_model}, base_url={new_api_base}")
            else:
                logger.debug("[Config] No valid LLM config in database, keeping current settings")

        except Exception as e:
            logger.warning(f"[Config] Error reloading LLM config: {e}")

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

            # Load user preferences and inject into evaluator prompt
            await self._inject_preferences(message)

            # LangGraph's graph.invoke() is synchronous — run it in a thread pool
            # to avoid blocking the asyncio event loop during LLM calls
            result = await asyncio.to_thread(
                self.evaluator_agent.run,
                message.title,
                message.content,
                message.url,
            )

            return result

        except Exception as e:
            logger.error(f"Error in agent evaluation: {e}")
            return None

    async def _handle_eval_failure(self, message: StreamMessage):
        """LLM 评估失败时：递增 eval_attempts，超过阈值则标记 DISCARDED"""
        try:
            row = await self.db_pool.fetchrow(
                "UPDATE content SET eval_attempts = eval_attempts + 1 WHERE id = $1 RETURNING eval_attempts",
                message.content_id
            )
            attempts = row["eval_attempts"] if row else 1
            if attempts >= settings.llm_max_eval_attempts:
                await self.db_service.update_content_status(message.content_id, "DISCARDED")
                await self.db_service.log_status_change(
                    content_id=message.content_id,
                    task_id=message.task_id,
                    from_status="PENDING",
                    to_status="DISCARDED",
                    reason=f"LLM evaluation failed after {attempts} attempts",
                )
                logger.warning(f"[Eval] Content {message.content_id} DISCARDED after {attempts} failed attempts")
            else:
                logger.info(f"[Eval] Content {message.content_id} kept PENDING (attempt {attempts}/{settings.llm_max_eval_attempts})")
        except Exception as e:
            logger.error(f"[Eval] Error handling eval failure for {message.content_id}: {e}")

    async def _requeue_pending_content(self):
        """将之前 LLM 失败但未超限的 PENDING 内容重新推入 stream"""
        try:
            rows = await self.db_pool.fetch(
                """SELECT id, task_id, title, original_url, clean_content, published_at, platform, author_name, content_hash
                   FROM content
                   WHERE status = 'PENDING' AND eval_attempts > 0 AND eval_attempts < $1
                   ORDER BY created_at ASC LIMIT $2""",
                settings.llm_max_eval_attempts,
                self.batch_size,
            )
            for row in rows:
                msg_data = json.dumps({
                    "content_id": row["id"],
                    "task_id": str(row["task_id"]),
                    "title": row["title"] or "",
                    "url": row["original_url"] or "",
                    "content": row["clean_content"] or "",
                    "published_at": row["published_at"].isoformat() if row["published_at"] else "",
                    "platform": row["platform"] or "blog",
                    "author_name": row["author_name"] or "",
                    "content_hash": row["content_hash"] or "",
                }, ensure_ascii=False)
                await self.redis.xadd(self.stream_name, {"data": msg_data})
            if rows:
                logger.info(f"[Requeue] Re-queued {len(rows)} PENDING content items for retry")
        except Exception as e:
            logger.error(f"[Requeue] Error re-queuing pending content: {e}")

    async def _should_notify(self, message: StreamMessage, result) -> bool:
        """Check notification settings from DB to decide whether to notify"""
        try:
            row = await self.db_pool.fetchrow(
                "SELECT min_innovation_score, min_depth_score, notify_on_interesting, watched_source_ids, enabled FROM notification_settings WHERE id = 1"
            )
            if not row:
                # Default behavior if no settings
                return result.decision == "INTERESTING" or (result.innovation_score >= 8 and result.depth_score >= 7)

            if not row["enabled"]:
                return False

            # Check watched sources filter
            watched = row["watched_source_ids"]
            if watched and isinstance(watched, list) and len(watched) > 0:
                # Get source_id for this content
                content_row = await self.db_pool.fetchrow(
                    "SELECT source_id FROM content WHERE id = $1", message.content_id
                )
                if content_row and content_row["source_id"] not in watched:
                    return False

            # Check decision-based rule
            if row["notify_on_interesting"] and result.decision == "INTERESTING":
                return True

            # Check score thresholds
            min_innovation = row["min_innovation_score"]
            min_depth = row["min_depth_score"]
            if result.innovation_score >= min_innovation and result.depth_score >= min_depth:
                return True

            return False
        except Exception as e:
            logger.debug(f"[Notification] Could not read settings: {e}")
            # Fallback to default
            return result.decision == "INTERESTING" or (result.innovation_score >= 8 and result.depth_score >= 7)

    async def _create_notification(self, message: StreamMessage, result):
        """Create a notification for high-value evaluated content"""
        try:
            await self.db_pool.execute(
                """INSERT INTO notifications (content_id, title, summary, innovation_score, depth_score, decision)
                   VALUES ($1, $2, $3, $4, $5, $6)""",
                message.content_id,
                message.title,
                result.tldr,
                result.innovation_score,
                result.depth_score,
                result.decision,
            )
            # Publish notification event via Redis Pub/Sub
            notification_data = json.dumps({
                "content_id": message.content_id,
                "title": message.title,
                "summary": result.tldr,
                "innovation_score": result.innovation_score,
                "depth_score": result.depth_score,
                "decision": result.decision,
            }, ensure_ascii=False)
            await self.redis.publish("notifications", notification_data)
            logger.info(f"[Notification] Created for content {message.content_id}: {message.title[:50]}")

            # Push to external channels (PushPlus, WxPusher, etc.)
            await push_to_channels(
                self.db_pool,
                title=message.title,
                summary=result.tldr,
                innovation_score=result.innovation_score,
                depth_score=result.depth_score,
                decision=result.decision,
                url=getattr(message, 'url', ''),
            )
        except Exception as e:
            logger.warning(f"[Notification] Failed to create: {e}")

    async def _inject_preferences(self, message: StreamMessage):
        """Load user preferences for the source and inject into evaluator prompt"""
        try:
            # Get source_id from content table
            source_id = None
            if hasattr(message, 'content_id') and message.content_id:
                row = await self.db_pool.fetchrow(
                    "SELECT source_id FROM content WHERE id = $1",
                    message.content_id,
                )
                if row:
                    source_id = row["source_id"]

            # Load preferences (source-specific first, then global fallback)
            prefs = {}
            if source_id:
                prefs = await _get_preferences_from_db(source_id)
            if not prefs:
                prefs = await _get_preferences_from_db(None)  # global

            # Inject into evaluator's system_prompt
            pref_text = format_preferences_for_prompt(prefs)
            if pref_text:
                base_prompt = self._base_system_prompt
                self.evaluator_agent.system_prompt = base_prompt + pref_text
            else:
                # No preferences — restore base prompt
                self.evaluator_agent.system_prompt = self._base_system_prompt

        except Exception as e:
            logger.debug(f"[Preferences] Could not inject preferences: {e}")

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

