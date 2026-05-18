package tui

import (
	"github.com/charmbracelet/huh"

	"github.com/siyuqian/gocraft/internal/prompt"
)

// Run interactively fills any empty fields on c.
// Fields already set via flags are skipped.
func Run(c *prompt.Config) error {
	var groups []*huh.Group

	if c.Name == "" {
		groups = append(groups, huh.NewGroup(
			huh.NewInput().Title("Project name").Value(&c.Name).
				Validate(func(s string) error {
					tmp := *c
					tmp.Name = s
					tmp.Module = "x"
					tmp.HTTP = prompt.HTTPChi
					tmp.Async = prompt.AsyncNone
					tmp.Output = "x"
					if err := tmp.Validate(); err != nil {
						return err
					}
					return nil
				}),
		))
	}

	if c.Module == "" {
		groups = append(groups, huh.NewGroup(
			huh.NewInput().Title("Module path (e.g. github.com/you/myapp)").Value(&c.Module),
		))
	}

	groups = append(groups, huh.NewGroup(
		huh.NewSelect[string]().Title("HTTP layer").
			Options(huh.NewOption("chi", prompt.HTTPChi), huh.NewOption("stdlib net/http", prompt.HTTPStdlib)).
			Value(&c.HTTP),
		huh.NewSelect[string]().Title("Async backend").
			Options(
				huh.NewOption("none", prompt.AsyncNone),
				huh.NewOption("river (Postgres)", prompt.AsyncRiver),
				huh.NewOption("goroutine pool", prompt.AsyncPool),
			).
			Value(&c.Async),
		huh.NewConfirm().Title("Include Sentry?").Value(&c.Sentry),
	))

	if c.Output == "" {
		groups = append(groups, huh.NewGroup(
			huh.NewInput().Title("Output directory").Placeholder("./" + c.Name).Value(&c.Output),
		))
	}

	return huh.NewForm(groups...).Run()
}
