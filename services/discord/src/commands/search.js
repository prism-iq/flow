import { SlashCommandBuilder, EmbedBuilder } from 'discord.js';
import { queryRag } from '../utils/api.js';
import { logger } from '../utils/logger.js';

export const searchCommand = {
  data: new SlashCommandBuilder()
    .setName('search')
    .setDescription('Search the knowledge base')
    .addStringOption(option =>
      option
        .setName('query')
        .setDescription('Search query')
        .setRequired(true)
    )
    .addIntegerOption(option =>
      option
        .setName('results')
        .setDescription('Number of results (1-10)')
        .setMinValue(1)
        .setMaxValue(10)
        .setRequired(false)
    ),

  async execute(interaction) {
    const query = interaction.options.getString('query');
    const numResults = interaction.options.getInteger('results') ?? 5;

    await interaction.deferReply();

    try {
      const result = await queryRag(query, numResults);

      if (!result.documents || result.documents.length === 0) {
        await interaction.editReply('No results found.');
        return;
      }

      const embed = new EmbedBuilder()
        .setTitle(`Search Results: "${query}"`)
        .setColor(0x5865F2)
        .setTimestamp();

      for (let i = 0; i < Math.min(result.documents.length, 5); i++) {
        const doc = result.documents[i];
        const preview = doc.content.slice(0, 200) + (doc.content.length > 200 ? '...' : '');

        embed.addFields({
          name: `Result ${i + 1} (Score: ${doc.score?.toFixed(3) || 'N/A'})`,
          value: preview,
          inline: false
        });
      }

      await interaction.editReply({ embeds: [embed] });

    } catch (error) {
      logger.error(error, 'Search command error');
      await interaction.editReply('Search failed. Please try again.');
    }
  }
};
