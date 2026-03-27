import { KeyParams, KeyValueParams, OperationStatus } from "./common";

type SnapshotEvent = {
  snapshot: Snapshot;
  status: OperationStatus;
};

interface SnapshotEventStoreApi {
  addEvent(event: SnapshotEvent): void;
  getAllEvents(): SnapshotEvent[];
  getLastSuccessfulEventByKey(key: KeyParams): SnapshotEvent | undefined;
  getLastEventByKey(key: KeyParams): SnapshotEvent | undefined;
  clear(): void;
}
