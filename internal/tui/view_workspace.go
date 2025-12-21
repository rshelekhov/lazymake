package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderWorkspaceView renders the workspace picker view
func (m Model) renderWorkspaceView() string {
	return m.renderWorkspaceList()
}

// renderWorkspaceList renders recent and discovered workspaces
func (m Model) renderWorkspaceList() string {
	var builder strings.Builder

	// Title
	title := TitleStyle.Render("Switch Workspace")
	builder.WriteString(title + "\n\n")

	// Render workspace list
	builder.WriteString(m.WorkspaceList.View())

	// Apply modern border style with subtle background and increased padding
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BorderColor).
		Background(BackgroundSubtle).
		Padding(2, 3).
		Width(m.Width - 4)

	content := containerStyle.Render(builder.String())

	// Status bar
	items := m.WorkspaceList.Items()
	recentCount := 0
	discoveredCount := 0
	for _, item := range items {
		if ws, ok := item.(WorkspaceItem); ok {
			if ws.Workspace.AccessCount > 0 {
				recentCount++
			} else {
				discoveredCount++
			}
		}
	}

	// Create status bar with background for entire bar
	statusBarStyle := lipgloss.NewStyle().
		Foreground(TextPrimary).
		Background(BackgroundSubtle)

	// Colored nugget style (only for first item)
	coloredNuggetStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"}).
		Background(PrimaryColor).
		Padding(0, 1).
		MarginRight(1)

	// Plain nugget style (inherits status bar background, just text)
	plainNuggetStyle := lipgloss.NewStyle().
		Foreground(TextSecondary).
		Padding(0, 1)

	// Recent workspaces nugget (only colored one)
	recentNugget := coloredNuggetStyle.Render(fmt.Sprintf("%d recent", recentCount))

	var sections []string
	sections = append(sections, recentNugget)

	// Discovered workspaces (plain text on status bar background)
	if discoveredCount > 0 {
		discoveredInfo := plainNuggetStyle.Render(fmt.Sprintf("%d discovered", discoveredCount))
		sections = append(sections, discoveredInfo)
	}

	leftBar := lipgloss.JoinHorizontal(lipgloss.Top, sections...)
	leftWidth := lipgloss.Width(leftBar)

	// Help text on the right
	helpText := "enter: switch • f: favorite • esc/w: cancel • q: quit"
	middleWidth := max(m.Width-leftWidth-lipgloss.Width(helpText)-6, 1)
	middle := lipgloss.NewStyle().Width(middleWidth).Render("")

	right := lipgloss.NewStyle().
		Foreground(TextMuted).
		Padding(0, 1).
		Render(helpText)

	bar := lipgloss.JoinHorizontal(lipgloss.Top, leftBar, middle, right)
	statusBar := statusBarStyle.Width(m.Width).Render(bar)

	return content + "\n" + statusBar
}
