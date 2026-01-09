<script>
  import { onMount, afterUpdate } from 'svelte';
  import { conversations, isStreaming } from '$lib/stores/chat';
  import { settings } from '$lib/stores/settings';
  import { streamMessage } from '$lib/utils/api';
  import Message from './Message.svelte';
  import ChatInput from './ChatInput.svelte';

  export let conversation;

  let messagesContainer;
  let scrollToBottom = true;

  afterUpdate(() => {
    if (scrollToBottom && messagesContainer) {
      messagesContainer.scrollTop = messagesContainer.scrollHeight;
    }
  });

  async function handleSend(event) {
    const { message } = event.detail;
    if (!message.trim() || $isStreaming) return;

    conversations.addMessage(conversation.id, {
      id: `msg_${Date.now()}`,
      role: 'user',
      content: message,
      timestamp: new Date().toISOString()
    });

    conversations.addMessage(conversation.id, {
      id: `msg_${Date.now() + 1}`,
      role: 'assistant',
      content: '',
      timestamp: new Date().toISOString()
    });

    isStreaming.set(true);

    try {
      let fullResponse = '';

      for await (const token of streamMessage(message, conversation.id, $settings.useRag)) {
        fullResponse += token;
        conversations.updateLastMessage(conversation.id, fullResponse);
      }

      if (conversation.messages.length === 2 && conversation.title === 'New Conversation') {
        const title = message.slice(0, 50) + (message.length > 50 ? '...' : '');
        conversations.setTitle(conversation.id, title);
      }

    } catch (error) {
      console.error('Stream error:', error);
      conversations.updateLastMessage(conversation.id, 'An error occurred. Please try again.');
    }

    isStreaming.set(false);
  }
</script>

<div class="chat">
  <div class="messages" bind:this={messagesContainer}>
    {#if conversation.messages.length === 0}
      <div class="empty">
        <p>Send a message to start the conversation.</p>
      </div>
    {:else}
      {#each conversation.messages as message (message.id)}
        <Message {message} />
      {/each}
    {/if}

    {#if $isStreaming}
      <div class="typing">
        <span></span>
        <span></span>
        <span></span>
      </div>
    {/if}
  </div>

  <ChatInput on:send={handleSend} disabled={$isStreaming} />
</div>

<style>
  .chat {
    height: 100%;
    display: flex;
    flex-direction: column;
  }

  .messages {
    flex: 1;
    overflow-y: auto;
    padding: 1.5rem;
  }

  .empty {
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-secondary);
  }

  .typing {
    display: flex;
    gap: 0.25rem;
    padding: 0.5rem 0;
  }

  .typing span {
    width: 8px;
    height: 8px;
    background: var(--text-secondary);
    border-radius: 50%;
    animation: pulse 1.5s infinite;
  }

  .typing span:nth-child(2) {
    animation-delay: 0.2s;
  }

  .typing span:nth-child(3) {
    animation-delay: 0.4s;
  }
</style>
