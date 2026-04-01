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
