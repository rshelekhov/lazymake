package tui

import (
	"reflect"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rshelekhov/lazymake/internal/history"
)

func TestParseEnvLine(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    map[string]string
		wantErr bool
	}{
		{name: "empty", in: "", want: nil},
		{name: "single", in: "FOO=bar", want: map[string]string{"FOO": "bar"}},
		{name: "multiple", in: "A=1 B=2 C=3", want: map[string]string{"A": "1", "B": "2", "C": "3"}},
		{name: "empty value", in: "FOO=", want: map[string]string{"FOO": ""}},
		{name: "quoted value with spaces", in: `MSG="hello world"`, want: map[string]string{"MSG": "hello world"}},
		{name: "leading whitespace", in: "  FOO=bar  ", want: map[string]string{"FOO": "bar"}},
		{name: "missing equals", in: "FOO", wantErr: true},
		{name: "missing key", in: "=bar", wantErr: true},
		{name: "key starts with digit", in: "1FOO=bar", wantErr: true},
		{name: "key with hyphen", in: "FOO-BAR=baz", wantErr: true},
		{name: "unterminated quote", in: `FOO="bar`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseEnvLine(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil (result: %v)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseFlagsLine(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    []string
		wantErr bool
	}{
		{name: "empty", in: "", want: nil},
		{name: "single short", in: "-j4", want: []string{"-j4"}},
		{name: "multiple", in: "-j4 -k --always-make", want: []string{"-j4", "-k", "--always-make"}},
		{name: "looks like variable", in: "FOO=bar", wantErr: true},
		{name: "no leading dash", in: "j4", wantErr: true},
		{name: "mixed valid and invalid", in: "-k j4", wantErr: true},
		{name: "unterminated quote", in: `-k "foo`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFlagsLine(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil (result: %v)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShellSplit(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    []string
		wantErr bool
	}{
		{name: "empty", in: "", want: nil},
		{name: "single", in: "foo", want: []string{"foo"}},
		{name: "multiple", in: "foo bar baz", want: []string{"foo", "bar", "baz"}},
		{name: "tabs", in: "foo\tbar", want: []string{"foo", "bar"}},
		{name: "quoted", in: `foo "bar baz"`, want: []string{"foo", "bar baz"}},
		{name: "embedded quote", in: `K="v 1"`, want: []string{"K=v 1"}},
		{name: "unterminated", in: `"foo`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := shellSplit(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

// makeParamsTestModel builds a minimal Model suitable for exercising the
// params form FSM. It avoids touching the disk by giving History an
// in-memory map and leaving Exporter/ShellIntegration nil.
func makeParamsTestModel(t *testing.T) Model {
	t.Helper()
	hist := &history.History{Entries: make(map[string][]history.Entry)}
	l := list.New(nil, NewItemDelegate(), 0, 0)
	return Model{
		List:            l,
		History:         hist,
		MakefilePath:    "/tmp/test-Makefile",
		StreamingOutput: &strings.Builder{},
		LastRunParams:   make(map[string]ExecutionParams),
	}
}

// TestParamsCancel verifies Esc returns to list with no history append.
// This covers the AC: "Cancel path returns to the list without execution."
func TestParamsCancel(t *testing.T) {
	m := makeParamsTestModel(t)
	target := Target{Name: "build"}
	m.initParamsForm(&target)
	m.State = StateRunParams

	updated, _ := m.updateRunParams(tea.KeyMsg{Type: tea.KeyEsc})
	mm := updated.(Model)

	if mm.State != StateList {
		t.Errorf("expected state=StateList after Esc, got %v", mm.State)
	}
	if mm.ParamsTarget != nil {
		t.Error("expected ParamsTarget to be cleared after Esc")
	}
	if got := len(mm.History.Entries[mm.MakefilePath]); got != 0 {
		t.Errorf("expected 0 history entries after cancel, got %d", got)
	}
}

// TestParamsSubmitValid verifies a valid submission transitions to
// StateExecuting and populates CurrentRunOpts with parsed values.
func TestParamsSubmitValid(t *testing.T) {
	m := makeParamsTestModel(t)
	target := Target{Name: "build"}
	m.initParamsForm(&target)
	m.State = StateRunParams
	m.ParamsEnvInput.SetValue("FOO=bar")
	m.ParamsFlagsInput.SetValue("-j4")

	updated, _ := m.updateRunParams(tea.KeyMsg{Type: tea.KeyEnter})
	mm := updated.(Model)

	if mm.State != StateExecuting {
		t.Errorf("expected state=StateExecuting after valid submit, got %v", mm.State)
	}
	if mm.CurrentRunOpts.Env["FOO"] != "bar" {
		t.Errorf("expected CurrentRunOpts.Env[FOO]=bar, got %v", mm.CurrentRunOpts.Env)
	}
	if len(mm.CurrentRunOpts.Flags) != 1 || mm.CurrentRunOpts.Flags[0] != "-j4" {
		t.Errorf("expected CurrentRunOpts.Flags=[-j4], got %v", mm.CurrentRunOpts.Flags)
	}
	cached, ok := mm.LastRunParams["build"]
	if !ok {
		t.Fatal("expected LastRunParams to be populated for build")
	}
	if cached.EnvRaw != "FOO=bar" || cached.FlagsRaw != "-j4" {
		t.Errorf("LastRunParams raw mismatch: %+v", cached)
	}
}

// TestParamsSubmitInvalidStaysInForm verifies a validation error keeps
// the user in StateRunParams with ParamsError set, and does NOT append
// to history.
func TestParamsSubmitInvalidStaysInForm(t *testing.T) {
	m := makeParamsTestModel(t)
	target := Target{Name: "build"}
	m.initParamsForm(&target)
	m.State = StateRunParams
	m.ParamsEnvInput.SetValue("not-a-valid-pair") // missing '='

	updated, _ := m.updateRunParams(tea.KeyMsg{Type: tea.KeyEnter})
	mm := updated.(Model)

	if mm.State != StateRunParams {
		t.Errorf("expected to remain in StateRunParams on invalid input, got %v", mm.State)
	}
	if mm.ParamsError == "" {
		t.Error("expected ParamsError to be set on invalid input")
	}
	if got := len(mm.History.Entries[mm.MakefilePath]); got != 0 {
		t.Errorf("expected 0 history entries on validation failure, got %d", got)
	}
}

// TestParamsTabTogglesFocus verifies Tab swaps focus between inputs.
func TestParamsTabTogglesFocus(t *testing.T) {
	m := makeParamsTestModel(t)
	target := Target{Name: "build"}
	m.initParamsForm(&target)
	m.State = StateRunParams

	if m.ParamsFocus != paramsFieldEnv {
		t.Fatalf("expected initial focus on env, got %d", m.ParamsFocus)
	}

	updated, _ := m.updateRunParams(tea.KeyMsg{Type: tea.KeyTab})
	mm := updated.(Model)
	if mm.ParamsFocus != paramsFieldFlags {
		t.Errorf("after Tab expected focus=flags, got %d", mm.ParamsFocus)
	}

	updated, _ = mm.updateRunParams(tea.KeyMsg{Type: tea.KeyTab})
	mm = updated.(Model)
	if mm.ParamsFocus != paramsFieldEnv {
		t.Errorf("after second Tab expected focus=env, got %d", mm.ParamsFocus)
	}
}

func TestFormatEnvRaw(t *testing.T) {
	tests := []struct {
		name string
		in   map[string]string
		want string
	}{
		{name: "empty", in: nil, want: ""},
		{name: "zero len", in: map[string]string{}, want: ""},
		{name: "single", in: map[string]string{"FOO": "bar"}, want: "FOO=bar"},
		{
			name: "sorted by key",
			in:   map[string]string{"B": "2", "A": "1", "C": "3"},
			want: "A=1 B=2 C=3",
		},
		{
			name: "value with space gets quoted",
			in:   map[string]string{"MSG": "hello world"},
			want: `MSG="hello world"`,
		},
		{
			name: "value with embedded quote",
			in:   map[string]string{"X": `a"b`},
			// The current parser doesn't support quote escapes, so we
			// strip them rather than emit something parseEnvLine would
			// reject when the user re-opens the form.
			want: `X="ab"`,
		},
		{
			name: "empty value preserved",
			in:   map[string]string{"E": ""},
			want: "E=",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatEnvRaw(tt.in)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatEnvRaw_RoundTripsThroughParseEnvLine(t *testing.T) {
	in := map[string]string{
		"FOO":   "bar",
		"MSG":   "hello world",
		"EMPTY": "",
	}
	got, err := parseEnvLine(formatEnvRaw(in))
	if err != nil {
		t.Fatalf("re-parse failed: %v", err)
	}
	if !reflect.DeepEqual(got, in) {
		t.Errorf("round-trip mismatch:\n  got  %v\n  want %v", got, in)
	}
}

func TestFormatFlagsRaw(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want string
	}{
		{name: "empty", in: nil, want: ""},
		{name: "zero len", in: []string{}, want: ""},
		{name: "single", in: []string{"-j4"}, want: "-j4"},
		{name: "multiple", in: []string{"-j4", "-k", "--always-make"}, want: "-j4 -k --always-make"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatFlagsRaw(tt.in)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// TestParamsPrefillsFromLastRunParams verifies opening the form for a
// target that was previously run with params restores those values.
func TestParamsPrefillsFromLastRunParams(t *testing.T) {
	m := makeParamsTestModel(t)
	m.LastRunParams["deploy"] = ExecutionParams{
		EnvRaw:   `ENV=prod TOKEN="ab cd"`,
		FlagsRaw: "-j2",
	}
	target := Target{Name: "deploy"}
	m.initParamsForm(&target)

	if got := m.ParamsEnvInput.Value(); got != `ENV=prod TOKEN="ab cd"` {
		t.Errorf("env not prefilled: got %q", got)
	}
	if got := m.ParamsFlagsInput.Value(); got != "-j2" {
		t.Errorf("flags not prefilled: got %q", got)
	}
}
