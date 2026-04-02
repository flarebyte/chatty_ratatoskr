# Contributing

## Repo Workflow

The `Makefile` is a first-class workflow entrypoint for this repository.

For humans:

- run `make review` periodically
- treat it as the main signal that format, docs, tests, e2e, and lint are aligned

For agents:

- run at least `make format` before handing work back

## Important Makefile Targets

- `make format`
  - runs `gofmt`
  - runs Biome formatting/checks when `./node_modules/.bin/biome` is available
- `make test`
  - runs `flyb` doc generation
  - runs Go unit and package-level tests
- `make e2e`
  - runs TypeScript end-to-end tests under `script/e2e` with Bun
- `make lint`
  - runs repo static checks
- `make review`
  - runs `format`, `test`, `e2e`, and `lint`
- `make build-dev`
  - builds `.e2e-bin/chatty`
- `make build`
  - builds release artifacts into `./build`
- `make release`
  - runs release checks and then the Bun-based release helper
- `make thoth-meta-go`
  - collects metadata for non-test Go files via `thoth`
- `make thoth-meta-go-test`
  - collects metadata for Go test files via `thoth`
- `make thoth-meta-ts-e2e`
  - collects metadata for TypeScript e2e tests via `thoth`
- `make thoth-lint-go`
  - reports large or complex Go functions via `thoth`
- `make thoth-meta-merge`
  - aggregates persisted `thoth` metadata

## Tool Roles

- `chatty`
  - product binary under development
- `thoth`
  - external metadata pipeline CLI used by this repo
  - operates on `.thoth.cue` pipeline files
- `flyb`
  - validates and generates design docs from `doc/design-meta`
- `bun`
  - runs TypeScript scripts and e2e tests
- `go`
  - builds, tests, and vets the Go code
- `golangci-lint`
  - Go lint aggregation
- `jq`
  - lightweight JSON inspection for some metadata/reporting targets
- `Biome`
  - optional local formatter/linter for JS/TS/JSON-style files when installed in `node_modules`

## Testing Policy

- Write unit tests in Go.
- Write package-level handler and parser tests in Go.
- Write end-to-end tests in TypeScript under `script/e2e`.
- Run end-to-end tests with Bun.
- Treat the TypeScript e2e suite as the feature ledger for what is actually supported.
- Do not mark a user-visible feature as fully working without matching acceptance coverage.

## Design And Planning Sources

Primary references:

- [doc/design/yggdrasil-mock-server.md](/Users/olivier/Documents/github/chatty-ratatoskr/doc/design/yggdrasil-mock-server.md)
- [doc/design-meta/app.cue](/Users/olivier/Documents/github/chatty-ratatoskr/doc/design-meta/app.cue)
- [scratch/prompt.md](/Users/olivier/Documents/github/chatty-ratatoskr/scratch/prompt.md)

The implementation plan in `scratch/prompt.md` is intentionally phased so one real workflow works early, e2e coverage becomes the feature ledger, and the repo workflow stays aligned with what humans and agents are expected to run.

## Repository Layout

- [cmd/chatty](/Users/olivier/Documents/github/chatty-ratatoskr/cmd/chatty): current Go CLI entrypoint
- [doc/design/yggdrasil-mock-server.md](/Users/olivier/Documents/github/chatty-ratatoskr/doc/design/yggdrasil-mock-server.md): primary product/design spec
- [doc/design-meta](/Users/olivier/Documents/github/chatty-ratatoskr/doc/design-meta): spec-first design inputs used by `flyb`
- [script/e2e](/Users/olivier/Documents/github/chatty-ratatoskr/script/e2e): Bun-powered TypeScript e2e tests
- [scratch/prompt.md](/Users/olivier/Documents/github/chatty-ratatoskr/scratch/prompt.md): architecture and phased implementation plan
- `pipeline-*.thoth.cue`: metadata/reporting pipelines run by the external `thoth` CLI
