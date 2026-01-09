import { TaskQueue } from '../services/task-queue.js';
import { config } from '../config.js';

const queue = new TaskQueue(config.queue);

export async function queueRoutes(server) {
  server.post('/enqueue', {
    schema: {
      body: {
        type: 'object',
        required: ['type', 'payload'],
        properties: {
          type: { type: 'string' },
          payload: { type: 'object' },
          priority: { type: 'number', default: 0 },
        }
      }
    }
  }, async (request) => {
    const { type, payload, priority } = request.body;
    const taskId = await queue.enqueue(type, payload, priority);
    return { taskId, status: 'queued' };
  });

  server.get('/status/:taskId', async (request) => {
    const { taskId } = request.params;
    const status = await queue.getStatus(taskId);

    if (!status) {
      return { error: 'Task not found' };
    }

    return status;
  });

  server.post('/cancel/:taskId', async (request) => {
    const { taskId } = request.params;
    const cancelled = await queue.cancel(taskId);
    return { taskId, cancelled };
  });

  server.get('/stats', async () => {
    return queue.getStats();
  });
}
