import { KeyParams, OperationStatus } from "./common";

type OnUpdateEvent = {
  id: string;
  eventList: [KeyParams, OperationStatus][];
};
