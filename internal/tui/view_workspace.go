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

	// Stats with badges
	var leftContent string
	if discoveredCount > 0 {
		recentBadge := Badge(fmt.Sprintf("%d recent", recentCount), TextSecondary, BackgroundSubtle)
		discoveredBadge := Badge(fmt.Sprintf("%d discovered", discoveredCount), TextSecondary, BackgroundSubtle)
		leftContent = recentBadge + "  " + discoveredBadge
	} else {
		leftContent = Badge(fmt.Sprintf("%d workspaces", recentCount), TextSecondary, BackgroundSubtle)
	}
	rightContent := "enter: switch • f: favorite • esc/w: cancel • q: quit"

	statusBar := renderStatusBar(m.Width, leftContent, rightContent)

	return content + "\n" + statusBar
}
