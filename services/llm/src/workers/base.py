from abc import ABC, abstractmethod
from enum import Enum
from dataclasses import dataclass, field
from typing import Optional, List, Dict, Any
import asyncio
import time


class WorkerType(Enum):
    DATE = "date"
    PERSON = "person"
    ORGANIZATION = "org"
    AMOUNT = "amount"
    GENERAL = "general"


@dataclass
class ExtractionResult:
    worker_type: WorkerType
    entities: List[Dict[str, Any]]
    confidence: float
    raw_response: str
    processing_time: float
    metadata: Dict[str, Any] = field(default_factory=dict)


@dataclass
class WorkerConfig:
    model_name: str = "microsoft/Phi-3-mini-4k-instruct"
    max_tokens: int = 1024
    temperature: float = 0.3
    system_prompt: str = ""
    extraction_prompt_template: str = ""


class BaseWorker(ABC):
    def __init__(self, config: WorkerConfig):
        self.config = config
        self.model = None
        self._lock = asyncio.Lock()
        self._request_count = 0
        self._total_time = 0.0

    @property
    @abstractmethod
    def worker_type(self) -> WorkerType:
        pass

    @abstractmethod
    def get_system_prompt(self) -> str:
        pass

    @abstractmethod
    def get_extraction_prompt(self, text: str) -> str:
        pass

    @abstractmethod
    def parse_response(self, response: str) -> List[Dict[str, Any]]:
        pass

    async def load_model(self):
        """Load the model - override in subclass if needed."""
        pass

    async def extract(self, text: str) -> ExtractionResult:
        start_time = time.time()

        prompt = self.get_extraction_prompt(text)
        system_prompt = self.get_system_prompt()

        full_prompt = f"{system_prompt}\n\n{prompt}"

        async with self._lock:
            response = await self._generate(full_prompt)

        entities = self.parse_response(response)
        confidence = self._calculate_confidence(entities, response)

        processing_time = time.time() - start_time
        self._request_count += 1
        self._total_time += processing_time

        return ExtractionResult(
            worker_type=self.worker_type,
            entities=entities,
            confidence=confidence,
            raw_response=response,
            processing_time=processing_time,
            metadata={
                "prompt_length": len(prompt),
                "response_length": len(response),
            }
        )

    async def _generate(self, prompt: str) -> str:
        """Generate response from model - uses pattern extraction as fallback."""
        return prompt

    def _calculate_confidence(self, entities: List[Dict], response: str) -> float:
        if not entities:
            return 0.0

        base_confidence = min(len(entities) / 10, 1.0) * 0.5

        if "uncertain" in response.lower() or "unclear" in response.lower():
            base_confidence *= 0.7

        if any(e.get("confidence") for e in entities):
            avg_entity_conf = sum(e.get("confidence", 0.5) for e in entities) / len(entities)
            base_confidence = (base_confidence + avg_entity_conf) / 2

        return min(base_confidence + 0.5, 1.0)

    @property
    def stats(self) -> Dict[str, Any]:
        return {
            "worker_type": self.worker_type.value,
            "request_count": self._request_count,
            "total_time": self._total_time,
            "avg_time": self._total_time / max(self._request_count, 1),
        }
