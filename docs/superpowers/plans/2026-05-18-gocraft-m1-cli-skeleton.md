# gocraft M1 — CLI Skeleton Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the gocraft CLI skeleton: cobra `new` and `version` subcommands, a `Config` struct with validation, flag binding, and a huh TUI form. No file generation yet — `new` should resolve a fully-validated `Config` and print it.

**Architecture:** A thin `cmd/gocraft` entry boots a cobra root with `new` and `version` subcommands. The `new` command collects options in three layers: (1) cobra flags, (2) huh TUI form filling in any blanks (unless `--no-tui`), (3) validation against a fixed allow-list. Output of M1 is a printed JSON dump of the resolved `Config`; M2 plugs the template renderer onto the same `Config`.

**Tech Stack:** Go 1.22+, `cobra`, `charmbracelet/huh`, stdlib `encoding/json`, stdlib `testing`.

---

## File Structure

| File | Responsibility |
|---|---|
| `go.mod` | module declaration, deps |
| `cmd/gocraft/main.go` | entry point; calls `cli.Execute()` |
| `internal/cli/root.go` | cobra root command + `version` |
| `internal/cli/new.go` | `new` subcommand: parses flags → Config → optional TUI → validate → print |
| `internal/cli/new_test.go` | end-to-end test of `new` via `cobra.Command.SetArgs` |
| `internal/prompt/config.go` | `Config` struct, `Validate`, constants |
| `internal/prompt/config_test.go` | unit tests for `Validate` |
| `internal/prompt/flags.go` | bind cobra flags onto a `*Config` |
| `internal/prompt/flags_test.go` | flag parsing tests |
| `internal/tui/form.go` | huh form that fills empty fields on a `*Config` |
| `Makefile` | `run` / `test` / `build` |

---

### Task 1: Initialize Go module and repo skeleton

**Files:**
- Create: `go.mod`
- Create: `Makefile`
- Create: `.gitignore`

- [ ] **Step 1: Initialize git repo and Go module**

Run:
```bash
cd /Users/siyu/Works/github.com/siyuqian/gocraft
git init
go mod init github.com/siyuqian/gocraft
```

- [ ] **Step 2: Create `.gitignore`**

Create `.gitignore`:
```
bin/
*.test
*.out
.DS_Store
```

- [ ] **Step 3: Create `Makefile`**

Create `Makefile`:
```makefile
.PHONY: build test run tidy

build:
	go build -o bin/gocraft ./cmd/gocraft

test:
	go test ./...

run:
	go run ./cmd/gocraft

tidy:
	go mod tidy
```

- [ ] **Step 4: Commit**

```bash
git add go.mod Makefile .gitignore
git commit -m "chore: init module and repo skeleton"
```

---

### Task 2: Define `Config` struct and constants

**Files:**
- Create: `internal/prompt/config.go`
- Test: `internal/prompt/config_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/prompt/config_test.go`:
```go
package prompt

import "testing"

func TestValidate_OK(t *testing.T) {
	c := Config{
		Name:   "myapp",
		Module: "github.com/you/myapp",
		HTTP:   HTTPChi,
		Async:  AsyncNone,
		Sentry: true,
		Output: "./myapp",
	}
	if err := c.Validate(); err != nil {
		t.Fatalf("expected ok, got %v", err)
	}
}

func TestValidate_MissingName(t *testing.T) {
	c := Config{Module: "github.com/you/myapp", HTTP: HTTPChi, Async: AsyncNone, Output: "./x"}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestValidate_MissingModule(t *testing.T) {
	c := Config{Name: "myapp", HTTP: HTTPChi, Async: AsyncNone, Output: "./x"}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for empty module")
	}
}

func TestValidate_BadHTTP(t *testing.T) {
	c := Config{Name: "myapp", Module: "m", HTTP: "echo", Async: AsyncNone, Output: "./x"}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for unknown http")
	}
}

func TestValidate_BadAsync(t *testing.T) {
	c := Config{Name: "myapp", Module: "m", HTTP: HTTPChi, Async: "celery", Output: "./x"}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for unknown async")
	}
}

func TestValidate_BadName(t *testing.T) {
	c := Config{Name: "My App!", Module: "m", HTTP: HTTPChi, Async: AsyncNone, Output: "./x"}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for non-identifier name")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/prompt/...`
Expected: FAIL — `undefined: Config`

- [ ] **Step 3: Implement `Config` and `Validate`**

Create `internal/prompt/config.go`:
```go
package prompt

import (
	"fmt"
	"regexp"
)

const (
	HTTPChi    = "chi"
	HTTPStdlib = "stdlib"

	AsyncNone  = "none"
	AsyncRiver = "river"
	AsyncPool  = "pool"
)

var (
	ValidHTTP  = []string{HTTPChi, HTTPStdlib}
	ValidAsync = []string{AsyncNone, AsyncRiver, AsyncPool}

	nameRe = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)
)

type Config struct {
	Name   string
	Module string
	HTTP   string
	Async  string
	Sentry bool
	Output string
	NoTUI  bool
	NoGit  bool
}

func (c Config) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("project name is required")
	}
	if !nameRe.MatchString(c.Name) {
		return fmt.Errorf("project name %q must match %s", c.Name, nameRe)
	}
	if c.Module == "" {
		return fmt.Errorf("module path is required")
	}
	if !contains(ValidHTTP, c.HTTP) {
		return fmt.Errorf("http=%q must be one of %v", c.HTTP, ValidHTTP)
	}
	if !contains(ValidAsync, c.Async) {
		return fmt.Errorf("async=%q must be one of %v", c.Async, ValidAsync)
	}
	if c.Output == "" {
		return fmt.Errorf("output directory is required")
	}
	return nil
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/prompt/...`
Expected: PASS — all 6 subtests.

- [ ] **Step 5: Commit**

```bash
git add internal/prompt/config.go internal/prompt/config_test.go
git commit -m "feat(prompt): add Config struct and Validate"
```

---

### Task 3: Cobra flag binding

**Files:**
- Create: `internal/prompt/flags.go`
- Test: `internal/prompt/flags_test.go`

- [ ] **Step 1: Add cobra dependency**

Run:
```bash
go get github.com/spf13/cobra@latest
```

- [ ] **Step 2: Write the failing test**

Create `internal/prompt/flags_test.go`:
```go
package prompt

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestBindFlags_Defaults(t *testing.T) {
	cmd := &cobra.Command{Use: "new"}
	var c Config
	BindFlags(cmd, &c)

	if err := cmd.ParseFlags(nil); err != nil {
		t.Fatal(err)
	}
	if c.HTTP != HTTPChi {
		t.Errorf("default HTTP = %q, want %q", c.HTTP, HTTPChi)
	}
	if c.Async != AsyncNone {
		t.Errorf("default Async = %q, want %q", c.Async, AsyncNone)
	}
	if !c.Sentry {
		t.Errorf("default Sentry = false, want true")
	}
}

func TestBindFlags_AllSet(t *testing.T) {
	cmd := &cobra.Command{Use: "new"}
	var c Config
	BindFlags(cmd, &c)

	err := cmd.ParseFlags([]string{
		"--module", "github.com/you/myapp",
		"--http", "stdlib",
		"--async", "river",
		"--no-sentry",
		"--output", "/tmp/myapp",
		"--no-tui",
		"--no-git",
	})
	if err != nil {
		t.Fatal(err)
	}
	if c.Module != "github.com/you/myapp" {
		t.Errorf("Module = %q", c.Module)
	}
	if c.HTTP != HTTPStdlib {
		t.Errorf("HTTP = %q", c.HTTP)
	}
	if c.Async != AsyncRiver {
		t.Errorf("Async = %q", c.Async)
	}
	if c.Sentry {
		t.Errorf("Sentry = true, want false")
	}
	if c.Output != "/tmp/myapp" {
		t.Errorf("Output = %q", c.Output)
	}
	if !c.NoTUI {
		t.Errorf("NoTUI = false")
	}
	if !c.NoGit {
		t.Errorf("NoGit = false")
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `go test ./internal/prompt/...`
Expected: FAIL — `undefined: BindFlags`.

- [ ] **Step 4: Implement `BindFlags`**

Create `internal/prompt/flags.go`:
```go
package prompt

import "github.com/spf13/cobra"

// BindFlags registers all gocraft `new` flags on cmd, writing into c.
// Defaults: --http=chi, --async=none, --sentry=true.
func BindFlags(cmd *cobra.Command, c *Config) {
	cmd.Flags().StringVar(&c.Module, "module", "", "Go module path (e.g. github.com/you/myapp)")
	cmd.Flags().StringVar(&c.HTTP, "http", HTTPChi, "HTTP layer: chi|stdlib")
	cmd.Flags().StringVar(&c.Async, "async", AsyncNone, "async backend: none|river|pool")
	cmd.Flags().BoolVar(&c.Sentry, "sentry", true, "include Sentry observability")
	cmd.Flags().StringVar(&c.Output, "output", "", "output directory (default ./<name>)")
	cmd.Flags().BoolVar(&c.NoTUI, "no-tui", false, "fail on missing options instead of prompting")
	cmd.Flags().BoolVar(&c.NoGit, "no-git", false, "skip git init")
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `go test ./internal/prompt/...`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/prompt/flags.go internal/prompt/flags_test.go go.mod go.sum
git commit -m "feat(prompt): bind cobra flags onto Config"
```

---

### Task 4: huh TUI form

**Files:**
- Create: `internal/tui/form.go`

- [ ] **Step 1: Add huh dependency**

Run:
```bash
go get github.com/charmbracelet/huh@latest
```

- [ ] **Step 2: Implement form**

huh's interactive form is hard to drive in a unit test (reads from a real TTY). We instead expose a single function `Run(c *prompt.Config) error` that fills only empty fields on `c`. M1 verifies this manually; M2+ may add a fake-tty harness.

Create `internal/tui/form.go`:
```go
package tui

import (
	"github.com/charmbracelet/huh"

	"github.com/siyuqian/gocraft/internal/prompt"
)

// Run interactively fills any empty fields on c.
// Fields already set via flags are skipped.
func Run(c *prompt.Config) error {
	var groups []*huh.Group

	if c.Name == "" {
		groups = append(groups, huh.NewGroup(
			huh.NewInput().Title("Project name").Value(&c.Name).
				Validate(func(s string) error {
					tmp := *c
					tmp.Name = s
					tmp.Module = "x"
					tmp.HTTP = prompt.HTTPChi
					tmp.Async = prompt.AsyncNone
					tmp.Output = "x"
					if err := tmp.Validate(); err != nil {
						return err
					}
					return nil
				}),
		))
	}

	if c.Module == "" {
		groups = append(groups, huh.NewGroup(
			huh.NewInput().Title("Module path (e.g. github.com/you/myapp)").Value(&c.Module),
		))
	}

	groups = append(groups, huh.NewGroup(
		huh.NewSelect[string]().Title("HTTP layer").
			Options(huh.NewOption("chi", prompt.HTTPChi), huh.NewOption("stdlib net/http", prompt.HTTPStdlib)).
			Value(&c.HTTP),
		huh.NewSelect[string]().Title("Async backend").
			Options(
				huh.NewOption("none", prompt.AsyncNone),
				huh.NewOption("river (Postgres)", prompt.AsyncRiver),
				huh.NewOption("goroutine pool", prompt.AsyncPool),
			).
			Value(&c.Async),
		huh.NewConfirm().Title("Include Sentry?").Value(&c.Sentry),
	))

	if c.Output == "" {
		groups = append(groups, huh.NewGroup(
			huh.NewInput().Title("Output directory").Placeholder("./" + c.Name).Value(&c.Output),
		))
	}

	return huh.NewForm(groups...).Run()
}
```

- [ ] **Step 3: Verify it compiles**

Run: `go build ./...`
Expected: build succeeds (no test yet — TUI is exercised end-to-end in Task 6).

- [ ] **Step 4: Commit**

```bash
git add internal/tui/form.go go.mod go.sum
git commit -m "feat(tui): add huh form filling empty Config fields"
```

---

### Task 5: cobra root and `version` subcommand

**Files:**
- Create: `internal/cli/root.go`
- Create: `cmd/gocraft/main.go`

- [ ] **Step 1: Implement root command and entry**

Create `internal/cli/root.go`:
```go
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

const Version = "0.1.0-dev"

func NewRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "gocraft",
		Short:         "Opinionated Go project scaffolder",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newVersionCmd())
	return root
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), Version)
			return nil
		},
	}
}

func Execute() error {
	return NewRoot().Execute()
}
```

Create `cmd/gocraft/main.go`:
```go
package main

import (
	"fmt"
	"os"

	"github.com/siyuqian/gocraft/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 2: Smoke test**

Run:
```bash
go run ./cmd/gocraft version
```
Expected output: `0.1.0-dev`

- [ ] **Step 3: Commit**

```bash
git add internal/cli/root.go cmd/gocraft/main.go
git commit -m "feat(cli): add cobra root and version subcommand"
```

---

### Task 6: `new` subcommand wiring

**Files:**
- Create: `internal/cli/new.go`
- Test: `internal/cli/new_test.go`
- Modify: `internal/cli/root.go` (register `new`)

- [ ] **Step 1: Write the failing test**

Create `internal/cli/new_test.go`:
```go
package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func runCmd(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	root := NewRoot()
	var out, errb bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errb)
	root.SetArgs(args)
	err := root.Execute()
	return out.String(), errb.String(), err
}

func TestNew_AllFlagsResolved(t *testing.T) {
	out, _, err := runCmd(t,
		"new", "myapp",
		"--module", "github.com/you/myapp",
		"--http", "chi",
		"--async", "river",
		"--sentry",
		"--output", "/tmp/myapp",
		"--no-tui",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, `"Name": "myapp"`) {
		t.Errorf("output missing name: %s", out)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("output not valid JSON: %v\n%s", err, out)
	}
	if got["Module"] != "github.com/you/myapp" {
		t.Errorf("Module = %v", got["Module"])
	}
	if got["Async"] != "river" {
		t.Errorf("Async = %v", got["Async"])
	}
}

func TestNew_NoTUI_MissingModule(t *testing.T) {
	_, _, err := runCmd(t, "new", "myapp", "--no-tui")
	if err == nil {
		t.Fatal("expected error when --no-tui and --module missing")
	}
	if !strings.Contains(err.Error(), "module") {
		t.Errorf("error should mention module: %v", err)
	}
}

func TestNew_OutputDefaultsFromName(t *testing.T) {
	out, _, err := runCmd(t,
		"new", "myapp",
		"--module", "m",
		"--no-tui",
	)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	var got map[string]any
	_ = json.Unmarshal([]byte(out), &got)
	if got["Output"] != "./myapp" {
		t.Errorf("Output = %v, want ./myapp", got["Output"])
	}
}

func TestNew_BadHTTP(t *testing.T) {
	_, _, err := runCmd(t,
		"new", "myapp",
		"--module", "m",
		"--http", "echo",
		"--no-tui",
	)
	if err == nil {
		t.Fatal("expected validation error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/cli/...`
Expected: FAIL — `new` command not registered.

- [ ] **Step 3: Implement `new` command**

Create `internal/cli/new.go`:
```go
package cli

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/siyuqian/gocraft/internal/prompt"
	"github.com/siyuqian/gocraft/internal/tui"
)

func newNewCmd() *cobra.Command {
	var cfg prompt.Config

	cmd := &cobra.Command{
		Use:   "new [project-name]",
		Short: "Scaffold a new Go project",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				cfg.Name = args[0]
			}

			if !cfg.NoTUI {
				if err := tui.Run(&cfg); err != nil {
					return err
				}
			}

			if cfg.Output == "" && cfg.Name != "" {
				cfg.Output = "./" + cfg.Name
			}

			if err := cfg.Validate(); err != nil {
				return err
			}

			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(cfg)
		},
	}

	prompt.BindFlags(cmd, &cfg)
	return cmd
}
```

- [ ] **Step 4: Register `new` on root**

Edit `internal/cli/root.go`. Change the `NewRoot` body to add `new`:
```go
func NewRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "gocraft",
		Short:         "Opinionated Go project scaffolder",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newVersionCmd())
	root.AddCommand(newNewCmd())
	return root
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `go test ./internal/cli/...`
Expected: PASS — all 4 subtests.

- [ ] **Step 6: Run full test suite**

Run: `go test ./...`
Expected: PASS.

- [ ] **Step 7: Manual TUI smoke test**

Run (in a real terminal):
```bash
go run ./cmd/gocraft new
```
Expected: huh form prompts for Project name, Module, HTTP, Async, Sentry, Output. On submission, prints JSON of the resolved Config.

- [ ] **Step 8: Commit**

```bash
git add internal/cli/new.go internal/cli/new_test.go internal/cli/root.go
git commit -m "feat(cli): add 'new' subcommand resolving Config via flags + TUI"
```

---

### Task 7: README and milestone wrap-up

**Files:**
- Create: `README.md`

- [ ] **Step 1: Write minimal README**

Create `README.md`:
```markdown
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
```

- [ ] **Step 2: Run final checks**

Run:
```bash
go vet ./...
go test ./...
go build ./cmd/gocraft
```
Expected: all pass; binary at `./gocraft` (delete or `make build` for `bin/gocraft`).

- [ ] **Step 3: Commit and tag M1**

```bash
git add README.md
git commit -m "docs: add README for M1"
git tag m1-cli-skeleton
```

---

## M1 Exit Criteria

- `go test ./...` passes with all tests from Tasks 2, 3, 6.
- `go run ./cmd/gocraft version` prints `0.1.0-dev`.
- `go run ./cmd/gocraft new myapp --module github.com/you/myapp --no-tui` prints valid JSON Config.
- `go run ./cmd/gocraft new` launches the huh TUI and produces a valid Config.
- `go vet ./...` clean.

## What this milestone deliberately does NOT do

- No file generation. The resolved `Config` is only printed.
- No template engine, no `embed.FS`.
- No post-processing (`goimports`, `go mod tidy`, `git init`).
- No automated test of the TUI (huh requires a TTY; manual smoke only).

M2 will add the template renderer and consume the same `Config` struct unchanged.
