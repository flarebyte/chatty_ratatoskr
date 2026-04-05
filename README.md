# chatty-ratatoskr

`chatty-ratatoskr` provides the `chatty` CLI, a lightweight Yggdrasil mock server for local development and CI.

![Ratatoskr climbing Yggdrasil's mighty trunk](./chatty-ratatoskr-hero.png)

## What `chatty` Does

`chatty` starts an in-memory mock server that exposes HTTP endpoints for:

- snapshot bootstrap
- targeted node reads and writes
- create operations with server-generated ids
- optional WebSocket events
- optional mock-only admin commands

The server is intended for client integration work and automated tests, not production persistence.

## Install And Build

Build a local development binary:

```bash
make build-dev
```

The binary is written to:

```bash
./.e2e-bin/chatty
```

Build release artifacts:

```bash
make build
```

## CLI Basics

Show help:

```bash
./.e2e-bin/chatty help
./.e2e-bin/chatty --help
```

Show version:

```bash
./.e2e-bin/chatty version
./.e2e-bin/chatty --version
```

Start the server on the default loopback address:

```bash
./.e2e-bin/chatty serve
```

Start the server on a chosen address:

```bash
./.e2e-bin/chatty serve --listen 127.0.0.1:19080
```

Start the server from config:

```bash
./.e2e-bin/chatty serve --config testdata/config/basic.cue
```

## Current Behavior

The current CLI supports:

- `help` and `version`
- `serve`
- `PUT` and `GET /snapshot`
- `PUT` and `GET /node`
- `POST /create`
- WebSocket `/events` when enabled by config
- admin `PUT /admin/commands` and `GET /admin/commands` when admin is enabled

Admin and WebSocket behavior are intentionally optional and controlled by config.

## Quick Test Workflow

Run the shipped review workflow:

```bash
make review
```

Run the end-to-end suite only:

```bash
make e2e
```

For a shorter command reference, see [CHEATSHEET.md](/Users/olivier/Documents/github/chatty-ratatoskr/CHEATSHEET.md).

## More Information

- design spec: [doc/design/yggdrasil-mock-server.md](/Users/olivier/Documents/github/chatty-ratatoskr/doc/design/yggdrasil-mock-server.md)
- quick command reference: [CHEATSHEET.md](/Users/olivier/Documents/github/chatty-ratatoskr/CHEATSHEET.md)
- maintainer workflow: [CONTRIBUTING.md](/Users/olivier/Documents/github/chatty-ratatoskr/CONTRIBUTING.md)
