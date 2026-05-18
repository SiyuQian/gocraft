package cli

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/siyuqian/gocraft/internal/prompt"
	"github.com/siyuqian/gocraft/internal/tui"
)

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

			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(cfg)
		},
	}

	prompt.BindFlags(cmd, &cfg)
	return cmd
}
