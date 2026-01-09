from fastapi import APIRouter, HTTPException
from pydantic import BaseModel
from typing import List, Dict, Any, Optional
import asyncio

from .workers.date_worker import DateWorker
from .workers.person_worker import PersonWorker
from .workers.org_worker import OrganizationWorker as OrgWorker
from .workers.amount_worker import AmountWorker
from .workers.base import WorkerType
from .router.query_router import QueryRouter
from .aggregator.haiku_aggregator import HaikuAggregator
from .aggregator.entity_merger import EntityMerger
from .utils.logger import logger


router = APIRouter(prefix="/extract")


class ExtractionRequest(BaseModel):
    text: str


class Entity(BaseModel):
    type: str
    value: Any
    confidence: float
    source: str
    metadata: Optional[Dict[str, Any]] = None


class ExtractionResponse(BaseModel):
    entities: List[Dict[str, Any]]


workers = {}
query_router = None
haiku_aggregator = None
entity_merger = None


def get_worker(worker_type: str):
    global workers
    if worker_type not in workers:
        if worker_type == "date":
            workers[worker_type] = DateWorker()
        elif worker_type == "person":
            workers[worker_type] = PersonWorker()
        elif worker_type == "org":
            workers[worker_type] = OrgWorker()
        elif worker_type == "amount":
            workers[worker_type] = AmountWorker()
        else:
            return None
    return workers[worker_type]


def get_router():
    global query_router
    if query_router is None:
        query_router = QueryRouter()
    return query_router


def get_aggregator():
    global haiku_aggregator
    if haiku_aggregator is None:
        haiku_aggregator = HaikuAggregator()
    return haiku_aggregator


def get_merger():
    global entity_merger
    if entity_merger is None:
        entity_merger = EntityMerger()
    return entity_merger


@router.post("/date", response_model=ExtractionResponse)
async def extract_date(body: ExtractionRequest):
    worker = get_worker("date")
    if not worker:
        raise HTTPException(status_code=500, detail="Date worker not available")

    try:
        result = await worker.extract(body.text)
        entities = [
            {
                "type": "date",
                "value": e.get("date_string") or e.get("value"),
                "confidence": e.get("confidence", 0.8),
                "source": "date",
                "metadata": {
                    "format": e.get("date_type"),
                    "normalized": e.get("normalized_date"),
                    "context": e.get("context")
                }
            }
            for e in result.entities
        ]
        return ExtractionResponse(entities=entities)
    except Exception as e:
        logger.error(f"Date extraction error: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@router.post("/person", response_model=ExtractionResponse)
async def extract_person(body: ExtractionRequest):
    worker = get_worker("person")
    if not worker:
        raise HTTPException(status_code=500, detail="Person worker not available")

    try:
        result = await worker.extract(body.text)
        entities = [
            {
                "type": "person",
                "value": e.get("name"),
                "confidence": e.get("confidence", 0.8),
                "source": "person",
                "metadata": {
                    "email": e.get("email"),
                    "title": e.get("title"),
                    "role": e.get("role"),
                    "organization": e.get("organization"),
                    "relationships": e.get("relationships", [])
                }
            }
            for e in result.entities
        ]
        return ExtractionResponse(entities=entities)
    except Exception as e:
        logger.error(f"Person extraction error: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@router.post("/org", response_model=ExtractionResponse)
async def extract_org(body: ExtractionRequest):
    worker = get_worker("org")
    if not worker:
        raise HTTPException(status_code=500, detail="Org worker not available")

    try:
        result = await worker.extract(body.text)
        entities = [
            {
                "type": "organization",
                "value": e.get("name"),
                "confidence": e.get("confidence", 0.8),
                "source": "org",
                "metadata": {
                    "type": e.get("type"),
                    "industry": e.get("industry"),
                    "aliases": e.get("aliases", []),
                    "relationships": e.get("relationships", [])
                }
            }
            for e in result.entities
        ]
        return ExtractionResponse(entities=entities)
    except Exception as e:
        logger.error(f"Org extraction error: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@router.post("/amount", response_model=ExtractionResponse)
async def extract_amount(body: ExtractionRequest):
    worker = get_worker("amount")
    if not worker:
        raise HTTPException(status_code=500, detail="Amount worker not available")

    try:
        result = await worker.extract(body.text)
        entities = [
            {
                "type": "amount",
                "value": e.get("value") or e.get("amount"),
                "confidence": e.get("confidence", 0.8),
                "source": "amount",
                "metadata": {
                    "currency": e.get("currency"),
                    "normalized": e.get("normalized"),
                    "context": e.get("context"),
                    "unit": e.get("unit")
                }
            }
            for e in result.entities
        ]
        return ExtractionResponse(entities=entities)
    except Exception as e:
        logger.error(f"Amount extraction error: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@router.post("/all", response_model=ExtractionResponse)
async def extract_all(body: ExtractionRequest):
    """Run all workers in parallel and merge results"""
    try:
        tasks = [
            extract_date(body),
            extract_person(body),
            extract_org(body),
            extract_amount(body)
        ]

        results = await asyncio.gather(*tasks, return_exceptions=True)

        all_entities = []
        for result in results:
            if isinstance(result, Exception):
                logger.warning(f"Worker failed: {result}")
                continue
            all_entities.extend(result.entities)

        merger = get_merger()
        merged = merger.merge_entities(all_entities)

        return ExtractionResponse(entities=merged)
    except Exception as e:
        logger.error(f"Full extraction error: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@router.post("/smart")
async def extract_smart(body: ExtractionRequest):
    """Use router to classify and dispatch to relevant workers only"""
    try:
        qr = get_router()
        classification = qr.classify(body.text)

        worker_types = classification.get("worker_types", ["date", "person", "org", "amount"])

        tasks = []
        for wt in worker_types:
            if wt == "date":
                tasks.append(("date", extract_date(body)))
            elif wt == "person":
                tasks.append(("person", extract_person(body)))
            elif wt == "org":
                tasks.append(("org", extract_org(body)))
            elif wt == "amount":
                tasks.append(("amount", extract_amount(body)))

        results = await asyncio.gather(*[t[1] for t in tasks], return_exceptions=True)

        all_entities = []
        workers_used = []
        for i, result in enumerate(results):
            if isinstance(result, Exception):
                logger.warning(f"Worker {tasks[i][0]} failed: {result}")
                continue
            workers_used.append(tasks[i][0])
            all_entities.extend(result.entities)

        merger = get_merger()
        merged = merger.merge_entities(all_entities)

        aggregator = get_aggregator()
        validated = await aggregator.validate(merged, body.text)

        return {
            "entities": validated,
            "classification": classification,
            "workers_used": workers_used
        }
    except Exception as e:
        logger.error(f"Smart extraction error: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@router.post("/classify")
async def classify_query(body: ExtractionRequest):
    """Classify query without extraction"""
    try:
        qr = get_router()
        return qr.classify(body.text)
    except Exception as e:
        logger.error(f"Classification error: {e}")
        raise HTTPException(status_code=500, detail=str(e))
