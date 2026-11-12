package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// renderWorkspaceView renders the workspace picker view
func (m Model) renderWorkspaceView() string {
	if m.Width == 0 || m.Height == 0 {
		return m.WorkspaceList.View()
	}

	// Calculate available space
	// Status bar takes 3 lines (border 2 + content 1)
	statusBarHeight := 3
	availableHeight := m.Height - statusBarHeight

	// Calculate width for the workspace list container
	// Use similar proportions as the target list
	containerWidth := m.Width

	// Render workspace list with border
	workspaceContainer := m.renderWorkspaceListContainer(containerWidth, availableHeight)

	// Render status bar
	statusBar := m.renderWorkspaceStatusBar()

	// Combine container with status bar
	return lipgloss.JoinVertical(
		lipgloss.Left,
		workspaceContainer,
		statusBar,
	)
}

// renderWorkspaceListContainer renders the workspace list with border (title is rendered by the list itself)
func (m Model) renderWorkspaceListContainer(width, height int) string {
	// Border adds 2 to height and width, padding adds more
	// Match the target list padding (2, 3) exactly
	contentWidth := width - 8   // 2 (border) + 6 (padding 3*2)
	contentHeight := height - 6 // 2 (border) + 4 (padding 2*2)

	// Set list size for this render - give full width for delegate to handle wrapping
	m.WorkspaceList.SetSize(contentWidth, contentHeight)

	// Get list content (title is rendered by the list itself)
	listContent := m.WorkspaceList.View()

	// Use lipgloss.Place to force content into exact dimensions
	// This ensures the content fills the entire space, even if list is shorter
	placedContent := lipgloss.Place(
		contentWidth,
		contentHeight,
		lipgloss.Left,
		lipgloss.Top,
		listContent,
	)

	// Apply modern border with padding matching target list
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BorderColor).
		Padding(2, 3).
		Margin(0, 0)

	return containerStyle.Render(placedContent)
}

// renderWorkspaceStatusBar renders the status bar for workspace view
func (m Model) renderWorkspaceStatusBar() string {
	items := m.WorkspaceList.Items()
	favoriteCount := 0
	workspaceCount := 0
	for _, item := range items {
		if ws, ok := item.(WorkspaceItem); ok {
			workspaceCount++
			if ws.Workspace.IsFavorite {
				favoriteCount++
			}
		}
	}

	var label string
	if workspaceCount == 1 {
		label = "1 workspace"
	} else {
		label = fmt.Sprintf("%d workspaces", workspaceCount)
	}

	sections := []string{coloredNugget(label)}

	if favoriteCount > 0 {
		var favLabel string
		if favoriteCount == 1 {
			favLabel = "1 favorite"
		} else {
			favLabel = fmt.Sprintf("%d favorites", favoriteCount)
		}
		sections = append(sections, plainNugget(favLabel))
	}

	leftBar := lipgloss.JoinHorizontal(lipgloss.Top, sections...)
	helpText := "enter: switch • f: favorite • esc/w: cancel • q: quit"

	return assembleStatusBar(m.Width, leftBar, helpText)
}
