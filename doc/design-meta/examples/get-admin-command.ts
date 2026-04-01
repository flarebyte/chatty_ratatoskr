import type { Command } from './common';
import type { ResponseEnvelope } from './envelope';

type GetCommandRequest = {
  command: Command;
};

type GetCommandResponse = ResponseEnvelope<{
  command: Command;
  content: string;
}>;

export interface CommandReadApi {
  getCommand(request: GetCommandRequest): GetCommandResponse;
}

export const readCommands: Command[] = [
  {
    id: 'read-logs',
    comment: 'Read the logs',
    arguments: ['logs'],
  },
];
