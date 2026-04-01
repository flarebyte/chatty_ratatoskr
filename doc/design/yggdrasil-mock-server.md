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

## 02 Use Cases

Current goals and constraints captured from the draft inputs.

### 01 Product Goals

Why the server should exist and what it must support.

#### Use Cases

| description | usecase |
| --- | --- |
| Simulate a more complex backend store server that would be developed separately. | simulate-backend-server |
| Support a hierarchical text-tree store. | hierarchical-text-tree |
| Allow a text node to represent free text, numbers, booleans, dates, and similar scalar values. | typed-text-node |
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
http: {
	port: 8080
	maxPayloadSizeKB: 400
}

websocket: {
	supported: true
	path:      "/events"
	pingIntervalSeconds: 60
	pongWaitSeconds:     10
	maxMessageSizeKB:    4
}

sync: {
	snapshotBootstrap: true
	eventStreaming:    true
	optimisticWrites:  true
	clientLocalStore:  true
}

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

events: {
	connectionPath: "/events"
	source:         "server-generated-after-state-change"
}

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
	id: {
		format: "uuid[:8]"
	}

	keyKind: {
		rootWithId: ['dashboard']
		rootWithoutId: ['profile']
		childrenWithId: ["note", "thumbnail"]
		childrenWithoutId: ["text", "user"]
		maxLevels: 20
	}

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

	optionKind: {
		boolean: ["--archived", "--sensitive"]
	}

	statusKind: {
		boolean: ["ok", "invalid", "unauthorised", "outdated"]
	}

	secureKeyId: {
		statusAsKey: true
	}

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
| A node in the hierarchical text tree. | TextNode | entity.text-node | data,design,entity | Text node |

#### Entity Fields

| category | dart_kind | entity_name | field_description | field_name | field_title | go_kind | required | ts_kind |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| core | String | TextNode | Client side logical stream/grouping identifier (e.g., project/topic path). | localKeyId | Local Key ID | string | false | string |
| core | String | TextNode | Logical stream or grouping identifier (for example, a project or topic path). | keyId | Key ID | string | true | string |
| core | String | TextNode | Encrypted representation of `keyId`. | secureKeyId | Secure Key ID | string | true | string |
| core | List<String>? | TextNode | Optional metadata flags (for example '--pinned', '--archived', '--sensitive'). | options | Options | []string | false | string[] |
| core | String | TextNode | Logical type of the message (e.g., 'note'), set by the client. | kind | Kind | string | false | string |
| core | String? | TextNode | Optional ISO language code for the value content. | language | Language | *string | false | string |
| core | String | TextNode | The message content or a reference to external data. | value | Value | string | true | string |
| core | String | TextNode | Server-generated version (e.g., UUID) | version | Version | string | true | string |
| core | DateTime | TextNode | UTC timestamp when the message was created (server clock). | created | Created | time.Time | true | Date |
| core | DateTime | TextNode | UTC timestamp when the message was updated (server clock). | updated | Updated | time.Time | true | Date |

### 04 Sync And Persistence

How the logical protocol model maps to server and client storage.

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

### 05 Open Inconsistencies

Known draft mismatches that should be resolved before implementation hardens.

#### Current Draft Inconsistencies

The draft material is now closer to a coherent protocol, but a few design questions remain open:

- The project intent is now clear: Yggdrasil is a hierarchical key/value and snapshot protocol. The remaining question is whether the current names such as `TextNode` are specific enough or should be generalized to a broader Yggdrasil node vocabulary.
- The WebSocket draft now defines `/events` as the connection path, but the exact message envelope for subscribe, unsubscribe, heartbeat, and event delivery is still only implied by the TypeScript examples rather than defined as a strict protocol contract.
- The examples describe both snapshots and event stores, but retention, overwrite semantics, archive behaviour, and snapshot rehydration rules are still under-specified.
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
import type { KeyParams, OperationStatus, UserParams } from './common';

type Subscription = {
  id: string;
  user: UserParams;
  eventList: KeyParams[];
};

type EventResponse = {
  id: string;
  user: UserParams;
  eventList: [KeyParams, OperationStatus][];
};

export interface EventApi {
  registerUser(user: UserParams): [UserParams, OperationStatus];
  unregisterUser(user: UserParams): [UserParams, OperationStatus];
  subscribe(subscription: Subscription): EventResponse;
  unsubscribe(subscription: Subscription): EventResponse;
  sendEvent(user: UserParams, key: KeyParams): void;
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

