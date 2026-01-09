import { request as undiciRequest } from 'undici';
import { config } from '../config.js';
import { logger } from '../utils/logger.js';

export class MessageRouter {
  constructor() {
    this.handlers = new Map();
    this.registerDefaultHandlers();
  }

  registerDefaultHandlers() {
    this.handlers.set('ping', this.handlePing.bind(this));
    this.handlers.set('chat', this.handleChat.bind(this));
    this.handlers.set('subscribe', this.handleSubscribe.bind(this));
    this.handlers.set('unsubscribe', this.handleUnsubscribe.bind(this));
  }

  register(type, handler) {
    this.handlers.set(type, handler);
  }

  async handle(clientId, message, socket) {
    const { type, ...payload } = message;

    const handler = this.handlers.get(type);
    if (!handler) {
      socket.send(JSON.stringify({
        type: 'error',
        error: `Unknown message type: ${type}`,
      }));
      return;
    }

    try {
      await handler(clientId, payload, socket);
    } catch (error) {
      logger.error({ clientId, type, error }, 'Handler error');
      socket.send(JSON.stringify({
        type: 'error',
        error: error.message,
      }));
    }
  }

  async handlePing(clientId, payload, socket) {
    socket.send(JSON.stringify({
      type: 'pong',
      timestamp: Date.now(),
    }));
  }

  async handleChat(clientId, payload, socket) {
    const { content, conversationId, useRag = false } = payload;

    socket.send(JSON.stringify({
      type: 'status',
      status: 'processing',
    }));

    try {
      let context = '';

      if (useRag) {
        socket.send(JSON.stringify({
          type: 'status',
          status: 'searching',
        }));

        const ragResponse = await undiciRequest(`${config.services.ragService}/query`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ query: content, top_k: 5 }),
        });

        const ragResult = await ragResponse.body.json();
        context = ragResult.documents.map(d => d.content).join('\n\n');
      }

      socket.send(JSON.stringify({
        type: 'status',
        status: 'generating',
      }));

      const prompt = context
        ? `Context:\n${context}\n\nQuery: ${content}`
        : content;

      const llmResponse = await undiciRequest(`${config.services.llmService}/generate/stream`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ prompt, max_tokens: 2048, stream: true }),
      });

      for await (const chunk of llmResponse.body) {
        socket.send(JSON.stringify({
          type: 'token',
          content: chunk.toString(),
        }));
      }

      socket.send(JSON.stringify({
        type: 'done',
        conversationId,
      }));

    } catch (error) {
      throw new Error(`Chat processing failed: ${error.message}`);
    }
  }

  async handleSubscribe(clientId, payload, socket) {
    const { channel } = payload;
    socket.send(JSON.stringify({
      type: 'subscribed',
      channel,
    }));
  }

  async handleUnsubscribe(clientId, payload, socket) {
    const { channel } = payload;
    socket.send(JSON.stringify({
      type: 'unsubscribed',
      channel,
    }));
  }
}
