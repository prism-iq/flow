import { healthRoutes } from './health.js';
import { queueRoutes } from './queue.js';
import { streamRoutes } from './stream.js';
import { proxyRoutes } from './proxy.js';

export function registerRoutes(server) {
  server.register(healthRoutes, { prefix: '/health' });
  server.register(queueRoutes, { prefix: '/api/queue' });
  server.register(streamRoutes, { prefix: '/api/stream' });
  server.register(proxyRoutes, { prefix: '/api/proxy' });
}
