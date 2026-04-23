import os
from pydantic_settings import BaseSettings
from pydantic import ConfigDict
import asyncpg
import logging

logger = logging.getLogger(__name__)

# 手动加载 .env 文件到 os.environ
# 优先加载 backend-python/.env，然后加载项目根目录 .env（后者覆盖前者）
env_files = [
    os.path.join(os.path.dirname(__file__), ".env"),  # backend-python/.env
    os.path.join(os.path.dirname(__file__), "..", ".env"),  # 项目根目录 .env
    ".env"  # 当前目录 .env
]

# 强制覆盖关键的 LLM 配置变量（以 .env 文件为准）
for env_file in env_files:
    if os.path.exists(env_file):
        with open(env_file, "r", encoding="utf-8") as f:
            for line in f:
                line = line.strip()
                if line and not line.startswith("#") and "=" in line:
                    key, value = line.split("=", 1)
                    key = key.strip()
                    value = value.strip()
                    # 强制覆盖 LLM 相关的环境变量
                    if key in ["LLM_MODEL_ID", "OPENAI_API_KEY", "LLM_BASE_URL", "LLM_TEMPERATURE", "LLM_MAX_TOKENS"]:
                        os.environ[key] = value
                    elif key not in os.environ:
                        os.environ[key] = value

class Settings(BaseSettings):
    """应用配置"""

    model_config = ConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=False
    )

    # 数据库配置 (P0 优化)
    db_host: str = "localhost"
    db_port: int = 5432
    db_user: str = "junkfilter"
    db_password: str = "junkfilter123"
    db_name: str = "junkfilter"
    db_pool_min_size: int = 10
    db_pool_max_size: int = 100  # ← P0: 从 20 改为 100 (支持更多并发)

    # Redis 配置
    redis_url: str = "redis://localhost:6379/0"
    redis_pool_size: int = 10

    # 评估配置
    evaluation_timeout: int = 30
    max_retries: int = 3
    batch_size: int = 10
    llm_max_workers: int = 1  # 串行，不并发
    llm_max_eval_attempts: int = 3  # 每篇文章最多尝试 LLM 评估次数，超限标记 DISCARDED

    # LLM 配置 (OpenAI)
    llm_provider: str = "openai"
    llm_model: str = ""
    openai_api_key: str = ""
    llm_base_url: str = ""
    llm_model_id: str = ""
    llm_temperature: float = 0.7
    llm_max_tokens: int = 2000
    llm_timeout: int = 60
    llm_request_interval: float = 3.0  # seconds between LLM API calls (3s ≈ 20 RPM, safe for Gemma free tier)

    # FastAPI 服务配置
    api_host: str = "0.0.0.0"
    api_port: int = 8083
    api_workers: int = 1

    # 日志配置
    log_level: str = "INFO"


settings = Settings()


# ============ 从数据库加载 LLM 配置 ============

async def load_llm_config_from_db(pool: asyncpg.pool.Pool) -> dict:
    """
    从数据库加载最新的 LLM 配置

    Args:
        pool: asyncpg 连接池

    Returns:
        dict: LLM 配置 {model_name, api_key, base_url, temperature, max_tokens}
        如果数据库中没有配置，返回 None
    """
    try:
        async with pool.acquire() as conn:
            # 从 ai_config 表获取配置
            row = await conn.fetchrow("""
                SELECT
                    default_model as model_name,
                    api_key,
                    base_url,
                    temperature,
                    max_tokens
                FROM ai_config
                LIMIT 1
            """)

            if row and row['api_key'] and row['api_key'] != 'sk-placeholder':
                config = {
                    'model_name': row['model_name'],
                    'api_key': row['api_key'],
                    'base_url': row['base_url'] or '',
                    'temperature': row['temperature'],
                    'max_tokens': row['max_tokens'],
                }
                logger.info(f"[Config] Loaded LLM config from DB: {config['model_name']}")
                return config
            else:
                logger.warning("[Config] No valid LLM config found in database (placeholder or empty)")
                return None
    except Exception as e:
        logger.error(f"[Config] Error loading LLM config from DB: {e}")
        return None


async def initialize_llm_config(pool: asyncpg.pool.Pool):
    """
    初始化 LLM 配置

    流程：
    1. 尝试从数据库加载配置
    2. 优先使用环境变量中的 base_url（支持中转站）
    3. 如果没有配置，使用环境变量/默认值
    """
    db_config = await load_llm_config_from_db(pool)

    if db_config and db_config.get('api_key'):
        # 使用数据库配置作为主要来源；DB base_url 优先于 .env 里的默认值
        settings.llm_model = db_config['model_name']
        settings.openai_api_key = db_config['api_key']
        settings.llm_base_url = db_config.get('base_url') or os.environ.get('LLM_BASE_URL', '')
        settings.llm_temperature = db_config.get('temperature', 0.7)
        settings.llm_max_tokens = db_config.get('max_tokens', 2000)
        logger.info(f"[Config] Using LLM config from database: {db_config['model_name']}, base_url: {settings.llm_base_url}")
    else:
        # 使用环境变量
        logger.info("[Config] Using LLM config from environment variables")
        if not settings.openai_api_key:
            logger.warning("[Config] No API key configured")



