// Package postproc runs cleanup steps after the template renderer has
// written files: dependency resolution (`go mod tidy`), import formatting
// (`goimports` with a `gofmt` fallback), and an optional initial git commit.
//
// Tidy returns errors verbatim — a failing `go mod tidy` almost always means
// a real template bug (bad import path, missing module requirement) that
// should not be silenced. Goimports and GitInit are best-effort: when their
// preferred tool is missing they log a note to stderr and either fall back
// to a stdlib alternative or skip the step, rather than failing the whole
// scaffold.
package postproc

import (
	"context"
	"fmt"
	"io"
	"os/exec"
)

// Tidy runs `go mod tidy` in dir. Errors are returned because a failing tidy
// usually means a real template bug (bad import path, missing module
// requirement) that should not be silenced.
func Tidy(ctx context.Context, dir string, stderr io.Writer) error {
	cmd := exec.CommandContext(ctx, "go", "mod", "tidy")
	cmd.Dir = dir
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go mod tidy in %s: %w", dir, err)
	}
	return nil
}

// Goimports runs `goimports -w .` to group and format imports. If goimports
// is not on PATH, falls back to `gofmt -w .` (which ships with the Go
// toolchain) so the generated project is at least syntactically formatted.
func Goimports(ctx context.Context, dir string, stderr io.Writer) error {
	tool := "goimports"
	if _, err := exec.LookPath("goimports"); err != nil {
		fmt.Fprintln(stderr, "note: goimports not found on PATH, falling back to gofmt")
		tool = "gofmt"
	}
	cmd := exec.CommandContext(ctx, tool, "-w", ".")
	cmd.Dir = dir
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s in %s: %w", tool, dir, err)
	}
	return nil
}

// GitInit initializes a git repository in dir on branch `main` and creates an
// initial commit of all files. Skipped silently if git is not on PATH.
// Identity is passed inline so the commit succeeds even when the host has no
// global git user configured (fresh containers, CI runners).
func GitInit(ctx context.Context, dir string, stderr io.Writer) error {
	if _, err := exec.LookPath("git"); err != nil {
		fmt.Fprintln(stderr, "note: git not found on PATH, skipping git init")
		return nil
	}
	for _, args := range [][]string{
		{"init", "--quiet", "-b", "main"},
		{"add", "."},
		{
			"-c", "user.email=gocraft@localhost",
			"-c", "user.name=gocraft",
			"commit", "--quiet", "-m", "initial scaffold",
		},
	} {
		cmd := exec.CommandContext(ctx, "git", args...)
		cmd.Dir = dir
		cmd.Stderr = stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git %v in %s: %w", args, dir, err)
		}
	}
	return nil
}
