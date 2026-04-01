import type { KeyParams, KeyValueParams } from './common';
import type { KeyStatusResult, ResponseEnvelope } from './envelope';

type SetKeyValueRequest = {
  rootKey: KeyParams; // required: keyId, secureKeyId
  keyValueList: KeyValueParams[]; // required: keyId, secureKeyId
};

type SetKeyValueResponse = ResponseEnvelope<{
  rootKey: KeyParams; // required: keyId
  keyList: KeyStatusResult[]; // required: keyId, and the remaining fields may depend on success or failure.
}>;

export interface KeyValueWriteApi {
  setKeyValueList(request: SetKeyValueRequest): SetKeyValueResponse;
}
