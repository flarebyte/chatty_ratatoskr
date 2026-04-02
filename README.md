# chatty-ratatoskr

`chatty-ratatoskr` is a Go CLI project for a lightweight Yggdrasil mock server.

The product binary is `chatty`.

The goal is to provide a cheap local and CI-friendly server that simulates a production Yggdrasil backend over HTTP and, later, optional WebSocket. The design is spec-first and currently lives under [doc/design/yggdrasil-mock-server.md](/Users/olivier/Documents/github/chatty-ratatoskr/doc/design/yggdrasil-mock-server.md) and [doc/design-meta](/Users/olivier/Documents/github/chatty-ratatoskr/doc/design-meta).

## What It Is For

The intended server supports:

- hierarchical logical keys
- snapshot bootstrap
- incremental sync events
- optimistic write conflict handling
- mock-only admin controls outside the protocol surface

The main use case is testing client code against a realistic mock Yggdrasil server in local development and CI.

## Current Status

The repository is still early.

It currently includes:

- a minimal `chatty` CLI entrypoint
- a minimal Bun e2e smoke test
- release/build scripts
- the spec and metadata pipeline setup

It does not yet implement the full Yggdrasil mock server described in the spec. Treat the design docs as the target behavior, not as a claim that all endpoints already exist.

## Install And Build

Core tools used by the current repo workflow:

- `go`
- `bun`
- `golangci-lint`
- `flyb`
- `thoth`
- `jq`

Build the local development binary:

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

## Current CLI Usage

Show the current version output:

```bash
./.e2e-bin/chatty version
```

Show the current help output:

```bash
./.e2e-bin/chatty help
```

At the moment, `chatty serve` is only a placeholder and the full mock-server behavior is still under implementation.

## End-To-End Usage

Run the TypeScript end-to-end suite:

```bash
make e2e
```

Or directly:

```bash
cd script/e2e && bun test
```

The e2e suite is intended to become the feature ledger for supported user-visible workflows.

## Design Sources

Primary references:

- [doc/design/yggdrasil-mock-server.md](/Users/olivier/Documents/github/chatty-ratatoskr/doc/design/yggdrasil-mock-server.md)
- [doc/design-meta/app.cue](/Users/olivier/Documents/github/chatty-ratatoskr/doc/design-meta/app.cue)
- [scratch/prompt.md](/Users/olivier/Documents/github/chatty-ratatoskr/scratch/prompt.md)

## Contributing

Contributor workflow, `Makefile` conventions, and maintenance/tooling notes live in [CONTRIBUTING.md](/Users/olivier/Documents/github/chatty-ratatoskr/CONTRIBUTING.md).
