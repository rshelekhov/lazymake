package tui

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rshelekhov/lazymake/internal/executor"
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
				m.State = StateExecuting
				m.ExecutingTarget = target.Name
				return m, executeTarget(target.Name)
			}
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
