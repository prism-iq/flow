export const config = {
  token: process.env.DISCORD_TOKEN,
  clientId: process.env.DISCORD_CLIENT_ID,
  guildId: process.env.DISCORD_GUILD_ID,

  services: {
    goOrchestrator: process.env.GO_ORCHESTRATOR_URL || 'http://localhost:8080',
    nodeService: process.env.NODE_SERVICE_URL || 'http://localhost:3001',
  },

  maxMessageLength: 2000,
  typingInterval: 5000,
  maxConversationHistory: 10,
};
