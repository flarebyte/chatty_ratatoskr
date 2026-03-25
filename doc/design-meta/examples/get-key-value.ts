import { KeyParams, KeyValueParams, OperationStatus } from "./common";

type GetKeyValueRequest = {
  rootKey: KeyParams;
  keyList: KeyParams[];
};

type KeyValueStatus = {
  keyId: string;
  status: OperationStatus;
  keyValue?: KeyValueParams;
};

type GetKeyValueResponse = {
  id: string;
  rootKey: KeyParams;
  keyValueList: KeyValueStatus[];
};
