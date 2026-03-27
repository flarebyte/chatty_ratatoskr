import type { KeyParams, KeyValueParams } from './common';

type GetSnapshotRequest = {
  key: KeyParams; //required: keyId, secureKeyId
};

type GetSnapshotResponse = {
  id: string;
  key: KeyParams; //required: keyId, and the rest may be depend on success/failure.
  keyValueList: KeyValueParams[]; //required: keyId, and the rest may be depend on success/failure.
};

export interface SnapshotReadApi {
  getSnapshot(request: GetSnapshotRequest): GetSnapshotResponse;
}
