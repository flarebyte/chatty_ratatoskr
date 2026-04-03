http: {
	port:             18080
	maxPayloadSizeKB: 400
}

websocket: {
	supported: true
	path:      "/events"
}

admin: {
	scope: "mock-server-only"
}
