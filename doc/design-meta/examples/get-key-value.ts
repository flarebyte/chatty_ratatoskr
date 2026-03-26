import { KeyParams, KeyValueParams, OperationStatus } from "./common";

type GetKeyValueRequest = {
  rootKey: KeyParams;
  keyList: KeyParams[];
};

type GetKeyValueResponse = {
  id: string;
  rootKey: KeyParams;
  keyValueList: [KeyValueParams, OperationStatus][];
};
