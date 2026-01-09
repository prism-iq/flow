import asyncio
import uvicorn
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from contextlib import asynccontextmanager

from .config import settings
from .routes import router
from .pipeline.manager import PipelineManager
from .utils.logger import logger


pipeline_manager: PipelineManager = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    global pipeline_manager
    logger.info("Initializing LLM pipeline...")
    pipeline_manager = PipelineManager(settings)
    await pipeline_manager.initialize()
    app.state.pipeline = pipeline_manager
    logger.info(f"Pipeline ready with {settings.num_workers} workers")
    yield
    logger.info("Shutting down pipeline...")
    await pipeline_manager.shutdown()


app = FastAPI(
    title="Flow LLM Service",
    version="1.0.0",
    lifespan=lifespan
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"]
)

app.include_router(router)


def main():
    uvicorn.run(
        "src.main:app",
        host=settings.host,
        port=settings.port,
        reload=False,
        workers=1
    )


if __name__ == "__main__":
    main()
