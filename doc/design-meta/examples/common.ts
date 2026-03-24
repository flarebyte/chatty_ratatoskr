export type KindOfTextNode =
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

export type OperationStatus = "ok" | "invalid" | "unauthorised";

export type KeyValueParams = {
  keyId: string;
  secureKeyId: string;
  value: string;
  kind: KindOfTextNode;
  flags?: string[];
  language?: string;
};