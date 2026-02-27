import asyncio
import logging
from typing import Optional
from models.evaluation import StreamMessage, EvaluationRequest
from utils.llm_client import LLMClient
from services.db_service import DBService

logger = logging.getLogger(__name__)


class EvaluatorService:
    """Service for evaluating content"""

    def __init__(self, llm_client: LLMClient, db_service: DBService):
        self.llm_client = llm_client
        self.db_service = db_service

    async def evaluate_content(
        self,
        message: StreamMessage,
    ) -> bool:
        """
        Evaluate a single content item and store the result.

        Returns True if successful, False otherwise.
        """
        try:
            # Update content status to PROCESSING
            await self.db_service.update_content_status(
                message.content_id,
                "PROCESSING",
            )

            # Call LLM for evaluation
            result = await self.llm_client.evaluate(
                title=message.title,
                content=message.content,
                url=message.url,
            )

            # Store evaluation result
            eval_id = await self.db_service.create_evaluation(
                content_id=message.content_id,
                task_id=message.task_id,
                result=result,
            )

            if eval_id is None:
                logger.warning(f"Could not create evaluation for {message.content_id}")
                return False

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
            return True

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
                    from_status="PROCESSING",
                    to_status="DISCARDED",
                    reason=f"Evaluation error: {str(e)}",
                )
            except Exception as log_e:
                logger.error(f"Error logging status change: {log_e}")
            return False

    async def evaluate_batch(
        self,
        messages: list[StreamMessage],
    ) -> tuple[int, int]:
        """
        Evaluate multiple items concurrently.

        Returns (success_count, failure_count)
        """
        tasks = [self.evaluate_content(msg) for msg in messages]
        results = await asyncio.gather(*tasks, return_exceptions=True)

        success_count = sum(1 for r in results if r is True)
        failure_count = len(results) - success_count

        return success_count, failure_count
