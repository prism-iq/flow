import { createServer } from './server.js';
import { config } from './config.js';
import { logger } from './utils/logger.js';

async function main() {
  const server = await createServer(config);

  try {
    await server.listen({
      port: config.port,
      host: config.host
    });

    logger.info({
      port: config.port,
      host: config.host
    }, 'Flow Async I/O service started');

  } catch (err) {
    logger.error(err, 'Failed to start server');
    process.exit(1);
  }

  const shutdown = async (signal) => {
    logger.info({ signal }, 'Shutting down');
    await server.close();
    process.exit(0);
  };

  process.on('SIGTERM', () => shutdown('SIGTERM'));
  process.on('SIGINT', () => shutdown('SIGINT'));
}

main();
