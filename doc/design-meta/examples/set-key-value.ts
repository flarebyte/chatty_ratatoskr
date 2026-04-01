import type { KeyParams, KeyValueParams, OperationStatus } from './common';

type SetKeyValueRequest = {
  rootKey: KeyParams; // required: keyId, secureKeyId
  keyValueList: KeyValueParams[]; // required: keyId, secureKeyId
};

type SetKeyValueResponse = {
  id: string;
  rootKey: KeyParams; // required: keyId
  keyList: [KeyParams, OperationStatus][]; // required: keyId, and the remaining fields may depend on success or failure.
};

export interface KeyValueWriteApi {
  setKeyValueList(request: SetKeyValueRequest): SetKeyValueResponse;
}
