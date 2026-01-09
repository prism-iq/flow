import { Events } from 'discord.js';
import { logger } from '../utils/logger.js';
import { handleMessage } from './message.js';
import { handleInteraction } from './interaction.js';

export async function loadEvents(client) {
  client.once(Events.ClientReady, (c) => {
    logger.info(`Logged in as ${c.user.tag}`);
    c.user.setActivity('Flow AI | /chat', { type: 0 });
  });

  client.on(Events.InteractionCreate, handleInteraction);

  client.on(Events.MessageCreate, handleMessage);

  client.on(Events.Error, (error) => {
    logger.error(error, 'Discord client error');
  });

  client.on(Events.Warn, (warning) => {
    logger.warn(warning, 'Discord client warning');
  });
}
