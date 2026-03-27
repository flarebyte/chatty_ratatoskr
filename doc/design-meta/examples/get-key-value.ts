import type { KeyParams, KeyValueParams, OperationStatus } from './common';

type GetKeyValueRequest = {
  rootKey: KeyParams; //required: keyId, secureKeyId
  keyList: KeyParams[]; //required: keyId, secureKeyId
};

type GetKeyValueResponse = {
  id: string;
  rootKey: KeyParams; //provide keyId, and optionally all other fields except localKeyId
  keyValueList: [KeyValueParams, OperationStatus][]; //provide keyId, and optionally all other fields except localKeyId
};

export interface KeyValueReadApi {
  getKeyValueList(request: GetKeyValueRequest): GetKeyValueResponse;
}
