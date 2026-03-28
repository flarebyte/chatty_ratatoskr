endpoints: {

	postCreate: {
		path: "/create"
        verb: "POST"
	}

    putNode: {
        path: "/node"
        verb: "PUT"
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

}
