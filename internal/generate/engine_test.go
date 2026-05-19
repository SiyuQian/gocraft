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
