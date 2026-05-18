package prompt

import "github.com/spf13/cobra"

// BindFlags registers all gocraft `new` flags on cmd, writing into c.
// Defaults: --http=chi, --async=none, --sentry=true.
func BindFlags(cmd *cobra.Command, c *Config) {
	cmd.Flags().StringVar(&c.Module, "module", "", "Go module path (e.g. github.com/you/myapp)")
	cmd.Flags().StringVar(&c.HTTP, "http", HTTPChi, "HTTP layer: chi|stdlib")
	cmd.Flags().StringVar(&c.Async, "async", AsyncNone, "async backend: none|river|pool")
	cmd.Flags().BoolVar(&c.Sentry, "sentry", true, "include Sentry observability")
	cmd.Flags().BoolFunc("no-sentry", "disable Sentry observability", func(s string) error {
		c.Sentry = false
		return nil
	})
	cmd.Flags().StringVar(&c.Output, "output", "", "output directory (default ./<name>)")
	cmd.Flags().BoolVar(&c.NoTUI, "no-tui", false, "fail on missing options instead of prompting")
	cmd.Flags().BoolVar(&c.NoGit, "no-git", false, "skip git init")
}
