export async function healthRoutes(server) {
  server.get('/', async () => {
    return {
      status: 'ok',
      service: 'flow-async-io',
      timestamp: new Date().toISOString(),
    };
  });

  server.get('/ready', async () => {
    return {
      status: 'ready',
      uptime: process.uptime(),
      memory: process.memoryUsage(),
    };
  });

  server.get('/live', async () => {
    return { status: 'live' };
  });
}
