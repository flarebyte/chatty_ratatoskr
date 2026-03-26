import { Command, OperationStatus } from "./common";


type CommandResponse = {
  command: Command;
  status: OperationStatus;
  message?: string;
}

type SetCommandsRequest = {
  commands: Command[];
};
 
type SetCommandsResponse = {
  id: string;
};
