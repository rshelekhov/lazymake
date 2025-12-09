package tui

import "github.com/charmbracelet/lipgloss"

// getContentWidth calculates responsive width for content blocks
// Uses 90% of terminal width with min/max constraints
func getContentWidth(terminalWidth int) int {
	width := min(max(int(float64(terminalWidth)*0.9), 40), 120)
	return width
}

func (m Model) View() string {
	if m.Err != nil {
		contentWidth := getContentWidth(m.Width)
		// Calculate inner width for text: subtract border (2) + padding (4) = 6
		innerWidth := max(contentWidth-6, 20)

		errorStyle := lipgloss.NewStyle().
			Foreground(ErrorColor).
			Bold(true).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ErrorColor).
			Padding(1, 2).
			Width(innerWidth)

		return "\n" + errorStyle.Render("Error: "+m.Err.Error()) + "\n\n  Press q to quit\n"
	}

	switch m.State {
	case StateExecuting:
		return "\n  ⏳ Executing: make " + m.ExecutingTarget + "\n\n  Please wait...\n"

	case StateOutput:
		var header string
		if m.ExecutionError != nil {
			header = ErrorStyle.Render("❌ Failed: make " + m.ExecutingTarget)
		} else {
			header = SuccessStyle.Render("✓ Success: make " + m.ExecutingTarget)
		}

		contentWidth := getContentWidth(m.Width)
		// Calculate inner width for text: subtract border (2) + padding (4) = 6
		innerWidth := max(contentWidth-6, 20)

		viewportStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(SecondaryColor).
			Padding(1, 2).
			Width(innerWidth)

		footer := lipgloss.NewStyle().
			Foreground(MutedColor).
			Render("\nPress esc to return • q to quit")

		return "\n" + header + "\n\n" + viewportStyle.Render(m.Viewport.View()) + footer

	default:
		return lipgloss.NewStyle().Margin(1, 2).Render(m.List.View())
	}
}
