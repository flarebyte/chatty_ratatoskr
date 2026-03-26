import { Command, OperationStatus } from "./common";

type GetCommandRequest = {
  command: Command;
};

type GetCommandResponse = {
  id: string;
  command: Command;
  status: OperationStatus;
  message?: string;
  content: string;
};
