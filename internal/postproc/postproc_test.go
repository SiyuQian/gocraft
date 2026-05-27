package postproc

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGitInit_CreatesRepo(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	if err := GitInit(context.Background(), dir, &stderr); err != nil {
		t.Fatalf("GitInit: %v\nstderr: %s", err, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
		t.Fatalf(".git missing: %v", err)
	}
	// Initial branch must be `main` regardless of the host's init.defaultBranch.
	head, err := os.ReadFile(filepath.Join(dir, ".git", "HEAD"))
	if err != nil {
		t.Fatalf("read HEAD: %v", err)
	}
	if !strings.Contains(string(head), "refs/heads/main") {
		t.Errorf("HEAD = %q, want refs/heads/main", strings.TrimSpace(string(head)))
	}
}

func TestTidy_ErrorsOnNoModule(t *testing.T) {
	// Running `go mod tidy` in a directory with no go.mod fails. We assert
	// the error surfaces rather than silently passing.
	dir := t.TempDir()
	var stderr bytes.Buffer
	if err := Tidy(context.Background(), dir, &stderr); err == nil {
		t.Fatal("expected error from go mod tidy in empty dir")
	}
}

func TestTidy_HappyPath(t *testing.T) {
	dir := t.TempDir()
	goMod := "module example.test\n\ngo 1.21\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	if err := Tidy(context.Background(), dir, &stderr); err != nil {
		t.Fatalf("Tidy: %v\nstderr: %s", err, stderr.String())
	}
}

func TestGoimports_MissingTool_FallsBackToGofmt(t *testing.T) {
	// Empty PATH directory hides `goimports` (and most other tools), so the
	// implementation should fall back to `gofmt`. We re-add the toolchain's
	// own bin directory so `gofmt` itself remains reachable.
	t.Setenv("PATH", gofmtBin(t))

	dir := t.TempDir()
	// Intentionally unformatted source — extra blank lines and bad indentation.
	src := "package x\n\n\n\nfunc  F( ){}\n"
	srcPath := filepath.Join(dir, "x.go")
	if err := os.WriteFile(srcPath, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	var stderr bytes.Buffer
	if err := Goimports(context.Background(), dir, &stderr); err != nil {
		t.Fatalf("Goimports: %v\nstderr: %s", err, stderr.String())
	}
	if !strings.Contains(stderr.String(), "falling back to gofmt") {
		t.Errorf("expected fallback note in stderr, got: %q", stderr.String())
	}
	out, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) == src {
		t.Errorf("gofmt did not rewrite file:\n%s", string(out))
	}
}

func TestGitInit_NoGit_Skips(t *testing.T) {
	// Empty PATH ⇒ git not found ⇒ best-effort skip with a note on stderr.
	t.Setenv("PATH", t.TempDir())

	dir := t.TempDir()
	var stderr bytes.Buffer
	if err := GitInit(context.Background(), dir, &stderr); err != nil {
		t.Fatalf("GitInit should skip silently, got: %v", err)
	}
	if !strings.Contains(stderr.String(), "git not found on PATH") {
		t.Errorf("expected skip note in stderr, got: %q", stderr.String())
	}
	if _, err := os.Stat(filepath.Join(dir, ".git")); !os.IsNotExist(err) {
		t.Errorf(".git should not exist when git is missing: err=%v", err)
	}
}

// gofmtBin returns the directory of the `gofmt` on the host PATH, so a test
// can build a minimal PATH that resolves `gofmt` but not `goimports`.
func gofmtBin(t *testing.T) string {
	t.Helper()
	p, err := exec.LookPath("gofmt")
	if err != nil {
		t.Skipf("gofmt not on PATH: %v", err)
	}
	return filepath.Dir(p)
}
