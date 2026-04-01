import type { KeyParams, KeyValueParams } from './common';
import type { ResponseEnvelope } from './envelope';

type SetSnapshotRequest = {
  key: KeyParams; // required: keyId, secureKeyId
  keyValueList: KeyValueParams[]; // required: keyId, secureKeyId
};

type SetSnapshotResponse = ResponseEnvelope<{
  key: KeyParams;
}>;

export interface SnapshotWriteApi {
  setSnapshot(request: SetSnapshotRequest): SetSnapshotResponse;
}
