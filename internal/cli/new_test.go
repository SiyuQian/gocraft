package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func runCmd(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	root := NewRoot()
	var out, errb bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errb)
	root.SetArgs(args)
	err := root.Execute()
	return out.String(), errb.String(), err
}

func TestNew_AllFlagsResolved(t *testing.T) {
	out, _, err := runCmd(t,
		"new", "myapp",
		"--module", "github.com/you/myapp",
		"--http", "chi",
		"--async", "river",
		"--sentry",
		"--output", "/tmp/myapp",
		"--no-tui",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, `"Name": "myapp"`) {
		t.Errorf("output missing name: %s", out)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("output not valid JSON: %v\n%s", err, out)
	}
	if got["Module"] != "github.com/you/myapp" {
		t.Errorf("Module = %v", got["Module"])
	}
	if got["Async"] != "river" {
		t.Errorf("Async = %v", got["Async"])
	}
}

func TestNew_NoTUI_MissingModule(t *testing.T) {
	_, _, err := runCmd(t, "new", "myapp", "--no-tui")
	if err == nil {
		t.Fatal("expected error when --no-tui and --module missing")
	}
	if !strings.Contains(err.Error(), "module") {
		t.Errorf("error should mention module: %v", err)
	}
}

func TestNew_OutputDefaultsFromName(t *testing.T) {
	out, _, err := runCmd(t,
		"new", "myapp",
		"--module", "m",
		"--no-tui",
	)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	var got map[string]any
	_ = json.Unmarshal([]byte(out), &got)
	if got["Output"] != "./myapp" {
		t.Errorf("Output = %v, want ./myapp", got["Output"])
	}
}

func TestNew_BadHTTP(t *testing.T) {
	_, _, err := runCmd(t,
		"new", "myapp",
		"--module", "m",
		"--http", "echo",
		"--no-tui",
	)
	if err == nil {
		t.Fatal("expected validation error")
	}
}
