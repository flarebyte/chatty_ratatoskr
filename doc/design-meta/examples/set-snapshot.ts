import type { KeyParams, KeyValueParams } from './common';
import type { RequestMetadata, ResponseEnvelope } from './envelope';

type SetSnapshotRequest = RequestMetadata & {
  key: KeyParams; // required: keyId, secureKeyId
  keyValueList: KeyValueParams[]; // required: keyId, secureKeyId
};

type SetSnapshotResponse = ResponseEnvelope<{
  key: KeyParams;
}>;

export interface SnapshotWriteApi {
  setSnapshot(request: SetSnapshotRequest): SetSnapshotResponse;
}
