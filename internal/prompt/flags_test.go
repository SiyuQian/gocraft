package prompt

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestBindFlags_Defaults(t *testing.T) {
	cmd := &cobra.Command{Use: "new"}
	var c Config
	BindFlags(cmd, &c)

	if err := cmd.ParseFlags(nil); err != nil {
		t.Fatal(err)
	}
	if c.HTTP != HTTPChi {
		t.Errorf("default HTTP = %q, want %q", c.HTTP, HTTPChi)
	}
	if c.Async != AsyncNone {
		t.Errorf("default Async = %q, want %q", c.Async, AsyncNone)
	}
	if !c.Sentry {
		t.Errorf("default Sentry = false, want true")
	}
}

func TestBindFlags_AllSet(t *testing.T) {
	cmd := &cobra.Command{Use: "new"}
	var c Config
	BindFlags(cmd, &c)

	err := cmd.ParseFlags([]string{
		"--module", "github.com/you/myapp",
		"--http", "stdlib",
		"--async", "river",
		"--no-sentry",
		"--output", "/tmp/myapp",
		"--no-tui",
		"--no-git",
	})
	if err != nil {
		t.Fatal(err)
	}
	if c.Module != "github.com/you/myapp" {
		t.Errorf("Module = %q", c.Module)
	}
	if c.HTTP != HTTPStdlib {
		t.Errorf("HTTP = %q", c.HTTP)
	}
	if c.Async != AsyncRiver {
		t.Errorf("Async = %q", c.Async)
	}
	if c.Sentry {
		t.Errorf("Sentry = true, want false")
	}
	if c.Output != "/tmp/myapp" {
		t.Errorf("Output = %q", c.Output)
	}
	if !c.NoTUI {
		t.Errorf("NoTUI = false")
	}
	if !c.NoGit {
		t.Errorf("NoGit = false")
	}
}
