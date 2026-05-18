package prompt

import "testing"

func TestValidate_OK(t *testing.T) {
	c := Config{
		Name:   "myapp",
		Module: "github.com/you/myapp",
		HTTP:   HTTPChi,
		Async:  AsyncNone,
		Sentry: true,
		Output: "./myapp",
	}
	if err := c.Validate(); err != nil {
		t.Fatalf("expected ok, got %v", err)
	}
}

func TestValidate_MissingName(t *testing.T) {
	c := Config{Module: "github.com/you/myapp", HTTP: HTTPChi, Async: AsyncNone, Output: "./x"}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestValidate_MissingModule(t *testing.T) {
	c := Config{Name: "myapp", HTTP: HTTPChi, Async: AsyncNone, Output: "./x"}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for empty module")
	}
}

func TestValidate_BadHTTP(t *testing.T) {
	c := Config{Name: "myapp", Module: "m", HTTP: "echo", Async: AsyncNone, Output: "./x"}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for unknown http")
	}
}

func TestValidate_BadAsync(t *testing.T) {
	c := Config{Name: "myapp", Module: "m", HTTP: HTTPChi, Async: "celery", Output: "./x"}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for unknown async")
	}
}

func TestValidate_BadName(t *testing.T) {
	c := Config{Name: "My App!", Module: "m", HTTP: HTTPChi, Async: AsyncNone, Output: "./x"}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for non-identifier name")
	}
}
