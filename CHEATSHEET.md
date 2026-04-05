# CHEATSHEET

Quick command reference for `chatty`.

## Build

```bash
make build-dev
./.e2e-bin/chatty version
```

## Help

```bash
./.e2e-bin/chatty help
./.e2e-bin/chatty --help
```

## Serve

```bash
./.e2e-bin/chatty serve
./.e2e-bin/chatty serve --listen 127.0.0.1:19080
./.e2e-bin/chatty serve --config testdata/config/basic.cue
```

## Review

```bash
make review
make e2e
make build
```

## Maintenance Gates

```bash
make dup
make complexity
make sec
```

## Notes

- `README.md` is for CLI users
- `CONTRIBUTING.md` is for maintainers and agents
- admin and websocket behavior depend on config
