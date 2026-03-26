import { KeyParams, KeyValueParams, OperationStatus } from "./common";

type SetSnapshotRequest = {
  key: KeyParams;
  keyValueList: KeyValueParams[];
};
 
type SetKeyValueResponse = {
  id: string;
  key: KeyParams;
  status: OperationStatus;
};
