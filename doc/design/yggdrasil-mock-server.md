# Yggdrasil Mock Server Design

First-pass design report for the lightweight HTTP and WebSocket mock server CLI.

## 01 Overview

Problem framing and the main architectural shape.

### 01 Intent

Why this server exists and how it should be used.

#### Logical Key Model

Yggdrasil should be treated as a protocol for hierarchical key/value, snapshot, and sync-oriented event operations.

The protocol should standardize the logical meaning of a key:
- hierarchical identity and parent/child structure
- kind and optional language or metadata facets
- versioning and conflict detection
- public and secure forms when needed

The protocol should not require a single serialized key format across products. A product may encode the same logical key as a Redis-friendly string, a structured Dart local-store record, or another storage-specific representation.

#### Spec-First Workflow

The source of truth for the design is intended to live in `doc/design-meta` as flyb input.

The examples under `doc/design-meta/examples` are design artefacts that capture the current protocol thinking:
- `config.cue` for draft protocol and schema configuration
- `.csv` files for use cases, action matrices, entities, and implementation constraints
- `.ts` files for request, response, and storage example shapes

#### Lightweight Go Mock Server

The project is shaping a lightweight CLI in Go that runs a mock Yggdrasil server over HTTP and, optionally, WebSocket.

Its purpose is to support testing and CI by simulating a production server that manages hierarchical key/value data, snapshots, and incremental updates while keeping behavior explicit, configurable, and cheap to run.

### 02 Transports

The protocol is primarily HTTP with optional WebSocket support.

#### HTTP Is The Primary Transport

The draft examples place most Yggdrasil interactions on a small HTTP surface: creating keys, reading and writing key/value data, and snapshot operations.

This aligns with the goal of a simple mock server that is easy to run locally and in CI.

#### WebSocket Support Is Optional

The protocol examples explicitly call out WebSocket support as optional rather than mandatory.

Current notes suggest a constrained event model with heartbeat support, bounded message sizes, and a preference for well-known identifiers for message safety.

A client may extend its active subscription set by sending additional `subscribe` messages for more root keys on the same connection. When a user unregisters or the connection is closed, all active subscriptions for that user or connection should be removed.

## 02 Use Cases

Current goals and constraints captured from the draft inputs.

### 01 Product Goals

Why the server should exist and what it must support.

#### Use Cases

| description | usecase |
| --- | --- |
| Simulate a more complex backend store server that would be developed separately. | simulate-backend-server |
| Support a hierarchical text-tree store. | hierarchical-text-tree |
| Allow a value node to carry free text, numbers, booleans, dates, and similar higher-level formats while remaining a protocol-level string. | typed-value-node |
| Reject writes when the submitted version is not based on the latest stored version. | optimistic-write-conflict |
| Run as a CLI that loads a CUE configuration file. | cue-config-cli |
| Allow behaviour and lightweight validation to be customised for different product needs. | customisable-behaviour |
| Run reliably inside CI pipelines. | ci-friendly |
| Be sufficient to test client code against a realistic mock server. | client-test-support |
| Provide a lightweight implementation of the evolving Yggdrasil protocol. | lightweight-yggdrasil |
| Keep WebSocket support optional because some clients may not support it. | optional-websocket |
| Limit WebSocket messages to well-known identifiers to reduce escaping and safety concerns. | constrained-websocket-messages |
| Use heartbeat ping/pong messages to keep WebSocket connections alive. | websocket-heartbeats |

### 02 Implementation Constraints

Preferred implementation choices for the Go CLI.

#### Implementation Constraints

| description | usecase |
| --- | --- |
| Write the CLI in Go. | go-cli |
| Use Cobra for CLI commands and argument parsing. | cobra-cli |
| Use CUE as the configuration source of truth. | cue-source-of-truth |
| Prefer early returns and guard clauses for errors. | guard-clauses |
| Keep functions small and single-purpose. | small-functions |
| Separate I/O from core logic. | separate-io |
| Use small structs to avoid long parameter lists. | tiny-structs |
| Replace boolean soup with named predicates. | named-predicates |
| Guarantee deterministic ordering in generated outputs. | deterministic-output |
| Use JSON for request and response payloads. | json-payloads |
| Prefer formats that are well supported across languages, especially Go, TypeScript, and Dart. | portable-formats |
| Keep `options` as lightweight representations of argument parsing, limited to a safe format. | safe-option-flags |
| Build the CLI for at least Linux and macOS. | target-platforms |
| Use `coder/websocket` for WebSocket support. | coder-websocket |
| Use `net/http` for the HTTP server. | net-http |

## 03 Protocol Surface

The current endpoint and payload sketches that inform the evolving Yggdrasil protocol.

### 01 Config Draft

CUE draft describing ports, endpoints, and schema constraints.

#### Draft Protocol Config

```cue
// Transport-level limits for the HTTP API.
http: {
	port: 8080
	maxPayloadSizeKB: 400
}

// WebSocket is optional and exists only for incremental updates.
websocket: {
	supported: true
	path:      "/events"
	pingIntervalSeconds: 60
	pongWaitSeconds:     10
	maxMessageSizeKB:    4
}

// Sync is based on authoritative snapshots plus incremental events.
sync: {
	snapshotBootstrap: true
	eventStreaming:    true
	optimisticWrites:  true
	clientLocalStore:  true
}

// Server and client may use different storage implementations while
// preserving the same logical key and version semantics.
storage: {
	server: {
		kind: "key-value"
		examples: ["redis", "other-compatible-store"]
	}

	client: {
		kind: "local-key-value"
		examples: ["dart-local-store"]
	}
}

// Production Yggdrasil endpoints.
endpoints: {
	create: {
		path: "/create"
		verb: "POST"
	}

	setKeyValueList: {
		path: "/node"
		verb: "PUT"
	}

	getKeyValueList: {
		path: "/node"
		verb: "GET"
	}

	getSnapshot: {
		path: "/snapshot"
		verb: "GET"
	}

	setSnapshot: {
		path: "/snapshot"
		verb: "PUT"
	}
}

// Events are emitted by the server after state changes.
events: {
	connectionPath: "/events"
	source:         "server-generated-after-state-change"
}

// Mock-server controls are intentionally outside the Yggdrasil protocol.
admin: {
	scope: "mock-server-only"

	endpoints: {
		commands: {
			path: "/admin/commands"
			verb: "PUT|GET"
		}
	}
}

schema: {
	// Logical identifiers remain product-specific in format, but the mock server
	// still validates enough structure to exercise client behavior.
	id: {
		format: "uuid[:8]"
	}

	// Key hierarchy rules define which kinds may appear at each level.
	// In normal operation, keyId is the source of truth and kind is derived from
	// the key schema with unique identifiers removed.
	keyKind: {
		rootWithId: ['dashboard']
		rootWithoutId: ['profile']
		childrenWithId: ["note", "thumbnail"]
		childrenWithoutId: ["text", "user"]
		maxLevels: 20
	}

	// The server should treat derived kind information as authoritative and
	// reject or correct any attempt to send corrupted kind data that conflicts
	// with keyId. This protects filtering logic and avoids security issues that
	// could arise from trusting client-supplied kind blindly.

	// The protocol standardizes key meaning, not one storage encoding.
	keyEncoding: {
		logicalModel: {
			productSpecificFormat: true
			preserveHierarchy:     true
			preserveIdentity:      true
			preserveVersioning:    true
		}

		serverExamples: [
			"dashboard:52ffe570:note:c401c269:text",
			"dashboard:52ffe570:note:c401c269:comment:e0ee7775",
		]

		clientExamples: [
			"{root:'dashboard', id:'52ffe570', path:['note','c401c269','text']}",
		]
	}

	// Options are lightweight string flags rather than a full dynamic schema.
	optionKind: {
		boolean: ["--archived", "--sensitive"]
	}

	// The protocol-level operation status vocabulary.
	statusKind: {
		boolean: ["ok", "invalid", "unauthorised", "outdated"]
	}

	// secureKeyId is mandatory in normal client requests and is used by servers
	// to detect whether the transmitted key identity has been corrupted or forged.
	//
	// Production intent:
	// - secureKeyId carries a signed or JWT-style check derived from keyId
	// - the server verifies it before accepting the request
	//
	// Mock-server intent:
	// - this CLI assumes the incoming value is otherwise valid
	// - when statusAsKey is true, secureKeyId may be used as a test hook to
	//   force a non-ok status such as invalid, unauthorised, or outdated
	// - this is intentionally a mock-only shortcut and not production behavior
	secureKeyId: {
		statusAsKey: true
	}

	// Value remains a protocol-level string even when it represents richer
	// formats such as dates, booleans, numbers, or encoded JSON.
	text: {
		maxCharLength: 1000
		checkVersion:  true
	}
}
```

### 02 Action Matrix

HTTP and WebSocket actions currently envisaged by the examples.

#### Protocol Actions

| action | description | endpoint | protocol | role | verb |
| --- | --- | --- | --- | --- | --- |
| create-new-keys | Create server-generated identifiers for requested nodes and their children. | /create | http | client | POST |
| set-key-value-list | Write one or more key/value entries under a root key. | /node | http | client | PUT |
| get-key-value-list | Read one or more key/value entries under a root key. | /node | http | client | GET |
| get-snapshot | Read the latest snapshot for a key. | /snapshot | http | client | GET |
| set-snapshot | Replace the current snapshot for a key. | /snapshot | http | client | PUT |
| receive-events | Open an event stream for subscriptions and server push updates. | /events | websocket | client | CONNECT |
| send-event | Emit a sync event to subscribed clients after a state change. | event-bus | internal | server | EMIT |

### 03 Entity Model

Current entity and field definitions used by the draft protocol.

#### Entities

| description | entity_name | entity_note_name | labels | title |
| --- | --- | --- | --- | --- |
| A hierarchical value node whose payload remains a protocol-level string. | ValueNode | entity.value-node | data,design,entity | Value node |

#### Entity Fields

| category | dart_kind | entity_name | field_description | field_name | field_title | go_kind | required | ts_kind |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| core | String | ValueNode | Client-side provisional identifier used while waiting for the official `keyId` returned by the server. Production servers should preserve it and may apply basic length or character checks for security, but that validation is outside the mock-server scope. | localKeyId | Local Key ID | string | false | string |
| core | String | ValueNode | Logical stream or grouping identifier (for example, a project or topic path). | keyId | Key ID | string | true | string |
| core | String | ValueNode | Mandatory integrity field derived from `keyId`. Production servers should verify it using a signed or JWT-style check; the mock server may also use it as a test hook to force a configured non-ok status. | secureKeyId | Secure Key ID | string | true | string |
| core | List<String>? | ValueNode | Optional metadata flags (for example '--pinned', '--archived', '--sensitive'). | options | Options | []string | false | string[] |
| core | String | ValueNode | Server-derived schema for the value node, inferred from `keyId` with unique identifiers removed. A client may use a temporary local hint, but the authoritative kind comes from the server-derived interpretation of `keyId`. | kind | Kind | string | false | string |
| core | String? | ValueNode | Optional ISO language code for the value content. | language | Language | *string | false | string |
| core | String | ValueNode | Protocol-level string payload. Higher-level formats such as free text, numbers, booleans, dates, or encoded JSON are still serialized as strings. | value | Value | string | true | string |
| core | String | ValueNode | Server-generated version used for optimistic sync checks so clients and servers can reject writes based on an older state. | version | Version | string | true | string |
| core | DateTime | ValueNode | UTC timestamp when the value node was created (server clock). Useful for ordering nodes by creation time, for example comments or chat messages. | created | Created | time.Time | true | Date |
| core | DateTime | ValueNode | UTC timestamp when the value node was last updated (server clock). Useful for support/debugging after version mismatches and for showing recent changes. | updated | Updated | time.Time | true | Date |

### 04 HTTP Status Rules

Minimal error and response semantics for client and mock-server interoperability.

#### HTTP Status Rules

| description | http_status | operation_status | scenario |
| --- | --- | --- | --- |
| Request completed successfully and returned the expected response payload. | 200 | ok | success |
| The request payload or key structure is invalid and the server rejects it. | 400 | invalid | invalid-key-or-payload |
| The client is not authorised to perform the requested operation. | 401 | unauthorised | unauthorised-access |
| The submitted version is older than the latest stored version and the write is rejected. | 409 | outdated | stale-write |
| The payload exceeds the configured HTTP request size limit. | 413 | invalid | payload-too-large |

### 05 Sync And Persistence

How the logical protocol model maps to server and client storage.

#### Event Rules

| description | rule | scope |
| --- | --- | --- |
| A `set` event must carry `rootKey`, `key`, `keyValue`, and `created`. | set-required-fields | event |
| A `snapshot-replaced` event must carry `rootKey`, `snapshotVersion`, and `created`. | snapshot-replaced-required-fields | event |
| `snapshot-replaced` is emitted after `setSnapshot`, not after every normal `set`. | snapshot-replaced-trigger | event |
| Archive is carried as record state, typically through `--archived`, rather than as a separate event operation. | archive-as-state | event |
| Unsubscribing a root key that is not currently subscribed is a no-op and does not raise an error. | unsubscribe-missing-key-noop | client |

#### Logical Keys Versus Storage Encoding

Yggdrasil should separate logical key semantics from storage encoding.

The logical model should define what a key means in the protocol. Storage-specific encodings should define how that key is serialized in a given implementation:
- Redis or similar server-side stores may prefer colon-delimited string keys
- a Dart local store may prefer a structured object or a different compact string format
- both encodings remain valid if they preserve the same hierarchy, identity, and version semantics

This separation lets the same protocol work across products without forcing every implementation into one storage-specific key format.

#### Sync Between Server And Client Stores

The protocol should support synchronisation between a production server and client-side local storage.

The intended model is:
- the server persists hierarchical key/value data, likely in a key-value store such as Redis or a similar backend
- the Dart client can store the same logical data in a local key-value store
- snapshots provide fast state rehydration
- events provide incremental updates after the last known state
- versions are used to reject stale writes and help clients converge on the latest state

This means sync is not a secondary implementation detail. It is a protocol concern that should shape snapshot semantics, event payloads, and conflict handling.

#### Sync Flow

| actor | description | step | transport |
| --- | --- | --- | --- |
| client | Fetch the latest authoritative snapshot for a root key. | fetch-snapshot | http |
| client | Store the snapshot in the local key-value store as the current baseline. | store-snapshot | local-store |
| client | Open the optional event stream for incremental updates. | open-event-stream | websocket |
| client | Apply `set` events as local upserts and treat `--archived` as record state rather than deletion. | apply-set-event | websocket |
| server | Emit `snapshot-replaced` after `setSnapshot` updates the authoritative snapshot for a root key. | emit-snapshot-replaced | websocket |
| client | When `snapshot-replaced` is received, discard the previous baseline for the root key and refetch the snapshot. | handle-snapshot-replaced | websocket |
| server | Reject writes whose submitted version is not based on the latest server version. | reject-stale-write | http |

#### Event Envelope

```ts
import type { KeyParams, KeyValueParams } from './common';

/**
 * Archive is resource state, not an event operation.
 * Archived records are still emitted as `set` events with `--archived`
 * present in the payload options when appropriate.
 *
 * `set` must carry `rootKey`, `key`, `keyValue`, and `created`.
 * `snapshot-replaced` must carry `rootKey`, `snapshotVersion`, and `created`.
 * `snapshot-replaced` is emitted after `setSnapshot`, not after every normal `set`.
 */
export type EventOperation = 'set' | 'snapshot-replaced';

export type EventEnvelope = {
  eventId: string;
  rootKey: KeyParams;
  operation: EventOperation;
  created: string;
  key?: KeyParams;
  keyValue?: KeyValueParams;
  snapshotVersion?: string;
};
```

#### Snapshot Envelope

```ts
import type { KeyParams, KeyValueParams } from './common';

export type SnapshotEnvelope = {
  snapshotId: string;
  rootKey: KeyParams;
  version: string;
  created: string;
  keyValueList: KeyValueParams[];
};
```

### 06 Open Inconsistencies

Known draft mismatches that should be resolved before implementation hardens.

#### Current Draft Inconsistencies

The draft material is now closer to a coherent protocol, but a few design questions remain open:

- The core sync rules are now defined, but retention policy and long-term snapshot storage semantics are still under-specified.
- Security is only sketched through `secureKeyId` and constrained event identifiers. Authentication, authorization, and trust boundaries are still intentionally unresolved in this draft.

## 04 Mock Server Administration

Test-only controls that are intentionally kept outside the Yggdrasil protocol.

### 01 Admin Surface

Operational controls for the mock server test harness.

#### Administration Is Outside The Protocol

Mock-server control operations such as clearing state, delaying responses, and reading logs are useful for testing, but they are not part of the Yggdrasil protocol itself.

These controls should remain on a separate administration surface so production clients do not depend on test-only capabilities such as `reset`-style actions. This reduces the risk of accidental exposure and keeps the protocol focused on domain data and synchronisation.

#### Admin Commands

| description | name |
| --- | --- |
| Clear all in-memory stores used by the mock server. | clear-state |
| Delay responses for timeout and retry testing. | delay-response |
| Read mock-server logs for debugging. | read-logs |

#### Get Admin Command Example

```ts
import type { Command, OperationStatus } from './common';

type GetCommandRequest = {
  command: Command;
};

type GetCommandResponse = {
  id: string;
  command: Command;
  status: OperationStatus;
  message?: string;
  content: string;
};

export interface CommandReadApi {
  getCommand(request: GetCommandRequest): GetCommandResponse;
}

export const readCommands: Command[] = [
  {
    id: 'read-logs',
    comment: 'Read the logs',
    arguments: ['logs'],
  },
];
```

#### Set Admin Commands Example

```ts
import type { Command, OperationStatus } from './common';

type CommandStatus = {
  command: Command;
  status: OperationStatus;
  message?: string;
};

type SetCommandsRequest = {
  commands: Command[];
};

type SetCommandsResponse = {
  id: string;
  results: CommandStatus[];
};

export interface CommandWriteApi {
  setCommands(request: SetCommandsRequest): SetCommandsResponse;
}

export const writeCommands: Command[] = [
  {
    id: 'clear-state',
    comment: 'Clear all mock-server in-memory stores',
    arguments: ['clear-state'],
  },
  {
    id: 'delay-response',
    comment: 'Delay the response for testing purposes',
    arguments: ['delay', '--seconds=10'],
  },
  {
    id: 'read-logs',
    comment: 'Read mock-server logs for debugging',
    arguments: ['logs'],
  },
];
```

## 05 TypeScript Examples

Reference examples that should be treated as protocol design inputs, not final implementation code.

### 01 Shared Types

Common primitives reused by the request/response examples.

#### Common TypeScript Types

```ts
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
```

### 02 HTTP APIs

REST-style request and response shapes.

#### Get Key Value Example

```ts
import type { KeyParams, KeyValueParams, OperationStatus } from './common';

type GetKeyValueRequest = {
  rootKey: KeyParams; // required: keyId, secureKeyId
  keyList: KeyParams[]; // required: keyId, secureKeyId
};

type GetKeyValueResponse = {
  id: string;
  rootKey: KeyParams; // provide keyId, and optionally all other fields except localKeyId
  keyValueList: [KeyValueParams, OperationStatus][]; // provide keyId, and optionally all other fields except localKeyId
};

export interface KeyValueReadApi {
  getKeyValueList(request: GetKeyValueRequest): GetKeyValueResponse;
}
```

#### Get Snapshot Example

```ts
import type { KeyParams, KeyValueParams } from './common';

type GetSnapshotRequest = {
  key: KeyParams; // required: keyId, secureKeyId
};

type GetSnapshotResponse = {
  id: string;
  key: KeyParams; // required: keyId, and the remaining fields may depend on success or failure.
  keyValueList: KeyValueParams[]; // required: keyId, and the remaining fields may depend on success or failure.
};

export interface SnapshotReadApi {
  getSnapshot(request: GetSnapshotRequest): GetSnapshotResponse;
}
```

#### Create New Keys Example

```ts
import type { KeyParams, NodeKindExample, OperationStatus } from './common';

type ChildParam = {
  localKeyId: string;
  expectedKind: NodeKindExample;
};

type NewKeyParams = {
  key: KeyParams;
  expectedKind: NodeKindExample;
  children: ChildParam[];
};

type SuggestedNewKeyParams = {
  key: KeyParams;
  status: OperationStatus;
  children: [KeyParams, OperationStatus][];
};

type NewKeysRequest = {
  rootKey: KeyParams;
  newKeys: NewKeyParams[];
};

type NewKeysResponse = {
  id: string;
  rootKey: KeyParams;
  newKeys: SuggestedNewKeyParams[];
};

export interface NewKeysApi {
  createNewKeys(request: NewKeysRequest): NewKeysResponse;
}
```

#### Set Key Value Example

```ts
import type { KeyParams, KeyValueParams, OperationStatus } from './common';

type SetKeyValueRequest = {
  rootKey: KeyParams; // required: keyId, secureKeyId
  keyValueList: KeyValueParams[]; // required: keyId, secureKeyId
};

type SetKeyValueResponse = {
  id: string;
  rootKey: KeyParams; // required: keyId
  keyList: [KeyParams, OperationStatus][]; // required: keyId, and the remaining fields may depend on success or failure.
};

export interface KeyValueWriteApi {
  setKeyValueList(request: SetKeyValueRequest): SetKeyValueResponse;
}
```

#### Set Snapshot Example

```ts
import type { KeyParams, KeyValueParams, OperationStatus } from './common';

type SetSnapshotRequest = {
  key: KeyParams; // required: keyId, secureKeyId
  keyValueList: KeyValueParams[]; // required: keyId, secureKeyId
};

type SetSnapshotResponse = {
  id: string;
  key: KeyParams;
  status: OperationStatus;
};

export interface SnapshotWriteApi {
  setSnapshot(request: SetSnapshotRequest): SetSnapshotResponse;
}
```

### 03 Event APIs

Optional WebSocket and event-store oriented examples.

#### Key Value Event Store Example

```ts
import type { KeyParams, KeyValueParams, OperationStatus } from './common';

type KeyValueEvent = {
  keyValue: KeyValueParams;
  status: OperationStatus;
};

export interface KeyValueEventStoreApi {
  addEvent(event: KeyValueEvent): void;
  getAllEvents(): KeyValueEvent[];
  getLastSuccessfulEventByKey(key: KeyParams): KeyValueEvent | undefined;
  getLastEventByKey(key: KeyParams): KeyValueEvent | undefined;
  clear(): void;
}
```

#### Receive Event Example

```ts
import type { OperationStatus, UserParams } from './common';
import type { EventEnvelope } from './event-envelope';
import type {
  ClientMessage,
  EventMessage,
  ServerMessage,
  SubscribeMessage,
  SubscribedMessage,
  UnsubscribeMessage,
  UnsubscribedMessage,
} from './websocket-messages';

type Subscription = {
  id: string;
  user: UserParams;
  rootKeys: string[];
};

type EventResponse = {
  id: string;
  user: UserParams;
  eventList: [EventEnvelope, OperationStatus][];
};

export interface EventApi {
  registerUser(user: UserParams): [UserParams, OperationStatus];
  // Unregistering a user clears all active subscriptions for that user.
  unregisterUser(user: UserParams): [UserParams, OperationStatus];
  subscribe(subscription: Subscription): EventResponse;
  // Unsubscribing a key that is not currently subscribed is a no-op and does not raise an error.
  unsubscribe(subscription: Subscription): EventResponse;
  receiveUserUpdate(user: UserParams): EventResponse;
}

export interface WebSocketEventApi {
  onClientMessage(message: ClientMessage): ServerMessage | EventMessage;
  // Repeated subscribe messages extend the active root-key set for the connection.
  // Duplicate root keys are normalized and the most recent entry wins.
  subscribe(message: SubscribeMessage): SubscribedMessage;
  unsubscribe(message: UnsubscribeMessage): UnsubscribedMessage;
  // Closing the connection clears all active subscriptions tied to that connection.
  disconnect(user: UserParams): void;
}

export type EventHandlingRule = {
  operation: 'set' | 'snapshot-replaced';
  clientAction: string;
};

export const eventHandlingRules: EventHandlingRule[] = [
  {
    operation: 'set',
    clientAction: 'Upsert the record locally. If options include --archived, treat archive as record state rather than a delete operation.',
  },
  {
    operation: 'snapshot-replaced',
    clientAction: 'This is emitted after setSnapshot. Refetch the authoritative snapshot for the root key and replace the local baseline.',
  },
];

export interface EventProducerApi {
  emit(event: EventEnvelope): void;
  receiveUserUpdate(user: UserParams): EventResponse;
}
```

#### Snapshot Event Store Example

```ts
import type { KeyParams, KeyValueParams } from './common';

type Snapshot = {
  key: KeyParams; //required: keyId, secureKeyId
  keyValueList: KeyValueParams[];
};

type SnapshotEvent = {
  snapshot: Snapshot;
};

export interface SnapshotEventStoreApi {
  addEvent(event: SnapshotEvent): void;
  getAllEvents(): SnapshotEvent[];
  getLastEventByKey(key: KeyParams): SnapshotEvent | undefined;
  clear(): void;
}
```

#### WebSocket Messages

```ts
import type { EventEnvelope } from './event-envelope';

export type SubscribeMessage = {
  kind: 'subscribe';
  // A client may send subscribe more than once to add further root keys
  // without reopening the WebSocket connection. Duplicate root keys are
  // normalized without error and the most recent subscription entry wins.
  rootKeys: string[];
};

export type UnsubscribeMessage = {
  kind: 'unsubscribe';
  rootKeys: string[];
};

export type PingMessage = {
  kind: 'ping';
};

export type ClientMessage =
  | SubscribeMessage
  | UnsubscribeMessage
  | PingMessage;

export type SubscribedMessage = {
  kind: 'subscribed';
  rootKeys: string[];
};

export type UnsubscribedMessage = {
  kind: 'unsubscribed';
  rootKeys: string[];
};

export type EventMessage = {
  kind: 'event';
  event: EventEnvelope;
};

export type PongMessage = {
  kind: 'pong';
};

export type ServerMessage =
  | SubscribedMessage
  | UnsubscribedMessage
  | EventMessage
  | PongMessage;
```

#### WebSocket Flow

| actor | description | step | transport |
| --- | --- | --- | --- |
| client | Open the WebSocket connection to the optional `/events` endpoint. | open-connection | websocket |
| client | Send a `subscribe` message with the root keys to watch. | subscribe-root-keys | websocket |
| client | Send another `subscribe` message later to add more root keys without reopening the connection. | extend-subscription | websocket |
| server | Normalize duplicate root keys without error and keep the most recent subscription entry. | deduplicate-subscription | internal |
| server | Reply with `subscribed` to confirm the active root-key subscriptions. | confirm-subscription | websocket |
| server | Send an `event` message containing the `EventEnvelope`. | receive-event | websocket |
| client-and-server | Use `ping` and `pong` messages to keep the connection alive. | ping-pong | websocket |
| client | Send `unsubscribe` when the client no longer wants updates for those root keys. | unsubscribe-root-keys | websocket |
| server | When a user unregisters or the connection closes, remove all active subscriptions tied to that user or connection. | clear-subscriptions-on-unregister | internal |

#### WebSocket Rules

| description | rule | scope |
| --- | --- | --- |
| Repeated `subscribe` messages extend the active root-key set for the same connection. | subscribe-extends-set | connection |
| Duplicate root keys are normalized without error. | duplicate-root-keys-normalized | connection |
| When duplicate root keys are received the most recent subscription entry wins. | most-recent-subscription-wins | connection |
| Unsubscribing a root key that is not currently subscribed is a no-op and does not raise an error. | unsubscribe-missing-key-noop | connection |
| When a user unregisters or the connection closes all active subscriptions tied to that user or connection are removed. | disconnect-clears-subscriptions | connection |

