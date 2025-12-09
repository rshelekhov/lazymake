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
	// If there's an error in the model, handle window size and quit commands
	if m.Err != nil {
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

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.State == StateOutput {
			// In output view, handle viewport scrolling and navigation
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "esc":
				m.State = StateList
				return m, nil
			}

			m.Viewport, cmd = m.Viewport.Update(msg)
			return m, cmd
		}

		if m.State == StateHelp {
			// In help view, handle navigation
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "esc", "?":
				m.State = StateList
				return m, nil
			}
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "?":
			if m.State == StateList {
				m.State = StateHelp
				return m, nil
			}
		case "enter":
			if m.State == StateList {
				selected := m.List.SelectedItem()
				if target, ok := selected.(Target); ok {
					m.State = StateExecuting
					m.ExecutingTarget = target.Name
					return m, executeTarget(target.Name)
				}
			}
		}

	case executeFinishedMsg:
		m.State = StateOutput
		m.Output = msg.result.Output
		m.ExecutionError = msg.result.Err

		m.initViewport(msg.result.Output)
		return m, nil

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.List.SetSize(msg.Width-4, msg.Height-4)

		if m.State == StateOutput {
			m.resizeViewport()
		}
	}

	// Let list handle its updates
	m.List, cmd = m.List.Update(msg)

	return m, cmd
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
