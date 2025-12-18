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

	// Apply border style matching other views
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(SecondaryColor).
		Padding(1, 2).
		Width(m.Width - 2)

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

	var leftContent string
	if discoveredCount > 0 {
		leftContent = fmt.Sprintf("%d recent • %d discovered", recentCount, discoveredCount)
	} else {
		leftContent = fmt.Sprintf("%d workspaces", recentCount)
	}
	rightContent := "enter: switch • f: favorite • esc/w: cancel • q: quit"

	statusBar := renderStatusBar(m.Width, leftContent, rightContent)

	return content + "\n" + statusBar
}
