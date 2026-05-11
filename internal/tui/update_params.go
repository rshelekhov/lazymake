package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rshelekhov/lazymake/internal/executor"
	"github.com/rshelekhov/lazymake/internal/safety"
)

// handleOpenParamsForm transitions from the list view to the params
// form for the currently selected target. Bound to the "e" key.
//
// The form is pre-filled from LastRunParams if the user has run this
// target with params before in the current session.
func (m Model) handleOpenParamsForm() (tea.Model, tea.Cmd) {
	selected := m.List.SelectedItem()
	target, ok := selected.(Target)
	if !ok {
		return m, nil
	}

	m.initParamsForm(&target)
	m.State = StateRunParams
	return m, textinput.Blink
}

// initParamsForm lazily creates the two textinputs and seeds their
// values from LastRunParams[target.Name] if present. Safe to call
// repeatedly; reinitializes each time so the form always opens fresh.
func (m *Model) initParamsForm(target *Target) {
	// Container border (2) + padding (4) + prompt width (8) = 14
	// chars of chrome before the input itself. Default to 60 when
	// the model hasn't been sized yet (mostly tests).
	inputWidth := 60
	if m.Width > 14 {
		inputWidth = m.Width - 14
	}

	envInput := textinput.New()
	envInput.Placeholder = "KEY=value KEY2=value"
	envInput.Prompt = "ENV   > "
	envInput.CharLimit = 1024
	envInput.Width = inputWidth

	flagsInput := textinput.New()
	flagsInput.Placeholder = "-j4 -k --always-make"
	flagsInput.Prompt = "FLAGS > "
	flagsInput.CharLimit = 1024
	flagsInput.Width = inputWidth

	// Pre-fill from session cache when available.
	if prev, ok := m.LastRunParams[target.Name]; ok {
		envInput.SetValue(prev.EnvRaw)
		flagsInput.SetValue(prev.FlagsRaw)
	}

	envInput.Focus()

	m.ParamsEnvInput = envInput
	m.ParamsFlagsInput = flagsInput
	m.ParamsFocus = paramsFieldEnv
	m.ParamsTarget = target
	m.ParamsError = ""
}

// updateRunParams handles key input while the params form is active.
//
// Esc cancels with no side effects. Tab toggles focus between the two
// inputs. Enter parses both fields; on success it stores the parsed
// values in LastRunParams and m.CurrentRunOpts, then transitions to
// the executor (or to dangerous confirmation, for critical targets).
func (m Model) updateRunParams(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			return m.cancelParamsForm(), nil
		case tea.KeyTab, tea.KeyShiftTab:
			m = m.toggleParamsFocus()
			return m, nil
		case tea.KeyEnter:
			return m.submitParamsForm()
		}
		// Ctrl+C still quits.
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		// Re-fit textinput widths to the new terminal width without
		// touching the current values or focus state.
		inputWidth := 60
		if m.Width > 14 {
			inputWidth = m.Width - 14
		}
		m.ParamsEnvInput.Width = inputWidth
		m.ParamsFlagsInput.Width = inputWidth
	}

	// Forward all other keys to the focused input.
	var cmd tea.Cmd
	if m.ParamsFocus == paramsFieldEnv {
		m.ParamsEnvInput, cmd = m.ParamsEnvInput.Update(msg)
	} else {
		m.ParamsFlagsInput, cmd = m.ParamsFlagsInput.Update(msg)
	}
	return m, cmd
}

// cancelParamsForm returns to the list view without running anything.
// Per the issue #37 acceptance criteria: cancel path performs no
// history, export, or shell-integration side effects.
func (m Model) cancelParamsForm() Model {
	m.State = StateList
	m.ParamsTarget = nil
	m.ParamsError = ""
	return m
}

// toggleParamsFocus moves focus between the env and flags inputs.
func (m Model) toggleParamsFocus() Model {
	if m.ParamsFocus == paramsFieldEnv {
		m.ParamsEnvInput.Blur()
		m.ParamsFlagsInput.Focus()
		m.ParamsFocus = paramsFieldFlags
	} else {
		m.ParamsFlagsInput.Blur()
		m.ParamsEnvInput.Focus()
		m.ParamsFocus = paramsFieldEnv
	}
	return m
}

// submitParamsForm validates the two inputs and either starts execution
// (or dangerous-confirmation) on success, or leaves the user in the
// form with ParamsError set on failure.
func (m Model) submitParamsForm() (tea.Model, tea.Cmd) {
	if m.ParamsTarget == nil {
		// Defensive: should not happen in practice, but treat as cancel.
		return m.cancelParamsForm(), nil
	}

	envRaw := m.ParamsEnvInput.Value()
	flagsRaw := m.ParamsFlagsInput.Value()

	env, err := parseEnvLine(envRaw)
	if err != nil {
		m.ParamsError = err.Error()
		return m, nil
	}
	flags, err := parseFlagsLine(flagsRaw)
	if err != nil {
		m.ParamsError = err.Error()
		return m, nil
	}

	target := *m.ParamsTarget
	parsed := ExecutionParams{
		EnvRaw:   envRaw,
		FlagsRaw: flagsRaw,
		Env:      env,
		Flags:    flags,
	}
	if m.LastRunParams == nil {
		m.LastRunParams = make(map[string]ExecutionParams)
	}
	m.LastRunParams[target.Name] = parsed

	m.CurrentRunOpts = executor.ExecutionOptions{
		DryRun: m.DryRun,
		Env:    env,
		Flags:  flags,
	}

	// Dangerous critical targets still go through confirmation, but
	// now with the freshly-parsed params already attached.
	if target.IsDangerous && target.DangerLevel == safety.SeverityCritical {
		targetCopy := target
		m.PendingTarget = &targetCopy
		m.State = StateConfirmDangerous
		m.ParamsError = ""
		return m, nil
	}

	return m.startExecution(target, m.CurrentRunOpts)
}

// startExecution is the single entry point into StateExecuting used by
// both the quick-run path (Enter from the list) and the parameterized
// path (form submission, optionally via dangerous confirmation).
//
// It updates history (skipped in dry-run, matching existing #38
// behavior) and kicks off the streaming executor.
func (m Model) startExecution(target Target, opts executor.ExecutionOptions) (tea.Model, tea.Cmd) {
	if !m.DryRun {
		m.History.RecordExecution(m.MakefilePath, target.Name)
		_ = m.History.Save()
		recentEntries := m.History.GetRecent(m.MakefilePath)
		m.RecentTargets = buildRecentTargets(recentEntries, m.Targets)
	}

	m.State = StateExecuting
	m.ExecutingTarget = target.Name
	m.ExecutionStartTime = time.Now()
	m.ExecutionElapsed = 0
	m.ParamsTarget = nil
	m.ParamsError = ""

	m.StreamingOutput = &strings.Builder{}
	m.initExecutingViewport()

	return m, tea.Batch(
		executeTargetStreaming(target.Name, m.MakefilePath, opts),
		tickTimer(),
		m.Spinner.Tick,
	)
}
