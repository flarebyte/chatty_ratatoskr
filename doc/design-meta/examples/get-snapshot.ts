import { KeyParams, KeyValueParams } from "./common";

type GetSnapshotRequest = {
  key: KeyParams;
};

type GetSnapshotResponse = {
  id: string;
  key: KeyParams;
  keyValueList: KeyValueParams[];
};
