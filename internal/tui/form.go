package tui

import (
	"github.com/charmbracelet/huh"

	"github.com/siyuqian/gocraft/internal/prompt"
)

// SkipMask marks fields whose values were explicitly supplied on the command
// line and should therefore not be re-prompted by the form. Fields with
// non-empty zero defaults (HTTP, Async, Sentry) need an explicit signal
// because their value alone cannot distinguish "user picked the default" from
// "user did not touch this flag."
type SkipMask struct {
	HTTP   bool
	Async  bool
	Sentry bool
}

// Run interactively fills fields on c that were not supplied via flags.
// Name/Module/Output are prompted only when empty. HTTP/Async/Sentry are
// prompted only when their corresponding SkipMask bit is false.
func Run(c *prompt.Config, skip SkipMask) error {
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
					return tmp.Validate()
				}),
		))
	}

	if c.Module == "" {
		groups = append(groups, huh.NewGroup(
			huh.NewInput().Title("Module path (e.g. github.com/you/myapp)").Value(&c.Module),
		))
	}

	var stackFields []huh.Field
	if !skip.HTTP {
		stackFields = append(stackFields,
			huh.NewSelect[string]().Title("HTTP layer").
				Options(huh.NewOption("chi", prompt.HTTPChi), huh.NewOption("stdlib net/http", prompt.HTTPStdlib)).
				Value(&c.HTTP),
		)
	}
	if !skip.Async {
		stackFields = append(stackFields,
			huh.NewSelect[string]().Title("Async backend").
				Options(
					huh.NewOption("none", prompt.AsyncNone),
					huh.NewOption("river (Postgres)", prompt.AsyncRiver),
					huh.NewOption("goroutine pool", prompt.AsyncPool),
				).
				Value(&c.Async),
		)
	}
	if !skip.Sentry {
		stackFields = append(stackFields,
			huh.NewConfirm().Title("Include Sentry?").Value(&c.Sentry),
		)
	}
	if len(stackFields) > 0 {
		groups = append(groups, huh.NewGroup(stackFields...))
	}

	if c.Output == "" {
		in := huh.NewInput().Title("Output directory").Value(&c.Output)
		// Placeholder is evaluated at construction time; only show one when
		// Name is already known (via flag or earlier group). Otherwise the
		// post-TUI defaulting in cli/new.go fills Output from Name.
		if c.Name != "" {
			in = in.Placeholder("./" + c.Name)
		}
		groups = append(groups, huh.NewGroup(in))
	}

	if len(groups) == 0 {
		return nil
	}
	return huh.NewForm(groups...).Run()
}
