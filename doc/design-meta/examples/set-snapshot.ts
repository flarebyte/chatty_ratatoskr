import type { KeyParams, KeyValueParams, OperationStatus } from './common';

type SetSnapshotRequest = {
  key: KeyParams; // required: keyId, secureKeyId
  keyValueList: KeyValueParams[]; // required: keyId, secureKeyId
};

type SetSnapshotResponse = {
  id: string;
  key: KeyParams;
  status: OperationStatus;
};

export interface SnapshotWriteApi {
  setSnapshot(request: SetSnapshotRequest): SetSnapshotResponse;
}
