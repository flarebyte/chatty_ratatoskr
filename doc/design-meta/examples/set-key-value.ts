import { KeyParams, KeyValueParams, OperationStatus } from "./common";

type SetKeyValueRequest = {
  rootKey: KeyParams;
  keyValueList: KeyValueParams[];
};

type KeyStatus = {
  key: KeyParams;
  version: string;
  status: OperationStatus;
};

type SetKeyValueResponse = {
  id: string;
  rootKey: KeyParams;
  keyList: KeyStatus[];
};
