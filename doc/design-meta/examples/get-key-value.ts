import { KeyValueParams, OperationStatus } from "./common";

type KeyParams = {
  keyId: string;
  secureKeyId: string;
};

type GetKeyValueRequest = {
  keyList: KeyParams[];
};

type KeyValueStatus = {
  keyId: string;
  status: OperationStatus;
  keyValue?: KeyValueParams;
};

type GetKeyValueResponse = {
  id: string;
  keyValueList: KeyValueStatus[];
};
