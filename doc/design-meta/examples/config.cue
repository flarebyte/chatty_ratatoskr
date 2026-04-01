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

	deleteSnapshot: {
		path: "/snapshot"
		verb: "DELETE"
	}

	sendEvent: {
		path: "/event"
		verb: "POST"
	}

	adminCommands: {
		path: "/admin/commands"
		verb: "PUT|GET"
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
