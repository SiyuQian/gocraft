package cli

import (
	"bytes"
	"os"
	"path/filepath"
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
	outDir := t.TempDir()
	out, _, err := runCmd(t,
		"new", "myapp",
		"--module", "github.com/you/myapp",
		"--http", "chi",
		"--async", "none",
		"--no-sentry",
		"--output", outDir,
		"--no-tui",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "created") {
		t.Errorf("expected success message, got: %s", out)
	}
	assertFile(t, outDir, "go.mod")
	assertFile(t, outDir, "cmd/myapp/main.go")
	assertFile(t, outDir, "internal/platform/httpserver/server.go")
	assertFile(t, outDir, "internal/health/handler.go")
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

func assertFile(t *testing.T, dir, rel string) {
	t.Helper()
	path := filepath.Join(dir, filepath.FromSlash(rel))
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected file %s: %v", rel, err)
	}
}
