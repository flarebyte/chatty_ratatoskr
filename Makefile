## Makefile: thin wrappers for local/dev/release commands.
## Keep target behavior explicit and deterministic.
## Human reviewers should periodically run `make review`.
## Agents should run at least `make format` before handing work back.
## Build artifacts are produced under ./build by build-go.ts.
## Release publishing is handled by release-go.ts.

.PHONY: lint format test test-race gen docs-gen build build-dev e2e release clean help bench perf-smoke contract-snapshots release-check complexity sec dup review thoth-meta-go thoth-meta-go-test thoth-meta-ts-e2e thoth-lint-go thoth-meta-merge

BIOME := ./node_modules/.bin/biome
JSCPD := ./node_modules/.bin/jscpd
BUN := bun
GO := go
GOLINT := golangci-lint
FLYB := flyb
THOTH := thoth
BIN := chatty
GO_CACHE_ENV := GOCACHE=$(PWD)/.gocache GOMODCACHE=$(PWD)/.gomodcache
GOLANGCI_LINT_CACHE := $(PWD)/.golangci-lint-cache
JSCPD_IGNORE := .git/**,.gocache/**,.gomodcache/**,script/e2e/.gocache/**,script/e2e/.gomodcache/**,node_modules/**,build/**,.e2e-bin/**,temp/**

lint:
	@if [ -x "$(BIOME)" ]; then \
		$(BIOME) check; \
	else \
		printf '%s\n' 'Skipping Biome check: @biomejs/biome is not installed locally.'; \
	fi
	$(GO_CACHE_ENV) $(GO) vet ./...
	GOLANGCI_LINT_CACHE=$(GOLANGCI_LINT_CACHE) $(GO_CACHE_ENV) $(GOLINT) run

format:
	find . -name '*.go' \
		-not -path './.git/*' \
		-not -path './.gocache/*' \
		-not -path './.gomodcache/*' \
		-not -path '*/.gocache/*' \
		-not -path '*/.gomodcache/*' \
		-exec gofmt -w {} +
	@if [ -x "$(BIOME)" ]; then \
		$(BIOME) format --write .; \
		$(BIOME) check --unsafe --write; \
	else \
		printf '%s\n' 'Skipping Biome format/check: @biomejs/biome is not installed locally.'; \
	fi

test: gen
	$(GO_CACHE_ENV) $(GO) test -coverprofile=coverage.out ./...
	$(GO_CACHE_ENV) $(GO) tool cover -func=coverage.out

test-race: gen
	$(GO_CACHE_ENV) $(GO) test -race ./...

bench:
	$(GO_CACHE_ENV) $(GO) test -bench=. -run=^$$ ./...

perf-smoke:
	$(GO_CACHE_ENV) $(GO) test ./... -run TestPerfSmoke_

contract-snapshots:
	$(GO_CACHE_ENV) $(GO) test ./... -run TestContract_

gen:
	$(MAKE) doc-gen

doc-gen:
	$(FLYB) validate --config doc/design-meta
	$(FLYB) generate markdown --config doc/design-meta

build:
	$(BUN) run build-go.ts

build-dev:
	mkdir -p .e2e-bin
	$(GO_CACHE_ENV) CGO_ENABLED=0 $(GO) build -o .e2e-bin/$(BIN) ./cmd/$(BIN)

e2e:
	cd script/e2e && $(BUN) test

release: release-check
	$(BUN) run release-go.ts

release-check: lint test e2e contract-snapshots

clean:
	rm -rf build .e2e-bin coverage.out temp thoth-meta .gocache .gomodcache .golangci-lint-cache

complexity:
	scc --sort complexity --by-file -i go . | head -n 15
	scc --sort complexity --by-file -i ts . | head -n 15

sec:
	@if semgrep --version >/dev/null 2>&1; then \
		semgrep scan --config auto; \
	else \
		printf '%s\n' 'Security scan unavailable: semgrep could not start in this environment (local X509 trust store issue).'; \
		printf '%s\n' 'Fix the local semgrep installation or trust store, then rerun `make sec`.'; \
		exit 1; \
	fi
dup:
	@if [ -x "$(JSCPD)" ]; then \
		$(JSCPD) --format go --min-lines 10 --gitignore --ignore "$(JSCPD_IGNORE)" .; \
		$(JSCPD) --format typescript --min-lines 15 --gitignore --ignore "$(JSCPD_IGNORE)" .; \
	else \
		printf '%s\n' 'Skipping duplication scan: jscpd is not installed locally.'; \
	fi

review: format test e2e lint

thoth-meta-go:
	$(THOTH) run --config ./pipeline-go-maat.thoth.cue

thoth-meta-go-test:
	$(THOTH) run --config ./pipeline-go-test-maat.thoth.cue

thoth-meta-ts-e2e:
	$(THOTH) run --config ./pipeline-ts-e2e-maat.thoth.cue

view-thoth-meta-ts-e2e:
	find thoth-meta/ts-e2e/script/e2e/*.thoth.yaml -exec yq '.meta.testcase_titles_list' {} \;

thoth-lint-go:
	$(THOTH) run --config ./pipeline-go-function-thresholds.thoth.cue
	cat temp/pipeline-go-function-thresholds.json | jq '.meta.reduced.worstOffenders'

thoth-meta-merge:
	$(THOTH) run --config ./pipeline-thoth-meta-aggregate.thoth.cue
	
help:
	@printf "Targets:\n"
	@printf "  (core tools: go, bun, golangci-lint, flyb, thoth, biome, jscpd)\n"
	@printf "  lint               Run linters (Biome + go vet + golangci-lint).\n"
	@printf "  format             Apply formatting (gofmt + Biome when locally installed).\n"
	@printf "  test               Run Go tests + coverage summary.\n"
	@printf "  test-race          Run Go tests with race detector.\n"
	@printf "  bench              Run Go benchmarks.\n"
	@printf "  perf-smoke         Run performance smoke tests.\n"
	@printf "  contract-snapshots Run contract snapshot tests.\n"
	@printf "  release-check      Run lint + tests + e2e + contract snapshots.\n"
	@printf "  gen                Generate repo artifacts.\n"
	@printf "  docs-gen           Generate design docs from flyb config.\n"
	@printf "  build              Build release binaries for chatty into ./build.\n"
	@printf "  build-dev          Build the local chatty binary into .e2e-bin/.\n"
	@printf "  e2e                Run Bun-powered end-to-end tests.\n"
	@printf "  release            Run release checks, build artifacts, publish GitHub release.\n"
	@printf "  clean              Remove build artifacts.\n"
	@printf "  complexity         Show top file complexity (Go/TS).\n"
	@printf "  sec                Run security scan (semgrep).\n"
	@printf "  dup                Run duplication scans with local jscpd (go/typescript).\n"
	@printf "  review             Human reviewer command: format + test + e2e + lint.\n"
	@printf "  thoth-meta-go      Collect metadata for non-test Go files via thoth.\n"
	@printf "  thoth-meta-go-test Collect metadata for Go test files via thoth.\n"
	@printf "  thoth-meta-ts-e2e  Collect metadata for TypeScript e2e tests via thoth.\n"
	@printf "  thoth-lint-go      Report large/complex Go functions via thoth.\n"
	@printf "  thoth-meta-merge   Aggregate persisted thoth metadata.\n"
