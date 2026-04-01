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

	// Key hierarchy rules define which kinds may appear at each level.
	// In normal operation, keyId is the source of truth and kind is derived from
	// the key schema with unique identifiers removed. KeyKind.language is an
	// optional ISO code, so multilingual documents may contain nodes in
	// different languages at the same time.
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
