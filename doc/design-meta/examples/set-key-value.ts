import { KeyValueParams, OperationStatus } from "./common";

type SetKeyValueRequest = {
  keyValueList: KeyValueParams[];
};

type KeyStatus = {
  keyId: string;
  status: OperationStatus;
};

type SetKeyValueResponse = {
  id: string;
  keyList: KeyStatus[];
};
