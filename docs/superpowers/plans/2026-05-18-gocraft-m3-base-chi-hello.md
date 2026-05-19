# gocraft M3 — Base + chi → Runnable Hello-World Plan

**Goal:** Add real template files for the `base/` and `http/chi/` layers, embed them, wire `generate.Render` into the `new` command, and verify the generated project compiles and starts a chi server with `/healthz`.

**Scope notes:** M5 adds pgx/migrations/oapi-codegen/pets. M6/M7 add async + sentry. So M3's generated project has: stdlib config, chi router, hand-written health handler, slog logging, no DB, no observability beyond log, no async, no codegen.

---

## Template files to create

All paths below are relative to `templates/`. Use `{{.Name}}`, `{{.Module}}` etc. Files ending in `.tmpl` go through `text/template`; others are verbatim.

### base/ layer

| Path | Notes |
|---|---|
| `base/.gitignore` | verbatim — `bin/\n*.test\n*.out\n.DS_Store\nvendor/\n` |
| `base/.golangci.yml` | verbatim minimal config — see snippet below |
| `base/go.mod.tmpl` | `module {{.Module}}\n\ngo 1.22\n` then conditional require blocks |
| `base/Makefile.tmpl` | targets: `run build test test-integration lint tidy` |
| `base/README.md.tmpl` | `# {{.Name}}` + a brief usage |
| `base/cmd/{{.Name}}/main.go.tmpl` | stdlib `net/http` stub (overridden by `http/chi/` layer) |
| `base/configs/config.example.env` | `APP_ADDR=:8080\nLOG_LEVEL=info\n` |
| `base/internal/platform/config/config.go.tmpl` | `Load()` reads env + parses flags |
| `base/internal/platform/httpserver/server.go.tmpl` | stub (overridden by chi) |
| `base/internal/platform/obs/obs.go.tmpl` | no-op `Init()` and `Flush()` (overridden by sentry layer) |
| `base/internal/health/handler.go.tmpl` | `Healthz(w, r)` writes `{"status":"ok"}` |
| `base/internal/health/handler_test.go.tmpl` | `httptest.NewRecorder` happy path |
| `base/api/openapi.yaml.tmpl` | tiny spec with `/healthz GET` operation |
| `base/deployments/Dockerfile` | verbatim multi-stage Go 1.22 |
| `base/deployments/docker-compose.yml.tmpl` | app + postgres (postgres unused in M3 but available) |
| `base/deployments/docker-compose.test.yml.tmpl` | postgres only on host port 55432 |
| `base/.github/workflows/test-and-lint.yml` | verbatim — Go 1.22, `go test ./...` + `golangci-lint` |

#### `base/go.mod.tmpl`

```
module {{.Module}}

go 1.22
{{- if eq .HTTP "chi"}}

require github.com/go-chi/chi/v5 v5.1.0
{{- end}}
```

#### `base/Makefile.tmpl`

```
.PHONY: run build test test-integration lint tidy

run:
	go run ./cmd/{{.Name}}

build:
	go build -o bin/{{.Name}} ./cmd/{{.Name}}

test:
	go test ./...

test-integration:
	docker compose -f deployments/docker-compose.test.yml up -d
	go test -tags=integration ./...

lint:
	golangci-lint run

tidy:
	go mod tidy
```

#### `base/cmd/{{.Name}}/main.go.tmpl` (stub)

```go
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"{{.Module}}/internal/platform/config"
	"{{.Module}}/internal/platform/httpserver"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("config load", "err", err)
		os.Exit(1)
	}

	srv := httpserver.New(cfg, logger)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("listening", "addr", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("listen", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
```

#### `base/internal/platform/config/config.go.tmpl`

```go
package config

import (
	"flag"
	"os"
)

type Config struct {
	Addr     string
	LogLevel string
}

func Load() (Config, error) {
	c := Config{
		Addr:     getenv("APP_ADDR", ":8080"),
		LogLevel: getenv("LOG_LEVEL", "info"),
	}
	flag.StringVar(&c.Addr, "addr", c.Addr, "HTTP listen address")
	flag.StringVar(&c.LogLevel, "log-level", c.LogLevel, "log level")
	flag.Parse()
	return c, nil
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
```

#### `base/internal/platform/httpserver/server.go.tmpl` (stub, overridden by chi)

```go
package httpserver

import (
	"log/slog"
	"net/http"

	"{{.Module}}/internal/platform/config"
)

func New(cfg config.Config, logger *slog.Logger) *http.Server {
	mux := http.NewServeMux()
	return &http.Server{Addr: cfg.Addr, Handler: mux}
}
```

#### `base/internal/platform/obs/obs.go.tmpl`

```go
package obs

import "log/slog"

func Init(logger *slog.Logger) {}
func Flush() {}
```

#### `base/internal/health/handler.go.tmpl`

```go
package health

import (
	"encoding/json"
	"net/http"
)

func Healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
```

#### `base/internal/health/handler_test.go.tmpl`

```go
package health

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthz(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	Healthz(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("code = %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"status":"ok"`) {
		t.Fatalf("body = %s", w.Body)
	}
}
```

#### `base/api/openapi.yaml.tmpl`

```yaml
openapi: 3.0.3
info:
  title: {{.Name}}
  version: 0.1.0
paths:
  /healthz:
    get:
      operationId: getHealthz
      responses:
        '200':
          description: ok
```

#### `base/deployments/Dockerfile`

```dockerfile
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download || true
COPY . .
RUN CGO_ENABLED=0 go build -o /out/app ./cmd/

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/app /app
USER nonroot:nonroot
ENTRYPOINT ["/app"]
```

#### `base/deployments/docker-compose.yml.tmpl`

```yaml
services:
  app:
    build:
      context: ..
      dockerfile: deployments/Dockerfile
    environment:
      APP_ADDR: ":8080"
    ports:
      - "8080:8080"
    depends_on:
      - db
  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: {{.Name}}
      POSTGRES_PASSWORD: {{.Name}}
      POSTGRES_DB: {{.Name}}
    ports:
      - "5432:5432"
```

#### `base/deployments/docker-compose.test.yml.tmpl`

```yaml
services:
  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: {{.Name}}
      POSTGRES_PASSWORD: {{.Name}}
      POSTGRES_DB: {{.Name}}_test
    ports:
      - "55432:5432"
```

#### `base/.github/workflows/test-and-lint.yml`

```yaml
name: test-and-lint
on:
  push: { branches: [main] }
  pull_request:
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22' }
      - run: go vet ./...
      - run: go test ./...
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22' }
      - uses: golangci/golangci-lint-action@v6
        with: { version: latest }
```

#### `base/.golangci.yml`

```yaml
run:
  timeout: 5m
linters:
  enable:
    - errcheck
    - govet
    - ineffassign
    - staticcheck
    - unused
```

### http/chi/ layer

#### `http/chi/internal/platform/httpserver/server.go.tmpl` (overrides base)

```go
package httpserver

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"{{.Module}}/internal/health"
	"{{.Module}}/internal/platform/config"
)

func New(cfg config.Config, logger *slog.Logger) *http.Server {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Get("/healthz", health.Healthz)
	return &http.Server{Addr: cfg.Addr, Handler: r}
}
```

The chi layer's `main.go.tmpl` is identical to base's — we leave the base version in place; chi only overlays `httpserver/server.go`.

---

## Engine wiring & embed

### `templates/embed.go` (new file)

```go
package templates

import "embed"

//go:embed all:base all:http
var FS embed.FS
```

The `all:` prefix forces embedding of dot-files (`.gitignore`, `.golangci.yml`, `.github/`).

### `internal/generate/embed.go` (new file)

Bridge so `Render` can take a sub-FS rooted at `templates/`:

```go
package generate

import (
	"io/fs"

	"github.com/siyuqian/gocraft/templates"
)

// EmbeddedFS returns the embedded template FS. Exposed for the `new` command
// and tests that exercise real templates.
func EmbeddedFS() fs.FS { return templates.FS }
```

### `internal/cli/new.go` changes

Replace the JSON-print logic with a real generation step:

```go
import (
	"fmt"
	"github.com/siyuqian/gocraft/internal/generate"
	// remove "encoding/json"
)

// in RunE, after Validate():
fsys := generate.EmbeddedFS()
if err := generate.Render(cfg, fsys, generate.Layers(cfg), cfg.Output); err != nil {
	return err
}
fmt.Fprintf(cmd.OutOrStdout(), "created %s\nnext: cd %s && make run\n", cfg.Output, cfg.Output)
return nil
```

### `internal/cli/new_test.go` changes

The old tests asserted JSON output; rewrite to assert files exist in a tempdir output. Use `t.TempDir()` for `--output`. Verify presence (not contents):
- `go.mod`
- `cmd/<name>/main.go`
- `internal/platform/httpserver/server.go`
- `internal/health/handler.go`

`TestNew_BadHTTP` and `TestNew_NoTUI_MissingModule` still assert errors (no output dir needed; validation runs before Render).

---

## Acceptance test

`internal/generate/acceptance_test.go` (new file, **build tag `acceptance`** so it doesn't run by default):

```go
//go:build acceptance

package generate_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/siyuqian/gocraft/internal/generate"
	"github.com/siyuqian/gocraft/internal/prompt"
)

func TestAcceptance_ChiBuilds(t *testing.T) {
	cfg := prompt.Config{
		Name:   "demoapp",
		Module: "example.com/demoapp",
		HTTP:   prompt.HTTPChi,
		Async:  prompt.AsyncNone,
		Sentry: false,
		Output: t.TempDir(),
	}
	if err := generate.Render(cfg, generate.EmbeddedFS(), generate.Layers(cfg), cfg.Output); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if _, err := os.Stat(filepath.Join(cfg.Output, "go.mod")); err != nil {
		t.Fatalf("go.mod missing: %v", err)
	}
	mustRun(t, cfg.Output, "go", "mod", "tidy")
	mustRun(t, cfg.Output, "go", "vet", "./...")
	mustRun(t, cfg.Output, "go", "build", "./...")
	mustRun(t, cfg.Output, "go", "test", "./...")
}

func mustRun(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v in %s: %v\n%s", name, args, dir, err, out)
	}
}
```

Acceptance test is run manually / in CI with `go test -tags=acceptance ./internal/generate/...`. Requires network access (`go mod tidy`).

---

## Tasks (sequence for implementation)

1. **Create all base/ templates** as listed above. Commit: `feat(templates): add base layer templates`.
2. **Create http/chi/ templates** as listed. Commit: `feat(templates): add http/chi layer`.
3. **Add templates/embed.go and internal/generate/embed.go**. Commit: `feat(generate): embed templates via embed.FS`.
4. **Rewrite new.go** to call Render; rewrite new_test.go to assert files. Run `go test ./...` until green. Commit: `feat(cli): wire generate.Render into 'new' command`.
5. **Add acceptance_test.go** behind `acceptance` build tag. Run `go test -tags=acceptance ./internal/generate/...` locally to verify (requires network). Commit: `test(generate): add acceptance test that generated project builds`.

## Exit criteria

- `go vet ./... && go test ./...` (default, no tag) passes
- `go build ./cmd/gocraft` succeeds
- `gocraft new demoapp --module example.com/demoapp --http chi --async none --no-sentry --no-tui --output /tmp/demoapp` produces a directory; in that directory `go mod tidy && go build ./...` succeed and `go test ./...` passes
- Acceptance test passes (manual or CI with `-tags=acceptance`)

## Out of scope (deferred)

- stdlib http layer (M4)
- pgx, migrations, oapi-codegen, pets slice (M5)
- async layers (M6)
- sentry layer (M7)
- post-processing automation: `goimports`, `git init`, automatic `go mod tidy` after Render (M8)
