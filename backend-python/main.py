import asyncio
import asyncpg
import redis.asyncio as aioredis
import logging
import signal
import sys
import os
import uvicorn
from threading import Thread

from config import settings
from services.stream_consumer import StreamConsumer

# 配置日志
logging.basicConfig(level=settings.log_level)
logger = logging.getLogger(__name__)


class Database:
    """数据库连接池"""
    _pool: asyncpg.pool.Pool = None

    @classmethod
    async def initialize(cls):
        """初始化数据库连接池"""
        dsn = f"postgresql://{settings.db_user}:{settings.db_password}@{settings.db_host}:{settings.db_port}/{settings.db_name}"
        cls._pool = await asyncpg.create_pool(
            dsn,
            min_size=settings.db_pool_min_size,
            max_size=settings.db_pool_max_size,
        )
        logger.info(f"✓ Database connected: {settings.db_host}:{settings.db_port}/{settings.db_name}")

    @classmethod
    async def close(cls):
        """关闭连接池"""
        if cls._pool:
            await cls._pool.close()
            logger.info("✓ Database connection closed")

    @classmethod
    def get_pool(cls) -> asyncpg.pool.Pool:
        """获取连接池"""
        return cls._pool


class Redis:
    """Redis 客户端"""
    _client: aioredis.Redis = None

    @classmethod
    async def initialize(cls):
        """初始化 Redis 客户端"""
        cls._client = await aioredis.from_url(
            settings.redis_url,
            encoding="utf8",
            decode_responses=True,
        )
        # 测试连接
        await cls._client.ping()
        logger.info(f"✓ Redis connected: {settings.redis_url}")

    @classmethod
    async def close(cls):
        """关闭 Redis 连接"""
        if cls._client:
            await cls._client.close()
            logger.info("✓ Redis connection closed")

    @classmethod
    def get_client(cls) -> aioredis.Redis:
        """获取 Redis 客户端"""
        return cls._client


class Application:
    """Main application container"""

    def __init__(self):
        self.consumer = None
        self.consumer_task = None
        self.fastapi_app = None
        self.uvicorn_server = None

    async def initialize(self):
        """Initialize all services"""
        logger.info("\n========== Junk Filter Python Evaluator ==========")

        # Initialize database
        await Database.initialize()

        # Initialize Redis
        await Redis.initialize()

        # Initialize consumer
        self.consumer = StreamConsumer(
            redis_client=Redis.get_client(),
            db_pool=Database.get_pool(),
            batch_size=settings.batch_size,
        )
        await self.consumer.initialize()

        # Import FastAPI app
        from api_server import app
        self.fastapi_app = app

        logger.info("================================================\n")

    async def run(self):
        """Run the application"""
        try:
            await self.initialize()

            # Start consumer in background
            self.consumer_task = asyncio.create_task(self.consumer.run())

            logger.info("Application running. Press Ctrl+C to stop.")
            logger.info(f"FastAPI server will start on http://0.0.0.0:{settings.api_port}")

            # Start FastAPI server in a separate thread
            def run_fastapi():
                uvicorn.run(
                    self.fastapi_app,
                    host="0.0.0.0",
                    port=settings.api_port,
                    log_level="info"
                )

            fastapi_thread = Thread(target=run_fastapi, daemon=True)
            fastapi_thread.start()

            # Wait for consumer task (keep the app running)
            await self.consumer_task

        except Exception as e:
            logger.error(f"Application error: {e}", exc_info=True)
            raise

    async def shutdown(self):
        """Graceful shutdown"""
        logger.info("Shutting down application...")

        if self.consumer_task and not self.consumer_task.done():
            self.consumer_task.cancel()
            try:
                await self.consumer_task
            except asyncio.CancelledError:
                pass

        if self.consumer:
            await self.consumer.stop()

        await Database.close()
        await Redis.close()

        logger.info("Application shutdown complete.")


async def main():
    """Main entry point"""
    app = Application()

    # Setup signal handlers for graceful shutdown (Windows compatible)
    loop = asyncio.get_event_loop()

    def signal_handler():
        logger.info("Signal received, initiating shutdown...")
        asyncio.create_task(app.shutdown())

    # Windows doesn't support add_signal_handler for all signals
    # Only set up signal handlers on Unix-like systems
    if sys.platform != "win32":
        try:
            loop.add_signal_handler(signal.SIGINT, signal_handler)
            loop.add_signal_handler(signal.SIGTERM, signal_handler)
        except NotImplementedError:
            pass

    try:
        await app.run()
    except KeyboardInterrupt:
        logger.info("Keyboard interrupt received")
    finally:
        await app.shutdown()


if __name__ == "__main__":
    asyncio.run(main())
