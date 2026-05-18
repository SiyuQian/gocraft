# gocraft M5 — pgx + migrations + pets slice Plan

**Goal:** Generated projects ship with a real `pets` slice that exercises the chosen DB stack: pgx pool, NamedArgs helper, migrations, and CRUD against Postgres. The openapi.yaml grows to document `/pets`. Codegen (oapi-codegen) wiring is **deferred** — types are hand-written in `internal/api/types.go.tmpl`. We'll note this in the PR; a follow-up can introduce true codegen without changing the rest.

**Architecture:** All changes are additions to `templates/base/`. No new layer needed — pgx is locked for all configs. New files include the platform/db helper, migrations, pets handler/service/repo, hand-written types, and Makefile targets. `cmd/main.go` is updated to open a pool and wire pets into the HTTP server. The chi and stdlib server overlays are updated to accept a `*pgxpool.Pool` and register pets routes.

---

## File changes (all under `templates/`)

### New: `base/internal/platform/db/db.go.tmpl`
- `Open(ctx, dsn) (*pgxpool.Pool, error)` thin wrapper around `pgxpool.New`
- `NamedArgs(struct) pgx.NamedArgs` reflect helper: walks struct fields with `db:"name"` tag and builds a `pgx.NamedArgs` map

### New: `base/internal/platform/db/db_test.go.tmpl`
- Unit tests for `NamedArgs` against a sample struct (no pool needed)

### New: `base/migrations/0001_init.up.sql`
```sql
CREATE TABLE pets (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    species    TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### New: `base/migrations/0001_init.down.sql`
```sql
DROP TABLE IF EXISTS pets;
```

### New: `base/internal/api/types.go.tmpl` (hand-written, replaces planned `gen/`)
```go
package api

import "time"

type Pet struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Species   string    `json:"species"`
    CreatedAt time.Time `json:"created_at"`
}

type CreatePetRequest struct {
    Name    string `json:"name"`
    Species string `json:"species"`
}
```

### New: `base/internal/pets/handler.go.tmpl`
- `Handler` struct with a `*Service`
- `Register(r Router)` where `Router` is a tiny interface satisfied by both chi and the stdlib mux
- Methods: `list`, `create`, `get` mapped to `/pets`, `POST /pets`, `/pets/{id}`

### New: `base/internal/pets/service.go.tmpl`
- `Service` with a `*Repo`; pass-through CRUD; UUID generation via `crypto/rand`

### New: `base/internal/pets/repo.go.tmpl`
- `Repo` wrapping `*pgxpool.Pool`
- `List`, `Insert`, `Get` using `pgx.NamedArgs` + `pgx.CollectRows[Pet]` + `pgx.RowToStructByName`
- Uses `= ANY(@ids)` style for any future bulk fetch (just documented in a comment)

### New: `base/internal/pets/repo_test.go.tmpl` (build tag `integration`)
- Connects to `postgres://...@localhost:55432/{{.Name}}_test` (matches `docker-compose.test.yml`)
- Runs migrations programmatically, exercises List/Insert/Get

### New: `base/internal/pets/service_test.go.tmpl`
- Pure unit tests using a fake repo

### New: `base/tools.go.tmpl`
```go
//go:build tools
package tools

import (
    _ "github.com/golang-migrate/migrate/v4/cmd/migrate"
)
```

### New: `base/scripts/migrate.sh`
```sh
#!/bin/sh
set -e
migrate -path migrations -database "$DATABASE_URL" "$@"
```

### Modified: `base/go.mod.tmpl`
Add unconditional pgx + migrate requires. Resulting template:
```
module {{.Module}}

go 1.22

require (
	github.com/jackc/pgx/v5 v5.7.0
	github.com/golang-migrate/migrate/v4 v4.18.1
{{- if eq .HTTP "chi"}}
	github.com/go-chi/chi/v5 v5.1.0
{{- end}}
)
```

### Modified: `base/Makefile.tmpl`
Add:
```
migrate-up:
	./scripts/migrate.sh up

migrate-down:
	./scripts/migrate.sh down 1
```

### Modified: `base/cmd/{{.Name}}/main.go.tmpl`
Open pgx pool from env `DATABASE_URL`; pass to httpserver. If DSN is empty, run with nil pool (pets routes register only when pool != nil).

### Modified: `base/internal/platform/httpserver/server.go.tmpl` (the stub)
Accept `*pgxpool.Pool` parameter (ignored in stub).

### Modified: `http/chi/internal/platform/httpserver/server.go.tmpl`
Take pool, wire pets handler when pool is non-nil.

### Modified: `http/stdlib/internal/platform/httpserver/server.go.tmpl`
Same.

### New: `base/configs/config.example.env` (overwrite)
```
APP_ADDR=:8080
LOG_LEVEL=info
DATABASE_URL=postgres://{{.Name}}:{{.Name}}@localhost:5432/{{.Name}}?sslmode=disable
```

---

## Engine considerations

The renderer's `text/template` runs file contents through Go's template engine, but **template files often need to emit literal `{{` for downstream tools** (e.g. embedded Go code generators). Not an issue here — none of M5's files emit raw braces. But if any are added later they need `{{"{{"}}` escaping.

The renderer currently runs `go test ./...` in the acceptance test, which will hit `repo_test.go` if no build tag — that's why `repo_test.go` uses `//go:build integration`.

---

## Acceptance test update

Modify `internal/generate/acceptance_test.go`:
- Continue testing both http layers
- Confirm `go vet`, `go build`, `go test ./...` pass against generated project (no integration tag — repo_test is excluded)
- Confirm `migrations/0001_init.up.sql` exists
- Do NOT attempt to run migrations or pets repo_test (would need a real postgres)

---

## Tasks

1. **Add all new files + update modified files** as listed (one big commit).
2. **Update acceptance test** to assert new files exist.
3. **Run** `go test ./...` (default) and `go test -tags=acceptance ./internal/generate/...` — both must pass.
4. **Commit** and push.

## Exit criteria

- `go vet ./... && go test ./...` passes (default tests)
- `go test -tags=acceptance ./internal/generate/...` passes for both chi and stdlib variants
- Generated project's go.mod includes pgx + migrate
- Generated project's `internal/pets/` has the slice
- `migrations/0001_init.{up,down}.sql` ships
- PR description notes oapi-codegen is deferred

## Out of scope (deferred)

- True oapi-codegen wiring (use generated code instead of hand-written types). Can be added in a follow-up without breaking existing template consumers.
- Pets `Update` and `Delete` endpoints (CRUD subset for brevity).
- Pagination for List.
- Integration test execution in `go test ./...` (requires running compose stack).
