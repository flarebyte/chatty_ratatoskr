import type { Command } from './common';
import type { RequestMetadata, ResponseEnvelope } from './envelope';

type GetCommandRequest = RequestMetadata & {
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
