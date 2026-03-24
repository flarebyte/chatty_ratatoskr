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

type SetKeyValueListRequestResponse = {
  keyList: KeyParamsState[];
};

type KeyParams = {
  keyId: string;
  secureKeyId: string;
  kind: KindOfTextNode;
};

type KeyParamsState = {
  key: KeyParams;
  status: SetStatus;
};

type GetKeyValueListRequest = {
  keyList: KeyParams[];
};

type GetKeyValueListRequestResponse = {
  keyValueList: KeyValueParamsState[];
};
