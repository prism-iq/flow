import asyncio
from typing import List, Optional, AsyncGenerator
from dataclasses import dataclass
from enum import Enum
import time

from ..models.phi3 import Phi3Model
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


class PipelineManager:
    def __init__(self, settings):
        self.settings = settings
        self.model_name = settings.model_name
        self.num_workers = settings.num_workers

        self.models: List[Phi3Model] = []
        self.task_queue: asyncio.Queue = asyncio.Queue()
        self.workers: List[asyncio.Task] = []
        self._task_counter = 0
        self._running = False

    async def initialize(self):
        logger.info(f"Initializing {self.num_workers} model workers")

        for i in range(self.num_workers):
            model = Phi3Model(
                model_name=self.model_name,
                cache_dir=self.settings.model_cache_dir,
                load_in_8bit=self.settings.load_in_8bit,
                load_in_4bit=self.settings.load_in_4bit,
                device_map=self.settings.device_map
            )
            await model.load()
            self.models.append(model)
            logger.info(f"Worker {i} initialized")

        self._running = True
        for i in range(self.num_workers):
            worker = asyncio.create_task(self._worker_loop(i))
            self.workers.append(worker)

    async def shutdown(self):
        self._running = False
        for worker in self.workers:
            worker.cancel()
        await asyncio.gather(*self.workers, return_exceptions=True)
        logger.info("Pipeline shutdown complete")

    async def _worker_loop(self, worker_id: int):
        model = self.models[worker_id]

        while self._running:
            try:
                task = await asyncio.wait_for(
                    self.task_queue.get(),
                    timeout=1.0
                )
            except asyncio.TimeoutError:
                continue
            except asyncio.CancelledError:
                break

            task.status = TaskStatus.PROCESSING

            try:
                result = await model.generate(
                    prompt=task.prompt,
                    max_tokens=task.max_tokens,
                    temperature=task.temperature,
                    top_p=task.top_p,
                    stop_sequences=task.stop_sequences
                )
                task.result = result
                task.status = TaskStatus.COMPLETED
            except Exception as e:
                logger.error(f"Worker {worker_id} error: {e}")
                task.error = str(e)
                task.status = TaskStatus.FAILED

            self.task_queue.task_done()

    async def generate(
        self,
        prompt: str,
        max_tokens: int = 2048,
        temperature: float = 0.7,
        top_p: float = 0.9,
        stop_sequences: Optional[List[str]] = None
    ) -> dict:
        self._task_counter += 1
        task = GenerationTask(
            id=f"task_{self._task_counter}",
            prompt=prompt,
            max_tokens=max_tokens,
            temperature=temperature,
            top_p=top_p,
            stop_sequences=stop_sequences
        )

        await self.task_queue.put(task)

        while task.status in (TaskStatus.PENDING, TaskStatus.PROCESSING):
            await asyncio.sleep(0.01)

        if task.status == TaskStatus.FAILED:
            raise Exception(task.error)

        return task.result

    async def generate_stream(
        self,
        prompt: str,
        max_tokens: int = 2048,
        temperature: float = 0.7,
        top_p: float = 0.9,
        stop_sequences: Optional[List[str]] = None
    ) -> AsyncGenerator[str, None]:
        # Use first available model for streaming
        model = self.models[0]

        async for token in model.generate_stream(
            prompt=prompt,
            max_tokens=max_tokens,
            temperature=temperature,
            top_p=top_p,
            stop_sequences=stop_sequences
        ):
            yield token

    async def generate_batch(
        self,
        prompts: List[str],
        max_tokens: int = 2048,
        temperature: float = 0.7
    ) -> List[dict]:
        tasks = []
        for prompt in prompts:
            task = self.generate(
                prompt=prompt,
                max_tokens=max_tokens,
                temperature=temperature
            )
            tasks.append(task)

        results = await asyncio.gather(*tasks, return_exceptions=True)

        return [
            r if not isinstance(r, Exception) else {"error": str(r)}
            for r in results
        ]

    @property
    def queue_size(self) -> int:
        return self.task_queue.qsize()
