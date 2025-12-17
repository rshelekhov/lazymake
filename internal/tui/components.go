package tui

import "github.com/charmbracelet/lipgloss"

// renderStatusBar creates a reusable status bar with left and right content
// This provides consistent layout and styling across all views
func renderStatusBar(width int, leftContent, rightContent string) string {
	contentWidth := width - 2
	leftWidth := len(leftContent) + 2

	// Ensure valid widths
	rightWidth := max(contentWidth-leftWidth, 0)

	// Apply styles with proper alignment
	left := StatusBarStyle.Width(leftWidth).Render(leftContent)
	right := StatusBarStyle.
		Width(rightWidth).
		Align(lipgloss.Right).
		Padding(0, 3, 0, 0).
		Render(rightContent)

	content := left + right

	// Use lipgloss.Place to force content into exact width
	placedContent := lipgloss.Place(
		contentWidth,
		1, // Single line height
		lipgloss.Left,
		lipgloss.Center,
		content,
	)

	return StatusBarStyle.Render(placedContent)
}
