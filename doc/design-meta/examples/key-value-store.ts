import { KeyParams, KeyValueParams, OperationStatus } from "./common";

type KeyValueEvent = {
  keyValue: KeyValueParams;
  status: OperationStatus;
};

interface KeyValueEventStoreApi {
  addEvent(event: KeyValueEvent): void;
  getAllEvents(): KeyValueEvent[];
  getLastSuccessfulEventByKey(key: KeyParams): KeyValueEvent | undefined;
  getLastEventByKey(key: KeyParams): KeyValueEvent | undefined;
  clear(): void;
}
