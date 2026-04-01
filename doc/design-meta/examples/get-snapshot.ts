import type { KeyParams, KeyValueParams } from './common';

type GetSnapshotRequest = {
  key: KeyParams; // required: keyId, secureKeyId
};

type GetSnapshotResponse = {
  id: string;
  key: KeyParams; // required: keyId, and the remaining fields may depend on success or failure.
  keyValueList: KeyValueParams[]; // required: keyId, and the remaining fields may depend on success or failure.
};

export interface SnapshotReadApi {
  getSnapshot(request: GetSnapshotRequest): GetSnapshotResponse;
}
