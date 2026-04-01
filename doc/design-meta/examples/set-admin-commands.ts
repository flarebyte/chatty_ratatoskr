import type { Command, OperationStatus } from './common';
import type { RequestMetadata, ResponseEnvelope } from './envelope';

type CommandStatus = {
  command: Command;
  status: OperationStatus;
  message?: string;
};

type SetCommandsRequest = RequestMetadata & {
  commands: Command[];
};

type SetCommandsResponse = ResponseEnvelope<{
  results: CommandStatus[];
}>;

export interface CommandWriteApi {
  setCommands(request: SetCommandsRequest): SetCommandsResponse;
}

export const writeCommands: Command[] = [
  {
    id: 'clear-state',
    comment: 'Clear all mock-server in-memory stores',
    arguments: ['clear-state'],
  },
  {
    id: 'delay-response',
    comment: 'Delay the response for testing purposes',
    arguments: ['delay', '--seconds=10'],
  },
  {
    id: 'read-logs',
    comment: 'Read mock-server logs for debugging',
    arguments: ['logs'],
  },
];
