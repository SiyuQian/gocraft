# gocraft — Design Spec

**Date:** 2026-05-18
**Status:** Approved, ready for implementation planning

## 1. Purpose

`gocraft` is a Go project scaffolder, similar in spirit to `go-blueprint`, but
opinionated around a single tech stack tailored to the author's workflow. It
generates a ready-to-run Go service with an OpenAPI spec-first API layer,
Postgres via pgx, and idiomatic project layout.

The scaffolder is intentionally narrow: instead of offering many framework
choices, it locks down most of the stack and exposes only a small number of
toggles. Forking is the escape hatch for unsupported stacks.

## 2. Scope

### In scope (v1)

- A single CLI command `gocraft new` that produces a fully runnable Go
  project on disk.
- TUI wizard (default when run without args) and flag-based mode (for CI/AI
  callers). Flags override TUI prompts.
- Locked stack: chi or stdlib `net/http`; pgx native; golang-migrate;
  oapi-codegen (strict server, committed); stdlib `flag` + `os.Getenv`;
  `log/slog`; stdlib `testing`; Sentry; Dockerfile + docker-compose; GitHub
  Actions test-and-lint.
- Optional toggles: HTTP layer (chi / stdlib), async backend
  (none / river / pool), Sentry (on / off).
- Post-generation: `goimports -w .`, `go mod tidy`, optional `git init`.
- Generated project must `make run` successfully and pass `make
  test-integration` against the bundled `docker-compose.test.yml`.

### Explicitly out of scope (v1)

- Incremental generators (`gocraft add slice|handler|migration`). Deferred to
  v2.
- Additional HTTP frameworks (echo, gin, fiber, etc.).
- Additional databases (MySQL, SQLite).
- Alternative ORMs (sqlx, sqlc, gorm, bun, ent).
- Prometheus / OpenTelemetry preset.
- Frontend / SPA scaffolding.
- Auto-generated AGENTS.md / CLAUDE.md context files (deferred).
- Monorepo or multi-module layouts.

## 3. Locked stack

| Layer | Choice | Toggle? |
|---|---|---|
| HTTP | `go-chi/chi` or stdlib `net/http` | yes (`--http`) |
| API contract | `oapi-codegen` (strict server, committed gen) | no |
| DB driver | `jackc/pgx/v5` + `pgxpool` | no |
| DB scan & params | `pgx.NamedArgs` (`@name`) + `pgx.CollectRows` + `RowToStructByName[T]` | no |
| Migrations | `golang-migrate/migrate` | no |
| Config | stdlib `flag` + `os.Getenv` | no |
| Logging | `log/slog` | no |
| Test | stdlib `testing` | no |
| Integration test infra | docker-compose (`docker-compose.test.yml`) — long-running stack, reused across runs | no |
| Observability | Sentry (`getsentry/sentry-go` + `sentryhttp`) | yes (`--sentry` / `--no-sentry`) |
| Async tasks | none / `riverqueue/river` / stdlib goroutine pool | yes (`--async`) |
| Container | Dockerfile + docker-compose | no, always |
| CI | GitHub Actions test-and-lint | no, always |

### Notable rationale

- **pgx native over sqlx**: pgx v5's `NamedArgs` (`@name` syntax) combined
  with `pgx.CollectRows` and `pgx.RowToStructByName[T]` provides struct scan +
  named binding with zero extra dependencies, and avoids the sqlx three-step
  `IN` clause dance (`= ANY(@ids)` works directly with a slice).
- **Integration tests via docker-compose, not testcontainers-go**: a
  persistent compose stack is reused across runs and is significantly faster
  than per-suite container startup.
- **Strict, committed oapi-codegen output**: PRs show contract changes; new
  contributors do not need to run codegen before opening the project in an
  IDE; type-safe handler signatures prevent status-code drift from the spec.

## 4. CLI UX

### Subcommands (v1)

```
gocraft new [project-name] [flags]
gocraft version
gocraft --help
```

### `new` interaction modes

**No args → TUI wizard** (bubbletea + huh):

```
gocraft new
  ▸ Project name:        ___________
  ▸ Module path:         github.com/you/myapp
  ▸ HTTP layer:          [chi]  stdlib
  ▸ Async tasks:         [none] river  pool
  ▸ Sentry:              [yes]  no
  ▸ Output directory:    ./myapp
  [ Create ]
```

**With flags → skips TUI** (intended for CI / AI callers):

```
gocraft new myapp \
  --module github.com/you/myapp \
  --http chi \
  --async river \
  --sentry \
  --output ./
```

Partial flags are allowed: any unspecified option is asked via TUI; if
`--no-tui` is passed, missing required options become errors.

### Flag set

| Flag | Type | Default | Notes |
|---|---|---|---|
| `--module` | string | — | required (TUI or flag) |
| `--http` | enum | `chi` | `chi` \| `stdlib` |
| `--async` | enum | `none` | `none` \| `river` \| `pool` |
| `--sentry` / `--no-sentry` | bool | `true` | |
| `--output` | path | `./<name>` | |
| `--no-tui` | bool | `false` | fails fast on missing options |
| `--no-git` | bool | `false` | skip `git init` |

## 5. Generated project layout

Example with `--http chi --async river --sentry`:

```
myapp/
├── api/
│   └── openapi.yaml
├── cmd/
│   └── myapp/main.go
├── configs/
│   └── config.example.env
├── deployments/
│   ├── Dockerfile
│   ├── docker-compose.yml
│   └── docker-compose.test.yml
├── internal/
│   ├── api/gen/                  # oapi-codegen output (committed)
│   │   ├── types.gen.go
│   │   └── server.gen.go
│   ├── platform/                 # cross-slice infrastructure
│   │   ├── config/
│   │   ├── db/                   # pgx pool + struct→NamedArgs helper
│   │   ├── httpserver/
│   │   ├── obs/                  # Sentry init
│   │   └── worker/               # river client (if async=river)
│   ├── health/                   # example slice (minimal)
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── handler_test.go
│   └── pets/                     # example slice (CRUD + DB)
│       ├── handler.go
│       ├── service.go
│       ├── repo.go
│       ├── repo_test.go          # integration test
│       └── service_test.go       # unit test
├── migrations/
│   ├── 0001_init.up.sql
│   └── 0001_init.down.sql
├── scripts/
│   ├── generate.sh
│   └── migrate.sh
├── .github/workflows/test-and-lint.yml
├── .gitignore
├── .golangci.yml
├── go.mod
├── go.sum
├── Makefile
├── README.md
└── tools.go
```

### Layout rules

- **Vertical slice** inside `internal/`: one folder per business feature
  (`pets/`, `users/`, `orders/`), each owning its handler + service + repo +
  tests.
- **Slices do not import each other.** Shared code lives in
  `internal/platform/` (infrastructure) or in a future `internal/sharedkernel/`
  (shared business primitives). Heuristic: if reusing this code from a new
  slice is natural, it belongs in platform.
- **All slice handlers implement interfaces from `internal/api/gen/`**.
  `cmd/myapp/main.go` wires every slice handler into a single strict server
  implementation and registers routes via the generated mux registrar.
- **`internal/api/gen/` is committed**, marked `// Code generated by
  oapi-codegen. DO NOT EDIT.`, and placed under `internal/` so external repos
  cannot import it.
- **`migrations/` at repo root** so CI and docker entrypoints can mount it
  directly. Filenames follow the golang-migrate convention
  (`<version>_<name>.{up,down}.sql`).
- **`tools.go`** with `//go:build tools` pins the versions of `oapi-codegen`
  and `migrate` via blank imports, ensuring all contributors run the same
  generator versions.

### Makefile targets (generated)

```
make run               # go run ./cmd/<name>
make build             # go build -o bin/<name> ./cmd/<name>
make test              # go test ./...
make test-integration  # docker compose -f deployments/docker-compose.test.yml up -d && go test -tags=integration ./...
make lint              # golangci-lint run
make gen               # go generate ./... (oapi-codegen)
make migrate-up        # migrate -path migrations -database $$DATABASE_URL up
make migrate-down      # migrate -path migrations -database $$DATABASE_URL down 1
make tidy              # go mod tidy && goimports -w .
```

## 6. gocraft tool architecture

### Stack of the tool itself

- **CLI framework**: `cobra`
- **TUI**: `bubbletea` + `charmbracelet/huh` (form builder, avoids hand-rolled
  state machines)
- **Templating**: stdlib `text/template` with templates embedded via
  `embed.FS`
- **Post-processing**: shell out to `goimports`, `go mod tidy`, `git init`

### Repo layout (the gocraft repo itself)

```
gocraft/
├── cmd/gocraft/main.go
├── internal/
│   ├── tui/         # bubbletea + huh forms
│   ├── prompt/      # merge flags + TUI answers into Config struct
│   ├── generate/    # template rendering + file writing
│   └── postproc/    # goimports / go mod tidy / git init
├── templates/       # embedded into binary
│   ├── base/                       # always generated
│   ├── http/{chi,stdlib}/          # one of
│   ├── async/{river,pool}/         # at most one
│   └── obs/sentry/                 # optional
└── go.mod
```

Template subfolders are organized hierarchically (subdomain style) rather
than flat with underscores. Adding a new option (e.g. a redis cache layer)
means adding `cache/redis/`, not a new `cache_redis_` prefix.

### Rendering algorithm — layered overlay

1. Always lay down `templates/base/` onto the output directory.
2. Append optional layers in a deterministic order:
   ```go
   layers := []string{"base"}
   layers = append(layers, "http/"+cfg.HTTP)
   if cfg.Async != "none" {
       layers = append(layers, "async/"+cfg.Async)
   }
   if cfg.Sentry {
       layers = append(layers, "obs/sentry")
   }
   ```
3. Within each layer, each `*.tmpl` file is rendered with the same `Config`
   struct (`{{.Name}}`, `{{.Module}}`, `{{.HTTPLayer}}`, `{{.HasRiver}}`,
   `{{.Sentry}}`, ...) and written to the path implied by its location, with
   `{{.Name}}` substitutable in path segments.
4. A later layer's file overwrites an earlier layer's file at the same
   relative path. This is how, e.g., `http/chi/cmd/{{.Name}}/main.go.tmpl`
   can specialize the main file from `base/`.
5. Non-`.tmpl` files (e.g. `.gitignore`) are copied verbatim.

### Post-processing pipeline

1. Walk output dir, render templates, write files.
2. `go mod tidy` in output dir.
3. `goimports -w .` in output dir.
4. If `!--no-git`: `git init && git add . && git commit -m "initial scaffold"`.
5. Print a success message with `cd <output> && make run` instructions.

### Validation strategy (development-time)

- Each layer combination (`chi|stdlib` × `none|river|pool` × `sentry|nosentry`
  = 12 combinations) is exercised by a **golden test**: generate the project
  into a tempdir, then `go vet`, `go build`, and `go test ./...` against it.
- Golden tests run in CI for the gocraft repo itself.

## 7. Milestones

1. **M1** — CLI skeleton: cobra command, huh form, Config struct, option
   validation. No file output yet.
2. **M2** — Template rendering engine: `embed.FS`, layered overlay,
   `text/template` execution, path interpolation.
3. **M3** — `base/` + `http/chi/` templates producing a runnable hello-world
   (health endpoint live, `make run` works).
4. **M4** — `http/stdlib/` path; both HTTP options validated.
5. **M5** — pgx integration, migrations, `pets/` example slice, integration
   test wiring with `docker-compose.test.yml`.
6. **M6** — `async/river/` and `async/pool/` layers.
7. **M7** — `obs/sentry/` layer.
8. **M8** — Post-processing (goimports, mod tidy, git init), error message
   polish, golden tests covering all 12 combinations.

## 8. Open questions / future work

- v2: `gocraft add slice <name>`, `gocraft add handler`, `gocraft add
  migration` — incremental generators using the same template engine.
- v2: optional AGENTS.md / CLAUDE.md generator with project-specific
  context for AI agents.
- v2: Prometheus and OpenTelemetry presets (decoupled from Sentry layer).
- v2: optional client codegen from the same OpenAPI spec.
- Plugin model for third-party layers (very speculative).
