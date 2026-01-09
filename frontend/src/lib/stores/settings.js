import { writable } from 'svelte/store';

const defaultSettings = {
  useRag: false,
  temperature: 0.7,
  maxTokens: 2048,
  theme: 'dark'
};

function createSettingsStore() {
  const stored = typeof localStorage !== 'undefined'
    ? localStorage.getItem('flow_settings')
    : null;

  const initial = stored ? JSON.parse(stored) : defaultSettings;
  const { subscribe, set, update } = writable(initial);

  return {
    subscribe,
    set: (value) => {
      if (typeof localStorage !== 'undefined') {
        localStorage.setItem('flow_settings', JSON.stringify(value));
      }
      set(value);
    },
    update: (fn) => {
      update(current => {
        const updated = fn(current);
        if (typeof localStorage !== 'undefined') {
          localStorage.setItem('flow_settings', JSON.stringify(updated));
        }
        return updated;
      });
    },
    reset: () => {
      if (typeof localStorage !== 'undefined') {
        localStorage.setItem('flow_settings', JSON.stringify(defaultSettings));
      }
      set(defaultSettings);
    }
  };
}

export const settings = createSettingsStore();
