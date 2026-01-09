import { logger } from '../utils/logger.js';

export class ConnectionManager {
  constructor() {
    this.connections = new Map();
    this.channels = new Map();
  }

  add(socket, channel = 'default') {
    const clientId = `client_${Date.now()}_${Math.random().toString(36).slice(2)}`;

    this.connections.set(clientId, {
      socket,
      channel,
      connectedAt: Date.now(),
      messageCount: 0,
    });

    if (!this.channels.has(channel)) {
      this.channels.set(channel, new Set());
    }
    this.channels.get(channel).add(clientId);

    return clientId;
  }

  remove(clientId) {
    const conn = this.connections.get(clientId);
    if (!conn) return;

    const channel = this.channels.get(conn.channel);
    if (channel) {
      channel.delete(clientId);
      if (channel.size === 0) {
        this.channels.delete(conn.channel);
      }
    }

    this.connections.delete(clientId);
  }

  get(clientId) {
    return this.connections.get(clientId);
  }

  send(clientId, message) {
    const conn = this.connections.get(clientId);
    if (!conn) return false;

    try {
      const data = typeof message === 'string'
        ? message
        : JSON.stringify(message);
      conn.socket.send(data);
      conn.messageCount++;
      return true;
    } catch (error) {
      logger.error({ clientId, error }, 'Failed to send message');
      return false;
    }
  }

  broadcast(message, channel = 'default') {
    const clients = this.channels.get(channel);
    if (!clients) return 0;

    let sent = 0;
    for (const clientId of clients) {
      if (this.send(clientId, message)) {
        sent++;
      }
    }

    return sent;
  }

  broadcastAll(message) {
    let sent = 0;
    for (const clientId of this.connections.keys()) {
      if (this.send(clientId, message)) {
        sent++;
      }
    }
    return sent;
  }

  getStats() {
    return {
      totalConnections: this.connections.size,
      channels: Object.fromEntries(
        Array.from(this.channels.entries()).map(([k, v]) => [k, v.size])
      ),
    };
  }

  getClientIds(channel) {
    if (channel) {
      return Array.from(this.channels.get(channel) || []);
    }
    return Array.from(this.connections.keys());
  }
}
