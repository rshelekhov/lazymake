package export

import (
	"strings"
	"testing"
	"time"

	"github.com/rshelekhov/lazymake/internal/executor"
)

// TestNewExecutionRecordWithParams asserts that params (env + flags)
// supplied by the interactive form (issue #37) are stored on the
// record and not silently dropped.
func TestNewExecutionRecordWithParams(t *testing.T) {
	res := executor.Result{
		Duration:  100 * time.Millisecond,
		StartTime: time.Now(),
		EndTime:   time.Now(),
		ExitCode:  0,
	}
	env := map[string]string{"FOO": "bar"}
	flags := []string{"-j4"}

	rec := NewExecutionRecordWithParams("/Makefile", "build", res, env, flags)
	if rec.Env["FOO"] != "bar" {
		t.Errorf("env: got %v, want FOO=bar", rec.Env)
	}
	if len(rec.Flags) != 1 || rec.Flags[0] != "-j4" {
		t.Errorf("flags: got %v, want [-j4]", rec.Flags)
	}
}

// TestNewExecutionRecordDefaultsEmpty makes sure the convenience
// NewExecutionRecord constructor keeps env/flags nil. JSON omitempty
// then prevents the legacy export shape from changing.
func TestNewExecutionRecordDefaultsEmpty(t *testing.T) {
	res := executor.Result{
		Duration:  100 * time.Millisecond,
		StartTime: time.Now(),
		EndTime:   time.Now(),
	}
	rec := NewExecutionRecord("/Makefile", "build", res)
	if rec.Env != nil {
		t.Errorf("expected nil env from legacy constructor, got %v", rec.Env)
	}
	if rec.Flags != nil {
		t.Errorf("expected nil flags from legacy constructor, got %v", rec.Flags)
	}
}

// TestFormatLogIncludesParams verifies the human-readable log contains
// the Env and Flags lines when present, sorted deterministically.
func TestFormatLogIncludesParams(t *testing.T) {
	rec := &ExecutionRecord{
		TargetName: "deploy",
		Timestamp:  time.Now(),
		Output:     "ok\n",
		Env:        map[string]string{"B": "2", "A": "1", "MSG": "hello world"},
		Flags:      []string{"-j4", "-k"},
	}
	got := rec.FormatLog()
	if !strings.Contains(got, "Env:") {
		t.Error("expected Env: line in log")
	}
	if !strings.Contains(got, "Flags:") {
		t.Error("expected Flags: line in log")
	}
	// Sorted: A first, then B, then MSG.
	if !strings.Contains(got, `A=1 B=2 MSG="hello world"`) {
		t.Errorf("env not formatted as expected (sorted, quoted-with-spaces):\n%s", got)
	}
	if !strings.Contains(got, "-j4 -k") {
		t.Errorf("flags not formatted as expected:\n%s", got)
	}
}

// TestFormatLogOmitsParamsWhenEmpty confirms the Env/Flags lines do not
// appear for plain runs.
func TestFormatLogOmitsParamsWhenEmpty(t *testing.T) {
	rec := &ExecutionRecord{
		TargetName: "build",
		Timestamp:  time.Now(),
		Output:     "ok\n",
	}
	got := rec.FormatLog()
	if strings.Contains(got, "Env:") {
		t.Error("Env: should not appear for runs without env")
	}
	if strings.Contains(got, "Flags:") {
		t.Error("Flags: should not appear for runs without flags")
	}
}
