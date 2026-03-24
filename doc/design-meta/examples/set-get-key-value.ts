type SetKeyValueParams = {
  keyId: string;
  secureKeyId: string;
  value: string;
  kind: KindOfTextNode;
  flags?: string[];
  language?: string;
};

type SetStatus = "ok" | "invalid" | "unauthorised";

type SetKeyValueParamsState = {
  keyValue: SetKeyValueParams;
  status: SetStatus;
};

type SetKeyValueListRequest = {
  keyValueList: SetKeyValueParams[];
};

type SetKeyValueListRequestResponse = {
  keyValueList: SetKeyValueParamsState[];
};
