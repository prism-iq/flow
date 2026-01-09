<script>
  import { createEventDispatcher } from 'svelte';

  export let disabled = false;

  const dispatch = createEventDispatcher();
  let message = '';

  function handleSubmit() {
    if (message.trim() && !disabled) {
      dispatch('send', { message: message.trim() });
      message = '';
    }
  }

  function handleKeydown(event) {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault();
      handleSubmit();
    }
  }
</script>

<form class="input-container" on:submit|preventDefault={handleSubmit}>
  <textarea
    bind:value={message}
    on:keydown={handleKeydown}
    placeholder="Send a message..."
    rows="1"
    {disabled}
  ></textarea>
  <button type="submit" disabled={!message.trim() || disabled}>
    Send
  </button>
</form>

<style>
  .input-container {
    display: flex;
    gap: 0.75rem;
    padding: 1rem 1.5rem;
    border-top: 1px solid var(--border);
    background: var(--bg-secondary);
  }

  textarea {
    flex: 1;
    min-height: auto;
    max-height: 200px;
    resize: none;
  }

  button {
    align-self: flex-end;
    padding: 0.75rem 1.5rem;
  }
</style>
