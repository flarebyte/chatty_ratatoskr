# Configuration

`chatty serve --config ...` accepts either:

- a JSON file that matches the runtime config shape directly
- a CUE file that `chatty` exports with `cue export ... --out json`

This document describes the runtime fields that are actually consumed by the current CLI.

## Current Runtime Contract

The current runtime config controls:

- HTTP listen address
- HTTP payload size limit
- whether WebSocket support is enabled
- WebSocket message size limit
- whether admin routes are enabled
- whether admin routes may be exposed on a non-loopback address

## JSON Shape

For JSON config files, `chatty` reads this shape:

```json
{
  "Listen": "127.0.0.1:18080",
  "WebSocketEnabled": true,
  "WebSocketMessageLimitBytes": 32768,
  "AdminEnabled": true,
  "AllowUnsafeAdminExposure": false,
  "HTTPPayloadLimitBytes": 1048576
}
```

Unknown JSON fields are rejected.

## CUE Shape

For CUE config files, `chatty` currently consumes this exported structure:

```cue
http: {
  port: int
  limitBytes?: int
}

websocket: {
  supported: bool
  limitBytes?: int
}

admin?: {
  unsafeExposure?: bool
}
```

The runtime derives:

- `listen` as `127.0.0.1:<http.port>`
- `websocketEnabled` from `websocket.supported`
- `websocketMessageLimitBytes` from `websocket.limitBytes`, or a default when omitted
- `adminEnabled` from the presence of the `admin` block
- `allowUnsafeAdminExposure` from `admin.unsafeExposure`
- `httpPayloadLimitBytes` from `http.limitBytes`, or a default when omitted

## Safe Defaults

When `--config` is omitted, `chatty` uses safe defaults:

- listen on `127.0.0.1:8080`
- WebSocket disabled
- admin disabled
- HTTP payload limit `1048576` bytes
- WebSocket message limit `32768` bytes

## Validation Rules

Startup fails before listening when:

- the listen address is empty
- the HTTP payload limit is not positive
- the WebSocket message limit is not positive
- `allowUnsafeAdminExposure` is true while admin is disabled
- admin is enabled on a non-loopback listen address without `unsafeExposure: true`

## Current Example

The repo includes:

- [testdata/config/basic.cue](/Users/olivier/Documents/github/chatty-ratatoskr/testdata/config/basic.cue)
- [testdata/config/basic.json](/Users/olivier/Documents/github/chatty-ratatoskr/testdata/config/basic.json)

Important note:

The current CUE runtime loader only consumes a subset of the available CUE fields. Extra spec-oriented fields may exist in example CUE files, but they are not necessarily used by the current runtime unless they export to the supported structure described above.
