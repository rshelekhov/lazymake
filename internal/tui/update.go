package tui

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rshelekhov/lazymake/internal/executor"
	"github.com/rshelekhov/lazymake/internal/export"
	"github.com/rshelekhov/lazymake/internal/safety"
)

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.Err != nil {
		return m.updateError(msg)
	}

	switch m.State {
	case StateList:
		return m.updateList(msg)
	case StateOutput:
		return m.updateOutput(msg)
	case StateHelp:
		return m.updateHelp(msg)
	case StateGraph:
		return m.updateGraph(msg)
	case StateConfirmDangerous:
		return m.updateConfirmDangerous(msg)
	case StateExecuting:
		return m.updateExecuting(msg)
	case StateVariables:
		return m.updateVariables(msg)
	case StateWorkspace:
		return m.updateWorkspace(msg)
	default:
		return m, nil
	}
}

func (m Model) updateError(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}
	return m, nil
}

func (m Model) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "?":
			m.State = StateHelp
			return m, nil

		case "v":
			// Toggle variable inspector view
			m.State = StateVariables
			m.VariableListIndex = 0
			return m, nil

		case "w":
			// Open workspace picker
			m.State = StateWorkspace
			m.initWorkspacePicker()
			return m, nil

		case "g":
			// Show graph view for selected target
			selected := m.List.SelectedItem()
			if target, ok := selected.(Target); ok {
				m.State = StateGraph
				m.GraphTarget = target.Name
				return m, nil
			}

		case "enter":
			selected := m.List.SelectedItem()
			if target, ok := selected.(Target); ok {
				// Check if target is critical and requires confirmation
				if target.IsDangerous && target.DangerLevel == safety.SeverityCritical {
					// Show confirmation dialog for critical targets
					targetCopy := target
					m.PendingTarget = &targetCopy
					m.State = StateConfirmDangerous
					return m, nil
				}

				// Safe or non-critical target - execute immediately
				// Record execution in history BEFORE starting
				m.History.RecordExecution(m.MakefilePath, target.Name)
				_ = m.History.Save() // Async, ignore errors (non-critical)

				// Refresh recent targets for next render
				recentEntries := m.History.GetRecent(m.MakefilePath)
				m.RecentTargets = buildRecentTargets(recentEntries, m.Targets)

				m.State = StateExecuting
				m.ExecutingTarget = target.Name
				m.ExecutionStartTime = time.Now()
				m.ExecutionElapsed = 0
				return m, tea.Batch(
					executeTarget(target.Name),
					tickTimer(), // Start timer
				)
			}

		case "down", "j":
			// Navigate down, skipping over separators and headers
			m = navigateToNextTarget(m, true)
			return m, nil

		case "up", "k":
			// Navigate up, skipping over separators and headers
			m = navigateToNextTarget(m, false)
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

		// Calculate list size for left column
		// Left column is 30% of width
		leftWidth := max(int(float64(msg.Width)*0.30), 30)
		// Account for border (2) on left column
		listWidth := leftWidth - 2
		// Account for status bar (3) and border (2) on columns
		listHeight := msg.Height - 3 - 2

		m.List.SetSize(listWidth, listHeight)
	}

	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)

	return m, cmd
}

// navigateToNextTarget moves the cursor to the next/previous target, skipping separators and headers
func navigateToNextTarget(m Model, down bool) Model {
	items := m.List.Items()
	currentIndex := m.List.Index()

	if down {
		// Search downward for next target
		for i := currentIndex + 1; i < len(items); i++ {
			if _, ok := items[i].(Target); ok {
				m.List.Select(i)
				return m
			}
		}
	} else {
		// Search upward for previous target
		for i := currentIndex - 1; i >= 0; i-- {
			if _, ok := items[i].(Target); ok {
				m.List.Select(i)
				return m
			}
		}
	}

	return m
}

func (m Model) updateOutput(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.State = StateList
			return m, nil
		}
		var cmd tea.Cmd
		m.Viewport, cmd = m.Viewport.Update(msg)
		return m, cmd

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.resizeViewport()
	}

	return m, nil
}

func (m Model) updateHelp(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc", "?":
			m.State = StateList
		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}
	return m, nil
}

// updateGraph handles the graph view state
func (m Model) updateGraph(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "esc", "g":
			// Return to list view and clear graph target
			m.State = StateList
			m.GraphTarget = ""
			return m, nil

		case "+", "=":
			// Increase depth (show more levels)
			if m.GraphDepth == -1 {
				// Already unlimited, do nothing
			} else {
				m.GraphDepth++
			}

		case "-", "_":
			// Decrease depth (show fewer levels)
			if m.GraphDepth == -1 {
				m.GraphDepth = 5 // Start with 5 levels when coming from unlimited
			} else if m.GraphDepth > 0 {
				m.GraphDepth--
			}

		case "o", "O":
			// Toggle order display
			m.ShowOrder = !m.ShowOrder

		case "c", "C":
			// Toggle critical path display
			m.ShowCritical = !m.ShowCritical

		case "p", "P":
			// Toggle parallel display
			m.ShowParallel = !m.ShowParallel
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}

	return m, nil
}

func (m Model) updateExecuting(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case timerTickMsg:
		// Update elapsed time
		if m.State == StateExecuting {
			m.ExecutionElapsed = time.Since(m.ExecutionStartTime)
			return m, tickTimer() // Continue ticking
		}
		return m, nil // Stop if not executing

	case executeFinishedMsg:
		// Timer stops automatically (state changes from StateExecuting)

		// Record execution with timing data
		success := msg.result.Err == nil
		m.History.RecordExecutionWithTiming(m.MakefilePath, m.ExecutingTarget, msg.result.Duration, success)
		_ = m.History.Save() // Async, ignore errors

		// Export execution result (async, non-blocking)
		if m.Exporter != nil {
			go func() {
				record := export.NewExecutionRecord(
					m.MakefilePath,
					m.ExecutingTarget,
					msg.result,
				)
				if err := m.Exporter.Export(record); err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "Export failed: %v\n", err)
				}
			}()
		}

		// Shell integration (async, non-blocking)
		if m.ShellIntegration != nil {
			go func() {
				if err := m.ShellIntegration.RecordExecution(m.ExecutingTarget); err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "Shell integration failed: %v\n", err)
				}
			}()
		}

		// Refresh performance stats for all targets
		enrichTargetsWithPerformance(m.History, m.MakefilePath, m.Targets)

		// Refresh recent targets to show updated timing
		recentEntries := m.History.GetRecent(m.MakefilePath)
		m.RecentTargets = buildRecentTargets(recentEntries, m.Targets)

		// Rebuild and update list items to reflect new performance stats
		updatedItems := rebuildListItems(m.RecentTargets, m.Targets)
		m.List.SetItems(updatedItems)

		m.State = StateOutput
		m.Output = msg.result.Output
		m.ExecutionError = msg.result.Err
		m.initViewport(msg.result.Output)

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}
	return m, nil
}

func (m *Model) initViewport(content string) {
	vw, vh := computeViewportSize(m.Width, m.Height)
	m.Viewport = viewport.New(vw, vh)
	m.Viewport.SetContent(content)
	m.Viewport.Style = lipgloss.NewStyle()
}

func (m *Model) resizeViewport() {
	vw, vh := computeViewportSize(m.Width, m.Height)
	m.Viewport.Width = vw
	m.Viewport.Height = vh
}

func computeViewportSize(winWidth, winHeight int) (int, int) {
	width := getContentWidth(winWidth)

	viewportWidth := width - 6 // 2 border + 4 padding

	// Account for UI elements in renderOutputView:
	// - Leading newline: 1 line
	// - Header: 1 line
	// - Potential regression alert: 0-2 lines
	// - Double newline: 2 lines
	// - Viewport border: 2 lines (top + bottom)
	// - Viewport padding: 2 lines (top + bottom)
	// - Footer: 2 lines
	// Total overhead: 10-12 lines
	// Use 14 to include safety margin for text wrapping
	viewportHeight := winHeight - 14

	if viewportWidth < 20 {
		viewportWidth = 20
	}
	if viewportHeight < 3 {
		viewportHeight = 3
	}

	return viewportWidth, viewportHeight
}

// updateConfirmDangerous handles the dangerous command confirmation dialog
func (m Model) updateConfirmDangerous(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "esc":
			// Cancel confirmation, return to list
			m.State = StateList
			m.PendingTarget = nil
			return m, nil

		case "enter":
			// Proceed with execution of dangerous target
			if m.PendingTarget != nil {
				target := *m.PendingTarget

				// Record execution in history
				m.History.RecordExecution(m.MakefilePath, target.Name)
				_ = m.History.Save()

				// Refresh recent targets
				recentEntries := m.History.GetRecent(m.MakefilePath)
				m.RecentTargets = buildRecentTargets(recentEntries, m.Targets)

				// Clear pending target and start execution
				m.PendingTarget = nil
				m.State = StateExecuting
				m.ExecutingTarget = target.Name
				m.ExecutionStartTime = time.Now()
				m.ExecutionElapsed = 0
				return m, tea.Batch(
					executeTarget(target.Name),
					tickTimer(), // Start timer
				)
			}
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}

	return m, nil
}

// Custom message for execution results
type executeFinishedMsg struct {
	result executor.Result
}

// Custom message for timer ticks
type timerTickMsg struct{}

func executeTarget(target string) tea.Cmd {
	return func() tea.Msg {
		result := executor.Execute(target)
		return executeFinishedMsg{result: result}
	}
}

func tickTimer() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return timerTickMsg{}
	})
}

// updateVariables handles the variable inspector view state
func (m Model) updateVariables(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "esc", "v":
			// Return to list view
			m.State = StateList
			return m, nil

		case "up", "k":
			// Navigate up in variable list
			if m.VariableListIndex > 0 {
				m.VariableListIndex--
			}

		case "down", "j":
			// Navigate down in variable list
			if m.VariableListIndex < len(m.Variables)-1 {
				m.VariableListIndex++
			}
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}

	return m, nil
}
