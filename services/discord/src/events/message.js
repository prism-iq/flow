import { sendChat } from '../utils/api.js';
import { config } from '../config.js';
import { logger } from '../utils/logger.js';

export async function handleMessage(message) {
  if (message.author.bot) return;

  const mentionedBot = message.mentions.has(message.client.user);
  const isDM = !message.guild;

  if (!mentionedBot && !isDM) return;

  let content = message.content;
  if (mentionedBot) {
    content = content.replace(/<@!?\d+>/g, '').trim();
  }

  if (!content) return;

  try {
    await message.channel.sendTyping();

    const typingInterval = setInterval(() => {
      message.channel.sendTyping().catch(() => {});
    }, config.typingInterval);

    const userId = message.author.id;
    const conversationId = message.client.conversations.get(userId);

    const result = await sendChat(content, conversationId, false);

    clearInterval(typingInterval);

    if (result.conversation_id) {
      message.client.conversations.set(userId, result.conversation_id);
    }

    const responseText = result.message?.content || 'I could not generate a response.';

    if (responseText.length > config.maxMessageLength) {
      const chunks = splitMessage(responseText, config.maxMessageLength);
      for (const chunk of chunks) {
        await message.reply(chunk);
      }
    } else {
      await message.reply(responseText);
    }

  } catch (error) {
    logger.error(error, 'Message handler error');
    await message.reply('Sorry, I encountered an error. Please try again.');
  }
}

function splitMessage(text, maxLength) {
  const chunks = [];
  let remaining = text;

  while (remaining.length > 0) {
    if (remaining.length <= maxLength) {
      chunks.push(remaining);
      break;
    }

    let splitIndex = remaining.lastIndexOf('\n', maxLength);
    if (splitIndex === -1 || splitIndex < maxLength / 2) {
      splitIndex = remaining.lastIndexOf(' ', maxLength);
    }
    if (splitIndex === -1 || splitIndex < maxLength / 2) {
      splitIndex = maxLength;
    }

    chunks.push(remaining.slice(0, splitIndex));
    remaining = remaining.slice(splitIndex).trim();
  }

  return chunks;
}
