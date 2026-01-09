import { request as undiciRequest } from 'undici';
import { config } from '../config.js';
import { logger } from '../utils/logger.js';

export async function proxyRoutes(server) {
  server.all('/go/*', async (request, reply) => {
    const path = request.url.replace('/api/proxy/go', '');
    const url = `${config.services.goOrchestrator}${path}`;

    try {
      const response = await undiciRequest(url, {
        method: request.method,
        headers: {
          ...request.headers,
          host: new URL(config.services.goOrchestrator).host,
        },
        body: request.body ? JSON.stringify(request.body) : undefined,
      });

      reply.code(response.statusCode);
      for (const [key, value] of Object.entries(response.headers)) {
        reply.header(key, value);
      }

      return response.body;

    } catch (error) {
      logger.error(error, 'Proxy error to Go service');
      reply.code(502).send({ error: 'Bad Gateway' });
    }
  });

  server.all('/llm/*', async (request, reply) => {
    const path = request.url.replace('/api/proxy/llm', '');
    const url = `${config.services.llmService}${path}`;

    try {
      const response = await undiciRequest(url, {
        method: request.method,
        headers: { 'Content-Type': 'application/json' },
        body: request.body ? JSON.stringify(request.body) : undefined,
      });

      reply.code(response.statusCode);
      return response.body;

    } catch (error) {
      logger.error(error, 'Proxy error to LLM service');
      reply.code(502).send({ error: 'Bad Gateway' });
    }
  });

  server.all('/rag/*', async (request, reply) => {
    const path = request.url.replace('/api/proxy/rag', '');
    const url = `${config.services.ragService}${path}`;

    try {
      const response = await undiciRequest(url, {
        method: request.method,
        headers: { 'Content-Type': 'application/json' },
        body: request.body ? JSON.stringify(request.body) : undefined,
      });

      reply.code(response.statusCode);
      return response.body;

    } catch (error) {
      logger.error(error, 'Proxy error to RAG service');
      reply.code(502).send({ error: 'Bad Gateway' });
    }
  });
}
