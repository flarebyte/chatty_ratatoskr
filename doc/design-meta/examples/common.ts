/**
 * NodeKind should be extensible and likely be a string internally,
 * but it should be validated against a supported list.
 */
export type NodeKindExample =
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

export type OptionExample =
  | '--pinned'
  | '--archived'
  | '--sensitive'
  | '--personal'
  | '--anonymous'
  | '--masked';
export type OperationStatus = 'ok' | 'invalid' | 'unauthorised' | 'outdated';

export type KeyKind = {
  hierarchy: NodeKindExample[];
  language?: string;
};

export type KeyParams = {
  keyId?: string;
  // In production this should carry a signed integrity check for keyId.
  // In the mock server it may also be used to force a non-ok response.
  secureKeyId?: string;
  localKeyId?: string;
  // The server should derive kind from keyId and treat that derived value as
  // authoritative. A client may use a temporary kind locally before the server
  // responds, but any mismatch must be corrected by the server-derived value.
  kind?: KeyKind;
  // Used by clients and servers for optimistic sync checks so writes are based
  // on the latest known state rather than an older version.
  version?: string;
  // Useful for ordering nodes by first appearance, for example comments or chat.
  created?: string;
  // Useful for support/debugging after a version mismatch and for showing that a
  // node has been updated more recently than its creation time.
  updated?: string;
};

export type KeyValueParams = {
  key: KeyParams;
  value?: string;
  options?: OptionExample[];
};

export type UserParams = {
  key: KeyParams;
};

export type Command = {
  id: string;
  comment: string;
  arguments: string[];
};

// Redis-key-compatible examples.
export const keyIdExamples = [
  'dashboard:52ffe570:note:c401c269:text',
  'dashboard:52ffe570:note:c401c269:comment:e0ee7775',
  'dashboard:52ffe570:note:c401c269:thumbnail:text',
  'dashboard:52ffe570:note:c401c269:like:user:_',
  'dashboard:52ffe570:note:c401c269:like:count',
  'dashboard:52ffe570:note:c401c269:comment:76f6d5e0:language:_',
];
