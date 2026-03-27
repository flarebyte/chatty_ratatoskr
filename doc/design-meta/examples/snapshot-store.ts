import type { KeyParams, KeyValueParams } from './common';

type Snapshot = {
  key: KeyParams; //required: keyId, secureKeyId
  keyValueList: KeyValueParams[];
};

type SnapshotEvent = {
  snapshot: Snapshot;
};

export interface SnapshotEventStoreApi {
  addEvent(event: SnapshotEvent): void;
  getAllEvents(): SnapshotEvent[];
  getLastEventByKey(key: KeyParams): SnapshotEvent | undefined;
  clear(): void;
}
