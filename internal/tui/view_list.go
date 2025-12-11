package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/rshelekhov/lazymake/internal/safety"
)

// renderListView renders the main two-column list view
func (m Model) renderListView() string {
	if m.Width == 0 || m.Height == 0 {
		return m.List.View()
	}

	// Get selected target for recipe preview
	var selectedTarget *Target
	if item := m.List.SelectedItem(); item != nil {
		if target, ok := item.(Target); ok {
			selectedTarget = &target
		}
	}

	// Calculate column widths (30% left, 70% right)
	leftWidth := max(int(float64(m.Width)*0.30), 30)
	rightWidth := m.Width - leftWidth - 4 // -4 for margins and spacing

	// Render left column (target list)
	leftColumn := m.renderTargetList(leftWidth)

	// Render right column (recipe preview)
	rightColumn := m.renderRecipePreview(selectedTarget, rightWidth)

	// Render status bar
	statusBar := m.renderStatusBar()

	// Calculate height for columns (leave space for status bar)
	columnHeight := m.Height - 3 // -3 for status bar and spacing

	// Apply height constraints
	leftStyle := lipgloss.NewStyle().
		Width(leftWidth).
		Height(columnHeight).
		MaxHeight(columnHeight)

	rightStyle := lipgloss.NewStyle().
		Width(rightWidth).
		Height(columnHeight).
		MaxHeight(columnHeight)

	// Join columns horizontally
	columns := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftStyle.Render(leftColumn),
		rightStyle.Render(rightColumn),
	)

	// Combine with status bar
	return lipgloss.JoinVertical(
		lipgloss.Left,
		columns,
		statusBar,
	)
}

// renderTargetList renders the left column with target list
func (m Model) renderTargetList(width int) string {
	// The list already handles its own styling
	return m.List.View()
}

// renderRecipePreview renders the right column with recipe and safety info
func (m Model) renderRecipePreview(target *Target, width int) string {
	if target == nil {
		return renderEmptyPreview(width)
	}

	var content strings.Builder

	// Target name header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(PrimaryColor).
		Render(target.Name + ":")
	content.WriteString(header + "\n\n")

	// Recipe commands
	if len(target.Recipe) > 0 {
		recipeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

		for _, line := range target.Recipe {
			// Indent each recipe line with tab
			content.WriteString(recipeStyle.Render("  " + line) + "\n")
		}
	} else {
		noRecipeStyle := lipgloss.NewStyle().
			Foreground(MutedColor).
			Italic(true)
		content.WriteString(noRecipeStyle.Render("  (no recipe - meta target)") + "\n")
	}

	// Safety warnings (if dangerous)
	if target.IsDangerous && len(target.SafetyMatches) > 0 {
		content.WriteString("\n")
		content.WriteString(renderSafetyWarnings(target.SafetyMatches))
	}

	// Wrap in container with padding and border
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(SecondaryColor).
		Padding(1, 2).
		Width(width - 2). // -2 for border
		Height(0)         // Auto-height

	return containerStyle.Render(content.String())
}

// renderSafetyWarnings renders safety match information
func renderSafetyWarnings(matches []safety.MatchResult) string {
	var content strings.Builder

	for i, match := range matches {
		if i > 0 {
			content.WriteString("\n")
		}

		// Severity indicator and rule ID
		var severityStr string
		var severityColor lipgloss.Color

		switch match.Severity {
		case safety.SeverityCritical:
			severityStr = "🚨 CRITICAL"
			severityColor = ErrorColor
		case safety.SeverityWarning:
			severityStr = "⚠️  WARNING"
			severityColor = lipgloss.Color("#FFA500") // Orange
		case safety.SeverityInfo:
			severityStr = "ℹ️  INFO"
			severityColor = SecondaryColor
		}

		severityHeader := lipgloss.NewStyle().
			Foreground(severityColor).
			Bold(true).
			Render(severityStr + ": " + match.Rule.ID)
		content.WriteString(severityHeader + "\n")

		// Matched line
		if match.MatchedLine != "" {
			matchedStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#666666")).
				Render("  Matched: " + match.MatchedLine)
			content.WriteString(matchedStyle + "\n")
		}

		// Description
		if match.Rule.Description != "" {
			descStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#AAAAAA"))
			content.WriteString(descStyle.Render("  " + match.Rule.Description) + "\n")
		}

		// Suggestion
		if match.Rule.Suggestion != "" {
			suggestionStyle := lipgloss.NewStyle().
				Foreground(SecondaryColor).
				Italic(true)
			content.WriteString(suggestionStyle.Render("  💡 " + match.Rule.Suggestion) + "\n")
		}
	}

	return content.String()
}

// renderEmptyPreview shows placeholder when no target selected
func renderEmptyPreview(width int) string {
	emptyStyle := lipgloss.NewStyle().
		Foreground(MutedColor).
		Italic(true).
		Align(lipgloss.Center).
		Width(width)

	return emptyStyle.Render("Select a target to preview recipe")
}

// renderStatusBar renders the bottom status bar
func (m Model) renderStatusBar() string {
	// Count stats
	totalTargets := 0
	dangerousCount := 0

	for _, item := range m.List.Items() {
		if target, ok := item.(Target); ok {
			totalTargets++
			if target.IsDangerous {
				dangerousCount++
			}
		}
	}

	// Left side: stats
	var leftContent string
	if dangerousCount > 0 {
		leftContent = fmt.Sprintf("%d targets • %d dangerous", totalTargets, dangerousCount)
	} else {
		leftContent = fmt.Sprintf("%d targets", totalTargets)
	}

	// Right side: shortcuts
	rightContent := "enter: run • g: graph • ?: help • q: quit"

	// If dangerous target selected, show warning
	if item := m.List.SelectedItem(); item != nil {
		if target, ok := item.(Target); ok && target.IsDangerous {
			if target.DangerLevel == safety.SeverityCritical {
				rightContent = "⚠️  Dangerous command • enter: confirm & run • q: quit"
			}
		}
	}

	// Build status bar with two sections
	leftStyle := lipgloss.NewStyle().
		Foreground(MutedColor)

	rightStyle := lipgloss.NewStyle().
		Foreground(MutedColor).
		Align(lipgloss.Right)

	// Calculate widths
	leftWidth := len(leftContent) + 2
	rightWidth := m.Width - leftWidth - 2

	left := leftStyle.Width(leftWidth).Render(leftContent)
	right := rightStyle.Width(rightWidth).Render(rightContent)

	statusBarStyle := lipgloss.NewStyle().
		Foreground(MutedColor).
		Background(lipgloss.Color("#2A2A2A")).
		Padding(0, 1).
		Width(m.Width)

	return statusBarStyle.Render(left + right)
}
