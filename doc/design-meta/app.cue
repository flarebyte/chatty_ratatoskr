package flyb

source: "yggdrasil-mock-server"
name:   "yggdrasil-mock-server"
modules: ["core"]

reports: [{
	title:       "Yggdrasil Mock Server Design"
	filepath:    "../design/yggdrasil-mock-server.md"
	description: "First-pass design report for the lightweight HTTP and WebSocket mock server CLI."
	sections: [{
		title:       "01 Overview"
		description: "Problem framing and the main architectural shape."
		sections: [{
			title:       "01 Intent"
			description: "Why this server exists and how it should be used."
			notes: ["yggdrasil.server", "yggdrasil.scope"]
		}, {
			title:       "02 Transports"
			description: "The protocol is primarily HTTP with optional WebSocket support."
			notes: ["yggdrasil.transport.http", "yggdrasil.transport.websocket"]
		}]
	}, {
		title:       "02 Use Cases"
		description: "Current goals and constraints captured from the draft inputs."
		sections: [{
			title:       "01 Product Goals"
			description: "Why the server should exist and what it must support."
			notes: ["yggdrasil.usecases"]
		}, {
			title:       "02 Implementation Constraints"
			description: "Preferred implementation choices for the Go CLI."
			notes: ["yggdrasil.implementation"]
		}]
	}, {
		title:       "03 Protocol Surface"
		description: "The current endpoint and payload sketches that inform the evolving Yggdrasil protocol."
		sections: [{
			title:       "01 Config Draft"
			description: "CUE draft describing ports, endpoints, and schema constraints."
			notes: ["yggdrasil.config"]
		}, {
			title:       "02 Action Matrix"
			description: "HTTP and WebSocket actions currently envisaged by the examples."
			notes: ["yggdrasil.protocol.actions"]
		}, {
			title:       "03 Entity Model"
			description: "Current entity and field definitions used by the draft protocol."
			notes: ["yggdrasil.entities", "yggdrasil.entity-fields"]
		}]
	}, {
		title:       "04 TypeScript Examples"
		description: "Reference examples that should be treated as protocol design inputs, not final implementation code."
		sections: [{
			title:       "01 Shared Types"
			description: "Common primitives reused by the request/response examples."
			notes: ["yggdrasil.ts.common"]
		}, {
			title:       "02 HTTP APIs"
			description: "REST-style request and response shapes."
			notes: [
				"yggdrasil.ts.new-keys",
				"yggdrasil.ts.set-key-value",
				"yggdrasil.ts.get-key-value",
				"yggdrasil.ts.set-snapshot",
				"yggdrasil.ts.get-snapshot",
				"yggdrasil.ts.set-admin-commands",
				"yggdrasil.ts.get-admin-command",
			]
		}, {
			title:       "03 Event APIs"
			description: "Optional WebSocket and event-store oriented examples."
			notes: [
				"yggdrasil.ts.receive-event",
				"yggdrasil.ts.key-value-store",
				"yggdrasil.ts.snapshot-store",
			]
		}]
	}]
}]

notes: [
	{
		name:  "yggdrasil.server"
		title: "Lightweight Go Mock Server"
		markdown: """
The project is shaping a lightweight CLI in Go that runs a mock backend over HTTP and, optionally, WebSocket.

Its purpose is to support testing and CI by simulating a more capable server that would exist elsewhere, while keeping behavior explicit, configurable, and cheap to run.
"""
		labels: ["overview", "server", "go"]
	},
	{
		name:  "yggdrasil.scope"
		title: "Spec-First Workflow"
		markdown: """
The source of truth for the design is intended to live in `doc/design-meta` as flyb input.

The examples under `doc/design-meta/examples` are design artefacts that capture the current protocol thinking:
- `config.cue` for draft protocol and schema configuration
- `.csv` files for use cases, action matrices, entities, and implementation constraints
- `.ts` files for request, response, and storage example shapes
"""
		labels: ["overview", "workflow", "flyb"]
	},
	{
		name:  "yggdrasil.transport.http"
		title: "HTTP Is The Primary Transport"
		markdown: """
The draft examples place most interactions on a small HTTP surface: creating keys, reading and writing key/value data, snapshot operations, event submission, and admin commands.

This aligns with the goal of a simple mock server that is easy to run locally and in CI.
"""
		labels: ["transport", "http"]
	},
	{
		name:  "yggdrasil.transport.websocket"
		title: "WebSocket Support Is Optional"
		markdown: """
The protocol examples explicitly call out WebSocket support as optional rather than mandatory.

Current notes suggest a constrained event model with heartbeat support, bounded message sizes, and a preference for well-known identifiers for message safety.
"""
		labels: ["transport", "websocket"]
	},
	{
		name:  "yggdrasil.config"
		title: "Draft Protocol Config"
		filepath: "examples/config.cue"
		labels: ["config", "cue", "protocol"]
	},
	{
		name:  "yggdrasil.usecases"
		title: "Use Cases"
		filepath: "examples/usecases.csv"
		arguments: ["format-csv=table"]
		labels: ["usecase", "csv"]
	},
	{
		name:  "yggdrasil.implementation"
		title: "Implementation Constraints"
		filepath: "examples/implementation.csv"
		arguments: ["format-csv=table"]
		labels: ["implementation", "csv"]
	},
	{
		name:  "yggdrasil.protocol.actions"
		title: "Protocol Actions"
		filepath: "examples/actions.csv"
		arguments: ["format-csv=table"]
		labels: ["protocol", "actions", "csv"]
	},
	{
		name:  "yggdrasil.entities"
		title: "Entities"
		filepath: "examples/entities.csv"
		arguments: ["format-csv=table"]
		labels: ["entity", "csv"]
	},
	{
		name:  "yggdrasil.entity-fields"
		title: "Entity Fields"
		filepath: "examples/entity-fields.csv"
		arguments: ["format-csv=table"]
		labels: ["entity", "fields", "csv"]
	},
	{
		name:  "yggdrasil.ts.common"
		title: "Common TypeScript Types"
		filepath: "examples/common.ts"
		labels: ["typescript", "example", "shared"]
	},
	{
		name:  "yggdrasil.ts.new-keys"
		title: "Create New Keys Example"
		filepath: "examples/new-keys.ts"
		labels: ["typescript", "example", "http"]
	},
	{
		name:  "yggdrasil.ts.set-key-value"
		title: "Set Key Value Example"
		filepath: "examples/set-key-value.ts"
		labels: ["typescript", "example", "http"]
	},
	{
		name:  "yggdrasil.ts.get-key-value"
		title: "Get Key Value Example"
		filepath: "examples/get-key-value.ts"
		labels: ["typescript", "example", "http"]
	},
	{
		name:  "yggdrasil.ts.set-snapshot"
		title: "Set Snapshot Example"
		filepath: "examples/set-snapshot.ts"
		labels: ["typescript", "example", "http"]
	},
	{
		name:  "yggdrasil.ts.get-snapshot"
		title: "Get Snapshot Example"
		filepath: "examples/get-snapshot.ts"
		labels: ["typescript", "example", "http"]
	},
	{
		name:  "yggdrasil.ts.set-admin-commands"
		title: "Set Admin Commands Example"
		filepath: "examples/set-admin-commands.ts"
		labels: ["typescript", "example", "admin"]
	},
	{
		name:  "yggdrasil.ts.get-admin-command"
		title: "Get Admin Command Example"
		filepath: "examples/get-admin-command.ts"
		labels: ["typescript", "example", "admin"]
	},
	{
		name:  "yggdrasil.ts.receive-event"
		title: "Receive Event Example"
		filepath: "examples/receive-event.ts"
		labels: ["typescript", "example", "websocket"]
	},
	{
		name:  "yggdrasil.ts.key-value-store"
		title: "Key Value Event Store Example"
		filepath: "examples/key-value-store.ts"
		labels: ["typescript", "example", "store"]
	},
	{
		name:  "yggdrasil.ts.snapshot-store"
		title: "Snapshot Event Store Example"
		filepath: "examples/snapshot-store.ts"
		labels: ["typescript", "example", "store"]
	},
]

relationships: [{
	from:   "yggdrasil.server"
	to:     "yggdrasil.transport.http"
	label:  "supports"
	labels: ["supports"]
}, {
	from:   "yggdrasil.server"
	to:     "yggdrasil.transport.websocket"
	label:  "supports_optional"
	labels: ["supports_optional"]
}, {
	from:   "yggdrasil.server"
	to:     "yggdrasil.config"
	label:  "configured_by"
	labels: ["configured_by"]
}, {
	from:   "yggdrasil.config"
	to:     "yggdrasil.protocol.actions"
	label:  "defines"
	labels: ["defines"]
}, {
	from:   "yggdrasil.protocol.actions"
	to:     "yggdrasil.ts.new-keys"
	label:  "illustrated_by"
	labels: ["illustrated_by"]
}, {
	from:   "yggdrasil.protocol.actions"
	to:     "yggdrasil.ts.set-key-value"
	label:  "illustrated_by"
	labels: ["illustrated_by"]
}, {
	from:   "yggdrasil.protocol.actions"
	to:     "yggdrasil.ts.get-key-value"
	label:  "illustrated_by"
	labels: ["illustrated_by"]
}, {
	from:   "yggdrasil.protocol.actions"
	to:     "yggdrasil.ts.set-snapshot"
	label:  "illustrated_by"
	labels: ["illustrated_by"]
}, {
	from:   "yggdrasil.protocol.actions"
	to:     "yggdrasil.ts.get-snapshot"
	label:  "illustrated_by"
	labels: ["illustrated_by"]
}, {
	from:   "yggdrasil.protocol.actions"
	to:     "yggdrasil.ts.receive-event"
	label:  "illustrated_by"
	labels: ["illustrated_by"]
}, {
	from:   "yggdrasil.protocol.actions"
	to:     "yggdrasil.ts.set-admin-commands"
	label:  "illustrated_by"
	labels: ["illustrated_by"]
}, {
	from:   "yggdrasil.protocol.actions"
	to:     "yggdrasil.ts.get-admin-command"
	label:  "illustrated_by"
	labels: ["illustrated_by"]
}, {
	from:   "yggdrasil.ts.common"
	to:     "yggdrasil.ts.new-keys"
	label:  "shared_types_for"
	labels: ["shared_types_for"]
}, {
	from:   "yggdrasil.ts.common"
	to:     "yggdrasil.ts.set-key-value"
	label:  "shared_types_for"
	labels: ["shared_types_for"]
}, {
	from:   "yggdrasil.ts.common"
	to:     "yggdrasil.ts.get-key-value"
	label:  "shared_types_for"
	labels: ["shared_types_for"]
}, {
	from:   "yggdrasil.ts.common"
	to:     "yggdrasil.ts.set-snapshot"
	label:  "shared_types_for"
	labels: ["shared_types_for"]
}, {
	from:   "yggdrasil.ts.common"
	to:     "yggdrasil.ts.get-snapshot"
	label:  "shared_types_for"
	labels: ["shared_types_for"]
}, {
	from:   "yggdrasil.ts.common"
	to:     "yggdrasil.ts.receive-event"
	label:  "shared_types_for"
	labels: ["shared_types_for"]
}, {
	from:   "yggdrasil.ts.common"
	to:     "yggdrasil.ts.key-value-store"
	label:  "shared_types_for"
	labels: ["shared_types_for"]
}, {
	from:   "yggdrasil.ts.common"
	to:     "yggdrasil.ts.snapshot-store"
	label:  "shared_types_for"
	labels: ["shared_types_for"]
}]

argumentRegistry: {
	version: "1"
	arguments: [{
		name:      "format-csv"
		valueType: "string"
		scopes:    ["note"]
		allowedValues: ["table", "raw"]
	}, {
		name:      "csv-include"
		valueType: "string"
		scopes:    ["note"]
	}, {
		name:      "csv-exclude"
		valueType: "string"
		scopes:    ["note"]
	}]
}
