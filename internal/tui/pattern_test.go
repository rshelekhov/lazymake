package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rshelekhov/lazymake/internal/safety"
)

func TestResolvePatternName(t *testing.T) {
	tests := []struct {
		pattern, stem, want string
	}{
		{"build-%", "foo", "build-foo"},
		{"%.o", "main", "main.o"},
		{"bin/%-tui", "app", "bin/app-tui"},
		{"%-%", "a", "a-%"}, // make only substitutes the first %
	}
	for _, tt := range tests {
		if got := resolvePatternName(tt.pattern, tt.stem); got != tt.want {
			t.Errorf("resolvePatternName(%q, %q) = %q, want %q", tt.pattern, tt.stem, got, tt.want)
		}
	}
}

func TestValidateStem(t *testing.T) {
	tests := []struct {
		stem    string
		wantErr bool
	}{
		{"foo", false},
		{"foo/bar", false},
		{"", true},
		{"foo bar", true},
		{"foo\tbar", true},
		{"fo%o", true},
	}
	for _, tt := range tests {
		err := validateStem(tt.stem)
		if (err != nil) != tt.wantErr {
			t.Errorf("validateStem(%q) error = %v, wantErr %v", tt.stem, err, tt.wantErr)
		}
	}
}

// makePatternTestModel returns a list-view model whose selected item is
// the given target.
func makePatternTestModel(t *testing.T, target Target) Model {
	t.Helper()
	m := makeParamsTestModel(t)
	m.List = newListWithTargets([]Target{target})
	m.State = StateList
	return m
}

func typeStem(t *testing.T, m Model, stem string) Model {
	t.Helper()
	m.StemInput.SetValue(stem)
	updated, _ := m.updatePatternStem(tea.KeyMsg{Type: tea.KeyEnter})
	return updated.(Model)
}

func TestEnterOnPatternTargetOpensStemPrompt(t *testing.T) {
	m := makePatternTestModel(t, Target{Name: "build-%", IsPatternRule: true})

	updated, _ := m.handleTargetSelection()
	mm := updated.(Model)

	if mm.State != StatePatternStem {
		t.Fatalf("expected StatePatternStem, got %v", mm.State)
	}
	if mm.StemTarget == nil || mm.StemTarget.Name != "build-%" {
		t.Errorf("expected StemTarget build-%%, got %+v", mm.StemTarget)
	}
	if got := len(mm.History.Entries[mm.MakefilePath]); got != 0 {
		t.Errorf("opening the prompt must not record history, got %d entries", got)
	}
}

func TestEnterOnRegularTargetRunsDirectly(t *testing.T) {
	m := makePatternTestModel(t, Target{Name: "build"})

	updated, _ := m.handleTargetSelection()
	mm := updated.(Model)

	if mm.State != StateExecuting || mm.ExecutingTarget != "build" {
		t.Errorf("expected to execute build, got state=%v target=%q", mm.State, mm.ExecutingTarget)
	}
}

func TestStemPromptSubmitRunsConcreteTarget(t *testing.T) {
	m := makePatternTestModel(t, Target{Name: "build-%", IsPatternRule: true})
	m = m.openStemPrompt(Target{Name: "build-%", IsPatternRule: true}, stemActionRun)

	mm := typeStem(t, m, "  foo ")

	if mm.State != StateExecuting {
		t.Fatalf("expected StateExecuting, got %v", mm.State)
	}
	if mm.ExecutingTarget != "build-foo" {
		t.Errorf("expected ExecutingTarget build-foo, got %q", mm.ExecutingTarget)
	}
	if mm.LastStems["build-%"] != "foo" {
		t.Errorf("expected stem to be remembered, got %v", mm.LastStems)
	}
	entries := mm.History.Entries[mm.MakefilePath]
	if len(entries) != 1 || entries[0].Name != "build-foo" {
		t.Errorf("expected history to record build-foo, got %+v", entries)
	}
}

func TestStemPromptInvalidStemStaysOpen(t *testing.T) {
	m := makePatternTestModel(t, Target{Name: "build-%", IsPatternRule: true})
	m = m.openStemPrompt(Target{Name: "build-%", IsPatternRule: true}, stemActionRun)

	mm := typeStem(t, m, "")

	if mm.State != StatePatternStem {
		t.Errorf("expected to stay in StatePatternStem, got %v", mm.State)
	}
	if mm.StemError == "" {
		t.Error("expected StemError to be set")
	}
}

func TestStemPromptEscCancels(t *testing.T) {
	m := makePatternTestModel(t, Target{Name: "build-%", IsPatternRule: true})
	m = m.openStemPrompt(Target{Name: "build-%", IsPatternRule: true}, stemActionRun)

	updated, _ := m.updatePatternStem(tea.KeyMsg{Type: tea.KeyEsc})
	mm := updated.(Model)

	if mm.State != StateList || mm.StemTarget != nil {
		t.Errorf("expected list with cleared StemTarget, got state=%v target=%+v", mm.State, mm.StemTarget)
	}
	if got := len(mm.History.Entries[mm.MakefilePath]); got != 0 {
		t.Errorf("cancel must not record history, got %d entries", got)
	}
}

func TestStemPromptPrefillsLastStem(t *testing.T) {
	m := makePatternTestModel(t, Target{Name: "build-%", IsPatternRule: true})
	m.LastStems = map[string]string{"build-%": "bar"}

	m = m.openStemPrompt(Target{Name: "build-%", IsPatternRule: true}, stemActionRun)

	if got := m.StemInput.Value(); got != "bar" {
		t.Errorf("expected prefilled stem bar, got %q", got)
	}
}

func TestStemPromptThenParamsForm(t *testing.T) {
	target := Target{Name: "build-%", IsPatternRule: true}
	m := makePatternTestModel(t, target)

	updated, _ := m.handleOpenParamsForm()
	m = updated.(Model)
	if m.State != StatePatternStem || m.StemNext != stemActionParams {
		t.Fatalf("expected stem prompt for params, got state=%v next=%v", m.State, m.StemNext)
	}

	mm := typeStem(t, m, "foo")

	if mm.State != StateRunParams {
		t.Fatalf("expected StateRunParams, got %v", mm.State)
	}
	if mm.ParamsTarget == nil || mm.ParamsTarget.Name != "build-foo" || mm.ParamsTarget.IsPatternRule {
		t.Errorf("expected concrete params target build-foo, got %+v", mm.ParamsTarget)
	}
}

func TestStemPromptDangerousTargetNeedsConfirmation(t *testing.T) {
	target := Target{
		Name:          "deploy-%",
		IsPatternRule: true,
		IsDangerous:   true,
		DangerLevel:   safety.SeverityCritical,
	}
	m := makePatternTestModel(t, target)
	m = m.openStemPrompt(target, stemActionRun)

	mm := typeStem(t, m, "prod")

	if mm.State != StateConfirmDangerous {
		t.Fatalf("expected StateConfirmDangerous, got %v", mm.State)
	}
	if mm.PendingTarget == nil || mm.PendingTarget.Name != "deploy-prod" {
		t.Errorf("expected pending target deploy-prod, got %+v", mm.PendingTarget)
	}
}

func TestPatternMarkerNotInFilterValue(t *testing.T) {
	target := Target{Name: "build-%", Description: "Build one module", IsPatternRule: true}
	if strings.Contains(target.FilterValue(), PatternMarker) {
		t.Errorf("FilterValue should not contain %q: %q", PatternMarker, target.FilterValue())
	}

	title := NewItemDelegate().buildTitle(target, "", TextPrimary, TextPrimary)
	if !strings.Contains(title, PatternMarker) {
		t.Errorf("expected title to contain %q, got %q", PatternMarker, title)
	}
}

func TestRenderPatternStemViewShowsPreview(t *testing.T) {
	m := makePatternTestModel(t, Target{Name: "build-%", IsPatternRule: true})
	m.Width, m.Height = 100, 30
	m = m.openStemPrompt(Target{Name: "build-%", Description: "Build one module", IsPatternRule: true}, stemActionRun)
	m.StemInput.SetValue("foo")

	out := m.renderPatternStemView()
	for _, want := range []string{"Target: build-%", "Build one module", "Runs: make build-foo", "enter: run"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected view to contain %q, got:\n%s", want, out)
		}
	}
}
