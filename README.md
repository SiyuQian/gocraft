# gocraft

Opinionated Go project scaffolder. Locks a single tech stack (chi/stdlib +
pgx + golang-migrate + oapi-codegen + slog + Sentry) and generates a
ready-to-run service.

## Status

Pre-alpha. M1 (CLI skeleton) only — `gocraft new` resolves a config but does
not yet generate files.

## Usage

```bash
go run ./cmd/gocraft new                   # interactive TUI
go run ./cmd/gocraft new myapp \
  --module github.com/you/myapp \
  --http chi --async none --sentry \
  --no-tui                                  # non-interactive
```

## Development

```bash
make test
make build
```
