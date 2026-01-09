from fastapi import APIRouter, Request, HTTPException
from fastapi.responses import StreamingResponse
from pydantic import BaseModel, Field
from typing import Optional, List
import json
import asyncio

from .utils.logger import logger


router = APIRouter()


class GenerateRequest(BaseModel):
    prompt: str
    max_tokens: int = Field(default=2048, ge=1, le=8192)
    temperature: float = Field(default=0.7, ge=0.0, le=2.0)
    top_p: float = Field(default=0.9, ge=0.0, le=1.0)
    stream: bool = False
    stop_sequences: Optional[List[str]] = None


class GenerateResponse(BaseModel):
    text: str
    token_count: int
    model: str
    finish_reason: str


class BatchRequest(BaseModel):
    prompts: List[str]
    max_tokens: int = 2048
    temperature: float = 0.7


@router.post("/generate", response_model=GenerateResponse)
async def generate(request: Request, body: GenerateRequest):
    pipeline = request.app.state.pipeline

    try:
        result = await pipeline.generate(
            prompt=body.prompt,
            max_tokens=body.max_tokens,
            temperature=body.temperature,
            top_p=body.top_p,
            stop_sequences=body.stop_sequences
        )
        return result
    except Exception as e:
        logger.error(f"Generation error: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@router.post("/generate/stream")
async def generate_stream(request: Request, body: GenerateRequest):
    pipeline = request.app.state.pipeline

    async def event_stream():
        try:
            async for token in pipeline.generate_stream(
                prompt=body.prompt,
                max_tokens=body.max_tokens,
                temperature=body.temperature,
                top_p=body.top_p,
                stop_sequences=body.stop_sequences
            ):
                yield f"data: {json.dumps({'token': token})}\n\n"
            yield f"data: {json.dumps({'done': True})}\n\n"
        except Exception as e:
            logger.error(f"Stream error: {e}")
            yield f"data: {json.dumps({'error': str(e)})}\n\n"

    return StreamingResponse(
        event_stream(),
        media_type="text/event-stream"
    )


@router.post("/generate/batch")
async def generate_batch(request: Request, body: BatchRequest):
    pipeline = request.app.state.pipeline

    try:
        results = await pipeline.generate_batch(
            prompts=body.prompts,
            max_tokens=body.max_tokens,
            temperature=body.temperature
        )
        return {"results": results}
    except Exception as e:
        logger.error(f"Batch generation error: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/health")
async def health(request: Request):
    pipeline = request.app.state.pipeline
    return {
        "status": "ok",
        "model": pipeline.model_name,
        "workers": pipeline.num_workers,
        "queue_size": pipeline.queue_size
    }


@router.get("/models")
async def list_models(request: Request):
    pipeline = request.app.state.pipeline
    return {
        "models": [pipeline.model_name],
        "active": pipeline.model_name
    }
