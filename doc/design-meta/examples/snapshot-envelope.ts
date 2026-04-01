import type { KeyParams, KeyValueParams } from './common';

export type SnapshotEnvelope = {
  snapshotId: string;
  rootKey: KeyParams;
  version: string;
  created: string;
  keyValueList: KeyValueParams[];
};
