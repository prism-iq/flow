import { chatCommand } from './chat.js';
import { searchCommand } from './search.js';
import { clearCommand } from './clear.js';
import { statusCommand } from './status.js';

const commands = [
  chatCommand,
  searchCommand,
  clearCommand,
  statusCommand
];

export async function loadCommands(client) {
  for (const command of commands) {
    client.commands.set(command.data.name, command);
  }
}

export function getCommandsData() {
  return commands.map(c => c.data.toJSON());
}
