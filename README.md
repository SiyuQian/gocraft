# gocraft

Opinionated Go project scaffolder. Locks a single tech stack (chi or stdlib
`net/http` + pgx + golang-migrate + slog + optional Sentry + optional async
backend) and generates a runnable service.

## Generated stack

| Layer | Choice | Toggle |
|---|---|---|
| HTTP | chi or stdlib `net/http` | `--http chi\|stdlib` |
| DB | `jackc/pgx/v5` + `pgxpool` + struct→NamedArgs helper | locked |
| Migrations | `golang-migrate` | locked |
| Config | stdlib `flag` + `os.Getenv` | locked |
| Logging | `log/slog` | locked |
| Test | stdlib `testing` + docker-compose stack | locked |
| Async | none, river (Postgres-backed), pool (goroutine pool) | `--async` |
| Observability | Sentry on/off | `--sentry` / `--no-sentry` |
| Container | Dockerfile + docker-compose | always |
| CI | GitHub Actions test-and-lint | always |

## Install

```bash
go install github.com/siyuqian/gocraft/cmd/gocraft@latest
```

Or from source:

```bash
git clone https://github.com/SiyuQian/gocraft
cd gocraft
make build
./bin/gocraft --help
```

## Usage

```bash
# Interactive TUI
gocraft new

# Non-interactive (for CI / AI callers)
gocraft new myapp \
  --module github.com/you/myapp \
  --http chi \
  --async river \
  --sentry

# Skip post-processing
gocraft new myapp --module github.com/you/myapp --no-tidy --no-git --no-tui
```

After generation:

```bash
cd myapp
make run          # starts the HTTP server on :8080
make test         # unit tests
make migrate-up   # apply migrations (needs DATABASE_URL)
```

## Development

```bash
make test                                            # unit tests
go test -tags=acceptance -timeout=20m ./internal/generate/...   # full matrix
```

## Project layout (generated)

See `docs/superpowers/specs/2026-05-18-gocraft-design.md` §5 for the full
file tree. Highlights:

- `internal/` uses vertical slices (`health/`, `pets/`, …) — each owns its
  handler, service, repo, and tests
- `internal/platform/` is the cross-slice infrastructure (`config`, `db`,
  `httpserver`, `obs`, `worker`)
- `internal/api/types.go` carries shared request/response types
- `api/openapi.yaml` is the spec document (oapi-codegen wiring is a planned
  follow-up — types are hand-written today)
- `migrations/` is at repo root for easy mount into CI / docker entrypoints

## Status

v1 complete (M1–M8). Roadmap for v2: `gocraft add slice|handler|migration`
incremental generators, true oapi-codegen wiring with committed gen output,
Prometheus / OpenTelemetry presets.
