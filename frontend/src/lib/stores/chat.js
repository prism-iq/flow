import { writable, derived } from 'svelte/store';

function createConversationsStore() {
  const { subscribe, set, update } = writable([]);

  return {
    subscribe,
    set,
    new: () => {
      const id = `conv_${Date.now()}`;
      update(convs => [{
        id,
        title: 'New Conversation',
        messages: [],
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString()
      }, ...convs]);
      activeConversation.set(id);
      return id;
    },
    addMessage: (conversationId, message) => {
      update(convs => convs.map(c => {
        if (c.id === conversationId) {
          return {
            ...c,
            messages: [...c.messages, message],
            updatedAt: new Date().toISOString()
          };
        }
        return c;
      }));
    },
    updateLastMessage: (conversationId, content) => {
      update(convs => convs.map(c => {
        if (c.id === conversationId && c.messages.length > 0) {
          const messages = [...c.messages];
          messages[messages.length - 1] = {
            ...messages[messages.length - 1],
            content
          };
          return { ...c, messages };
        }
        return c;
      }));
    },
    delete: (id) => {
      update(convs => convs.filter(c => c.id !== id));
    },
    setTitle: (id, title) => {
      update(convs => convs.map(c =>
        c.id === id ? { ...c, title } : c
      ));
    },
    clear: () => set([])
  };
}

export const conversations = createConversationsStore();
export const activeConversation = writable(null);
export const isStreaming = writable(false);
