package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rshelekhov/lazymake/internal/executor"
	"github.com/rshelekhov/lazymake/internal/history"
	"github.com/rshelekhov/lazymake/internal/safety"
)

// makeRerunTestModel builds a Model wired with a History store and a
// single-target list. It mirrors the pattern from makePresetsTestModel
// so the test setup stays familiar.
func makeRerunTestModel(t *testing.T, target Target) Model {
	t.Helper()

	l := list.New([]list.Item{target}, NewItemDelegate(), 0, 0)
	l.Select(0)

	return Model{
		List:            l,
		History:         &history.History{Entries: make(map[string][]history.Entry)},
		MakefilePath:    "/tmp/test-Makefile",
		StreamingOutput: &strings.Builder{},
		LastRunParams:   make(map[string]ExecutionParams),
		Targets:         []Target{target},
		AllTargets:      []Target{target},
		// Width/Height irrelevant for handler logic; viewport rendering
		// isn't exercised by these tests.
	}
}

// TestHandleRerunLast_NoHistory: pressing ctrl+r on a target that has
// never been executed is a silent no-op so the user's cursor stays
// put and nothing surprising happens.
func TestHandleRerunLast_NoHistory(t *testing.T) {
	m := makeRerunTestModel(t, Target{Name: "build"})

	updated, cmd := m.handleRerunLast()
	mm := updated.(Model)

	if mm.State != StateList {
		t.Errorf("state=%v, want StateList (no-op)", mm.State)
	}
	if mm.PendingTarget != nil {
		t.Errorf("PendingTarget=%+v, want nil", mm.PendingTarget)
	}
	if cmd != nil {
		t.Errorf("cmd=%v, want nil", cmd)
	}
}

// TestHandleRerunLast_PopulatesFromHistory: ctrl+r with prior history
// pins the target, captures env/flags from the latest record, and
// switches to StateConfirmRerun.
func TestHandleRerunLast_PopulatesFromHistory(t *testing.T) {
	target := Target{Name: "build"}
	m := makeRerunTestModel(t, target)

	env := map[string]string{"FOO": "bar"}
	flags := []string{"-j4", "-k"}
	m.History.RecordExecutionWithParams(m.MakefilePath, target.Name, 250*time.Millisecond, true, env, flags)

	updated, _ := m.handleRerunLast()
	mm := updated.(Model)

	if mm.State != StateConfirmRerun {
		t.Errorf("state=%v, want StateConfirmRerun", mm.State)
	}
	if mm.PendingTarget == nil || mm.PendingTarget.Name != "build" {
		t.Errorf("PendingTarget=%+v, want pinned build", mm.PendingTarget)
	}
	if mm.CurrentRunOpts.Env["FOO"] != "bar" {
		t.Errorf("CurrentRunOpts.Env=%v, want FOO=bar", mm.CurrentRunOpts.Env)
	}
	if len(mm.CurrentRunOpts.Flags) != 2 {
		t.Errorf("CurrentRunOpts.Flags=%v, want 2 flags", mm.CurrentRunOpts.Flags)
	}
}

// TestHandleRerunLast_HonorsDryRun: the DryRun bit must come from the
// session flag, not from whatever the previous execution recorded —
// ExecutionOptions doesn't even persist DryRun in history, but this
// guards against accidental future regressions where someone wires up
// such persistence and forgets to override here.
func TestHandleRerunLast_HonorsDryRun(t *testing.T) {
	target := Target{Name: "build"}
	m := makeRerunTestModel(t, target)
	m.DryRun = true
	m.History.RecordExecutionWithTiming(m.MakefilePath, target.Name, time.Millisecond, true)

	updated, _ := m.handleRerunLast()
	mm := updated.(Model)

	if !mm.CurrentRunOpts.DryRun {
		t.Error("CurrentRunOpts.DryRun should mirror session DryRun=true")
	}
}

// TestUpdateConfirmRerun_EscReturnsToList: canceling the confirm
// dialog clears PendingTarget and CurrentRunOpts so later code paths
// don't pick up a stale rerun plan.
func TestUpdateConfirmRerun_EscReturnsToList(t *testing.T) {
	target := Target{Name: "build"}
	m := makeRerunTestModel(t, target)
	m.History.RecordExecutionWithParams(m.MakefilePath, target.Name, time.Millisecond, true,
		map[string]string{"FOO": "bar"}, []string{"-j4"})

	updated, _ := m.handleRerunLast()
	mm := updated.(Model)

	updated, _ = mm.updateConfirmRerun(tea.KeyMsg{Type: tea.KeyEsc})
	mm = updated.(Model)

	if mm.State != StateList {
		t.Errorf("state=%v, want StateList after Esc", mm.State)
	}
	if mm.PendingTarget != nil {
		t.Errorf("PendingTarget=%+v, want nil after cancel", mm.PendingTarget)
	}
	if !equalOptions(mm.CurrentRunOpts, executor.ExecutionOptions{}) {
		t.Errorf("CurrentRunOpts=%+v, want zero value after cancel", mm.CurrentRunOpts)
	}
}

// TestUpdateConfirmRerun_EnterStartsExecution: Enter on the confirm
// dialog hands control to startExecution and forwards the snapshotted
// env/flags through to the executor.
func TestUpdateConfirmRerun_EnterStartsExecution(t *testing.T) {
	target := Target{Name: "build"}
	m := makeRerunTestModel(t, target)
	env := map[string]string{"FOO": "bar"}
	flags := []string{"-j4"}
	m.History.RecordExecutionWithParams(m.MakefilePath, target.Name, time.Millisecond, true, env, flags)

	updated, _ := m.handleRerunLast()
	mm := updated.(Model)

	updated, _ = mm.updateConfirmRerun(tea.KeyMsg{Type: tea.KeyEnter})
	mm = updated.(Model)

	if mm.State != StateExecuting {
		t.Errorf("state=%v, want StateExecuting after Enter", mm.State)
	}
	if mm.ExecutingTarget != "build" {
		t.Errorf("ExecutingTarget=%q, want build", mm.ExecutingTarget)
	}
}

// TestHandleRerunLast_DangerousTargetStaysInRerunConfirm: when the
// last-executed target is critical-dangerous, ctrl+r still routes to
// StateConfirmRerun (not StateConfirmDangerous). The combined modal
// will display the safety details inline; users only press Enter once.
func TestHandleRerunLast_DangerousTargetStaysInRerunConfirm(t *testing.T) {
	target := Target{
		Name:        "nuke",
		IsDangerous: true,
		DangerLevel: safety.SeverityCritical,
		SafetyMatches: []safety.MatchResult{{
			Rule:        safety.Rule{ID: "rm-rf", Description: "removes everything"},
			MatchedLine: "rm -rf /",
			Severity:    safety.SeverityCritical,
		}},
	}
	m := makeRerunTestModel(t, target)
	m.History.RecordExecutionWithTiming(m.MakefilePath, target.Name, time.Millisecond, true)

	updated, _ := m.handleRerunLast()
	mm := updated.(Model)

	if mm.State != StateConfirmRerun {
		t.Errorf("state=%v, want StateConfirmRerun (combined modal), got dangerous-only", mm.State)
	}
	if mm.PendingTarget == nil || !mm.PendingTarget.IsDangerous {
		t.Errorf("PendingTarget should carry IsDangerous=true, got %+v", mm.PendingTarget)
	}
}

// TestRenderConfirmRerunView_ContainsPlan smoke-tests that the view
// renders the previous env/flags in the dialog so the user can verify
// what is about to run. Wide enough that wrapping doesn't break the
// substring assertions.
func TestRenderConfirmRerunView_ContainsPlan(t *testing.T) {
	target := Target{Name: "build"}
	m := makeRerunTestModel(t, target)
	m.Width = 120
	m.Height = 40
	m.History.RecordExecutionWithParams(m.MakefilePath, target.Name, time.Millisecond, true,
		map[string]string{"ENV": "prod"}, []string{"-j4"})

	updated, _ := m.handleRerunLast()
	mm := updated.(Model)

	out := mm.renderConfirmRerunView()
	for _, want := range []string{"RERUN LAST EXECUTION", "build", "ENV=prod", "-j4"} {
		if !strings.Contains(out, want) {
			t.Errorf("rendered view missing %q\n---\n%s", want, out)
		}
	}
}

// TestRenderConfirmRerunView_DangerBlock asserts that the combined
// modal surfaces the safety description when the target is critical.
func TestRenderConfirmRerunView_DangerBlock(t *testing.T) {
	target := Target{
		Name:        "nuke",
		IsDangerous: true,
		DangerLevel: safety.SeverityCritical,
		SafetyMatches: []safety.MatchResult{{
			Rule:        safety.Rule{ID: "rm-rf", Description: "removes everything"},
			MatchedLine: "rm -rf /",
			Severity:    safety.SeverityCritical,
		}},
	}
	m := makeRerunTestModel(t, target)
	m.Width = 120
	m.Height = 40
	m.History.RecordExecutionWithTiming(m.MakefilePath, target.Name, time.Millisecond, true)

	updated, _ := m.handleRerunLast()
	mm := updated.(Model)

	out := mm.renderConfirmRerunView()
	for _, want := range []string{"rm-rf", "rm -rf /", "removes everything"} {
		if !strings.Contains(out, want) {
			t.Errorf("rendered view missing danger fragment %q\n---\n%s", want, out)
		}
	}
}

// equalOptions compares ExecutionOptions field-by-field. We avoid
// reflect.DeepEqual here so Env/Flags nil-vs-empty doesn't trip the
// assertion in subtle ways across Go versions.
func equalOptions(a, b executor.ExecutionOptions) bool {
	if a.DryRun != b.DryRun {
		return false
	}
	if len(a.Env) != len(b.Env) {
		return false
	}
	for k, v := range a.Env {
		if b.Env[k] != v {
			return false
		}
	}
	if len(a.Flags) != len(b.Flags) {
		return false
	}
	for i, v := range a.Flags {
		if b.Flags[i] != v {
			return false
		}
	}
	return true
}
