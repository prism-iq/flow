<script>
  import { conversations, activeConversation } from '$lib/stores/chat';

  function newConversation() {
    conversations.new();
  }

  function selectConversation(id) {
    activeConversation.set(id);
  }

  function deleteConversation(id, event) {
    event.stopPropagation();
    conversations.delete(id);
    if ($activeConversation === id) {
      activeConversation.set($conversations[0]?.id || null);
    }
  }
</script>

<aside>
  <div class="header">
    <span class="logo">Flow</span>
    <button class="new-btn" on:click={newConversation}>+</button>
  </div>

  <nav>
    {#each $conversations as conv (conv.id)}
      <button
        class="conv-item"
        class:active={$activeConversation === conv.id}
        on:click={() => selectConversation(conv.id)}
      >
        <span class="title">{conv.title}</span>
        <button
          class="delete-btn"
          on:click={(e) => deleteConversation(conv.id, e)}
        >
          x
        </button>
      </button>
    {/each}
  </nav>

  <div class="footer">
    <span>v1.0.0</span>
  </div>
</aside>

<style>
  aside {
    width: 260px;
    background: var(--bg-secondary);
    border-right: 1px solid var(--border);
    display: flex;
    flex-direction: column;
  }

  .header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 1rem;
    border-bottom: 1px solid var(--border);
  }

  .logo {
    font-weight: 700;
    font-size: 1.25rem;
    background: linear-gradient(135deg, var(--accent), #ec4899);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
  }

  .new-btn {
    width: 32px;
    height: 32px;
    padding: 0;
    font-size: 1.25rem;
    line-height: 1;
  }

  nav {
    flex: 1;
    overflow-y: auto;
    padding: 0.5rem;
  }

  .conv-item {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
    background: transparent;
    padding: 0.75rem;
    margin-bottom: 0.25rem;
    border-radius: var(--radius);
    text-align: left;
  }

  .conv-item:hover {
    background: var(--bg-tertiary);
  }

  .conv-item.active {
    background: var(--accent);
  }

  .conv-item .title {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .delete-btn {
    width: 24px;
    height: 24px;
    padding: 0;
    font-size: 0.875rem;
    background: transparent;
    opacity: 0;
    transition: opacity var(--transition);
  }

  .conv-item:hover .delete-btn {
    opacity: 0.5;
  }

  .delete-btn:hover {
    opacity: 1 !important;
    background: var(--error);
  }

  .footer {
    padding: 1rem;
    border-top: 1px solid var(--border);
    color: var(--text-secondary);
    font-size: 0.75rem;
  }
</style>
