import { SlashCommandBuilder, EmbedBuilder } from 'discord.js';
import { request } from 'undici';
import { config } from '../config.js';
import { logger } from '../utils/logger.js';

export const statusCommand = {
  data: new SlashCommandBuilder()
    .setName('status')
    .setDescription('Check Flow service status'),

  async execute(interaction) {
    await interaction.deferReply();

    const services = [
      { name: 'Go Orchestrator', url: `${config.services.goOrchestrator}/api/v1/health` },
      { name: 'Node Service', url: `${config.services.nodeService}/health` },
    ];

    const statuses = [];

    for (const service of services) {
      try {
        const start = Date.now();
        const response = await request(service.url, { method: 'GET' });
        const latency = Date.now() - start;

        statuses.push({
          name: service.name,
          status: response.statusCode === 200 ? '✅ Online' : '⚠️ Degraded',
          latency: `${latency}ms`
        });
      } catch {
        statuses.push({
          name: service.name,
          status: '❌ Offline',
          latency: 'N/A'
        });
      }
    }

    const embed = new EmbedBuilder()
      .setTitle('Flow Service Status')
      .setColor(0x5865F2)
      .setTimestamp();

    for (const s of statuses) {
      embed.addFields({
        name: s.name,
        value: `Status: ${s.status}\nLatency: ${s.latency}`,
        inline: true
      });
    }

    embed.addFields({
      name: 'Active Conversations',
      value: `${interaction.client.conversations.size}`,
      inline: true
    });

    await interaction.editReply({ embeds: [embed] });
  }
};
