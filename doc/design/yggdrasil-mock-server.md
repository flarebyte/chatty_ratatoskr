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
| Support a hierarchical value-tree store addressed by logical keys. | hierarchical-value-tree |
| Allow a value node to carry free text, numbers, booleans, dates, and similar higher-level formats while remaining a protocol-level string. | typed-value-node |
| Reject writes when the submitted version is not based on the latest stored version. | optimistic-write-conflict |
| Allow clients to bootstrap local state from an authoritative snapshot for a root key. | snapshot-bootstrap |
| Allow clients to stay in sync through incremental events after the snapshot baseline. | incremental-sync |
| Allow clients to keep the same logical data in a local key-value store for sync and offline-friendly flows. | client-local-store |
| Allow clients to use provisional localKeyId values until the server returns the official keyId. | provisional-local-keys |
| Derive kind from keyId on the server so filtering and validation use the authoritative key schema. | server-derived-kind |
| Allow different value nodes in the same document to use different ISO language codes. | multilingual-nodes |
| Require secureKeyId as an integrity check so corrupted or forged key identities can be detected. | secure-key-integrity |
| Allow the mock server to force protocol status responses through test-oriented configuration hooks. | mock-status-forcing |
| Be sufficient to test client code against a realistic mock server. | client-test-support |
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
| Prefer a lightweight single-binary server suitable for local development and CI. | single-binary |
| Use in-memory state for the mock server rather than production-grade persistence. | in-memory-mock-state |
| Keep mock-server administration outside the protocol-facing Yggdrasil surface. | separate-admin-surface |

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
	// Subscriptions are allowed only for predefined document-level roots.
	// A root subscription covers that root and all readable descendants.
	subscription: {
		predefinedRootsOnly: true
		includeDescendants:  true
		allowSubtreeRoots:   false
		invalidStatusForNonAllowedRoot: "invalid"
		// Overlapping configured roots should not exist in practice. The mock
		// server may warn, while stricter production validation may reject them.
		warnOnOverlappingRoots: true
		allowedRoots:          ["dashboard", "profile"]
	}
}

// Sync is based on authoritative snapshots plus incremental events.
sync: {
	snapshotBootstrap: true
	eventStreaming:    true
	optimisticWrites:  true
	clientLocalStore:  true
	// A valid root subscription may still omit unreadable descendants instead of
	// failing the whole subscription.
	filterUnreadableDescendants: true
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
	// Event subscriptions are rooted at predefined top-level documents rather
	// than arbitrary nested subtrees.
	rootSubscriptionModel: "predefined-roots-with-descendants"
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

	// Key hierarchy is positional for v1. Each level defines which labels are
	// accepted there, what kind of value follows, and whether special aliases
	// such as `_` may be resolved by the server. This keeps the mock-server
	// schema practical without introducing a more advanced rule engine.
	valueKind: ["id", "leaf", "derived"]

	keyPart: {
		maxLevels: 9

		level1: {
			labels:    ["tenant", "department"]
			valueKind: "id"
			aliases:   ["_"]
		}

		level2: {
			labels:    ["group", "team", "region"]
			valueKind: "id"
			aliases:   ["_"]
			optional:  true
		}

		level3: {
			labels: ["user", "member", "subscriber", "dashboard", "profile"]
			valueKindByLabel: {
				user:       "id"
				member:     "id"
				subscriber: "id"
				dashboard:  "id"
				profile:    "id"
			}
			aliasesByLabel: {
				user:       ["_"]
				member:     ["_"]
				subscriber: ["_"]
			}
		}

		level4: {
			labels: ["dashboard", "profile", "note", "comment", "thumbnail", "language", "like"]
			valueKindByLabel: {
				dashboard: "id"
				profile:   "id"
				note:      "id"
				comment:   "id"
				thumbnail: "leaf"
				language:  "leaf"
				like:      "leaf"
			}
			optional: true
		}

		level5: {
			labels: ["note", "comment", "thumbnail", "language", "like", "text", "count", "user", "member", "subscriber"]
			valueKindByLabel: {
				note:       "id"
				comment:    "id"
				thumbnail:  "leaf"
				language:   "leaf"
				like:       "leaf"
				text:       "leaf"
				count:      "derived"
				user:       "id"
				member:     "id"
				subscriber: "id"
			}
			aliasesByLabel: {
				user:       ["_"]
				member:     ["_"]
				subscriber: ["_"]
			}
			optional: true
		}

		level6: {
			labels: ["comment", "like", "text", "language", "count", "user", "member", "subscriber"]
			valueKindByLabel: {
				comment:    "id"
				like:       "leaf"
				text:       "leaf"
				language:   "leaf"
				count:      "derived"
				user:       "id"
				member:     "id"
				subscriber: "id"
			}
			aliasesByLabel: {
				user:       ["_"]
				member:     ["_"]
				subscriber: ["_"]
			}
			optional: true
		}

		level7: {
			labels: ["user", "member", "subscriber", "text", "language", "count"]
			valueKindByLabel: {
				user:       "id"
				member:     "id"
				subscriber: "id"
				text:       "leaf"
				language:   "leaf"
				count:      "derived"
			}
			aliasesByLabel: {
				user:       ["_"]
				member:     ["_"]
				subscriber: ["_"]
			}
			optional: true
		}

		level8: {
			labels: ["text", "language", "count"]
			valueKindByLabel: {
				text:     "leaf"
				language: "leaf"
				count:    "derived"
			}
			optional: true
		}

		level9: {
			labels: ["text", "language", "count"]
			valueKindByLabel: {
				text:     "leaf"
				language: "leaf"
				count:    "derived"
			}
			optional: true
		}
	}

	// The server should treat derived kind information as authoritative and
	// reject or correct any attempt to send corrupted kind data that conflicts
	// with keyId. This protects filtering logic and avoids security issues that
	// could arise from trusting client-supplied kind blindly.
	//
	// Access scope should also be structurally visible in keyId. Group hierarchy
	// is fixed-depth rather than recursive so access can be validated cheaply and
	// consistently from the key schema. The exact labels are product-specific:
	// for example tenant/group/user, department/region/member, or similar.
	accessScope: {
		visibleInKeyId: true
		scopeLevels:    ["tenant", "group", "user"]
		scopeAliases: {
			level1: ["tenant", "department"]
			level2: ["group", "team", "region"]
			principal: ["user", "member", "subscriber"]
		}
		examples: [
			"tenant:t8f3a1c2:group:g4b7d9e1:user:u42c91ab:dashboard:d1e52f07:note:n7c401c2:text",
			"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07",
			"department:d9a8c102:region:r6be41f0:member:m17aa9f3:dashboard:d1e52f07:note:n7c401c2:text",
			"department:d9a8c102:dashboard:d1e52f07",
		]
		recursiveGroups: false
		principalScopedKeysRequirePrincipalId: true
		groupScopedKeysRequirePath:            true
		invalidStatusForMalformedKey: "invalid"
		unauthorisedStatusForDeniedAccess: "unauthorised"
		currentPrincipalPlaceholders: ["principal:_", "user:_"]
	}

	// The protocol standardizes key meaning, not one storage encoding.
	keyEncoding: {
		logicalModel: {
			productSpecificFormat: true
			preserveHierarchy:     true
			preserveIdentity:      true
			preserveVersioning:    true
			preserveAccessScope:   true
			serverResolvesCurrentPrincipal: true
			supportsDerivedLeaves:          true
		}

		serverExamples: [
			"dashboard:52ffe570:note:c401c269:text",
			"dashboard:52ffe570:note:c401c269:comment:e0ee7775",
		]

		clientExamples: [
			"{root:'dashboard', id:'52ffe570', path:['note','c401c269','text']}",
		]

		reservedLeaves: {
			currentPrincipal: ["principal:_", "user:_"]
			derivedAggregate: ["count"]
		}
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

#### Key Parsing Rules

| description | rule | scope |
| --- | --- | --- |
| When a label has `valueKind` or `valueKindByLabel` set to `id`, the parser should consume the following token as the label value. | id-consumes-next-token | key |
| When a label has `valueKind` or `valueKindByLabel` set to `leaf`, that label is terminal for the key path. | leaf-terminates | key |
| When a label has `valueKind` or `valueKindByLabel` set to `derived`, that label is terminal and server-managed. | derived-terminates | key |
| The `_` placeholder should be accepted only for labels that explicitly allow it through `aliases` or `aliasesByLabel`. | aliases-are-label-specific | key |
| When `valueKindByLabel` exists for a level, the server should use the per-label rule rather than assuming one behavior for the whole level. | label-behaviour-overrides-level-default | key |

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
| core | String | ValueNode | Access scope is structurally visible in `keyId`. User-scoped nodes must include user identity in the key, and group-scoped nodes must include a fixed-depth group path such as tenant or team. | accessScope | Access Scope | string | true | string |
| core | String | ValueNode | Mandatory integrity field derived from `keyId`. Production servers should verify it using a signed or JWT-style check; the mock server may also use it as a test hook to force a configured non-ok status. | secureKeyId | Secure Key ID | string | true | string |
| core | List<String>? | ValueNode | Optional metadata flags (for example '--pinned', '--archived', '--sensitive'). | options | Options | []string | false | string[] |
| core | String | ValueNode | Server-derived schema for the value node, inferred from `keyId` with unique identifiers removed. A client may use a temporary local hint, but the authoritative kind comes from the server-derived interpretation of `keyId`. | kind | Kind | string | false | string |
| core | String? | ValueNode | Optional ISO language code for the value content. Different value nodes in the same document may use different languages. | language | Language | *string | false | string |
| core | String | ValueNode | Protocol-level string payload. Higher-level formats such as free text, numbers, booleans, dates, or encoded JSON are still serialized as strings. | value | Value | string | true | string |
| core | String | ValueNode | Server-generated version used for optimistic sync checks so clients and servers can reject writes based on an older state. | version | Version | string | true | string |
| core | DateTime | ValueNode | UTC timestamp when the value node was created (server clock). Useful for ordering nodes by creation time, for example comments or chat messages. | created | Created | time.Time | true | Date |
| core | DateTime | ValueNode | UTC timestamp when the value node was last updated (server clock). Useful for support/debugging after version mismatches and for showing recent changes. | updated | Updated | time.Time | true | Date |

### 04 HTTP Status Rules

Minimal error and response semantics for client and mock-server interoperability.

#### Batch Write Rules

| description | rule | scope |
| --- | --- | --- |
| `createNewKeys` and `setKeyValueList` should process items independently rather than failing the entire request when one item fails. | best-effort-processing | collection-write |
| Batch write responses should return an item-level status for every requested item so clients can distinguish successful and unsuccessful writes. | per-item-status-required | collection-write |
| Batch write responses should preserve the request order of items to keep client-side correlation simple. | request-order-preserved | collection-write |
| The top-level response `status` should describe whether the batch request was accepted and processed, not whether every item succeeded. | top-level-status-is-transport-level | collection-write |
| The protocol should not imply transactional rollback across items; successful item writes remain valid even if other items in the same request fail. | no-implicit-rollback | collection-write |

#### Envelope Rules

| description | rule | scope |
| --- | --- | --- |
| HTTP responses should use a shared envelope with `id`, top-level `status`, optional `message`, and a shallow `data` payload. | response-envelope-minimal | response |
| Requests should remain domain-shaped and should not be wrapped in an additional envelope unless correlation metadata becomes necessary later. | requests-remain-flat | request |
| Operations with a single authoritative outcome should rely on the top-level `status` rather than duplicating the same status inside `data`. | single-outcome-uses-top-level-status | response |
| Collection operations may include per-item statuses inside `data` while keeping the top-level envelope consistent. | collection-statuses-live-in-data | collection |
| The envelope should stay shallow so the same contract maps cleanly to a Go struct with embedded or nested payload data. | go-friendly-shape | implementation |

#### HTTP Response Shape

| description | rule | scope |
| --- | --- | --- |
| HTTP responses should include a top-level `id` for request tracing and correlation. | top-level-id | response |
| HTTP responses should include a top-level `status` when the operation has a single authoritative outcome. | top-level-status | response |
| HTTP responses should carry the domain payload under a shallow top-level `data` field. | top-level-data | response |
| `message` is optional and should be used when human-readable context helps explain failures or unusual outcomes. | optional-message | response |
| Collection operations may return per-item statuses in addition to top-level fields. | item-status-for-collections | response |
| Structured fields should carry protocol meaning first; `message` is supplementary rather than the primary machine-readable contract. | structured-over-freeform | response |

#### HTTP Status Rules

| description | http_status | operation_status | scenario |
| --- | --- | --- | --- |
| Request completed successfully and returned the expected response payload. | 200 | ok | success |
| The request payload or key structure is invalid and the server rejects it. | 400 | invalid | invalid-key-or-payload |
| The client is not authorised to perform the requested operation. | 401 | unauthorised | unauthorised-access |
| The submitted version is older than the latest stored version and the write is rejected. | 409 | outdated | stale-write |
| The payload exceeds the configured HTTP request size limit. | 413 | invalid | payload-too-large |

#### Partial Success Rules

| description | rule | scope |
| --- | --- | --- |
| List-style operations may contain a mix of successful and unsuccessful item-level results. | list-operations-may-mix | collection |
| For mixed results, item-level statuses are authoritative for each returned item. | per-item-status-authoritative | collection |
| The top-level status describes whether the request itself was processed, not whether every item succeeded. | top-level-status-reflects-request | collection |
| A response may still be top-level `ok` when the request was processed successfully but some returned items carry non-ok statuses. | ok-with-item-failures | collection |
| Clients must inspect item-level statuses for collection operations rather than relying only on the top-level status. | client-must-inspect-items | collection |
| Collection responses should preserve request order so clients can correlate item-level outcomes without additional matching logic. | responses-follow-request-order | collection |

#### Request Correlation

| description | rule | scope |
| --- | --- | --- |
| HTTP requests may include an optional top-level `id` for correlation and tracing. | request-id-optional | request |
| When a request includes `id`, the server should echo the same `id` in the response envelope. | response-echoes-request-id | response |
| When a request omits `id`, the server should generate a response `id`. | server-generates-response-id | response |
| Request correlation should not require an additional envelope; `id` remains a shallow optional field on the request itself. | flat-request-shape | request |
| The same request-correlation rule should apply to both the mock server and a production server so tests can rely on stable tracing behavior. | mock-and-production-consistent | implementation |

### 05 Trust Model

Which fields are trusted, which are hints, and which are server-derived.

#### Access Evaluation Rules

| description | operation | rule |
| --- | --- | --- |
| A level-1 scoped key is readable when the authenticated principal belongs to the same configured top-level scope identified in `keyId`, for example `tenant` or `department`. | read | level-1-scope-read |
| A level-1 scoped key is writable only when the authenticated principal has write permission for that configured top-level scope. | write | level-1-scope-write |
| A level-2 scoped key is readable when the authenticated principal belongs to the same configured level-1 and level-2 scopes identified in `keyId`, for example `tenant/team` or `department/region`. | read | level-2-scope-read |
| A level-2 scoped key is writable only when the authenticated principal has write permission for that configured level-2 scope. | write | level-2-scope-write |
| A principal-scoped key is readable when the authenticated principal id matches the principal identity encoded in `keyId`, regardless of whether the product calls it `user`, `member`, `subscriber`, or another term. | read | principal-scope-read |
| A principal-scoped key is writable only when the authenticated principal id matches the principal identity encoded in `keyId` or another explicit server policy allows it. | write | principal-scope-write |
| If the access scope implied by `keyId` is malformed or missing required ownership segments, the request should return `invalid`. | validation | invalid-key-scope |
| If `keyId` is structurally valid but the authenticated user lacks read or write permission for that scope, the request should return `unauthorised`. | authorization | unauthorised-access |

#### Access Scope Rules

| description | rule | scope |
| --- | --- | --- |
| Access scope should be visible in the authoritative `keyId` structure rather than hidden in separate metadata. | ownership-visible-in-key | access |
| If a key is readable or writable only by one principal, that principal identity must be present in `keyId`. The concrete label may be `user`, `member`, `subscriber`, or another product-specific term. | principal-scope-requires-principal-id | access |
| If a key is readable or writable by a group scope, the group path must be present in `keyId`. | group-scope-requires-group-path | access |
| Access hierarchy should use predefined fixed levels configured per product, for example `tenant/group/user`, `department/region/member`, or similar, rather than recursive arbitrary nesting. | fixed-scope-levels | access |
| The protocol requires structural visibility and fixed-depth scope, but the exact scope labels remain product-specific and configurable. | configurable-scope-labels | access |
| Access checks should be derived from the authoritative key structure and its server-derived kind, not from client-supplied hints. | kind-maps-to-scope | access |
| A valid root subscription may silently omit descendants that the user is not allowed to read. | subscription-filters-unreadable-descendants | subscription |

#### Derived Key Rules

| description | rule | scope |
| --- | --- | --- |
| `principal:_` or a product-specific alias such as `user:_` is a reserved placeholder meaning the current authenticated principal. | current-principal-placeholder | key |
| The server should resolve the current-principal placeholder to the authenticated principal identity before applying access checks and storage rules. | server-resolves-placeholder | key |
| If a request uses the current-principal placeholder without an authenticated principal context, the request should return `unauthorised`. | placeholder-requires-authentication | key |
| `count` is a reserved derived leaf representing aggregate server-managed state such as like totals. | derived-count-leaf | key |
| Derived aggregate leaves such as `count` are readable by normal clients but are not directly writable by them. | derived-leaf-read-only | key |
| Server logic should update derived aggregate leaves as a consequence of principal-scoped actions such as like and unlike. | principal-action-updates-aggregate | key |
| Principal-scoped action records such as likes should remain isolated to the acting principal and should not expose other principals' identities to ordinary readers. | private-principal-actions | key |

#### Field Runtime Behaviour

| field | mock_behaviour | production_behaviour | scenario |
| --- | --- | --- | --- |
| secureKeyId | Assume the input is otherwise valid unless test configuration forces a different status. | Verify integrity token derived from keyId before accepting the request | normal-request |
| secureKeyId | Allow secureKeyId to force a non-ok status such as invalid unauthorised or outdated. | Ignore mock-only forcing and use normal integrity verification | statusAsKey-enabled |
| localKeyId | Preserve the provisional value while returning the official keyId. | Preserve the provisional value while returning the official keyId | client-waits-for-official-keyId |
| kind | Correct or reject the mismatch based on the server-derived interpretation. | Reject or correct the mismatch and use the kind derived from keyId | client-sends-kind-that-does-not-match-keyId |
| version | Reject the write as outdated. | Reject the write as outdated | client-sends-stale-version |
| created | Order by creation time when that is the chosen view. | Order by creation time when that is the chosen view | ordering-comments-or-chat |
| updated | Use updated to inspect what changed since the client last synced. | Use updated to inspect what changed since the client last synced | support-after-version-mismatch |

#### Field Trust

| client_role | field | security_note | server_role | source_of_truth | trust_level |
| --- | --- | --- | --- | --- | --- |
| Use returned official identity and reference paths | keyId | Identity, schema interpretation, and access-scope derivation must rely on keyId rather than client hints. | Generate and validate authoritative identity | server | authoritative |
| Send integrity token derived from keyId | secureKeyId | Production should use a signed or JWT-style check to detect corruption or forgery. | Verify integrity token before accepting the request | production-server | verified-required |
| Use provisional local identifier until the official keyId is returned | localKeyId | Production servers may apply basic validation such as length and character checks. | Preserve value without treating it as authoritative identity | client | client-hint |
| Use as a temporary local hint before the server responds | kind | Do not trust client-supplied kind when it conflicts with keyId. | Derive authoritative kind from keyId with unique identifiers removed | server-derived-from-keyId | server-derived |
| Optionally provide an ISO language code per value node | language | Different nodes in the same document may use different languages. | Preserve and validate optional language metadata | client-and-server | optional-metadata |
| Send latest known version when writing | version | Reject writes based on older versions. | Generate and compare versions for optimistic sync | server | server-managed |
| Use for ordering and display | created | Useful for comments, chat, and other chronological views. | Generate creation timestamp | server | server-managed |
| Use for ordering, support, and stale-state diagnosis | updated | Useful for understanding what changed after version mismatches. | Generate update timestamp | server | server-managed |
| Send and consume protocol-level string payload | value | Interpret richer formats above the protocol layer without changing the transport type. | Store and return protocol-level string payload | client-and-server | shared-payload |

#### Scoped Key Examples

| description | example_name | key_example | product_variant | scope_shape |
| --- | --- | --- | --- | --- |
| Example of a principal-scoped key where access is visible through a tenant/group/user hierarchy using stable generated ids rather than mutable names. | tenant-group-user-note | tenant:t8f3a1c2:group:g4b7d9e1:user:u42c91ab:dashboard:d1e52f07:note:n7c401c2:text | tenant-group-user | level1/level2/principal |
| Example of a group-scoped document root where all members of the configured group scope may read it. | tenant-group-root | tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07 | tenant-group-user | level1/level2 |
| Same fixed-depth model with product-specific labels instead of tenant/group/user, still using stable generated ids. | department-region-member-note | department:d9a8c102:region:r6be41f0:member:m17aa9f3:dashboard:d1e52f07:note:n7c401c2:text | department-region-member | level1/level2/principal |
| Example of a top-level scoped document root using a department label instead of tenant. | department-root | department:d9a8c102:dashboard:d1e52f07 | department-region-member | level1 |

### 06 Sync And Persistence

How the logical protocol model maps to server and client storage.

#### Archive Retention

| description | rule | scope |
| --- | --- | --- |
| Archive remains record state carried by the value node rather than a delete operation. | archived-in-state | archive |
| Archived value nodes may still appear in snapshots so clients can reconstruct the full known state. | archived-in-snapshots | archive |
| Archived value nodes may still be emitted in `set` events with `--archived` present in the payload options. | archived-in-events | archive |
| Archive retention is indefinite by default at the protocol level. | archive-indefinite-default | archive |
| Physical cleanup or compaction of archived nodes is an implementation-specific concern rather than a protocol guarantee. | cleanup-implementation-specific | archive |

#### Event Rules

| description | rule | scope |
| --- | --- | --- |
| A `set` event must carry `rootKey`, `key`, `keyValue`, and `created`. | set-required-fields | event |
| A `snapshot-replaced` event must carry `rootKey`, `snapshotVersion`, and `created`. | snapshot-replaced-required-fields | event |
| `snapshot-replaced` is emitted after `setSnapshot`, not after every normal `set`. | snapshot-replaced-trigger | event |
| Archive is carried as record state, typically through `--archived`, rather than as a separate event operation. | archive-as-state | event |
| Unsubscribing a root key that is not currently subscribed is a no-op and does not raise an error. | unsubscribe-missing-key-noop | client |

#### Snapshot Lifecycle

| description | rule | scope |
| --- | --- | --- |
| `getSnapshot` returns the full subtree for the requested root key rather than a partial fragment. | full-subtree-read | snapshot |
| A snapshot may be empty when the root key exists but currently has no child value nodes. | empty-snapshot-valid | snapshot |
| `setSnapshot` replaces the full authoritative subtree baseline for the requested root key. | replace-full-baseline | snapshot |
| The current snapshot is replaced in place for sync purposes rather than requiring historical snapshot retention in the protocol. | replace-in-place | snapshot |
| After `setSnapshot` succeeds the server emits `snapshot-replaced` for that root key. | emit-snapshot-replaced | snapshot |

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

### 07 Open Inconsistencies

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
import type { Command } from './common';
import type { RequestMetadata, ResponseEnvelope } from './envelope';

type GetCommandRequest = RequestMetadata & {
  command: Command;
};

type GetCommandResponse = ResponseEnvelope<{
  command: Command;
  content: string;
}>;

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
import type { RequestMetadata, ResponseEnvelope } from './envelope';

type CommandStatus = {
  command: Command;
  status: OperationStatus;
  message?: string;
};

type SetCommandsRequest = RequestMetadata & {
  commands: Command[];
};

type SetCommandsResponse = ResponseEnvelope<{
  results: CommandStatus[];
}>;

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

### 02 Production Versus Mock

Where the mock server intentionally diverges from production behavior.

#### Production Versus Mock Behavior

The mock server intentionally shares the protocol shape of a production Yggdrasil server while simplifying some runtime behavior.

- production servers should verify `secureKeyId` as an integrity check derived from `keyId`
- the mock server may use `secureKeyId` as a test hook to force a non-`ok` status when `statusAsKey` is enabled
- production servers may apply stricter validation to `localKeyId`, `kind`, and other inputs
- mock-server administration commands exist only for tests and are outside the Yggdrasil protocol

This keeps client-facing protocol semantics realistic while still allowing the CLI to simulate failure modes cheaply during tests.

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
  'tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07',
  'tenant:t8f3a1c2:group:g4b7d9e1:user:u42c91ab:dashboard:d1e52f07:note:n7c401c2:text',
  'tenant:t8f3a1c2:group:g4b7d9e1:user:u42c91ab:dashboard:d1e52f07:note:n7c401c2:comment:c38dd201:text',
  'tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n7c401c2:like:user:_',
  'tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n7c401c2:like:count',
  'department:d9a8c102:region:r6be41f0:member:m17aa9f3:dashboard:d1e52f07:note:n7c401c2:language:_',
];
```

#### Shared Response Envelope Example

```ts
import type { KeyParams, KeyValueParams, OperationStatus } from './common';

export type ResponseEnvelope<T> = {
  id: string;
  status: OperationStatus;
  message?: string;
  data: T;
};

export type RequestMetadata = {
  // Optional client-provided correlation identifier. If omitted, the server
  // should generate a response id.
  id?: string;
};

export type KeyStatusResult = {
  key: KeyParams;
  status: OperationStatus;
  message?: string;
};

export type KeyValueStatusResult = {
  keyValue: KeyValueParams;
  status: OperationStatus;
  message?: string;
};
```

### 02 HTTP APIs

REST-style request and response shapes.

#### Get Key Value Example

```ts
import type { KeyParams } from './common';
import type { KeyValueStatusResult, RequestMetadata, ResponseEnvelope } from './envelope';

type GetKeyValueRequest = RequestMetadata & {
  rootKey: KeyParams; // required: keyId, secureKeyId
  keyList: KeyParams[]; // required: keyId, secureKeyId
};

type GetKeyValueResponse = ResponseEnvelope<{
  rootKey: KeyParams; // provide keyId, and optionally all other fields except localKeyId
  keyValueList: KeyValueStatusResult[]; // provide keyId, and optionally all other fields except localKeyId
}>;

export interface KeyValueReadApi {
  getKeyValueList(request: GetKeyValueRequest): GetKeyValueResponse;
}
```

#### Get Snapshot Example

```ts
import type { KeyParams, KeyValueParams } from './common';
import type { RequestMetadata, ResponseEnvelope } from './envelope';

type GetSnapshotRequest = RequestMetadata & {
  key: KeyParams; // required: keyId, secureKeyId
};

type GetSnapshotResponse = ResponseEnvelope<{
  key: KeyParams; // required: keyId, and the remaining fields may depend on success or failure.
  keyValueList: KeyValueParams[]; // required: keyId, and the remaining fields may depend on success or failure.
}>;

export interface SnapshotReadApi {
  getSnapshot(request: GetSnapshotRequest): GetSnapshotResponse;
}
```

#### Create New Keys Example

```ts
import type { KeyParams, NodeKindExample, OperationStatus } from './common';
import type { KeyStatusResult, RequestMetadata, ResponseEnvelope } from './envelope';

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
  children: KeyStatusResult[];
};

type NewKeysRequest = RequestMetadata & {
  rootKey: KeyParams;
  newKeys: NewKeyParams[]; // processed independently and returned in request order
};

type NewKeysResponse = ResponseEnvelope<{
  rootKey: KeyParams;
  newKeys: SuggestedNewKeyParams[]; // every requested item should receive a corresponding per-item status
}>;

export interface NewKeysApi {
  createNewKeys(request: NewKeysRequest): NewKeysResponse;
}
```

#### Set Key Value Example

```ts
import type { KeyParams, KeyValueParams } from './common';
import type { KeyStatusResult, RequestMetadata, ResponseEnvelope } from './envelope';

type SetKeyValueRequest = RequestMetadata & {
  rootKey: KeyParams; // required: keyId, secureKeyId
  keyValueList: KeyValueParams[]; // required: keyId, secureKeyId, processed independently and returned in request order
};

type SetKeyValueResponse = ResponseEnvelope<{
  rootKey: KeyParams; // required: keyId
  keyList: KeyStatusResult[]; // required: keyId, with one per-item status for each requested write
}>;

export interface KeyValueWriteApi {
  setKeyValueList(request: SetKeyValueRequest): SetKeyValueResponse;
}
```

#### Set Snapshot Example

```ts
import type { KeyParams, KeyValueParams } from './common';
import type { RequestMetadata, ResponseEnvelope } from './envelope';

type SetSnapshotRequest = RequestMetadata & {
  key: KeyParams; // required: keyId, secureKeyId
  keyValueList: KeyValueParams[]; // required: keyId, secureKeyId
};

type SetSnapshotResponse = ResponseEnvelope<{
  key: KeyParams;
}>;

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
import type { OperationStatus, PrincipalParams } from './common';
import type { EventEnvelope } from './event-envelope';
import type {
  ClientMessage,
  EventMessage,
  ServerMessage,
  StatusMessage,
  SubscribeMessage,
  SubscribedMessage,
  UnsubscribeMessage,
  UnsubscribedMessage,
} from './websocket-messages';

type Subscription = {
  id: string;
  principal: PrincipalParams;
  rootKeys: string[];
};

type EventResponse = {
  id: string;
  principal: PrincipalParams;
  eventList: [EventEnvelope, OperationStatus][];
};

export interface EventApi {
  registerPrincipal(principal: PrincipalParams): [PrincipalParams, OperationStatus];
  // Unregistering a principal clears all active subscriptions for that principal.
  unregisterPrincipal(principal: PrincipalParams): [PrincipalParams, OperationStatus];
  subscribe(subscription: Subscription): EventResponse;
  // Unsubscribing a key that is not currently subscribed is a no-op and does not raise an error.
  unsubscribe(subscription: Subscription): EventResponse;
  receivePrincipalUpdate(principal: PrincipalParams): EventResponse;
}

export interface WebSocketEventApi {
  onClientMessage(message: ClientMessage): ServerMessage | EventMessage;
  // Repeated subscribe messages extend the active root-key set for the connection.
  // Root subscriptions are predefined and apply to the full readable descendant subtree.
  // A subscribe request for a non-allowed root should return a status message with invalid.
  // When present, the client command id should be echoed by the matching reply.
  // Duplicate root keys are normalized and the most recent entry wins.
  subscribe(message: SubscribeMessage): SubscribedMessage | StatusMessage;
  // A valid unsubscribe request should return unsubscribed even when some keys were not active.
  unsubscribe(message: UnsubscribeMessage): UnsubscribedMessage | StatusMessage;
  // Closing the connection clears all active subscriptions tied to that connection.
  disconnect(principal: PrincipalParams): void;
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
  receivePrincipalUpdate(principal: PrincipalParams): EventResponse;
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
import type { OperationStatus } from './common';
import type { EventEnvelope } from './event-envelope';

export type CommandMetadata = {
  // Optional client-provided correlation identifier echoed by the matching
  // acknowledgement or status response for the same command.
  id?: string;
};

export type SubscribeMessage = CommandMetadata & {
  kind: 'subscribe';
  // A client may send subscribe more than once to add further root keys
  // without reopening the WebSocket connection. Duplicate root keys are
  // normalized without error and the most recent subscription entry wins.
  rootKeys: string[];
};

export type UnsubscribeMessage = CommandMetadata & {
  kind: 'unsubscribe';
  rootKeys: string[];
};

export type PingMessage = CommandMetadata & {
  kind: 'ping';
};

export type ClientMessage =
  | SubscribeMessage
  | UnsubscribeMessage
  | PingMessage;

export type SubscribedMessage = {
  kind: 'subscribed';
  id?: string;
  rootKeys: string[];
};

export type UnsubscribedMessage = {
  kind: 'unsubscribed';
  id?: string;
  rootKeys: string[];
};

export type EventMessage = {
  kind: 'event';
  event: EventEnvelope;
};

export type StatusMessage = {
  kind: 'status';
  id?: string;
  status: OperationStatus;
  message?: string;
};

export type PongMessage = {
  kind: 'pong';
  id?: string;
};

export type ServerMessage =
  | SubscribedMessage
  | UnsubscribedMessage
  | EventMessage
  | StatusMessage
  | PongMessage;
```

#### WebSocket Correlation

| description | rule | scope |
| --- | --- | --- |
| Client WebSocket commands may include an optional shallow `id` for correlation. | command-id-optional | websocket |
| When a client command includes `id`, the matching `subscribed`, `unsubscribed`, `pong`, or `status` response should echo the same `id`. | ack-echoes-command-id | websocket |
| Server-pushed `event` messages do not need a command-correlation `id` because they are not direct replies to a client command. | server-may-omit-id-for-push | websocket |
| WebSocket command correlation should not require an additional wrapper; `id` remains a shallow optional field on command messages. | no-extra-command-envelope | websocket |
| WebSocket command correlation should follow the same echo-or-generate intent as HTTP request correlation where practical. | consistent-with-http | implementation |

#### WebSocket Flow

| actor | description | step | transport |
| --- | --- | --- | --- |
| client | Open the WebSocket connection to the optional `/events` endpoint. | open-connection | websocket |
| client | Send a `subscribe` message with one or more allowed root keys to watch and an optional command `id`. | subscribe-root-keys | websocket |
| server | Return a `status` message with `invalid` for subscriptions that target arbitrary subtrees or root keys outside the predefined subscribable set. | validate-roots | internal |
| server | Apply each accepted root subscription to that root and all readable descendants. | expand-to-descendants | internal |
| client | Send another `subscribe` message later to add more root keys without reopening the connection. | extend-subscription | websocket |
| server | Normalize duplicate root keys without error and keep the most recent subscription entry. | deduplicate-subscription | internal |
| server | Reply with `subscribed` to confirm the active root-key subscriptions and echo the command `id` when present. | confirm-subscription | websocket |
| server | Send an `event` message containing the `EventEnvelope`. | receive-event | websocket |
| client-and-server | Use `ping` and `pong` messages to keep the connection alive and echo the command `id` when present. | ping-pong | websocket |
| client | Send `unsubscribe` with an optional command `id` when the client no longer wants updates for those root keys. | unsubscribe-root-keys | websocket |
| server | Reply with `unsubscribed` for valid unsubscribe requests even when some requested root keys were not active and echo the command `id` when present. | confirm-unsubscribe | websocket |
| server | When a user unregisters or the connection closes, remove all active subscriptions tied to that user or connection. | clear-subscriptions-on-unregister | internal |

#### WebSocket Rules

| description | rule | scope |
| --- | --- | --- |
| Subscribing to an allowed root key includes events for that root and all of its readable descendants. | root-scope-includes-descendants | subscription |
| A subscription only yields descendant events that the user is allowed to read. | read-access-still-applies | subscription |
| The list of subscribable root keys should be predefined by configuration rather than accepted as arbitrary subtree keys. | roots-predefined | configuration |
| Clients should be able to subscribe to a document root but not to an arbitrary section or nested subtree inside that root. | subtree-subscriptions-not-allowed | subscription |
| Subscribing to a root key outside the predefined allowed set should return `invalid`. | non-allowed-root-invalid | subscription |
| When a client WebSocket command cannot be accepted, the server should return a `status` message with the relevant `OperationStatus` and optional `message`. | status-message-for-command-errors | connection |
| When a client WebSocket command includes `id`, the matching acknowledgement or `status` response should echo that same `id`. | command-id-echoed | connection |
| Configured subscribable roots should not overlap. This should be treated as a configuration warning or validation error rather than a runtime merge rule. | overlapping-roots-disallowed | configuration |
| Repeated `subscribe` messages extend the active root-key set for the same connection. | subscribe-extends-set | connection |
| Duplicate root keys are normalized without error. | duplicate-root-keys-normalized | connection |
| When duplicate root keys are received the most recent subscription entry wins. | most-recent-subscription-wins | connection |
| Unsubscribing a root key that is not currently subscribed is a no-op and does not raise an error. | unsubscribe-missing-key-noop | connection |
| A valid `unsubscribe` command should return `unsubscribed` even when some requested root keys were not active. | unsubscribe-valid-ack | connection |
| When a user unregisters or the connection closes all active subscriptions tied to that user or connection are removed. | disconnect-clears-subscriptions | connection |

