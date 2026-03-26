import { KeyParams, KeyValueParams, OperationStatus } from "./common";

type SetKeyValueRequest = {
  rootKey: KeyParams;
  keyValueList: KeyValueParams[];
};
 
type SetKeyValueResponse = {
  id: string;
  rootKey: KeyParams;
  keyList: [KeyParams, OperationStatus][];
};
