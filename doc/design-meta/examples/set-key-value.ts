import type { KeyParams, KeyValueParams } from './common';
import type {
  KeyStatusResult,
  RequestMetadata,
  ResponseEnvelope,
} from './envelope';

type SetKeyValueRequest = RequestMetadata & {
  rootKey: KeyParams; // required: keyId, secureKeyId
  keyValueList: KeyValueParams[]; // required: keyId, secureKeyId, processed independently and returned in request order
};

type SetKeyValueResponse = ResponseEnvelope<{
  rootKey: KeyParams; // required: keyId
  keyList: KeyStatusResult[]; // required: keyId, with one per-item status for each requested write
}>;

export interface KeyValueWriteApi {
  setKeyValueList(request: SetKeyValueRequest): SetKeyValueResponse;
}
