import { ConnectionManager } from '../services/connection-manager.js';
import { MessageRouter } from '../services/message-router.js';
import { logger } from '../utils/logger.js';

const connections = new ConnectionManager();
const router = new MessageRouter();

export async function registerWebSocket(server) {
  server.get('/ws', { websocket: true }, (socket, request) => {
    const clientId = connections.add(socket);
    logger.info({ clientId }, 'WebSocket client connected');

    socket.on('message', async (data) => {
      try {
        const message = JSON.parse(data.toString());
        await router.handle(clientId, message, socket);
      } catch (error) {
        logger.error(error, 'WebSocket message error');
        socket.send(JSON.stringify({
          type: 'error',
          error: 'Invalid message format'
        }));
      }
    });

    socket.on('close', () => {
      connections.remove(clientId);
      logger.info({ clientId }, 'WebSocket client disconnected');
    });

    socket.on('error', (error) => {
      logger.error({ clientId, error }, 'WebSocket error');
      connections.remove(clientId);
    });

    socket.send(JSON.stringify({
      type: 'connected',
      clientId
    }));
  });

  server.get('/ws/broadcast', { websocket: true }, (socket) => {
    const clientId = connections.add(socket, 'broadcast');

    socket.on('message', (data) => {
      connections.broadcast(data.toString(), 'broadcast');
    });

    socket.on('close', () => {
      connections.remove(clientId);
    });
  });
}
