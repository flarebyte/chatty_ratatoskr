# Contributing

## Scope

This file is for maintainers and coding agents working inside the repository.

`README.md` is for users of the `chatty` CLI.
`CHEATSHEET.md` is for quick command usage.

## Main Workflow

The `Makefile` is the first-class workflow entrypoint.

For humans:

- run `make review` periodically
- treat it as the main signal that docs, formatting, tests, e2e, and lint still align

For agents:

- run at least `make format` before handing work back
- prefer the existing `Makefile` targets over ad hoc command sequences when the workflow is recurring

## Important Targets

- `make format`
  - runs `gofmt`
  - runs Biome formatting/checks when `./node_modules/.bin/biome` is available
- `make test`
  - runs `flyb` doc generation
  - runs Go unit and package-level tests
- `make e2e`
  - runs Bun-powered TypeScript end-to-end tests under `script/e2e`
- `make lint`
  - runs repo static checks
- `make review`
  - runs `format`, `test`, `e2e`, and `lint`
- `make build-dev`
  - builds `.e2e-bin/chatty`
- `make build`
  - builds release artifacts into `./build`
- `make dup`
  - runs duplication scans with the local `jscpd` binary
- `make complexity`
  - shows the current complexity hotspots
- `make sec`
  - runs the configured security scan
  - if `semgrep` cannot start because of a local trust-store problem, treat that as a blocking environment issue rather than a passing scan

## Tool Roles

- `chatty`
  - product binary under development
- `thoth`
  - external metadata pipeline CLI used by this repo
- `flyb`
  - validates and generates design docs from `doc/design-meta`
- `bun`
  - runs TypeScript scripts and e2e tests
- `go`
  - builds, tests, and vets the Go code
- `golangci-lint`
  - Go lint aggregation
- `Biome`
  - local JS/TS/JSON formatter and checker when installed
- `jscpd`
  - local duplication scanner used by `make dup`

## Testing Policy

- write unit tests in Go
- write package-level handler and parser tests in Go
- write end-to-end tests in TypeScript under `script/e2e`
- keep shared Go test helpers for repeated handler and transport setup
- keep shared TypeScript e2e helpers for repeated mock-server interactions
- do not mark a user-visible feature as working without matching acceptance coverage

## Maintenance Gates

- `make dup`
  - ignore cache and generated directories; review only repo-owned files
  - current duplication baseline is `0` clones for both Go and TypeScript
- `make complexity`
  - current accepted Go hotspots are `internal/yggkey/parse.go`, `internal/httpapi/events.go`, `internal/httpapi/snapshot.go`, and `internal/httpapi/node.go`
  - these remain acceptable because they hold the main parsing and protocol contracts; refactor only when readability or change risk justifies it
- `make sec`
  - review the scan result, not just the command exit
- `make build`
  - check the artifacts under `build/`, not just the command exit

## Primary References

- [doc/design/yggdrasil-mock-server.md](/Users/olivier/Documents/github/chatty-ratatoskr/doc/design/yggdrasil-mock-server.md)
- [doc/design-meta/app.cue](/Users/olivier/Documents/github/chatty-ratatoskr/doc/design-meta/app.cue)
- [scratch/prompt.md](/Users/olivier/Documents/github/chatty-ratatoskr/scratch/prompt.md)

## Repository Layout

- [cmd/chatty](/Users/olivier/Documents/github/chatty-ratatoskr/cmd/chatty)
- [internal/httpapi](/Users/olivier/Documents/github/chatty-ratatoskr/internal/httpapi)
- [internal/runtimeconfig](/Users/olivier/Documents/github/chatty-ratatoskr/internal/runtimeconfig)
- [script/e2e](/Users/olivier/Documents/github/chatty-ratatoskr/script/e2e)
- [testdata](/Users/olivier/Documents/github/chatty-ratatoskr/testdata)
- `pipeline-*.thoth.cue`
