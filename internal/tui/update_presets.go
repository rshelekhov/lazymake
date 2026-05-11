package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rshelekhov/lazymake/internal/executor"
	"github.com/rshelekhov/lazymake/internal/presets"
	"github.com/rshelekhov/lazymake/internal/safety"
)

// presetNameMaxLen caps preset names so the picker columns stay
// predictable. 64 is generous for "prod-deploy-eu-west-2" style
// names without letting users paste a paragraph.
const presetNameMaxLen = 64

// handleOpenPresetsPicker transitions to the presets picker for the
// currently selected target. Bound to the "p" key in the list view.
//
// The picker shows only presets whose target still exists in the
// current Makefile; presets for renamed/removed targets stay in the
// JSON file so they're not lost if the target comes back, but they
// don't clutter the UI.
func (m Model) handleOpenPresetsPicker() (tea.Model, tea.Cmd) {
	selected := m.List.SelectedItem()
	target, ok := selected.(Target)
	if !ok {
		return m, nil
	}
	return m.openPresetsPicker(&target, StateList), nil
}

// openPresetsPicker is the shared entry point used both by the list-
// view "p" hotkey and by the params-form Ctrl+P. It pins the target
// and the return state, then snapshots the current list of presets.
func (m Model) openPresetsPicker(target *Target, returnTo AppState) Model {
	m.State = StateRunPresets
	m.PresetsTarget = target
	m.PresetsReturnTo = returnTo
	m.PresetsSubMode = presetsSubModeList
	m.PresetsPendingDelete = ""
	m.PresetsError = ""
	m.PresetsStatus = ""
	m.PresetsCursor = 0

	if m.Presets != nil {
		m.PresetsItems = m.Presets.List(m.MakefilePath, target.Name)
	} else {
		m.PresetsItems = nil
	}
	return m
}

// updateRunPresets handles key input while the presets picker is
// active. Sub-modes (list vs. delete confirm) are dispatched first
// so the main switch stays focused on the common case.
func (m Model) updateRunPresets(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		switch m.PresetsSubMode {
		case presetsSubModeDeleteConfirm:
			return m.updatePresetsDeleteConfirm(keyMsg)
		default:
			return m.updatePresetsList(keyMsg)
		}
	}
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		m.Width = size.Width
		m.Height = size.Height
	}
	return m, nil
}

// updatePresetsList handles navigation and actions on the main list
// of presets for a target.
func (m Model) updatePresetsList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return m.cancelPresetsPicker(), nil
	case "q":
		return m, tea.Quit
	case "j", "down":
		if m.PresetsCursor < len(m.PresetsItems)-1 {
			m.PresetsCursor++
		}
		return m, nil
	case "k", "up":
		if m.PresetsCursor > 0 {
			m.PresetsCursor--
		}
		return m, nil
	case "g":
		m.PresetsCursor = 0
		return m, nil
	case "G":
		if len(m.PresetsItems) > 0 {
			m.PresetsCursor = len(m.PresetsItems) - 1
		}
		return m, nil
	case "d":
		if len(m.PresetsItems) == 0 {
			return m, nil
		}
		m.PresetsPendingDelete = m.PresetsItems[m.PresetsCursor].Name
		m.PresetsSubMode = presetsSubModeDeleteConfirm
		return m, nil
	case "e":
		// "Edit" loads the preset into the params form so the user can
		// tweak values and re-save. This is the natural workflow when
		// you want to derive "prod-eu" from "prod".
		return m.loadPresetIntoForm()
	case "enter":
		return m.handlePresetSelect()
	}
	return m, nil
}

// updatePresetsDeleteConfirm handles the inline "y/n" confirmation
// shown when the user presses 'd' on a preset.
func (m Model) updatePresetsDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "enter":
		if m.Presets != nil && m.PresetsTarget != nil {
			name := m.PresetsPendingDelete
			if m.Presets.Delete(m.MakefilePath, m.PresetsTarget.Name, name) {
				_ = m.Presets.Save()
				m.PresetsStatus = fmt.Sprintf("Deleted preset %q", name)
				// Refresh the snapshot so the deleted item vanishes
				// immediately, and keep the cursor inside the new bounds.
				m.PresetsItems = m.Presets.List(m.MakefilePath, m.PresetsTarget.Name)
				if m.PresetsCursor >= len(m.PresetsItems) && m.PresetsCursor > 0 {
					m.PresetsCursor = len(m.PresetsItems) - 1
				}
			}
		}
		m.PresetsSubMode = presetsSubModeList
		m.PresetsPendingDelete = ""
		return m, nil
	case "n", "N", "esc":
		m.PresetsSubMode = presetsSubModeList
		m.PresetsPendingDelete = ""
		return m, nil
	}
	return m, nil
}

// cancelPresetsPicker routes Esc back to the state that opened the
// picker (list or params form), clearing transient picker state.
func (m Model) cancelPresetsPicker() Model {
	returnTo := m.PresetsReturnTo
	if returnTo == 0 && m.State == StateRunPresets {
		returnTo = StateList
	}
	m.State = returnTo
	m.PresetsTarget = nil
	m.PresetsPendingDelete = ""
	m.PresetsSubMode = presetsSubModeList
	m.PresetsError = ""
	m.PresetsStatus = ""
	return m
}

// handlePresetSelect handles Enter on a preset row. The destination
// depends on where the picker was opened from:
//   - StateList → start execution with the preset's env/flags (and
//     mark it as last-used so 'R' picks it up later).
//   - StateRunParams → populate the form inputs and return to it so
//     the user can tweak before running.
func (m Model) handlePresetSelect() (tea.Model, tea.Cmd) {
	if len(m.PresetsItems) == 0 || m.PresetsTarget == nil {
		return m, nil
	}
	p := m.PresetsItems[m.PresetsCursor]

	if m.PresetsReturnTo == StateRunParams {
		return m.applyPresetToParamsForm(p), nil
	}
	return m.runWithPreset(*m.PresetsTarget, p)
}

// loadPresetIntoForm opens the params form pre-filled with the
// currently selected preset's values. Used by the "e" key on a preset
// row. The form is opened in "return to picker" mode? No — opening
// the params form replaces the picker on the state stack, which is
// the more useful behavior for an edit-then-save workflow.
func (m Model) loadPresetIntoForm() (tea.Model, tea.Cmd) {
	if len(m.PresetsItems) == 0 || m.PresetsTarget == nil {
		return m, nil
	}
	p := m.PresetsItems[m.PresetsCursor]
	target := *m.PresetsTarget

	m.initParamsForm(&target)
	// Override the LastRunParams prefill with the preset's values.
	m.ParamsEnvInput.SetValue(formatEnvRaw(p.Env))
	m.ParamsFlagsInput.SetValue(formatFlagsRaw(p.Flags))
	// Carry the name into PendingPresetName so a subsequent Ctrl+S
	// can default to the same name (and trigger an overwrite prompt).
	m.PendingPresetName = p.Name
	m.State = StateRunParams
	return m, textinput.Blink
}

// applyPresetToParamsForm fills the params form inputs with the
// preset's values, then returns to the form so the user can edit.
// Called when the picker was opened from inside the params form.
func (m Model) applyPresetToParamsForm(p presets.Preset) Model {
	m.ParamsEnvInput.SetValue(formatEnvRaw(p.Env))
	m.ParamsFlagsInput.SetValue(formatFlagsRaw(p.Flags))
	m.PendingPresetName = p.Name
	m.ParamsError = ""
	m.State = StateRunParams
	m.PresetsTarget = nil
	m.PresetsPendingDelete = ""
	m.PresetsSubMode = presetsSubModeList
	return m
}

// runWithPreset records the preset as last-used and dispatches to
// startExecution, taking the same dangerous-confirmation detour as
// the quick-run path so safety rules apply uniformly.
func (m Model) runWithPreset(target Target, p presets.Preset) (tea.Model, tea.Cmd) {
	if m.Presets != nil {
		m.Presets.MarkUsed(m.MakefilePath, target.Name, p.Name)
		_ = m.Presets.Save()
	}

	// Mirror submitParamsForm: cache as the in-session last run so a
	// subsequent 'e' opens the form pre-filled with the same values.
	parsed := ExecutionParams{
		EnvRaw:   formatEnvRaw(p.Env),
		FlagsRaw: formatFlagsRaw(p.Flags),
		Env:      p.Env,
		Flags:    p.Flags,
	}
	if m.LastRunParams == nil {
		m.LastRunParams = make(map[string]ExecutionParams)
	}
	m.LastRunParams[target.Name] = parsed

	m.CurrentRunOpts = executor.ExecutionOptions{
		DryRun: m.DryRun,
		Env:    p.Env,
		Flags:  p.Flags,
	}

	if target.IsDangerous && target.DangerLevel == safety.SeverityCritical {
		targetCopy := target
		m.PendingTarget = &targetCopy
		m.State = StateConfirmDangerous
		m.PresetsTarget = nil
		m.PresetsSubMode = presetsSubModeList
		return m, nil
	}
	return m.startExecution(target, m.CurrentRunOpts)
}

// handleRerunLastUsed runs the last-used preset for the currently
// selected target, if any. Bound to capital 'R' in the list view.
// On an empty store / no last-used / missing target the call is a
// no-op so it never disrupts the list.
func (m Model) handleRerunLastUsed() (tea.Model, tea.Cmd) {
	if m.Presets == nil {
		return m, nil
	}
	selected := m.List.SelectedItem()
	target, ok := selected.(Target)
	if !ok {
		return m, nil
	}
	p, ok := m.Presets.LastUsed(m.MakefilePath, target.Name)
	if !ok {
		return m, nil
	}
	return m.runWithPreset(target, p)
}

// initPresetNameInput primes the textinput for the save-as-preset
// wizard with the previous PendingPresetName when available, so
// re-saving an edited preset defaults to the same name.
func (m *Model) initPresetNameInput() {
	inputWidth := 60
	if m.Width > 14 {
		inputWidth = m.Width - 14
	}
	ti := textinput.New()
	ti.Placeholder = "preset name"
	ti.Prompt = "NAME  > "
	ti.CharLimit = presetNameMaxLen
	ti.Width = inputWidth
	if m.PendingPresetName != "" {
		ti.SetValue(m.PendingPresetName)
	}
	ti.Focus()
	m.PresetNameInput = ti
}

// handleParamsSavePreset transitions the params form into the
// save-name sub-mode after validating the current env/flags. We
// validate first so the user can't save a syntactically broken
// preset that they'd never be able to use.
func (m Model) handleParamsSavePreset() (tea.Model, tea.Cmd) {
	if m.ParamsTarget == nil {
		return m, nil
	}
	// Validate the form contents — same checks submitParamsForm runs.
	if _, err := parseEnvLine(m.ParamsEnvInput.Value()); err != nil {
		m.ParamsError = err.Error()
		return m, nil
	}
	if _, err := parseFlagsLine(m.ParamsFlagsInput.Value()); err != nil {
		m.ParamsError = err.Error()
		return m, nil
	}
	m.ParamsError = ""
	m.initPresetNameInput()
	m.ParamsSubMode = paramsSubModeSaveName
	return m, textinput.Blink
}

// updateParamsSaveName drives the inline "type a preset name" prompt.
// Esc returns to the form, Enter triggers save (or overwrite-confirm
// when the name is taken).
func (m Model) updateParamsSaveName(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.Type {
		case tea.KeyEsc:
			m.ParamsSubMode = paramsSubModeForm
			m.ParamsError = ""
			return m, nil
		case tea.KeyEnter:
			return m.commitPresetName()
		}
		if keyMsg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.PresetNameInput, cmd = m.PresetNameInput.Update(msg)
	return m, cmd
}

// commitPresetName validates the typed name and either saves the
// preset immediately or shifts into the overwrite-confirmation
// sub-mode when the name already exists for this target.
func (m Model) commitPresetName() (tea.Model, tea.Cmd) {
	name := strings.TrimSpace(m.PresetNameInput.Value())
	if name == "" {
		m.ParamsError = "preset name cannot be empty"
		return m, nil
	}
	if m.Presets == nil {
		m.ParamsError = "presets store unavailable"
		return m, nil
	}
	if m.ParamsTarget == nil {
		// Defensive: shouldn't happen because we entered SaveName from
		// a valid form.
		m.ParamsSubMode = paramsSubModeForm
		return m, nil
	}
	target := m.ParamsTarget.Name
	if m.Presets.Exists(m.MakefilePath, target, name) && name != m.PendingPresetName {
		// Existing preset with a different original — confirm.
		m.PendingPresetName = name
		m.ParamsSubMode = paramsSubModeOverwriteConfirm
		m.ParamsError = ""
		return m, nil
	}
	return m.savePreset(name)
}

// updateParamsOverwriteConfirm handles the inline "y/n" prompt that
// fires when the user typed an already-taken preset name.
func (m Model) updateParamsOverwriteConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch keyMsg.String() {
	case "y", "Y", "enter":
		return m.savePreset(m.PendingPresetName)
	case "n", "N", "esc":
		// Back to the name prompt so the user can pick a different name.
		m.ParamsSubMode = paramsSubModeSaveName
		m.ParamsError = ""
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

// savePreset persists the form's current env/flags under name, then
// returns the user to the form with a brief status message echoed
// inline.
func (m Model) savePreset(name string) (tea.Model, tea.Cmd) {
	if m.Presets == nil || m.ParamsTarget == nil {
		m.ParamsSubMode = paramsSubModeForm
		return m, nil
	}
	env, err := parseEnvLine(m.ParamsEnvInput.Value())
	if err != nil {
		m.ParamsError = err.Error()
		m.ParamsSubMode = paramsSubModeForm
		return m, nil
	}
	flags, err := parseFlagsLine(m.ParamsFlagsInput.Value())
	if err != nil {
		m.ParamsError = err.Error()
		m.ParamsSubMode = paramsSubModeForm
		return m, nil
	}

	m.Presets.Upsert(m.MakefilePath, m.ParamsTarget.Name, presets.Preset{
		Name:  name,
		Env:   env,
		Flags: flags,
	})
	if err := m.Presets.Save(); err != nil {
		m.ParamsError = "failed to save preset: " + err.Error()
		m.ParamsSubMode = paramsSubModeForm
		return m, nil
	}

	m.PendingPresetName = name
	m.ParamsSubMode = paramsSubModeForm
	m.ParamsError = "" // clear; success status is shown via a separate field
	m.PresetsStatus = fmt.Sprintf("Saved preset %q", name)
	return m, nil
}
