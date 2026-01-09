import { request as undiciRequest } from 'undici';
import { config } from '../config.js';
import { logger } from '../utils/logger.js';

export async function streamRoutes(server) {
  server.post('/llm', async (request, reply) => {
    const { prompt, maxTokens = 2048 } = request.body;

    reply.raw.setHeader('Content-Type', 'text/event-stream');
    reply.raw.setHeader('Cache-Control', 'no-cache');
    reply.raw.setHeader('Connection', 'keep-alive');

    try {
      const response = await undiciRequest(`${config.services.llmService}/generate/stream`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ prompt, max_tokens: maxTokens, stream: true }),
      });

      for await (const chunk of response.body) {
        reply.raw.write(`data: ${chunk.toString()}\n\n`);
      }

      reply.raw.write('data: [DONE]\n\n');
      reply.raw.end();

    } catch (error) {
      logger.error(error, 'LLM stream error');
      reply.raw.write(`data: ${JSON.stringify({ error: error.message })}\n\n`);
      reply.raw.end();
    }
  });

  server.post('/rag', async (request, reply) => {
    const { query, topK = 5 } = request.body;

    reply.raw.setHeader('Content-Type', 'text/event-stream');
    reply.raw.setHeader('Cache-Control', 'no-cache');

    try {
      reply.raw.write(`data: ${JSON.stringify({ step: 'searching' })}\n\n`);

      const ragResponse = await undiciRequest(`${config.services.ragService}/query`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ query, top_k: topK }),
      });

      const ragResult = await ragResponse.body.json();
      reply.raw.write(`data: ${JSON.stringify({ step: 'found', documents: ragResult.documents })}\n\n`);

      const context = ragResult.documents.map(d => d.content).join('\n\n');
      const prompt = `Context:\n${context}\n\nQuery: ${query}`;

      reply.raw.write(`data: ${JSON.stringify({ step: 'generating' })}\n\n`);

      const llmResponse = await undiciRequest(`${config.services.llmService}/generate/stream`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ prompt, max_tokens: 2048, stream: true }),
      });

      for await (const chunk of llmResponse.body) {
        reply.raw.write(`data: ${JSON.stringify({ step: 'token', content: chunk.toString() })}\n\n`);
      }

      reply.raw.write(`data: ${JSON.stringify({ step: 'done' })}\n\n`);
      reply.raw.end();

    } catch (error) {
      logger.error(error, 'RAG stream error');
      reply.raw.write(`data: ${JSON.stringify({ error: error.message })}\n\n`);
      reply.raw.end();
    }
  });
}
