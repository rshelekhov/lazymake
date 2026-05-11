package tui

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rshelekhov/lazymake/internal/history"
	"github.com/rshelekhov/lazymake/internal/presets"
)

// makePresetsTestModel mirrors makeParamsTestModel but additionally
// wires a real presets.Manager pointed at t.TempDir(). The List is
// seeded with a single "build" target so handleOpenPresetsPicker and
// handleRerunLastUsed both have something to operate on.
func makePresetsTestModel(t *testing.T) Model {
	t.Helper()

	presetsPath := filepath.Join(t.TempDir(), "presets.json")
	pmgr, err := presets.LoadFrom(presetsPath)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}

	target := Target{Name: "build"}
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
		Presets:         pmgr,
	}
}

// TestPresetsPickerOpensFromList: pressing 'p' (via handleOpenPresetsPicker)
// flips state to StateRunPresets and snapshots the list of presets.
func TestPresetsPickerOpensFromList(t *testing.T) {
	m := makePresetsTestModel(t)
	m.Presets.Upsert(m.MakefilePath, "build", presets.Preset{
		Name:  "fast",
		Flags: []string{"-j4"},
	})

	updated, _ := m.handleOpenPresetsPicker()
	mm := updated.(Model)

	if mm.State != StateRunPresets {
		t.Errorf("state=%v, want StateRunPresets", mm.State)
	}
	if mm.PresetsTarget == nil || mm.PresetsTarget.Name != "build" {
		t.Errorf("expected PresetsTarget=build, got %+v", mm.PresetsTarget)
	}
	if mm.PresetsReturnTo != StateList {
		t.Errorf("PresetsReturnTo=%v, want StateList", mm.PresetsReturnTo)
	}
	if len(mm.PresetsItems) != 1 || mm.PresetsItems[0].Name != "fast" {
		t.Errorf("PresetsItems=%+v, want [fast]", mm.PresetsItems)
	}
}

// TestPresetsPickerEnterRunsTarget: Enter on a preset row marks it as
// last-used, populates CurrentRunOpts, and transitions to execution.
func TestPresetsPickerEnterRunsTarget(t *testing.T) {
	m := makePresetsTestModel(t)
	m.Presets.Upsert(m.MakefilePath, "build", presets.Preset{
		Name:  "fast",
		Env:   map[string]string{"FOO": "bar"},
		Flags: []string{"-j4"},
	})
	updated, _ := m.handleOpenPresetsPicker()
	mm := updated.(Model)

	updated, _ = mm.updateRunPresets(tea.KeyMsg{Type: tea.KeyEnter})
	mm = updated.(Model)

	if mm.State != StateExecuting {
		t.Errorf("state=%v, want StateExecuting after Enter on preset", mm.State)
	}
	if mm.CurrentRunOpts.Env["FOO"] != "bar" {
		t.Errorf("CurrentRunOpts.Env mismatch: %v", mm.CurrentRunOpts.Env)
	}
	if len(mm.CurrentRunOpts.Flags) != 1 || mm.CurrentRunOpts.Flags[0] != "-j4" {
		t.Errorf("CurrentRunOpts.Flags mismatch: %v", mm.CurrentRunOpts.Flags)
	}
	if lu, ok := mm.Presets.LastUsed(mm.MakefilePath, "build"); !ok || lu.Name != "fast" {
		t.Errorf("LastUsed=%+v ok=%v, want fast/true", lu, ok)
	}
}

// TestPresetsDeleteWithConfirmation: 'd' enters confirm sub-mode, 'y'
// removes the preset and saves; 'n' cancels.
func TestPresetsDeleteWithConfirmation(t *testing.T) {
	m := makePresetsTestModel(t)
	m.Presets.Upsert(m.MakefilePath, "build", presets.Preset{Name: "old"})
	m.Presets.Upsert(m.MakefilePath, "build", presets.Preset{Name: "keep"})

	updated, _ := m.handleOpenPresetsPicker()
	mm := updated.(Model)

	// Cursor starts at 0; the first preset in the picker is the
	// most-recently-updated one. We don't care which name appears
	// first, but we DO want to assert the delete pipeline removes
	// whatever the cursor points at.
	cursorName := mm.PresetsItems[mm.PresetsCursor].Name

	updated, _ = mm.updateRunPresets(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	mm = updated.(Model)
	if mm.PresetsSubMode != presetsSubModeDeleteConfirm {
		t.Fatalf("expected delete-confirm sub-mode, got %d", mm.PresetsSubMode)
	}
	if mm.PresetsPendingDelete != cursorName {
		t.Errorf("PendingDelete=%q, want %q", mm.PresetsPendingDelete, cursorName)
	}

	// Confirm with 'y'.
	updated, _ = mm.updateRunPresets(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	mm = updated.(Model)
	if mm.PresetsSubMode != presetsSubModeList {
		t.Errorf("expected list sub-mode after confirm, got %d", mm.PresetsSubMode)
	}
	if mm.Presets.Exists(mm.MakefilePath, "build", cursorName) {
		t.Errorf("preset %q still exists after delete-confirm", cursorName)
	}
	if got := mm.Presets.Count(mm.MakefilePath, "build"); got != 1 {
		t.Errorf("Count=%d after delete, want 1", got)
	}
}

// TestPresetsDeleteCanceled: pressing 'n' in confirm leaves the preset
// untouched and returns to the list.
func TestPresetsDeleteCanceled(t *testing.T) {
	m := makePresetsTestModel(t)
	m.Presets.Upsert(m.MakefilePath, "build", presets.Preset{Name: "p"})
	updated, _ := m.handleOpenPresetsPicker()
	mm := updated.(Model)
	updated, _ = mm.updateRunPresets(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	mm = updated.(Model)
	updated, _ = mm.updateRunPresets(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	mm = updated.(Model)

	if mm.PresetsSubMode != presetsSubModeList {
		t.Errorf("expected list sub-mode after cancel, got %d", mm.PresetsSubMode)
	}
	if !mm.Presets.Exists(mm.MakefilePath, "build", "p") {
		t.Error("preset was deleted despite cancel")
	}
}

// TestPresetsSaveFromParamsForm: Ctrl+S → type name → Enter saves the
// preset and returns to the form with a success message.
func TestPresetsSaveFromParamsForm(t *testing.T) {
	m := makePresetsTestModel(t)
	target := Target{Name: "build"}
	m.initParamsForm(&target)
	m.State = StateRunParams
	m.ParamsEnvInput.SetValue("FOO=bar")
	m.ParamsFlagsInput.SetValue("-j4")

	// Ctrl+S is delivered as tea.KeyCtrlS (msg.String() == "ctrl+s").
	updated, _ := m.updateRunParams(tea.KeyMsg{Type: tea.KeyCtrlS})
	mm := updated.(Model)
	if mm.ParamsSubMode != paramsSubModeSaveName {
		t.Fatalf("expected SaveName sub-mode after Ctrl+S, got %d", mm.ParamsSubMode)
	}

	// Type a name into the prompt.
	mm.PresetNameInput.SetValue("fast")

	// Enter to commit.
	updated, _ = mm.updateRunParams(tea.KeyMsg{Type: tea.KeyEnter})
	mm = updated.(Model)
	if mm.ParamsSubMode != paramsSubModeForm {
		t.Errorf("expected back to form sub-mode, got %d", mm.ParamsSubMode)
	}
	if !mm.Presets.Exists(mm.MakefilePath, "build", "fast") {
		t.Error("preset 'fast' was not saved")
	}
	got, _ := mm.Presets.Get(mm.MakefilePath, "build", "fast")
	if got.Env["FOO"] != "bar" {
		t.Errorf("preset Env mismatch: %v", got.Env)
	}
	if len(got.Flags) != 1 || got.Flags[0] != "-j4" {
		t.Errorf("preset Flags mismatch: %v", got.Flags)
	}
}

// TestPresetsSaveOverwritePromptYes: saving with an existing name
// shifts to overwrite confirm; 'y' overwrites.
func TestPresetsSaveOverwritePromptYes(t *testing.T) {
	m := makePresetsTestModel(t)
	m.Presets.Upsert(m.MakefilePath, "build", presets.Preset{
		Name: "fast",
		Env:  map[string]string{"OLD": "1"},
	})
	target := Target{Name: "build"}
	m.initParamsForm(&target)
	m.State = StateRunParams
	m.ParamsEnvInput.SetValue("NEW=2")

	updated, _ := m.updateRunParams(tea.KeyMsg{Type: tea.KeyCtrlS})
	mm := updated.(Model)
	mm.PresetNameInput.SetValue("fast")

	updated, _ = mm.updateRunParams(tea.KeyMsg{Type: tea.KeyEnter})
	mm = updated.(Model)
	if mm.ParamsSubMode != paramsSubModeOverwriteConfirm {
		t.Fatalf("expected OverwriteConfirm sub-mode, got %d", mm.ParamsSubMode)
	}

	updated, _ = mm.updateRunParams(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	mm = updated.(Model)
	if mm.ParamsSubMode != paramsSubModeForm {
		t.Errorf("expected back to form sub-mode after overwrite, got %d", mm.ParamsSubMode)
	}
	got, _ := mm.Presets.Get(mm.MakefilePath, "build", "fast")
	if _, ok := got.Env["OLD"]; ok {
		t.Errorf("old env still present after overwrite: %v", got.Env)
	}
	if got.Env["NEW"] != "2" {
		t.Errorf("new env missing after overwrite: %v", got.Env)
	}
}

// TestPresetsSaveOverwritePromptNo: 'n' in overwrite confirm bounces
// back to the name prompt without writing.
func TestPresetsSaveOverwritePromptNo(t *testing.T) {
	m := makePresetsTestModel(t)
	m.Presets.Upsert(m.MakefilePath, "build", presets.Preset{Name: "fast", Env: map[string]string{"OLD": "1"}})
	target := Target{Name: "build"}
	m.initParamsForm(&target)
	m.State = StateRunParams
	m.ParamsEnvInput.SetValue("NEW=2")
	updated, _ := m.updateRunParams(tea.KeyMsg{Type: tea.KeyCtrlS})
	mm := updated.(Model)
	mm.PresetNameInput.SetValue("fast")
	updated, _ = mm.updateRunParams(tea.KeyMsg{Type: tea.KeyEnter})
	mm = updated.(Model)

	updated, _ = mm.updateRunParams(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	mm = updated.(Model)

	if mm.ParamsSubMode != paramsSubModeSaveName {
		t.Errorf("expected SaveName after 'n', got %d", mm.ParamsSubMode)
	}
	got, _ := mm.Presets.Get(mm.MakefilePath, "build", "fast")
	if got.Env["OLD"] != "1" || got.Env["NEW"] == "2" {
		t.Errorf("preset mutated despite cancel: %v", got.Env)
	}
}

// TestPresetsSaveValidatesFirst: Ctrl+S with malformed env input
// reports the error and does NOT enter the save-name sub-mode.
func TestPresetsSaveValidatesFirst(t *testing.T) {
	m := makePresetsTestModel(t)
	target := Target{Name: "build"}
	m.initParamsForm(&target)
	m.State = StateRunParams
	m.ParamsEnvInput.SetValue("not-an-env-pair")

	updated, _ := m.updateRunParams(tea.KeyMsg{Type: tea.KeyCtrlS})
	mm := updated.(Model)

	if mm.ParamsSubMode != paramsSubModeForm {
		t.Errorf("expected to stay in form on validation error, got %d", mm.ParamsSubMode)
	}
	if mm.ParamsError == "" {
		t.Error("expected ParamsError to be set on invalid env")
	}
}

// TestPresetsLoadFromParamsForm: Ctrl+P opens the picker with
// PresetsReturnTo=StateRunParams; selecting a preset populates the
// form and returns there.
func TestPresetsLoadFromParamsForm(t *testing.T) {
	m := makePresetsTestModel(t)
	m.Presets.Upsert(m.MakefilePath, "build", presets.Preset{
		Name:  "fast",
		Env:   map[string]string{"FOO": "bar"},
		Flags: []string{"-j4"},
	})
	target := Target{Name: "build"}
	m.initParamsForm(&target)
	m.State = StateRunParams

	updated, _ := m.updateRunParams(tea.KeyMsg{Type: tea.KeyCtrlP})
	mm := updated.(Model)
	if mm.State != StateRunPresets {
		t.Fatalf("expected StateRunPresets after Ctrl+P, got %v", mm.State)
	}
	if mm.PresetsReturnTo != StateRunParams {
		t.Errorf("PresetsReturnTo=%v, want StateRunParams", mm.PresetsReturnTo)
	}

	updated, _ = mm.updateRunPresets(tea.KeyMsg{Type: tea.KeyEnter})
	mm = updated.(Model)
	if mm.State != StateRunParams {
		t.Errorf("expected return to StateRunParams, got %v", mm.State)
	}
	if mm.ParamsEnvInput.Value() != "FOO=bar" {
		t.Errorf("env not populated: %q", mm.ParamsEnvInput.Value())
	}
	if mm.ParamsFlagsInput.Value() != "-j4" {
		t.Errorf("flags not populated: %q", mm.ParamsFlagsInput.Value())
	}
}

// TestRerunLastUsed: 'R' picks the last-used preset for the cursor
// target and starts execution.
func TestRerunLastUsed(t *testing.T) {
	m := makePresetsTestModel(t)
	m.Presets.Upsert(m.MakefilePath, "build", presets.Preset{
		Name:  "fast",
		Env:   map[string]string{"E": "1"},
		Flags: []string{"-j2"},
	})
	m.Presets.MarkUsed(m.MakefilePath, "build", "fast")

	updated, _ := m.handleRerunLastUsed()
	mm := updated.(Model)

	if mm.State != StateExecuting {
		t.Errorf("expected StateExecuting after R, got %v", mm.State)
	}
	if mm.CurrentRunOpts.Env["E"] != "1" {
		t.Errorf("CurrentRunOpts not populated from preset: %+v", mm.CurrentRunOpts)
	}
}

// TestRerunLastUsed_NoPresetNoOp: 'R' is a no-op when there's no
// last-used preset — never crashes, never transitions state.
func TestRerunLastUsed_NoPresetNoOp(t *testing.T) {
	m := makePresetsTestModel(t)
	updated, _ := m.handleRerunLastUsed()
	mm := updated.(Model)
	if mm.State != m.State {
		t.Errorf("expected no state change, got %v", mm.State)
	}
}

// TestPresetsPickerEscReturnsToOpenerState: Esc returns to the state
// that originally opened the picker (StateList or StateRunParams).
func TestPresetsPickerEscReturnsToOpenerState(t *testing.T) {
	m := makePresetsTestModel(t)
	updated, _ := m.handleOpenPresetsPicker()
	mm := updated.(Model)
	updated, _ = mm.updateRunPresets(tea.KeyMsg{Type: tea.KeyEsc})
	mm = updated.(Model)
	if mm.State != StateList {
		t.Errorf("expected StateList after Esc, got %v", mm.State)
	}
}

// TestPresetsForMissingTargetNotShownInList: a preset stored against
// a target name not present in m.Targets is hidden but NOT deleted
// (so the user keeps it for when the target comes back). The picker
// is opened from the current list cursor, which only knows about
// targets that exist — verifying the absence path is enough.
func TestPresetsForMissingTargetNotShownInList(t *testing.T) {
	m := makePresetsTestModel(t)
	// Store a preset for "deploy", but Targets only contains "build".
	m.Presets.Upsert(m.MakefilePath, "deploy", presets.Preset{Name: "prod"})

	updated, _ := m.handleOpenPresetsPicker()
	mm := updated.(Model)
	for _, p := range mm.PresetsItems {
		if p.Name == "prod" {
			t.Errorf("preset for missing target was shown: %+v", p)
		}
	}
	// Confirm the underlying store still has it.
	if !mm.Presets.Exists(mm.MakefilePath, "deploy", "prod") {
		t.Error("preset for missing target was deleted; should be kept for future use")
	}
}

// TestPresetsPickerDoesNotCrashWithNilManager: a nil Presets manager
// (Load failed at startup) must not panic when 'p' is pressed.
func TestPresetsPickerDoesNotCrashWithNilManager(t *testing.T) {
	m := makePresetsTestModel(t)
	m.Presets = nil
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("panic: %v", r)
		}
	}()
	updated, _ := m.handleOpenPresetsPicker()
	mm := updated.(Model)
	if mm.State != StateRunPresets {
		t.Errorf("expected picker to open even with nil store, got %v", mm.State)
	}
	if len(mm.PresetsItems) != 0 {
		t.Errorf("expected empty items list, got %d", len(mm.PresetsItems))
	}
}
