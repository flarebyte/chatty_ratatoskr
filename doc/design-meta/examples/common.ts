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
  // Optional ISO language code. Different nodes in the same document may use
  // different languages, so multilingual documents are supported.
  language?: string;
};

export type KeyParams = {
  keyId?: string;
  // In production this should carry a signed integrity check for keyId.
  // In the mock server it may also be used to force a non-ok response.
  secureKeyId?: string;
  // Client-side provisional identifier used while waiting for the official
  // keyId returned by the server. The server should preserve localKeyId rather
  // than rewriting it.
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

export type PrincipalParams = {
  key: KeyParams;
  // Product-specific top-level scope identifier, for example tenant or department.
  level1ScopeId?: string;
  // Product-specific fixed-depth group scope identifiers, for example team or region.
  level2ScopeIdList?: string[];
  // Product-specific principal identity, often a user but it may also be
  // called member, subscriber, or another product term.
  principalId?: string;
};

export type Command = {
  id: string;
  comment: string;
  arguments: string[];
};

// Redis-key-compatible examples. These include:
// - scoped document and node examples
// - the reserved current-principal placeholder `user:_`
// - the derived aggregate leaf `count`
export const keyIdExamples = [
  'tenant:acme:group:editorial:dashboard:d1',
  'tenant:acme:group:editorial:user:u42:dashboard:d1:note:n7:text',
  'tenant:acme:group:editorial:user:u42:dashboard:d1:note:n7:comment:c3:text',
  'tenant:acme:group:editorial:dashboard:d1:note:n7:like:user:_',
  'tenant:acme:group:editorial:dashboard:d1:note:n7:like:count',
  'department:news:region:emea:member:m17:dashboard:d1:note:n7:language:_',
];
