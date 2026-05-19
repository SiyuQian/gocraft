# gocraft M2 — Template Rendering Engine Implementation Plan

**Goal:** Implement the layered overlay template renderer in `internal/generate/`. The renderer takes a `prompt.Config`, an `fs.FS` rooted at the templates tree, and an output directory; it walks layers in order, renders `.tmpl` files via `text/template`, copies non-`.tmpl` files verbatim, and interpolates `{{.Name}}` segments in destination paths.

**Architecture:** Two small files. `layers.go` derives the ordered layer list from `Config`. `engine.go` walks each layer and writes files. Tests use `fstest.MapFS` so the engine has zero dependency on real templates (those land in M3+).

**Tech Stack:** stdlib `io/fs`, `text/template`, `fstest`, `embed` (declared but contents minimal until M3).

---

## File Structure

| File | Responsibility |
|---|---|
| `internal/generate/layers.go` | `Layers(cfg) []string` — deterministic layer order |
| `internal/generate/layers_test.go` | unit tests for `Layers` |
| `internal/generate/engine.go` | `Render(cfg, fsys, layers, outDir) error` |
| `internal/generate/engine_test.go` | tests with `fstest.MapFS` |
| `templates/.keep` | placeholder so `templates/` directory exists for future M3 |

---

### Task 1: `Layers` function (deterministic layer order)

**Files:**
- Create: `internal/generate/layers.go`
- Test: `internal/generate/layers_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/generate/layers_test.go`:
```go
package generate

import (
	"reflect"
	"testing"

	"github.com/siyuqian/gocraft/internal/prompt"
)

func TestLayers(t *testing.T) {
	cases := []struct {
		name string
		cfg  prompt.Config
		want []string
	}{
		{
			name: "chi none no-sentry",
			cfg:  prompt.Config{HTTP: prompt.HTTPChi, Async: prompt.AsyncNone, Sentry: false},
			want: []string{"base", "http/chi"},
		},
		{
			name: "stdlib none sentry",
			cfg:  prompt.Config{HTTP: prompt.HTTPStdlib, Async: prompt.AsyncNone, Sentry: true},
			want: []string{"base", "http/stdlib", "obs/sentry"},
		},
		{
			name: "chi river sentry",
			cfg:  prompt.Config{HTTP: prompt.HTTPChi, Async: prompt.AsyncRiver, Sentry: true},
			want: []string{"base", "http/chi", "async/river", "obs/sentry"},
		},
		{
			name: "stdlib pool no-sentry",
			cfg:  prompt.Config{HTTP: prompt.HTTPStdlib, Async: prompt.AsyncPool, Sentry: false},
			want: []string{"base", "http/stdlib", "async/pool"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Layers(tc.cfg)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Layers() = %v, want %v", got, tc.want)
			}
		})
	}
}
```

- [ ] **Step 2: Implement `Layers`**

Create `internal/generate/layers.go`:
```go
package generate

import "github.com/siyuqian/gocraft/internal/prompt"

// Layers returns the ordered list of template layer roots for cfg.
// Order is: base, http/<choice>, async/<choice> (if not none), obs/sentry (if enabled).
// Later layers overwrite earlier layers when the same relative path appears.
func Layers(cfg prompt.Config) []string {
	out := []string{"base", "http/" + cfg.HTTP}
	if cfg.Async != prompt.AsyncNone {
		out = append(out, "async/"+cfg.Async)
	}
	if cfg.Sentry {
		out = append(out, "obs/sentry")
	}
	return out
}
```

- [ ] **Step 3: Verify**

Run: `go test ./internal/generate/...`
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add internal/generate/layers.go internal/generate/layers_test.go
git commit -m "feat(generate): add Layers function for ordered template overlay"
```

---

### Task 2: `Render` engine (overlay walker)

**Files:**
- Create: `internal/generate/engine.go`
- Test: `internal/generate/engine_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/generate/engine_test.go`:
```go
package generate

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/siyuqian/gocraft/internal/prompt"
)

func TestRender_BaseLayerOnly(t *testing.T) {
	fsys := fstest.MapFS{
		"base/README.md.tmpl":            {Data: []byte("# {{.Name}}\n")},
		"base/cmd/{{.Name}}/main.go.tmpl": {Data: []byte("package main // {{.Module}}\n")},
		"base/.gitignore":                 {Data: []byte("bin/\n")},
	}
	cfg := prompt.Config{Name: "myapp", Module: "example.com/myapp"}
	out := t.TempDir()

	if err := Render(cfg, fsys, []string{"base"}, out); err != nil {
		t.Fatalf("Render: %v", err)
	}

	mustReadFile(t, filepath.Join(out, "README.md"), "# myapp\n")
	mustReadFile(t, filepath.Join(out, "cmd/myapp/main.go"), "package main // example.com/myapp\n")
	mustReadFile(t, filepath.Join(out, ".gitignore"), "bin/\n")
}

func TestRender_LaterLayerOverwrites(t *testing.T) {
	fsys := fstest.MapFS{
		"base/main.go.tmpl":     {Data: []byte("// base {{.Name}}\n")},
		"http/chi/main.go.tmpl": {Data: []byte("// chi {{.Name}}\n")},
	}
	cfg := prompt.Config{Name: "myapp"}
	out := t.TempDir()

	if err := Render(cfg, fsys, []string{"base", "http/chi"}, out); err != nil {
		t.Fatalf("Render: %v", err)
	}
	mustReadFile(t, filepath.Join(out, "main.go"), "// chi myapp\n")
}

func TestRender_NonTmplCopiedVerbatim(t *testing.T) {
	// .tmpl is stripped; non-.tmpl files keep their name and contents (no template execution).
	raw := "{{ this should not be interpreted }}\n"
	fsys := fstest.MapFS{
		"base/raw.txt": {Data: []byte(raw)},
	}
	cfg := prompt.Config{Name: "x"}
	out := t.TempDir()

	if err := Render(cfg, fsys, []string{"base"}, out); err != nil {
		t.Fatalf("Render: %v", err)
	}
	mustReadFile(t, filepath.Join(out, "raw.txt"), raw)
}

func TestRender_PathInterpolation(t *testing.T) {
	fsys := fstest.MapFS{
		"base/cmd/{{.Name}}/main.go.tmpl": {Data: []byte("x")},
		"base/{{.Name}}.md":                {Data: []byte("y")},
	}
	cfg := prompt.Config{Name: "svc"}
	out := t.TempDir()

	if err := Render(cfg, fsys, []string{"base"}, out); err != nil {
		t.Fatalf("Render: %v", err)
	}
	mustReadFile(t, filepath.Join(out, "cmd/svc/main.go"), "x")
	mustReadFile(t, filepath.Join(out, "svc.md"), "y")
}

func TestRender_TemplateErrorSurfaces(t *testing.T) {
	fsys := fstest.MapFS{
		"base/bad.tmpl": {Data: []byte("{{ .Nope.Field }}")},
	}
	cfg := prompt.Config{Name: "x"}
	out := t.TempDir()
	if err := Render(cfg, fsys, []string{"base"}, out); err == nil {
		t.Fatal("expected template error")
	}
}

func TestRender_MissingLayerIsSkipped(t *testing.T) {
	// A declared layer with no files in fsys should not error.
	fsys := fstest.MapFS{
		"base/x.txt": {Data: []byte("ok")},
	}
	cfg := prompt.Config{Name: "x"}
	out := t.TempDir()
	if err := Render(cfg, fsys, []string{"base", "http/chi"}, out); err != nil {
		t.Fatalf("Render: %v", err)
	}
	mustReadFile(t, filepath.Join(out, "x.txt"), "ok")
}

func mustReadFile(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if string(got) != want {
		t.Errorf("%s = %q, want %q", path, got, want)
	}
}
```

- [ ] **Step 2: Implement engine**

Create `internal/generate/engine.go`:
```go
package generate

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/siyuqian/gocraft/internal/prompt"
)

// Render walks each layer in order under fsys and writes files into outDir.
// Files ending in ".tmpl" have the suffix stripped and are rendered as
// text/template with cfg as the data context. Other files are copied
// byte-for-byte. Path segments may contain template actions (e.g. "{{.Name}}")
// which are expanded against cfg. Later layers overwrite earlier layers when
// they target the same relative output path.
func Render(cfg prompt.Config, fsys fs.FS, layers []string, outDir string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", outDir, err)
	}
	for _, layer := range layers {
		if err := renderLayer(cfg, fsys, layer, outDir); err != nil {
			return fmt.Errorf("layer %q: %w", layer, err)
		}
	}
	return nil
}

func renderLayer(cfg prompt.Config, fsys fs.FS, layer, outDir string) error {
	// A layer with no files (entry missing in fsys) is silently skipped.
	if _, err := fs.Stat(fsys, layer); err != nil {
		return nil
	}
	return fs.WalkDir(fsys, layer, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(layer, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		destRel, err := expandPath(rel, cfg)
		if err != nil {
			return fmt.Errorf("path %q: %w", rel, err)
		}
		destRel = strings.TrimSuffix(destRel, ".tmpl")
		destAbs := filepath.Join(outDir, filepath.FromSlash(destRel))

		if err := os.MkdirAll(filepath.Dir(destAbs), 0o755); err != nil {
			return err
		}

		data, err := fs.ReadFile(fsys, p)
		if err != nil {
			return err
		}
		if strings.HasSuffix(rel, ".tmpl") {
			rendered, err := execTemplate(rel, string(data), cfg)
			if err != nil {
				return fmt.Errorf("render %q: %w", rel, err)
			}
			data = rendered
		}
		return os.WriteFile(destAbs, data, 0o644)
	})
}

func expandPath(p string, cfg prompt.Config) (string, error) {
	if !strings.Contains(p, "{{") {
		return p, nil
	}
	segs := strings.Split(p, "/")
	for i, s := range segs {
		if !strings.Contains(s, "{{") {
			continue
		}
		out, err := execTemplate("path:"+s, s, cfg)
		if err != nil {
			return "", err
		}
		segs[i] = string(out)
	}
	return path.Join(segs...), nil
}

func execTemplate(name, src string, cfg prompt.Config) ([]byte, error) {
	tmpl, err := template.New(name).Option("missingkey=error").Parse(src)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, cfg); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
```

- [ ] **Step 3: Verify**

Run: `go test ./internal/generate/...`
Expected: all 6 subtests pass.

- [ ] **Step 4: Commit**

```bash
git add internal/generate/engine.go internal/generate/engine_test.go
git commit -m "feat(generate): implement layered overlay template renderer"
```

---

### Task 3: Templates directory placeholder

**Files:**
- Create: `templates/.keep`

- [ ] **Step 1: Create placeholder**

```bash
mkdir -p templates && echo "M3+ will populate this directory." > templates/.keep
```

- [ ] **Step 2: Commit**

```bash
git add templates/.keep
git commit -m "chore: add templates/ placeholder for M3+"
```

---

### Task 4: Run full suite and tidy

- [ ] **Step 1: Full checks**

```bash
go vet ./... && go test ./... && go build ./cmd/gocraft
```
Expected: all pass.

---

## M2 Exit Criteria

- `Layers(cfg)` returns deterministic order per spec
- `Render` correctly: renders `.tmpl`, copies non-`.tmpl`, expands path segments, overwrites in layer order, surfaces template errors
- All tests pass; vet clean; binary builds
- M3 can drop real template files into `templates/<layer>/` and `Render` will pick them up unchanged

## Out of scope (deferred to M3+)

- The actual template content (`base/`, `http/chi/`, etc.)
- Wiring `Render` into the `new` command (still prints JSON in M2)
- `embed.FS` directive (declared in M3 once `templates/` has content)
- Post-processing (`goimports`, `go mod tidy`)
