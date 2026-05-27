package prompt

import (
	"fmt"
	"regexp"
)

const (
	HTTPChi    = "chi"
	HTTPStdlib = "stdlib"

	AsyncNone  = "none"
	AsyncRiver = "river"
	AsyncPool  = "pool"
)

var (
	ValidHTTP  = []string{HTTPChi, HTTPStdlib}
	ValidAsync = []string{AsyncNone, AsyncRiver, AsyncPool}

	nameRe = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)
)

type Config struct {
	Name   string
	Module string
	HTTP   string
	Async  string
	Sentry bool
	Output string
	NoTUI  bool
	NoGit  bool
	NoTidy bool
}

func (c Config) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("project name is required")
	}
	if !nameRe.MatchString(c.Name) {
		return fmt.Errorf("project name %q must match %s", c.Name, nameRe)
	}
	if c.Module == "" {
		return fmt.Errorf("module path is required")
	}
	if !contains(ValidHTTP, c.HTTP) {
		return fmt.Errorf("http=%q must be one of %v", c.HTTP, ValidHTTP)
	}
	if !contains(ValidAsync, c.Async) {
		return fmt.Errorf("async=%q must be one of %v", c.Async, ValidAsync)
	}
	if c.Output == "" {
		return fmt.Errorf("output directory is required")
	}
	return nil
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}
