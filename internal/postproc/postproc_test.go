package postproc

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestGitInit_CreatesRepo(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	if err := GitInit(dir, &stderr); err != nil {
		t.Fatalf("GitInit: %v\nstderr: %s", err, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
		t.Fatalf(".git missing: %v", err)
	}
}

func TestTidy_ErrorsOnNoModule(t *testing.T) {
	// Running `go mod tidy` in a directory with no go.mod fails. We assert
	// the error surfaces rather than silently passing.
	dir := t.TempDir()
	var stderr bytes.Buffer
	if err := Tidy(dir, &stderr); err == nil {
		t.Fatal("expected error from go mod tidy in empty dir")
	}
}
