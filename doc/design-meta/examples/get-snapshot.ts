import { KeyParams, KeyValueParams } from "./common";

type GetSnapshotRequest = {
  key: KeyParams; //required: keyId, secureKeyId
};

type GetSnapshotResponse = {
  id: string;
  key: KeyParams;
  keyValueList: KeyValueParams[];
};
