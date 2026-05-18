package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

const Version = "0.1.0-dev"

func NewRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "gocraft",
		Short:         "Opinionated Go project scaffolder",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newVersionCmd())
	root.AddCommand(newNewCmd())
	return root
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), Version)
			return nil
		},
	}
}

func Execute() error {
	return NewRoot().Execute()
}
