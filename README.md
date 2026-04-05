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

## Install

Install `chatty` with Homebrew:

```bash
brew install flarebyte/tap/chatty-ratatoskr
```

Then check the CLI:

```bash
chatty --help
```

For source builds and maintainer workflow commands, see [CONTRIBUTING.md](/Users/olivier/Documents/github/chatty-ratatoskr/CONTRIBUTING.md).

## CLI Basics

Show help:

```bash
chatty help
chatty --help
```

Show version:

```bash
chatty version
chatty --version
```

Start the server on the default loopback address:

```bash
chatty serve
```

Start the server on a chosen address:

```bash
chatty serve --listen 127.0.0.1:19080
```

Start the server from config:

```bash
chatty serve --config testdata/config/basic.cue
```

See the config reference for the currently supported `--config` fields:

- [doc/configuration.md](/Users/olivier/Documents/github/chatty-ratatoskr/doc/configuration.md)

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

## Quick Reference

For a shorter command reference, see [CHEATSHEET.md](/Users/olivier/Documents/github/chatty-ratatoskr/CHEATSHEET.md).

## More Information

- design spec: [doc/design/yggdrasil-mock-server.md](/Users/olivier/Documents/github/chatty-ratatoskr/doc/design/yggdrasil-mock-server.md)
- config reference: [doc/configuration.md](/Users/olivier/Documents/github/chatty-ratatoskr/doc/configuration.md)
- quick command reference: [CHEATSHEET.md](/Users/olivier/Documents/github/chatty-ratatoskr/CHEATSHEET.md)
- maintainer workflow and repo commands: [CONTRIBUTING.md](/Users/olivier/Documents/github/chatty-ratatoskr/CONTRIBUTING.md)
