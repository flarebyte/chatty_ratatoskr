/**
 * NodeKind should be extendible and likely be a string internally
 * but it should be validated against a supported list.
 */
export type NodeKindExample =
  | "avatar"
  | "comment"
  | "emoticon"
  | "image"
  | "label"
  | "language"
  | "like"
  | "note"
  | "style"
  | "table"
  | "text"
  | "thumbnail"
  | "url"
  | "user";

export type OptionExample =
  | "--pinned"
  | "--archived"
  | "--sensitive"
  | "--personal"
  | "--anonymous"
  | "--masked";
export type OperationStatus = "ok" | "invalid" | "unauthorised" | "outdated";

export type KeyKind = {
  hierarchy: NodeKindExample[];
  language?: string;
};

export type KeyParams = {
  keyId?: string;
  secureKeyId?: string;
  localKeyId?: string;
  kind?: KeyKind;
  version?: string;
  created?: string;
  updated?: string;
};

export type KeyValueParams = {
  key: KeyParams;
  value?: string;
  options?: OptionExample[];
};

export type Command = {
  id: string;
  comment: string;
  arguments: string[];
};

//Redis key compatible
export const keyIdExamples = [
  "dashboard:52ffe570:note:c401c269:text",
  "dashboard:52ffe570:note:c401c269:comment:e0ee7775",
  "dashboard:52ffe570:note:c401c269:thumbnail:text",
  "dashboard:52ffe570:note:c401c269:like:user:_",
  "dashboard:52ffe570:note:c401c269:like:count",
  "dashboard:52ffe570:note:c401c269:comment:76f6d5e0:language:_",
];
