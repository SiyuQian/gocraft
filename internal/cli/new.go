package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/siyuqian/gocraft/internal/generate"
	"github.com/siyuqian/gocraft/internal/postproc"
	"github.com/siyuqian/gocraft/internal/prompt"
	"github.com/siyuqian/gocraft/internal/tui"
)

func newNewCmd() *cobra.Command {
	var cfg prompt.Config
	var skipTidy bool

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

			stderr := cmd.ErrOrStderr()
			if !skipTidy {
				if err := postproc.Tidy(cfg.Output, stderr); err != nil {
					return err
				}
				if err := postproc.Goimports(cfg.Output, stderr); err != nil {
					return err
				}
			}
			if !cfg.NoGit {
				if err := postproc.GitInit(cfg.Output, stderr); err != nil {
					return err
				}
			}

			fmt.Fprintf(cmd.OutOrStdout(), "created %s\nnext: cd %s && make run\n", cfg.Output, cfg.Output)
			return nil
		},
	}

	prompt.BindFlags(cmd, &cfg)
	cmd.Flags().BoolVar(&skipTidy, "no-tidy", false, "skip 'go mod tidy' after rendering (faster, but generated project will not build until you run it)")
	return cmd
}
