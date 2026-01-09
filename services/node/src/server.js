import Fastify from 'fastify';
import cors from '@fastify/cors';
import websocket from '@fastify/websocket';
import { logger } from './utils/logger.js';
import { registerRoutes } from './routes/index.js';
import { registerWebSocket } from './routes/websocket.js';
import { errorHandler } from './middleware/error.js';

export async function createServer(config) {
  const server = Fastify({
    logger: logger,
    trustProxy: true,
  });

  await server.register(cors, {
    origin: true,
    credentials: true,
  });

  await server.register(websocket, {
    options: {
      maxPayload: 1048576,
      clientTracking: true,
    }
  });

  server.setErrorHandler(errorHandler);

  registerRoutes(server);
  await registerWebSocket(server);

  server.addHook('onRequest', async (request) => {
    request.startTime = Date.now();
  });

  server.addHook('onResponse', async (request, reply) => {
    const duration = Date.now() - request.startTime;
    logger.debug({
      method: request.method,
      url: request.url,
      statusCode: reply.statusCode,
      duration,
    }, 'Request completed');
  });

  return server;
}
