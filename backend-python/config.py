import os
from pydantic_settings import BaseSettings

class Settings(BaseSettings):
    """应用配置"""

    # 数据库配置
    db_host: str = "localhost"
    db_port: int = 5432
    db_user: str = "junkfilter"
    db_password: str = "junkfilter123"
    db_name: str = "junkfilter"
    db_pool_min_size: int = 5
    db_pool_max_size: int = 20

    # Redis 配置
    redis_url: str = "redis://localhost:6379/0"
    redis_pool_size: int = 10

    # 评估配置
    evaluation_timeout: int = 30
    max_retries: int = 3
    batch_size: int = 10

    # 日志配置
    log_level: str = "INFO"

    class Config:
        env_file = ".env"
        env_file_encoding = "utf-8"

settings = Settings()
