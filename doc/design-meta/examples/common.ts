/**
 * NodeKind should be extendible and likely be a string internally
 * but it should be validated against a supported list.
 */
export type NodeKind =
  | 'note'
  | 'comment'
  | 'label'
  | 'like'
  | 'user'
  | 'avatar'
  | 'emoticon'
  | 'style'
  | 'table'
  | 'image'
  | 'thumbnail'
  | 'url';

export type OperationStatus = 'ok' | 'invalid' | 'unauthorised' | 'outdated';

export type KeyKind = {
  hierarchy: NodeKind[];
  language?: string;
};

export type KeyParams = {
  keyId?: string;
  secureKeyId?: string;
  localKeyId?: string;
  kind?: KeyKind;
  version?: string;
  updated?: string;
};

export type KeyValueParams = {
  key: KeyParams;
  value?: string;
  flags?: string[];
};

export type Command = {
  id: string;
  comment: string;
  arguments: string[];
};
