import type { KeyParams, KeyValueParams, OperationStatus } from './common';

type KeyValueEvent = {
  keyValue: KeyValueParams;
  status: OperationStatus;
};

export interface KeyValueEventStoreApi {
  addEvent(event: KeyValueEvent): void;
  getAllEvents(): KeyValueEvent[];
  getLastSuccessfulEventByKey(key: KeyParams): KeyValueEvent | undefined;
  getLastEventByKey(key: KeyParams): KeyValueEvent | undefined;
  clear(): void;
}
