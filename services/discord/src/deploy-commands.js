import { REST, Routes } from 'discord.js';
import { config } from './config.js';
import { getCommandsData } from './commands/index.js';
import { logger } from './utils/logger.js';

async function deployCommands() {
  const commands = getCommandsData();

  const rest = new REST().setToken(config.token);

  try {
    logger.info(`Deploying ${commands.length} commands...`);

    if (config.guildId) {
      await rest.put(
        Routes.applicationGuildCommands(config.clientId, config.guildId),
        { body: commands }
      );
      logger.info(`Commands deployed to guild ${config.guildId}`);
    } else {
      await rest.put(
        Routes.applicationCommands(config.clientId),
        { body: commands }
      );
      logger.info('Commands deployed globally');
    }

  } catch (error) {
    logger.error(error, 'Failed to deploy commands');
    process.exit(1);
  }
}

deployCommands();
