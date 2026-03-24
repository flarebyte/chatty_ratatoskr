type KeyValueParams = {
  keyId: string;
  secureKeyId: string;
  value: string;
  kind: KindOfTextNode;
  flags?: string[];
  language?: string;
};

type SetStatus = "ok" | "invalid" | "unauthorised";

type KeyValueParamsState = {
  keyValue: KeyValueParams;
  status: SetStatus;
};

type SetKeyValueListRequest = {
  keyValueList: KeyValueParams[];
};

type GetKeyParams = {
  keyId: string;
  secureKeyId: string;
  kind: KindOfTextNode;
};

type GetKeyValueListRequest = {
  keyList: GetKeyParams[];
};

type KeyValueListRequestResponse = {
  keyValueList: KeyValueParamsState[];
};
