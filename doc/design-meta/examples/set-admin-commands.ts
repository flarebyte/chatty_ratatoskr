import type { Command, OperationStatus } from './common';

type CommandStatus = {
  command: Command;
  status: OperationStatus;
  message?: string;
};

type SetCommandsRequest = {
  commands: Command[];
};

type SetCommandsResponse = {
  id: string;
  results: CommandStatus[];
};

export interface CommandWriteApi {
  setCommands(request: SetCommandsRequest): SetCommandsResponse;
}
