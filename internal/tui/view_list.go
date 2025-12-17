package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/rshelekhov/lazymake/internal/safety"
	"github.com/rshelekhov/lazymake/internal/util"
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

	// Calculate available space
	// Status bar takes 3 lines (border 2 + content 1)
	statusBarHeight := 3
	availableHeight := m.Height - statusBarHeight

	// Calculate left column width (30% of terminal width, minimum 30 chars)
	leftWidthPercent := 0.30
	minLeftWidth := 30

	leftWidth := int(float64(m.Width) * leftWidthPercent)
	if leftWidth < minLeftWidth && m.Width >= minLeftWidth*2 {
		leftWidth = minLeftWidth
	} else if leftWidth < minLeftWidth {
		// Terminal too narrow for minimum, use percentage
		leftWidth = int(float64(m.Width) * leftWidthPercent)
	}

	// Safety check: ensure left width is valid
	if leftWidth < 10 {
		leftWidth = 10
	}

	// Render left column first
	leftColumn := m.renderTargetList(leftWidth, availableHeight)

	// Measure actual rendered width of left column
	actualLeftWidth := lipgloss.Width(leftColumn)

	// Calculate right column width based on ACTUAL measured left width
	// This prevents any rounding errors or overflow
	rightWidth := max(m.Width-actualLeftWidth, 10)

	// Render right column with measured width
	rightColumn := m.renderRecipePreview(selectedTarget, rightWidth, availableHeight)

	// Render status bar
	statusBar := m.renderStatusBar()

	// Join columns horizontally (both same height)
	columns := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftColumn,
		rightColumn,
	)

	// Combine with status bar
	return lipgloss.JoinVertical(
		lipgloss.Left,
		columns,
		statusBar,
	)
}

// renderTargetList renders the left column with target list and border
func (m Model) renderTargetList(width, height int) string {
	// Border adds 2 to height (1 top + 1 bottom) and 2 to width (1 left + 1 right)
	contentWidth := width - 2
	contentHeight := height - 2

	// Set list size for this render - give full width for delegate to handle wrapping
	m.List.SetSize(contentWidth, contentHeight)

	// Get list content
	listContent := m.List.View()

	// Use lipgloss.Place to force content into exact dimensions
	// This ensures the content fills the entire space, even if list is shorter
	placedContent := lipgloss.Place(
		contentWidth,
		contentHeight,
		lipgloss.Left,
		lipgloss.Top,
		listContent,
	)

	// Apply border WITHOUT Width/Height (let it wrap the placed content naturally)
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(SecondaryColor)

	return containerStyle.Render(placedContent)
}

// renderRecipePreview renders the right column with recipe and safety info
func (m Model) renderRecipePreview(target *Target, width, height int) string {
	if target == nil {
		return renderEmptyPreview(width, height)
	}

	var builder strings.Builder

	// Target name header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(PrimaryColor).
		Render(target.Name + ":")
	util.WriteString(&builder, header+"\n\n")

	// Recipe commands
	if len(target.Recipe) > 0 {
		recipeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

		for _, line := range target.Recipe {
			// Indent each recipe line with tab
			util.WriteString(&builder, recipeStyle.Render("  "+line)+"\n")
		}

		// Graph view hint
		util.WriteString(&builder, "\n")
		hintStyle := lipgloss.NewStyle().
			Foreground(SecondaryColor).
			Italic(true)
		util.WriteString(&builder, hintStyle.Render("  💡 Press 'g' to view full dependency graph")+"\n")
	} else {
		noRecipeStyle := lipgloss.NewStyle().
			Foreground(MutedColor).
			Italic(true)
		util.WriteString(&builder, noRecipeStyle.Render("  (no recipe - meta target)")+"\n")
	}

	// Variables section (if any variables used by this target)
	targetVariables := m.getVariablesForTarget(target.Name)
	if len(targetVariables) > 0 {
		util.WriteString(&builder, "\n")
		util.WriteString(&builder, renderVariablesSection(targetVariables))
	}

	// Safety warnings (if dangerous)
	if target.IsDangerous && len(target.SafetyMatches) > 0 {
		util.WriteString(&builder, "\n")
		util.WriteString(&builder, renderSafetyWarnings(target.SafetyMatches))
	}

	// Performance section (context-aware)
	perfSection := renderPerformanceSection(*target)
	if perfSection != "" {
		util.WriteString(&builder, "\n")
		util.WriteString(&builder, perfSection)
	}

	// Padding(1,2) = 2 vertical + 4 horizontal
	// Border = 2 vertical + 2 horizontal
	// Total overhead: 4 vertical, 6 horizontal
	contentWidth := width - 6
	contentHeight := height - 4

	// Use lipgloss.Place to force content into exact dimensions
	placedContent := lipgloss.Place(
		contentWidth,
		contentHeight,
		lipgloss.Left,
		lipgloss.Top,
		builder.String(),
	)

	// Apply padding and border WITHOUT Width/Height (let it wrap naturally)
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(SecondaryColor).
		Padding(1, 2)

	return containerStyle.Render(placedContent)
}

// renderSafetyWarnings renders safety match information
func renderSafetyWarnings(matches []safety.MatchResult) string {
	var builder strings.Builder

	for i, match := range matches {
		if i > 0 {
			util.WriteString(&builder, "\n")
		}

		// Severity indicator and rule ID
		var severityStr string
		var severityColor lipgloss.AdaptiveColor

		switch match.Severity {
		case safety.SeverityCritical:
			severityStr = "🚨 CRITICAL"
			severityColor = ErrorColor
		case safety.SeverityWarning:
			severityStr = "⚠️  WARNING"
			severityColor = lipgloss.AdaptiveColor{Light: "#FFA500", Dark: "#FFA500"}
		case safety.SeverityInfo:
			severityStr = "ℹ️  INFO"
			severityColor = SecondaryColor
		}

		severityHeader := lipgloss.NewStyle().
			Foreground(severityColor).
			Bold(true).
			Render(severityStr + ": " + match.Rule.ID)
		util.WriteString(&builder, severityHeader+"\n")

		// Matched line
		if match.MatchedLine != "" {
			matchedStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#666666")).
				Render("  Matched: " + match.MatchedLine)
			util.WriteString(&builder, matchedStyle+"\n")
		}

		// Description
		if match.Rule.Description != "" {
			descStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#AAAAAA"))
			util.WriteString(&builder, descStyle.Render("  "+match.Rule.Description)+"\n")
		}

		// Suggestion
		if match.Rule.Suggestion != "" {
			suggestionStyle := lipgloss.NewStyle().
				Foreground(SecondaryColor).
				Italic(true)
			util.WriteString(&builder, suggestionStyle.Render("  💡 "+match.Rule.Suggestion)+"\n")
		}
	}

	return builder.String()
}

// renderEmptyPreview shows placeholder when no target selected
func renderEmptyPreview(width, height int) string {
	emptyText := "Select a target to preview recipe"

	emptyStyle := lipgloss.NewStyle().
		Foreground(MutedColor).
		Italic(true)

	content := emptyStyle.Render(emptyText)

	// Same dimensions as recipe preview
	contentWidth := width - 6
	contentHeight := height - 4

	// Use lipgloss.Place to force content into exact dimensions
	// Center the text within the space
	placedContent := lipgloss.Place(
		contentWidth,
		contentHeight,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)

	// Wrap in border with padding WITHOUT Width/Height
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(SecondaryColor).
		Padding(1, 2)

	return borderStyle.Render(placedContent)
}

// renderStatusBar renders the bottom status bar
func (m Model) renderStatusBar() string {
	// Count stats
	totalTargets := 0
	dangerousCount := 0
	regressedCount := 0

	for _, item := range m.List.Items() {
		if target, ok := item.(Target); ok {
			totalTargets++
			if target.IsDangerous {
				dangerousCount++
			}
			if target.PerfStats != nil && target.PerfStats.IsRegressed {
				regressedCount++
			}
		}
	}

	// Left side: stats
	leftStats := []string{fmt.Sprintf("%d targets", totalTargets)}

	if dangerousCount > 0 {
		leftStats = append(leftStats, fmt.Sprintf("%d dangerous", dangerousCount))
	}

	if regressedCount > 0 {
		leftStats = append(leftStats, fmt.Sprintf("%d regressed 📈", regressedCount))
	}

	leftContent := strings.Join(leftStats, " • ")

	// Right side: shortcuts - dynamically build from key bindings
	var rightContent string

	// If dangerous target selected, show warning with specific keys
	if item := m.List.SelectedItem(); item != nil {
		if target, ok := item.(Target); ok && target.IsDangerous {
			if target.DangerLevel == safety.SeverityCritical {
				rightContent = "⚠️  Dangerous command • enter: confirm & run • q: quit"
			} else {
				// Non-critical dangerous target, show normal shortcuts
				rightContent = formatKeyBindings(m.KeyBindings)
			}
		} else {
			// Normal target, show all shortcuts
			rightContent = formatKeyBindings(m.KeyBindings)
		}
	} else {
		// No target selected, show all shortcuts
		rightContent = formatKeyBindings(m.KeyBindings)
	}

	// Use the reusable status bar component
	return renderStatusBar(m.Width, leftContent, rightContent)
}

// formatKeyBindings formats key bindings as "key: description • key: description • ..."
func formatKeyBindings(bindings []key.Binding) string {
	var parts []string
	for _, binding := range bindings {
		help := binding.Help()
		parts = append(parts, help.Key+": "+help.Desc)
	}
	return strings.Join(parts, " • ")
}

// renderPerformanceSection returns context-aware performance info
// Returns empty string if no performance data or not relevant to show
func renderPerformanceSection(target Target) string {
	if target.PerfStats == nil {
		return "" // No data, no section
	}

	stats := target.PerfStats

	// Context-aware content
	if stats.IsRegressed {
		return renderRegressionAlert(target)
	} else if target.IsRecent && stats.ExecutionCount > 0 {
		return renderRecentTargetInfo(target)
	}

	return "" // Default: no section
}

// renderRegressionAlert renders a warning about performance regression
func renderRegressionAlert(target Target) string {
	stats := target.PerfStats
	change := int(((float64(stats.LastDuration) - float64(stats.AvgDuration)) / float64(stats.AvgDuration)) * 100)

	var builder strings.Builder

	// Separator
	separator := lipgloss.NewStyle().
		Foreground(MutedColor).
		Render("  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	util.WriteString(&builder, separator+"\n\n")

	// Header
	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color("220")).
		Bold(true).
		Render("  📈 Performance Regression Detected")
	util.WriteString(&builder, header+"\n\n")

	// Stats
	statsStyle := lipgloss.NewStyle().Foreground(TextColor)
	util.WriteString(&builder, statsStyle.Render(fmt.Sprintf("    Current:  %s\n", formatDuration(stats.LastDuration))))
	util.WriteString(&builder, statsStyle.Render(fmt.Sprintf("    Average:  %s (last %d runs)\n", formatDuration(stats.AvgDuration), stats.ExecutionCount)))

	changeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)
	util.WriteString(&builder, statsStyle.Render("    Change:   "))
	util.WriteString(&builder, changeStyle.Render(fmt.Sprintf("+%d%% slower", change))+"\n\n")

	// Hint
	hintStyle := lipgloss.NewStyle().
		Foreground(SecondaryColor).
		Italic(true)
	util.WriteString(&builder, hintStyle.Render("    💡 This target recently got slower - investigate recent changes"))

	return builder.String()
}

// renderRecentTargetInfo renders performance info for recently executed targets
func renderRecentTargetInfo(target Target) string {
	stats := target.PerfStats

	var builder strings.Builder

	// Separator
	separator := lipgloss.NewStyle().
		Foreground(MutedColor).
		Render("  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	util.WriteString(&builder, separator+"\n\n")

	// Header
	header := lipgloss.NewStyle().
		Foreground(SecondaryColor).
		Bold(true).
		Render("  ⏱ Performance")
	util.WriteString(&builder, header+"\n\n")

	// Stats
	statsStyle := lipgloss.NewStyle().Foreground(TextColor)
	util.WriteString(&builder, statsStyle.Render(fmt.Sprintf("    Last run: %s\n", formatDuration(stats.LastDuration))))
	util.WriteString(&builder, statsStyle.Render(fmt.Sprintf("    Average:  %s (%d runs)\n", formatDuration(stats.AvgDuration), stats.ExecutionCount)))

	return builder.String()
}

// getVariablesForTarget returns variables used by a specific target
func (m Model) getVariablesForTarget(targetName string) []string {
	var result []string

	for _, variable := range m.Variables {
		for _, usedTarget := range variable.UsedByTargets {
			if usedTarget == targetName {
				// Format: NAME = value
				result = append(result, fmt.Sprintf("%s = %s", variable.Name, variable.ExpandedValue))
				break
			}
		}
	}

	return result
}

// renderVariablesSection renders the variables used by a target
func renderVariablesSection(vars []string) string {
	if len(vars) == 0 {
		return ""
	}

	var builder strings.Builder

	// Separator
	separator := lipgloss.NewStyle().
		Foreground(MutedColor).
		Render("  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	util.WriteString(&builder, separator+"\n\n")

	// Header
	header := lipgloss.NewStyle().
		Foreground(SecondaryColor).
		Bold(true).
		Render("  📦 Variables Used")
	util.WriteString(&builder, header+"\n\n")

	// List variables (max 5, show "and N more")
	displayCount := min(len(vars), 5)
	varStyle := lipgloss.NewStyle().Foreground(TextColor)

	for i := 0; i < displayCount; i++ {
		util.WriteString(&builder, varStyle.Render("    "+vars[i])+"\n")
	}

	if len(vars) > 5 {
		moreStyle := lipgloss.NewStyle().
			Foreground(MutedColor).
			Italic(true)
		util.WriteString(&builder, moreStyle.Render(fmt.Sprintf("    ... and %d more\n", len(vars)-5)))
	}

	// Hint to view all
	util.WriteString(&builder, "\n")
	hintStyle := lipgloss.NewStyle().
		Foreground(SecondaryColor).
		Italic(true)
	util.WriteString(&builder, hintStyle.Render("    💡 Press 'v' to view all variables")+"\n")

	return builder.String()
}
