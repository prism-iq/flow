import asyncio
from typing import Dict, List, Optional, Any
from dataclasses import dataclass, field
from enum import Enum

from ..workers import (
    BaseWorker, WorkerType,
    DateWorker, PersonWorker, OrganizationWorker,
    AmountWorker, GeneralWorker
)
from ..workers.base import ExtractionResult
from .classifier import QueryClassifier, QueryType, ClassificationResult
from ..utils.logger import logger


@dataclass
class RoutingResult:
    query: str
    classification: ClassificationResult
    results: Dict[WorkerType, ExtractionResult]
    combined_entities: List[Dict[str, Any]]
    total_time: float
    workers_used: List[str]


class QueryRouter:
    def __init__(self):
        self.classifier = QueryClassifier()
        self.workers: Dict[WorkerType, BaseWorker] = {}
        self._initialized = False

    async def initialize(self):
        logger.info("Initializing query router with specialized workers...")

        self.workers = {
            WorkerType.DATE: DateWorker(),
            WorkerType.PERSON: PersonWorker(),
            WorkerType.ORGANIZATION: OrganizationWorker(),
            WorkerType.AMOUNT: AmountWorker(),
            WorkerType.GENERAL: GeneralWorker(),
        }

        for worker_type, worker in self.workers.items():
            await worker.load_model()
            logger.info(f"Worker {worker_type.value} initialized")

        self._initialized = True
        logger.info("Query router ready")

    async def route(self, query: str, force_types: List[WorkerType] = None) -> RoutingResult:
        if not self._initialized:
            await self.initialize()

        import time
        start_time = time.time()

        classification = self.classifier.classify(query)

        if force_types:
            target_types = force_types
        elif classification.primary_type == QueryType.MULTI:
            target_types = [
                t for t, s in classification.all_types
                if s >= 0.2 and t != QueryType.GENERAL
            ]
            if not target_types:
                target_types = [WorkerType.GENERAL]
        else:
            type_mapping = {
                QueryType.DATE: WorkerType.DATE,
                QueryType.PERSON: WorkerType.PERSON,
                QueryType.ORGANIZATION: WorkerType.ORGANIZATION,
                QueryType.AMOUNT: WorkerType.AMOUNT,
                QueryType.GENERAL: WorkerType.GENERAL,
            }
            target_types = [type_mapping.get(classification.primary_type, WorkerType.GENERAL)]

        worker_types = []
        for t in target_types:
            if isinstance(t, QueryType):
                type_mapping = {
                    QueryType.DATE: WorkerType.DATE,
                    QueryType.PERSON: WorkerType.PERSON,
                    QueryType.ORGANIZATION: WorkerType.ORGANIZATION,
                    QueryType.AMOUNT: WorkerType.AMOUNT,
                    QueryType.GENERAL: WorkerType.GENERAL,
                }
                worker_types.append(type_mapping.get(t, WorkerType.GENERAL))
            else:
                worker_types.append(t)

        tasks = []
        for worker_type in worker_types:
            if worker_type in self.workers:
                tasks.append(self._run_worker(worker_type, query))

        results_list = await asyncio.gather(*tasks, return_exceptions=True)

        results: Dict[WorkerType, ExtractionResult] = {}
        for worker_type, result in zip(worker_types, results_list):
            if isinstance(result, Exception):
                logger.error(f"Worker {worker_type.value} failed: {result}")
            else:
                results[worker_type] = result

        combined = self._combine_entities(results)
        total_time = time.time() - start_time

        return RoutingResult(
            query=query,
            classification=classification,
            results=results,
            combined_entities=combined,
            total_time=total_time,
            workers_used=[wt.value for wt in worker_types]
        )

    async def _run_worker(self, worker_type: WorkerType, query: str) -> ExtractionResult:
        worker = self.workers[worker_type]
        return await worker.extract(query)

    def _combine_entities(self, results: Dict[WorkerType, ExtractionResult]) -> List[Dict[str, Any]]:
        combined = []

        for worker_type, result in results.items():
            for entity in result.entities:
                combined.append({
                    **entity,
                    '_source_worker': worker_type.value,
                    '_extraction_confidence': result.confidence,
                })

        return combined

    async def extract_all(self, query: str) -> RoutingResult:
        all_types = [
            WorkerType.DATE,
            WorkerType.PERSON,
            WorkerType.ORGANIZATION,
            WorkerType.AMOUNT,
        ]
        return await self.route(query, force_types=all_types)

    def classify(self, query: str) -> Dict[str, Any]:
        """Classify a query and return worker types to use."""
        result = self.classifier.classify(query)

        worker_types = []
        for qt, score in result.all_types:
            if score >= 0.2 and qt != QueryType.GENERAL:
                type_mapping = {
                    QueryType.DATE: "date",
                    QueryType.PERSON: "person",
                    QueryType.ORGANIZATION: "org",
                    QueryType.AMOUNT: "amount",
                }
                if qt in type_mapping:
                    worker_types.append(type_mapping[qt])

        if not worker_types:
            worker_types = ["date", "person", "org", "amount"]

        return {
            "primary_type": result.primary_type.name.lower(),
            "confidence": result.confidence,
            "worker_types": worker_types,
            "scores": {qt.name.lower(): score for qt, score in result.all_types}
        }

    def get_stats(self) -> Dict[str, Any]:
        return {
            'initialized': self._initialized,
            'workers': {
                wt.value: self.workers[wt].stats
                for wt in self.workers
            }
        }
