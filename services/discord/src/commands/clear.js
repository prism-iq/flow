import { SlashCommandBuilder } from 'discord.js';

export const clearCommand = {
  data: new SlashCommandBuilder()
    .setName('clear')
    .setDescription('Clear your conversation history'),

  async execute(interaction) {
    const userId = interaction.user.id;

    if (interaction.client.conversations.has(userId)) {
      interaction.client.conversations.delete(userId);
      await interaction.reply({
        content: 'Conversation history cleared.',
        ephemeral: true
      });
    } else {
      await interaction.reply({
        content: 'No conversation history to clear.',
        ephemeral: true
      });
    }
  }
};
