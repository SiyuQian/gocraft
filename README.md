# gocraft

Opinionated Go project scaffolder. Locks a single tech stack (chi/stdlib +
pgx + golang-migrate + oapi-codegen + slog + Sentry) and generates a
ready-to-run service.

## Install

```bash
go install github.com/siyuqian/gocraft/cmd/gocraft@latest
```

Or from a clone:

```bash
make install   # go install ./cmd/gocraft → $GOBIN (defaults to ~/go/bin)
make build     # local build to ./bin/gocraft
```

Make sure `$GOBIN` (or `~/go/bin`) is on your `PATH`.

## Quick start

```bash
gocraft new                  # interactive TUI — prompts for everything
gocraft new myapp            # name on the CLI, TUI for the rest
cd myapp && make run
```

`gocraft new` runs `go mod tidy` for you after scaffolding, so the
generated project is ready to build immediately. If tidy fails (e.g. no
network), gocraft prints a warning and you can rerun it manually.

## Non-interactive

Pass every option as a flag and add `--no-tui`:

```bash
gocraft new myapp \
  --module github.com/you/myapp \
  --http chi \
  --async none \
  --sentry \
  --no-tui
```

Any flag you pass also skips its TUI step, so you can mix the two:

```bash
gocraft new myapp --module github.com/you/myapp --http stdlib --async river
# TUI still prompts for: Sentry, output directory
```

## Flags

| Flag | Values | Default | Description |
|------|--------|---------|-------------|
| `--module` | Go module path | — | e.g. `github.com/you/myapp` |
| `--http` | `chi`, `stdlib` | `chi` | HTTP layer |
| `--async` | `none`, `river`, `pool` | `none` | Background job backend |
| `--sentry` / `--no-sentry` | — | on | Sentry observability |
| `--output` | path | `./<name>` | Output directory |
| `--no-tui` | — | off | Fail on missing options instead of prompting |

Project names must match `^[a-z][a-z0-9_-]*$`.

## What you get

A generated project with:

```sh
make run        # run locally
make build      # build binary
make test       # unit tests
make lint       # golangci-lint
make tidy       # go mod tidy
```

…plus migrations, OpenAPI codegen wiring, structured logging, and (if
enabled) Sentry initialisation.

## Development

```bash
make test
make build
```
