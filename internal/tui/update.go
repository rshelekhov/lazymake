package tui

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rshelekhov/lazymake/internal/executor"
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
				return m, executeTarget(target.Name)
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
		m.List.SetSize(msg.Width-4, msg.Height-4)
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
	case executeFinishedMsg:
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

	viewportWidth := width - 6      // 2 border + 4 padding
	viewportHeight := winHeight - 8 // header/footer + borders

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
				return m, executeTarget(target.Name)
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

func executeTarget(target string) tea.Cmd {
	return func() tea.Msg {
		result := executor.Execute(target)
		return executeFinishedMsg{result: result}
	}
}
