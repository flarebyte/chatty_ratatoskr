/**
 * NodeKind should be extendible and likely be a string internally
 * but it should be validated against a supported list.
 */
export type NodeKind =
  | 'avatar'
  | 'comment'
  | 'emoticon'
  | 'image'
  | 'label'
  | 'language'
  | 'like'
  | 'note'
  | 'style'
  | 'table'
  | 'text'
  | 'thumbnail'
  | 'url'
  | 'user';

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
  created?: string;
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

//Redis key compatible
export const keyIdExamples = [
  'dashboard:52ffe570:note:c401c269:text',
  'dashboard:52ffe570:note:c401c269:comment:e0ee7775',
  'dashboard:52ffe570:note:c401c269:thumbnail:text',
  'dashboard:52ffe570:note:c401c269:like:user:_',
  'dashboard:52ffe570:note:c401c269:like:count',
  'dashboard:52ffe570:note:c401c269:comment:76f6d5e0:language:_',
];
