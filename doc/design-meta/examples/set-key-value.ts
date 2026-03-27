import { KeyParams, KeyValueParams, OperationStatus } from "./common";

type SetKeyValueRequest = {
  rootKey: KeyParams; //required: keyId, secureKeyId
  keyValueList: KeyValueParams[]; //required: keyId, secureKeyId
};

type SetKeyValueResponse = {
  id: string;
  rootKey: KeyParams; //required: keyId
  keyList: [KeyParams, OperationStatus][]; //required: keyId, and the rest may be depend on success/failure.
};
