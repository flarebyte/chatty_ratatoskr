type KindOfTextNode =
  | "note"
  | "comment"
  | "label"
  | "like"
  | "avatar"
  | "emoticon"
  | "style"
  | "table"
  | "image/jpeg"
  | "altText"
  | "thumbnail"
  | "url";

type KeyValueParams = {
  keyId: string;
  secureKeyId: string;
  value: string;
  kind: KindOfTextNode;
  flags?: string[];
  language?: string;
};

type KeyParams = {
  keyId: string;
  secureKeyId: string;
};

type OperationStatus = "ok" | "invalid" | "unauthorised";

type SetKeyValueRequest = {
  keyValueList: KeyValueParams[];
};

type GetKeyValueListRequest = {
  keyList: KeyParams[];
};

type KeyStatus = {
  keyId: string;
  status: OperationStatus;
};

type KeyValueResponse = {
  keyList: KeyStatus[];
};
