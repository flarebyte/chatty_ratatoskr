port: 8080

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

    receiveEvent: {
        path: "/event"
        verb: "GET"
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
        checkVersion: true
    }
}
