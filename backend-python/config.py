import os
from pydantic_settings import BaseSettings
from pydantic import ConfigDict

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

    # 数据库配置
    db_host: str = "localhost"
    db_port: int = 5432
    db_user: str = "truesignal"
    db_password: str = "truesignal123"
    db_name: str = "truesignal"
    db_pool_min_size: int = 5
    db_pool_max_size: int = 20

    # Redis 配置
    redis_url: str = "redis://localhost:6379/0"
    redis_pool_size: int = 10

    # 评估配置
    evaluation_timeout: int = 30
    max_retries: int = 3
    batch_size: int = 10

    # LLM 配置 (OpenAI)
    llm_provider: str = "openai"
    llm_model: str = ""
    openai_api_key: str = ""
    llm_base_url: str = ""
    llm_model_id: str = ""
    llm_temperature: float = 0.7
    llm_max_tokens: int = 2000
    llm_timeout: int = 30

    # FastAPI 服务配置
    api_host: str = "0.0.0.0"
    api_port: int = 8081
    api_workers: int = 4

    # 日志配置
    log_level: str = "INFO"


settings = Settings()


