import { SlashCommandBuilder } from 'discord.js';
import { sendChat, streamChat } from '../utils/api.js';
import { config } from '../config.js';
import { logger } from '../utils/logger.js';

export const chatCommand = {
  data: new SlashCommandBuilder()
    .setName('chat')
    .setDescription('Chat with Flow AI')
    .addStringOption(option =>
      option
        .setName('message')
        .setDescription('Your message')
        .setRequired(true)
    )
    .addBooleanOption(option =>
      option
        .setName('rag')
        .setDescription('Use RAG for context')
        .setRequired(false)
    )
    .addBooleanOption(option =>
      option
        .setName('stream')
        .setDescription('Stream the response')
        .setRequired(false)
    ),

  async execute(interaction) {
    const message = interaction.options.getString('message');
    const useRag = interaction.options.getBoolean('rag') ?? false;
    const useStream = interaction.options.getBoolean('stream') ?? true;

    await interaction.deferReply();

    const userId = interaction.user.id;
    let conversationId = interaction.client.conversations.get(userId);

    try {
      if (useStream) {
        let response = '';
        let lastUpdate = Date.now();

        for await (const token of streamChat(message, conversationId, useRag)) {
          response += token;

          if (Date.now() - lastUpdate > 1000 && response.length > 0) {
            const displayText = truncateMessage(response);
            await interaction.editReply(displayText + ' ▌');
            lastUpdate = Date.now();
          }
        }

        await interaction.editReply(truncateMessage(response));

      } else {
        const result = await sendChat(message, conversationId, useRag);

        if (result.conversation_id) {
          interaction.client.conversations.set(userId, result.conversation_id);
        }

        await interaction.editReply(truncateMessage(result.message.content));
      }

    } catch (error) {
      logger.error(error, 'Chat command error');
      await interaction.editReply('An error occurred. Please try again.');
    }
  }
};

function truncateMessage(text) {
  if (text.length <= config.maxMessageLength) {
    return text;
  }
  return text.slice(0, config.maxMessageLength - 3) + '...';
}
