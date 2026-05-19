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

func TestAcceptance_HTTPLayers(t *testing.T) {
	cases := []struct {
		name string
		http string
	}{
		{"chi", prompt.HTTPChi},
		{"stdlib", prompt.HTTPStdlib},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := prompt.Config{
				Name:   "demoapp",
				Module: "example.com/demoapp",
				HTTP:   tc.http,
				Async:  prompt.AsyncNone,
				Sentry: false,
				Output: t.TempDir(),
			}
			if err := generate.Render(cfg, generate.EmbeddedFS(), generate.Layers(cfg), cfg.Output); err != nil {
				t.Fatalf("Render: %v", err)
			}
			for _, rel := range []string{
				"go.mod",
				".gitignore",
				".golangci.yml",
				".github/workflows/test-and-lint.yml",
			} {
				if _, err := os.Stat(filepath.Join(cfg.Output, rel)); err != nil {
					t.Fatalf("%s missing: %v", rel, err)
				}
			}
			if _, err := os.Stat(filepath.Join(cfg.Output, "migrations", "0001_init.up.sql")); err != nil {
				t.Fatalf("migrations/0001_init.up.sql missing: %v", err)
			}
			mustRun(t, cfg.Output, "go", "mod", "tidy")
			mustRun(t, cfg.Output, "go", "vet", "./...")
			mustRun(t, cfg.Output, "go", "build", "./...")
			mustRun(t, cfg.Output, "go", "test", "./...")
		})
	}
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
