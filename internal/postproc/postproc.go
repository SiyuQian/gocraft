// Package postproc runs cleanup steps after the template renderer has
// written files: dependency resolution (`go mod tidy`), import grouping
// (`goimports`), and an optional initial git commit. Each step is best-effort
// — when a tool is unavailable, the step logs a warning and returns nil
// rather than failing the whole scaffold.
package postproc

import (
	"fmt"
	"io"
	"os/exec"
)

// Tidy runs `go mod tidy` in dir. Errors are returned because a failing tidy
// usually means a real template bug (bad import path, missing module
// requirement) that should not be silenced.
func Tidy(dir string, stderr io.Writer) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = dir
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go mod tidy in %s: %w", dir, err)
	}
	return nil
}

// Goimports runs `goimports -w .` to normalize imports. If goimports is not
// on PATH, the step is silently skipped: stdlib `gofmt` ran via `go mod
// tidy` already handles syntactic formatting, and goimports is a nicety.
func Goimports(dir string, stderr io.Writer) error {
	if _, err := exec.LookPath("goimports"); err != nil {
		fmt.Fprintln(stderr, "note: goimports not found on PATH, skipping import grouping")
		return nil
	}
	cmd := exec.Command("goimports", "-w", ".")
	cmd.Dir = dir
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("goimports in %s: %w", dir, err)
	}
	return nil
}

// GitInit initializes a git repository in dir and creates an initial commit
// of all files. Skipped silently if git is not on PATH.
func GitInit(dir string, stderr io.Writer) error {
	if _, err := exec.LookPath("git"); err != nil {
		fmt.Fprintln(stderr, "note: git not found on PATH, skipping git init")
		return nil
	}
	for _, args := range [][]string{
		{"init", "--quiet"},
		{"add", "."},
		{"commit", "--quiet", "-m", "initial scaffold"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Stderr = stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git %v in %s: %w", args, dir, err)
		}
	}
	return nil
}
