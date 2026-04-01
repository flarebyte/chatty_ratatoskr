import type { KeyParams, KeyValueParams } from './common';

export type SnapshotEnvelope = {
  snapshotId: string;
  rootKey: KeyParams;
  version: string;
  serverTimestamp: string;
  keyValueList: KeyValueParams[];
  authoritative: true;
};
