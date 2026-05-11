package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rshelekhov/lazymake/internal/executor"
)

// handleRerunLast implements the ctrl+r shortcut (issue #36): repeat
// the last recorded execution of the target under the cursor, reusing
// the env vars and make flags that were applied last time.
//
// Behavior:
//   - cursor not on a Target → no-op (the key press is silently absorbed).
//   - target has no execution history yet → no-op, mirroring the way
//     'R' (rerun last preset) ignores the press when nothing is saved.
//   - target has history → pin the target on PendingTarget, snapshot
//     env/flags into CurrentRunOpts (DryRun comes from the session
//     flag, never from history), and switch to StateConfirmRerun so
//     the user sees what is about to run before it starts.
//
// The dangerous-confirmation step is intentionally bypassed for the
// rerun path. Per the #36 design we render the safety details inside
// renderConfirmRerunView so the user confirms both "rerun" and "this
// is critical" in a single keystroke instead of stacking two modals.
func (m Model) handleRerunLast() (tea.Model, tea.Cmd) {
	selected := m.List.SelectedItem()
	target, ok := selected.(Target)
	if !ok {
		return m, nil
	}
	if m.History == nil {
		return m, nil
	}

	rec, ok := m.History.LastExecution(m.MakefilePath, target.Name)
	if !ok {
		return m, nil
	}

	targetCopy := target
	m.PendingTarget = &targetCopy
	m.CurrentRunOpts = executor.ExecutionOptions{
		DryRun: m.DryRun,
		Env:    rec.Env,
		Flags:  rec.Flags,
	}
	m.State = StateConfirmRerun
	return m, nil
}

// updateConfirmRerun drives the rerun-last confirmation dialog. The
// modal has only two actions — confirm and cancel — so the message
// switch is intentionally small.
//
// Confirm path (Enter): hand off to startExecution with the params
// already snapshotted by handleRerunLast. startExecution writes a new
// history entry, so the next ctrl+r will pick up this run as the
// freshest record.
//
// Cancel path (Esc): clear PendingTarget and CurrentRunOpts so the
// next dangerous-confirm or quick-Enter doesn't inherit a stale plan.
func (m Model) updateConfirmRerun(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.State = StateList
			m.PendingTarget = nil
			m.CurrentRunOpts = executor.ExecutionOptions{}
			return m, nil
		case "enter":
			if m.PendingTarget == nil {
				// Defensive: shouldn't happen because we entered this
				// state only after pinning a target.
				m.State = StateList
				return m, nil
			}
			target := *m.PendingTarget
			m.PendingTarget = nil
			opts := m.CurrentRunOpts
			opts.DryRun = m.DryRun // session flag always wins
			return m.startExecution(target, opts)
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}
	return m, nil
}
