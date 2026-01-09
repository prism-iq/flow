import { request } from 'undici';
import { config } from '../config.js';
import { logger } from './logger.js';

export async function sendChat(message, conversationId = null, useRag = false) {
  try {
    const response = await request(`${config.services.goOrchestrator}/api/v1/chat/send`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        message,
        conversation_id: conversationId,
        use_rag: useRag
      })
    });

    if (response.statusCode !== 200) {
      throw new Error(`API returned ${response.statusCode}`);
    }

    return await response.body.json();
  } catch (error) {
    logger.error(error, 'API error');
    throw error;
  }
}

export async function* streamChat(message, conversationId = null, useRag = false) {
  try {
    const response = await request(`${config.services.nodeService}/api/stream/llm`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        prompt: message,
        maxTokens: 2048
      })
    });

    const decoder = new TextDecoder();
    let buffer = '';

    for await (const chunk of response.body) {
      buffer += decoder.decode(chunk, { stream: true });

      const lines = buffer.split('\n');
      buffer = lines.pop() || '';

      for (const line of lines) {
        if (line.startsWith('data: ')) {
          const data = line.slice(6);
          if (data === '[DONE]') {
            return;
          }
          try {
            const parsed = JSON.parse(data);
            if (parsed.token) {
              yield parsed.token;
            }
          } catch {}
        }
      }
    }
  } catch (error) {
    logger.error(error, 'Stream error');
    throw error;
  }
}

export async function queryRag(query, topK = 5) {
  try {
    const response = await request(`${config.services.goOrchestrator}/api/v1/rag/query`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ query, top_k: topK })
    });

    return await response.body.json();
  } catch (error) {
    logger.error(error, 'RAG query error');
    throw error;
  }
}
