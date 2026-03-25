import { KeyParams, KeyValueParams, OperationStatus } from "./common";

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
