package cli

import (
	"context"
	"fmt"
	"io"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/siyuqian/gocraft/internal/generate"
	"github.com/siyuqian/gocraft/internal/prompt"
	"github.com/siyuqian/gocraft/internal/tui"
)

// runTidy fetches and tidies module dependencies in dir. Overridable in tests.
var runTidy = func(ctx context.Context, dir string, stdout, stderr io.Writer) error {
	c := exec.CommandContext(ctx, "go", "mod", "tidy")
	c.Dir = dir
	c.Stdout = stdout
	c.Stderr = stderr
	return c.Run()
}

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
				skip := tui.SkipMask{
					HTTP:   cmd.Flags().Changed("http"),
					Async:  cmd.Flags().Changed("async"),
					Sentry: cmd.Flags().Changed("sentry") || cmd.Flags().Changed("no-sentry"),
				}
				if err := tui.Run(&cfg, skip); err != nil {
					return err
				}
			}

			if cfg.Output == "" && cfg.Name != "" {
				cfg.Output = "./" + cfg.Name
			}

			if err := cfg.Validate(); err != nil {
				return err
			}

			fsys := generate.EmbeddedFS()
			if err := generate.Render(cfg, fsys, generate.Layers(cfg), cfg.Output); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "created %s\n", cfg.Output)

			fmt.Fprintln(cmd.OutOrStdout(), "running go mod tidy...")
			if err := runTidy(cmd.Context(), cfg.Output, cmd.OutOrStdout(), cmd.ErrOrStderr()); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "warning: go mod tidy failed: %v\nrun it manually in %s\n", err, cfg.Output)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "next: cd %s && make run\n", cfg.Output)
			return nil
		},
	}

	prompt.BindFlags(cmd, &cfg)
	return cmd
}
