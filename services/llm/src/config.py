from pydantic_settings import BaseSettings
from typing import List, Optional
import os


class Settings(BaseSettings):
    host: str = "0.0.0.0"
    port: int = 8001

    # Model settings
    model_name: str = "microsoft/Phi-3-mini-4k-instruct"
    model_cache_dir: str = "./model_cache"

    # Parallelization
    num_workers: int = 4
    max_batch_size: int = 8

    # Generation defaults
    default_max_tokens: int = 2048
    default_temperature: float = 0.7
    default_top_p: float = 0.9

    # Memory
    load_in_8bit: bool = True
    load_in_4bit: bool = False
    device_map: str = "auto"

    # Haiku validation
    haiku_api_key: Optional[str] = None
    haiku_validation_enabled: bool = False

    class Config:
        env_prefix = "LLM_"


settings = Settings()
