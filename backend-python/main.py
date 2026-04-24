import asyncio
import asyncpg
import redis.asyncio as aioredis
import logging
import signal
import sys
import os

from config import settings
from services.stream_consumer import StreamConsumer

# 配置日志
logging.basicConfig(level=settings.log_level)
logger = logging.getLogger(__name__)


class Database:
    """
    数据库连接池（单例模式）。

    使用类变量 _pool + 类方法，确保全局只有一个连接池实例。
    优点：避免多处创建连接池导致连接数爆炸；asyncpg 连接池自带复用和回收。
    """
    _pool: asyncpg.pool.Pool = None

    @classmethod
    async def initialize(cls):
        """初始化连接池。须在 Application.initialize() 中调用一次。"""
        dsn = f"postgresql://{settings.db_user}:{settings.db_password}@{settings.db_host}:{settings.db_port}/{settings.db_name}"
        cls._pool = await asyncpg.create_pool(
            dsn,
            min_size=settings.db_pool_min_size,
            max_size=settings.db_pool_max_size,
        )
        logger.info(f"✓ Database connected: {settings.db_host}:{settings.db_port}/{settings.db_name}")

    @classmethod
    async def close(cls):
        """关闭连接池，释放所有连接。须在 shutdown() 中调用。"""
        if cls._pool:
            await cls._pool.close()
            logger.info("✓ Database connection closed")

    @classmethod
    def get_pool(cls) -> asyncpg.pool.Pool:
        """获取连接池引用，供 stream_consumer 和 db_service 使用。"""
        return cls._pool


class Redis:
    """
    Redis 客户端（单例模式）。

    decode_responses=True：Redis 返回的 bytes 自动转为 str，
    省去下游 everywhere 的 .decode("utf-8") 调用。
    """
    _client: aioredis.Redis = None

    @classmethod
    async def initialize(cls):
        """初始化 Redis 客户端并验证连通性。"""
        cls._client = await aioredis.from_url(
            settings.redis_url,
            encoding="utf8",
            decode_responses=True,
        )
        await cls._client.ping()
        logger.info(f"✓ Redis connected: {settings.redis_url}")

    @classmethod
    async def close(cls):
        """关闭 Redis 连接。"""
        if cls._client:
            await cls._client.close()
            logger.info("✓ Redis connection closed")

    @classmethod
    def get_client(cls) -> aioredis.Redis:
        """获取 Redis 客户端引用。"""
        return cls._client


class Application:
    """
    应用生命周期管理器。

    三个阶段：
      initialize() → 建 DB / Redis / Consumer 连接
      run()        → 启动 Consumer 协程，进入事件循环
      shutdown()   → 收到信号后优雅关闭（带超时保护，防止 hang 死）
    """

    def __init__(self):
        self.consumer = None
        self.consumer_task = None

    async def initialize(self):
        """初始化所有依赖服务（DB、Redis、Consumer）。"""
        logger.info("\n========== Junk Filter Python Evaluator ==========")

        await Database.initialize()
        await Redis.initialize()

        # 从数据库加载 LLM 配置（供 evaluator 热加载使用）
        from config import initialize_llm_config
        await initialize_llm_config(Database.get_pool())

        self.consumer = StreamConsumer(
            redis_client=Redis.get_client(),
            db_pool=Database.get_pool(),
            batch_size=settings.batch_size,
        )
        await self.consumer.initialize()

        logger.info("================================================\n")

    async def run(self):
        """
        启动主循环。

        create_task 将 Consumer 的 run() 放入事件循环，不阻塞当前协程。
        随后 await self.consumer_task 挂起，保持进程存活直到收到取消信号。
        """
        try:
            await self.initialize()

            self.consumer_task = asyncio.create_task(self.consumer.run())

            logger.info("Stream consumer running. Press Ctrl+C to stop.")

            await self.consumer_task

        except Exception as e:
            logger.error(f"Application error: {e}", exc_info=True)
            raise

    async def shutdown(self):
        """
        优雅关闭（带多级超时保护）。

        关闭顺序：
          1. 取消 Consumer 协程（不再从 Stream 读取新消息）
          2. 给进行中的评估留出一点时间
          3. 关闭 DB 连接池
          4. 关闭 Redis

        每步都有独立超时，避免某一步 hang 死导致整个关闭流程卡住。
        """
        logger.info("Shutting down application...")

        # 1. 取消 Consumer 协程，等待它自然退出（最多 10 秒）
        if self.consumer_task and not self.consumer_task.done():
            self.consumer_task.cancel()
            try:
                await asyncio.wait_for(self.consumer_task, timeout=10)
            except asyncio.TimeoutError:
                logger.warning("Consumer shutdown timeout (10s)")
            except asyncio.CancelledError:
                pass

        # 2. 短暂等待让进行中的评估有机会完成 DB 写入
        try:
            await asyncio.sleep(0.5)
        except:
            pass

        # 3. 通知 Consumer 做清理（如释放资源）
        if self.consumer:
            await self.consumer.stop()

        # 4. 关闭数据库连接池（最多 5 秒，防止 asyncpg 连接 hang 住）
        try:
            await asyncio.wait_for(Database.close(), timeout=5)
        except asyncio.TimeoutError:
            logger.error("Database pool close timeout (5s)")

        # 4. 关闭 Redis（最多 3 秒）
        try:
            await asyncio.wait_for(Redis.close(), timeout=3)
        except asyncio.TimeoutError:
            logger.error("Redis close timeout (3s)")

        logger.info("Application shutdown complete.")


async def main():
    """
    入口函数。

    注册 SIGINT/SIGTERM 信号处理器，捕获 Ctrl+C 或 kill 信号后触发 shutdown()。
    Windows 不支持 add_signal_handler，因此做了平台兼容处理。
    """
    app = Application()

    loop = asyncio.get_event_loop()

    def signal_handler():
        logger.info("Signal received, initiating shutdown...")
        asyncio.create_task(app.shutdown())

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
    # asyncio.run() 创建新的事件循环，运行 main() 协程，结束后自动关闭循环
    # 比直接 get_event_loop().run_until_complete() 更简洁，且自动处理循环生命周期
    asyncio.run(main())
