<script>
  import { activeConversation, conversations } from '$lib/stores/chat';
  import { settings } from '$lib/stores/settings';

  $: conversation = $conversations.find(c => c.id === $activeConversation);

  function toggleRag() {
    settings.update(s => ({ ...s, useRag: !s.useRag }));
  }
</script>

<header>
  <div class="title">
    {#if conversation}
      <h1>{conversation.title}</h1>
    {:else}
      <h1>Flow</h1>
    {/if}
  </div>

  <div class="actions">
    <button
      class="toggle"
      class:active={$settings.useRag}
      on:click={toggleRag}
      title="Toggle RAG"
    >
      RAG {$settings.useRag ? 'ON' : 'OFF'}
    </button>
  </div>
</header>

<style>
  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 1rem 1.5rem;
    border-bottom: 1px solid var(--border);
    background: var(--bg-secondary);
  }

  .title h1 {
    font-size: 1.125rem;
    font-weight: 600;
  }

  .actions {
    display: flex;
    gap: 0.5rem;
  }

  .toggle {
    background: var(--bg-tertiary);
    font-size: 0.75rem;
    padding: 0.375rem 0.75rem;
  }

  .toggle.active {
    background: var(--accent);
  }
</style>
