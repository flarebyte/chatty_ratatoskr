import { KeyParams, KeyValueParams, OperationStatus } from "./common";

type SetKeyValueRequest = {
  keyValueList: KeyValueParams[];
};

type KeyStatus = {
  key: KeyParams;
  version: string;
  status: OperationStatus;
};

type SetKeyValueResponse = {
  id: string;
  keyList: KeyStatus[];
};
