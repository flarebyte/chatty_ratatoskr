import type { KeyParams } from './common';
import type {
  KeyValueStatusResult,
  RequestMetadata,
  ResponseEnvelope,
} from './envelope';

type GetKeyValueRequest = RequestMetadata & {
  rootKey: KeyParams; // required: keyId, secureKeyId
  keyList: KeyParams[]; // required: keyId, secureKeyId
};

type GetKeyValueResponse = ResponseEnvelope<{
  rootKey: KeyParams; // provide keyId, and optionally all other fields except localKeyId
  keyValueList: KeyValueStatusResult[]; // provide keyId, and optionally all other fields except localKeyId
}>;

export interface KeyValueReadApi {
  getKeyValueList(request: GetKeyValueRequest): GetKeyValueResponse;
}
