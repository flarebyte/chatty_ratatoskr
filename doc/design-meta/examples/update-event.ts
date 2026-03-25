import { KeyParams, KeyVersionParams, OperationStatus } from "./common";

type UpdateEvent = {
  key: KeyParams;
  status: OperationStatus;
  keyValue?: KeyVersionParams;
};

type OnUpdateEvent = {
  id: string;
  eventList: UpdateEvent[];
};
