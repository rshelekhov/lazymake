package tui

import (
	"fmt"
	"os"
	"path/filepath"
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

	// Calculate left column width (35% of terminal width, minimum 35 chars)
	// Increased from 30% for better balance and more room for target names
	leftWidthPercent := 0.35
	minLeftWidth := 35

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
	// Add 1-char gap between columns for breathing room
	rightWidth := max(m.Width-actualLeftWidth-1, 10)

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
	// Border adds 2 to height and width, padding adds more
	// Increased padding from (1, 2) implicit to (2, 3) for breathing room
	contentWidth := width - 8  // 2 (border) + 6 (padding 3*2)
	contentHeight := height - 6 // 2 (border) + 4 (padding 2*2)

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

	// Apply modern border with increased padding and margin
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BorderColor).
		Padding(2, 3).
		Margin(0, 0) // Left margin handled by join

	return containerStyle.Render(placedContent)
}

// renderRecipePreview renders the right column with recipe and safety info
func (m Model) renderRecipePreview(target *Target, width, height int) string {
	if target == nil {
		return renderEmptyPreview(width, height)
	}

	var builder strings.Builder

	// Target name header with bottom border
	contentWidth := width - 12 // Account for container padding and border
	header := lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(BorderColor).
		PaddingBottom(1).
		Width(contentWidth).
		Render(target.Name)
	util.WriteString(&builder, header+"\n\n")

	// Recipe section with label
	if len(target.Recipe) > 0 {
		// Section label
		recipeLabel := lipgloss.NewStyle().
			Foreground(TextSecondary).
			Bold(true).
			Render("Recipe:")
		util.WriteString(&builder, recipeLabel+"\n\n")

		// Detect language for syntax highlighting
		language := m.Highlighter.DetectLanguage(target.Recipe, target.LanguageOverride)

		// Highlight each line with subtle background
		for _, line := range target.Recipe {
			highlighted := m.Highlighter.HighlightLine(line, language)
			codeLine := lipgloss.NewStyle().
				Background(BackgroundSubtle).
				Padding(0, 1).
				Render("  " + highlighted)
			util.WriteString(&builder, codeLine+"\n")
		}

		// Show language badge for non-bash languages
		if language != "bash" && language != "" {
			langBadge := Badge(language, TextSecondary, BackgroundSubtle)
			util.WriteString(&builder, "\n  "+langBadge+"\n")
		}

		// Graph view hint with icon
		util.WriteString(&builder, "\n")
		hintStyle := lipgloss.NewStyle().
			Foreground(TextMuted).
			Italic(true)
		util.WriteString(&builder, hintStyle.Render("  "+IconInfo+" Press 'g' to view dependency graph")+"\n")
	} else {
		noRecipeStyle := lipgloss.NewStyle().
			Foreground(TextMuted).
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

	// Padding(2,3) = 4 vertical + 6 horizontal
	// Border = 2 vertical + 2 horizontal
	// Total overhead: 6 vertical, 8 horizontal
	containerContentWidth := width - 10
	containerContentHeight := height - 8

	// Use lipgloss.Place to force content into exact dimensions
	placedContent := lipgloss.Place(
		containerContentWidth,
		containerContentHeight,
		lipgloss.Left,
		lipgloss.Top,
		builder.String(),
	)

	// Apply modern border with subtle background and increased padding
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BorderColor).
		Background(BackgroundSubtle).
		Padding(2, 3).
		Margin(0, 0)

	return containerStyle.Render(placedContent)
}

// renderSafetyWarnings renders safety match information
func renderSafetyWarnings(matches []safety.MatchResult) string {
	var builder strings.Builder

	for i, match := range matches {
		if i > 0 {
			util.WriteString(&builder, "\n")
		}

		// Severity indicator and rule ID with modern icons
		var severityStr string
		var severityIcon string
		var severityColor lipgloss.AdaptiveColor

		switch match.Severity {
		case safety.SeverityCritical:
			severityIcon = IconDangerCritical
			severityStr = "CRITICAL"
			severityColor = ErrorColor
		case safety.SeverityWarning:
			severityIcon = IconDangerWarning
			severityStr = "WARNING"
			severityColor = WarningColor
		case safety.SeverityInfo:
			severityIcon = IconInfo
			severityStr = "INFO"
			severityColor = SecondaryColor
		}

		icon := lipgloss.NewStyle().
			Foreground(severityColor).
			Bold(true).
			Render(severityIcon)

		severityBadge := StatusPill(strings.ToLower(severityStr))

		severityHeader := icon + " " + severityBadge + " " +
			lipgloss.NewStyle().
				Foreground(TextSecondary).
				Render(match.Rule.ID)
		util.WriteString(&builder, "  "+severityHeader+"\n")

		// Matched line
		if match.MatchedLine != "" {
			matchedStyle := lipgloss.NewStyle().
				Foreground(TextMuted).
				Render("    Matched: " + match.MatchedLine)
			util.WriteString(&builder, matchedStyle+"\n")
		}

		// Description
		if match.Rule.Description != "" {
			descStyle := lipgloss.NewStyle().
				Foreground(TextSecondary)
			util.WriteString(&builder, descStyle.Render("    "+match.Rule.Description)+"\n")
		}

		// Suggestion
		if match.Rule.Suggestion != "" {
			suggestionStyle := lipgloss.NewStyle().
				Foreground(SecondaryColor).
				Italic(true)
			util.WriteString(&builder, suggestionStyle.Render("    "+IconInfo+" "+match.Rule.Suggestion)+"\n")
		}
	}

	return builder.String()
}

// renderEmptyPreview shows placeholder when no target selected
func renderEmptyPreview(width, height int) string {
	emptyText := "Select a target to preview recipe"

	emptyStyle := lipgloss.NewStyle().
		Foreground(TextMuted).
		Italic(true)

	content := emptyStyle.Render(emptyText)

	// Same dimensions as recipe preview (with new padding)
	contentWidth := width - 10
	contentHeight := height - 8

	// Use lipgloss.Place to force content into exact dimensions
	// Center the text within the space
	placedContent := lipgloss.Place(
		contentWidth,
		contentHeight,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)

	// Wrap in modern border with subtle background
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BorderColor).
		Background(BackgroundSubtle).
		Padding(2, 3)

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

	// Left side: workspace path + stats
	workspacePath := m.getWorkspaceDisplayPath()
	leftStats := []string{workspacePath, fmt.Sprintf("%d targets", totalTargets)}

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

	// Modern separator with subtle color
	separator := lipgloss.NewStyle().
		Foreground(BorderColor).
		Render("  " + strings.Repeat("─", 50))
	util.WriteString(&builder, separator+"\n\n")

	// Header with icon and badge
	icon := lipgloss.NewStyle().
		Foreground(WarningColor).
		Bold(true).
		Render(IconRegression)

	badge := StatusPill("warning")

	header := lipgloss.NewStyle().
		Foreground(TextPrimary).
		Bold(true).
		Render("  " + icon + " Performance Regression " + badge)
	util.WriteString(&builder, header+"\n\n")

	// Stats in a subtle box
	statsContent := fmt.Sprintf(
		"Current:  %s\nAverage:  %s (%d runs)\nChange:   %s",
		formatDuration(stats.LastDuration),
		formatDuration(stats.AvgDuration),
		stats.ExecutionCount,
		lipgloss.NewStyle().Foreground(WarningColor).Bold(true).Render(fmt.Sprintf("+%d%% slower", change)),
	)

	statsBox := lipgloss.NewStyle().
		Foreground(TextSecondary).
		Background(BackgroundSubtle).
		Border(lipgloss.NormalBorder()).
		BorderForeground(BorderColor).
		Padding(1, 2).
		Render(statsContent)

	util.WriteString(&builder, "  "+statsBox+"\n\n")

	// Helpful hint
	hint := lipgloss.NewStyle().
		Foreground(TextMuted).
		Italic(true).
		Render("  " + IconInfo + " This target recently got slower - investigate recent changes")

	util.WriteString(&builder, hint)

	return builder.String()
}

// renderRecentTargetInfo renders performance info for recently executed targets
func renderRecentTargetInfo(target Target) string {
	stats := target.PerfStats

	var builder strings.Builder

	// Modern separator with subtle color
	separator := lipgloss.NewStyle().
		Foreground(BorderColor).
		Render("  " + strings.Repeat("─", 50))
	util.WriteString(&builder, separator+"\n\n")

	// Header with icon
	icon := lipgloss.NewStyle().
		Foreground(SecondaryColor).
		Render(IconRecent)

	header := lipgloss.NewStyle().
		Foreground(TextSecondary).
		Bold(true).
		Render("  " + icon + " Performance")
	util.WriteString(&builder, header+"\n\n")

	// Stats
	statsStyle := lipgloss.NewStyle().Foreground(TextSecondary)
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

// getWorkspaceDisplayPath returns the relative path to the current Makefile for display in status bar
func (m Model) getWorkspaceDisplayPath() string {
	if m.WorkspaceManager == nil {
		return filepath.Base(m.MakefilePath)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return filepath.Base(m.MakefilePath)
	}

	relPath := m.WorkspaceManager.GetRelativePath(m.MakefilePath, cwd)
	if relPath == "" {
		return filepath.Base(m.MakefilePath)
	}

	return relPath
}
