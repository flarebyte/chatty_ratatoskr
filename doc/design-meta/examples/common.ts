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

export type KeyValueParams = {
  keyId: string;
  secureKeyId: string;
  value: string;
  kind: KindOfTextNode;
  flags?: string[];
  language?: string;
  version?: string;
  updated?: string;
};