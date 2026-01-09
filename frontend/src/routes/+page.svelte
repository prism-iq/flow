<script>
  import Chat from '$lib/components/Chat.svelte';
  import { conversations, activeConversation } from '$lib/stores/chat';

  $: conversation = $conversations.find(c => c.id === $activeConversation);
</script>

<svelte:head>
  <title>Flow - Chat</title>
</svelte:head>

<div class="chat-page">
  {#if conversation}
    <Chat {conversation} />
  {:else}
    <div class="welcome">
      <h1>Welcome to Flow</h1>
      <p>Start a new conversation or select one from the sidebar.</p>
      <button on:click={() => conversations.new()}>
        New Conversation
      </button>
    </div>
  {/if}
</div>

<style>
  .chat-page {
    height: 100%;
    display: flex;
    flex-direction: column;
  }

  .welcome {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 1rem;
    text-align: center;
    padding: 2rem;
  }

  .welcome h1 {
    font-size: 2rem;
    font-weight: 600;
  }

  .welcome p {
    color: var(--text-secondary);
    max-width: 400px;
  }
</style>
