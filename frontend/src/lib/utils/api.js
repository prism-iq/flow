const API_BASE = '/api/v1';
const WS_BASE = `ws://${typeof window !== 'undefined' ? window.location.host : 'localhost:8080'}/ws`;

export async function sendMessage(message, conversationId, useRag = false) {
  const response = await fetch(`${API_BASE}/chat/send`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      message,
      conversation_id: conversationId,
      use_rag: useRag
    })
  });

  if (!response.ok) {
    throw new Error(`API error: ${response.status}`);
  }

  return response.json();
}

export async function* streamMessage(message, conversationId, useRag = false) {
  const response = await fetch(`${API_BASE}/chat/stream`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      message,
      conversation_id: conversationId,
      use_rag: useRag
    })
  });

  if (!response.ok) {
    throw new Error(`API error: ${response.status}`);
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = '';

  while (true) {
    const { done, value } = await reader.read();
    if (done) break;

    buffer += decoder.decode(value, { stream: true });
    const lines = buffer.split('\n');
    buffer = lines.pop() || '';

    for (const line of lines) {
      if (line.startsWith('event: ')) continue;
      if (line.startsWith('data: ')) {
        const data = line.slice(6);
        try {
          const parsed = JSON.parse(data);
          if (parsed.content) {
            yield parsed.content;
          }
          if (parsed.done) {
            return;
          }
        } catch {}
      }
    }
  }
}

export function createWebSocket(onMessage) {
  const ws = new WebSocket(WS_BASE);

  ws.onopen = () => {
    console.log('WebSocket connected');
  };

  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data);
      onMessage(data);
    } catch (e) {
      console.error('WebSocket message error:', e);
    }
  };

  ws.onerror = (error) => {
    console.error('WebSocket error:', error);
  };

  ws.onclose = () => {
    console.log('WebSocket closed');
  };

  return {
    send: (type, payload) => {
      ws.send(JSON.stringify({ type, ...payload }));
    },
    close: () => ws.close()
  };
}

export async function getHealth() {
  const response = await fetch(`${API_BASE}/health`);
  return response.json();
}

export async function getStatus() {
  const response = await fetch(`${API_BASE}/status`);
  return response.json();
}
