export type KindOfTextNode =
  | "note"
  | "note/comment"
  | "note/label"
  | "note/like/user"
  | "note/avatar"
  | "note/emoticon"
  | "note/style"
  | "note/table"
  | "note/image"
  | "note/image/text"
  | "note/thumbnail"
  | "note/url";

export type OperationStatus = "ok" | "invalid" | "unauthorised";

export type KeyParams = {
  keyId: string;
  secureKeyId: string;
  localKeyId?: string;
  kind?: KindOfTextNode;
} 

export type KeyVersionParams = {
  key: KeyParams;
  version: string;
  updated: string;
};

export type KeyValueParams = {
  key: KeyParams;
  value: string;
  flags?: string[];
  language?: string;
  version?: string;
  updated?: string;
};
