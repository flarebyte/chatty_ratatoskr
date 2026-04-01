import type { KeyParams, KeyValueParams } from './common';
import type { RequestMetadata, ResponseEnvelope } from './envelope';

type GetSnapshotRequest = RequestMetadata & {
  key: KeyParams; // required: keyId, secureKeyId
};

type GetSnapshotResponse = ResponseEnvelope<{
  key: KeyParams; // required: keyId, and the remaining fields may depend on success or failure.
  keyValueList: KeyValueParams[]; // required: keyId, and the remaining fields may depend on success or failure.
}>;

export interface SnapshotReadApi {
  getSnapshot(request: GetSnapshotRequest): GetSnapshotResponse;
}
