export const config = {
  port: parseInt(process.env.PORT || '3001', 10),
  host: process.env.HOST || '0.0.0.0',

  redis: {
    host: process.env.REDIS_HOST || 'localhost',
    port: parseInt(process.env.REDIS_PORT || '6379', 10),
    password: process.env.REDIS_PASSWORD || undefined,
  },

  postgres: {
    host: process.env.PG_HOST || 'localhost',
    port: parseInt(process.env.PG_PORT || '5432', 10),
    database: process.env.PG_DATABASE || 'flow',
    user: process.env.PG_USER || 'flow',
    password: process.env.PG_PASSWORD || 'flow',
  },

  services: {
    goOrchestrator: process.env.GO_ORCHESTRATOR_URL || 'http://localhost:8080',
    llmService: process.env.LLM_SERVICE_URL || 'http://localhost:8001',
    ragService: process.env.RAG_SERVICE_URL || 'http://localhost:8002',
  },

  queue: {
    maxConcurrency: parseInt(process.env.QUEUE_CONCURRENCY || '10', 10),
    retryAttempts: parseInt(process.env.QUEUE_RETRIES || '3', 10),
    retryDelay: parseInt(process.env.QUEUE_RETRY_DELAY || '1000', 10),
  },

  logLevel: process.env.LOG_LEVEL || 'info',
};
