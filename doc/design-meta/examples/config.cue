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
