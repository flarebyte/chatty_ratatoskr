import type { KeyParams, KeyValueParams } from './common';
import type { ResponseEnvelope } from './envelope';

type GetSnapshotRequest = {
  key: KeyParams; // required: keyId, secureKeyId
};

type GetSnapshotResponse = ResponseEnvelope<{
  key: KeyParams; // required: keyId, and the remaining fields may depend on success or failure.
  keyValueList: KeyValueParams[]; // required: keyId, and the remaining fields may depend on success or failure.
}>;

export interface SnapshotReadApi {
  getSnapshot(request: GetSnapshotRequest): GetSnapshotResponse;
}
