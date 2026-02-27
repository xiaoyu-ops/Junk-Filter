import asyncpg
import logging
from datetime import datetime
from typing import Optional
from models.evaluation import EvaluationResult

logger = logging.getLogger(__name__)


class DBService:
    """Database service for storing evaluation results"""

    def __init__(self, pool: asyncpg.pool.Pool):
        self.pool = pool

    async def create_evaluation(
        self,
        content_id: int,
        task_id: str,
        result: EvaluationResult,
    ) -> Optional[int]:
        """Create an evaluation record"""
        async with self.pool.acquire() as conn:
            try:
                eval_id = await conn.fetchval(
                    """
                    INSERT INTO evaluation (
                        content_id, task_id, innovation_score, depth_score,
                        decision, reasoning, tldr, key_concepts, evaluated_at,
                        evaluator_version, created_at, updated_at
                    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
                    RETURNING id
                    """,
                    content_id,
                    task_id,
                    result.innovation_score,
                    result.depth_score,
                    result.decision,
                    result.reasoning,
                    result.tldr,
                    result.key_concepts,
                    datetime.utcnow(),
                    result.evaluator_version,
                    datetime.utcnow(),
                    datetime.utcnow(),
                )
                return eval_id
            except asyncpg.UniqueViolationError:
                logger.warning(f"Evaluation already exists for content_id {content_id}")
                return None
            except Exception as e:
                logger.error(f"Error creating evaluation: {e}")
                raise

    async def update_content_status(
        self,
        content_id: int,
        status: str,
    ) -> bool:
        """Update content status"""
        async with self.pool.acquire() as conn:
            try:
                result = await conn.execute(
                    """
                    UPDATE content SET status = $1, updated_at = $2
                    WHERE id = $3
                    """,
                    status,
                    datetime.utcnow(),
                    content_id,
                )
                return result == "UPDATE 1"
            except Exception as e:
                logger.error(f"Error updating content status: {e}")
                raise

    async def update_content_status_by_task_id(
        self,
        task_id: str,
        status: str,
    ) -> bool:
        """Update content status by task ID"""
        async with self.pool.acquire() as conn:
            try:
                result = await conn.execute(
                    """
                    UPDATE content SET status = $1, updated_at = $2
                    WHERE task_id = $3
                    """,
                    status,
                    datetime.utcnow(),
                    task_id,
                )
                return result == "UPDATE 1"
            except Exception as e:
                logger.error(f"Error updating content status by task_id: {e}")
                raise

    async def log_status_change(
        self,
        content_id: int,
        task_id: str,
        from_status: str,
        to_status: str,
        reason: Optional[str] = None,
    ) -> bool:
        """Log status change in status_log table"""
        async with self.pool.acquire() as conn:
            try:
                await conn.execute(
                    """
                    INSERT INTO status_log (content_id, task_id, from_status, to_status, reason, logged_at)
                    VALUES ($1, $2, $3, $4, $5, $6)
                    """,
                    content_id,
                    task_id,
                    from_status,
                    to_status,
                    reason,
                    datetime.utcnow(),
                )
                return True
            except Exception as e:
                logger.error(f"Error logging status change: {e}")
                raise
