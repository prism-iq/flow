import asyncio
from typing import List, Optional, AsyncGenerator
from dataclasses import dataclass
from enum import Enum
import time

from ..utils.logger import logger


class TaskStatus(Enum):
    PENDING = "pending"
    PROCESSING = "processing"
    COMPLETED = "completed"
    FAILED = "failed"


@dataclass
class GenerationTask:
    id: str
    prompt: str
    max_tokens: int
    temperature: float
    top_p: float
    stop_sequences: Optional[List[str]]
    status: TaskStatus = TaskStatus.PENDING
    result: Optional[dict] = None
    error: Optional[str] = None
    created_at: float = None

    def __post_init__(self):
        if self.created_at is None:
            self.created_at = time.time()


class MockPipelineManager:
    """Mock pipeline manager for testing without torch/transformers."""

    def __init__(self, settings):
        self.settings = settings
        self.model_name = "mock-phi3"
        self.num_workers = 1
        self._task_counter = 0

    async def initialize(self):
        logger.info("Initializing mock pipeline manager (no GPU/torch)")

    async def shutdown(self):
        logger.info("Mock pipeline shutdown complete")

    async def generate(
        self,
        prompt: str,
        max_tokens: int = 2048,
        temperature: float = 0.7,
        top_p: float = 0.9,
        stop_sequences: Optional[List[str]] = None
    ) -> dict:
        self._task_counter += 1

        await asyncio.sleep(0.01)

        response_text = f"[Mock response for prompt length {len(prompt)}]"

        return {
            "text": response_text,
            "token_count": len(response_text.split()),
            "model": self.model_name,
            "finish_reason": "stop"
        }

    async def generate_stream(
        self,
        prompt: str,
        max_tokens: int = 2048,
        temperature: float = 0.7,
        top_p: float = 0.9,
        stop_sequences: Optional[List[str]] = None
    ) -> AsyncGenerator[str, None]:
        response = f"[Mock streaming response for prompt length {len(prompt)}]"
        for word in response.split():
            await asyncio.sleep(0.05)
            yield word + " "

    async def generate_batch(
        self,
        prompts: List[str],
        max_tokens: int = 2048,
        temperature: float = 0.7
    ) -> List[dict]:
        results = []
        for prompt in prompts:
            result = await self.generate(
                prompt=prompt,
                max_tokens=max_tokens,
                temperature=temperature
            )
            results.append(result)
        return results

    @property
    def queue_size(self) -> int:
        return 0
