<script>
  export let message;

  $: isUser = message.role === 'user';
</script>

<div class="message" class:user={isUser} class:assistant={!isUser}>
  <div class="avatar">
    {isUser ? 'U' : 'AI'}
  </div>
  <div class="content">
    <div class="text">{message.content}</div>
    <div class="meta">
      {new Date(message.timestamp).toLocaleTimeString()}
    </div>
  </div>
</div>

<style>
  .message {
    display: flex;
    gap: 1rem;
    margin-bottom: 1.5rem;
    animation: fadeIn 0.3s ease;
  }

  .message.user {
    flex-direction: row-reverse;
  }

  .avatar {
    width: 36px;
    height: 36px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.75rem;
    font-weight: 600;
    flex-shrink: 0;
  }

  .user .avatar {
    background: var(--accent);
  }

  .assistant .avatar {
    background: var(--bg-tertiary);
    border: 1px solid var(--border);
  }

  .content {
    max-width: 70%;
  }

  .text {
    background: var(--bg-secondary);
    padding: 0.75rem 1rem;
    border-radius: var(--radius);
    white-space: pre-wrap;
    word-break: break-word;
  }

  .user .text {
    background: var(--accent);
  }

  .user .content {
    text-align: right;
  }

  .meta {
    font-size: 0.75rem;
    color: var(--text-secondary);
    margin-top: 0.25rem;
  }
</style>
