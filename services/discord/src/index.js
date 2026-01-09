import { Client, GatewayIntentBits, Collection, Events } from 'discord.js';
import { config } from './config.js';
import { logger } from './utils/logger.js';
import { loadCommands } from './commands/index.js';
import { loadEvents } from './events/index.js';

const client = new Client({
  intents: [
    GatewayIntentBits.Guilds,
    GatewayIntentBits.GuildMessages,
    GatewayIntentBits.MessageContent,
    GatewayIntentBits.DirectMessages
  ]
});

client.commands = new Collection();
client.conversations = new Map();

async function main() {
  try {
    await loadCommands(client);
    await loadEvents(client);

    await client.login(config.token);

    logger.info('Flow Discord bot started');
  } catch (error) {
    logger.error(error, 'Failed to start bot');
    process.exit(1);
  }
}

process.on('SIGTERM', async () => {
  logger.info('Shutting down...');
  client.destroy();
  process.exit(0);
});

process.on('SIGINT', async () => {
  logger.info('Shutting down...');
  client.destroy();
  process.exit(0);
});

main();
