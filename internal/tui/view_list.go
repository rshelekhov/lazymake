package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
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
	contentWidth := width - 8   // 2 (border) + 6 (padding 3*2)
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

// buildRecipeContent generates the full text content for the recipe preview
// This is separate from rendering so we can set viewport content in the update cycle
func (m Model) buildRecipeContent(target *Target, width int) string {
	var builder strings.Builder

	// Target name header with bottom border
	contentWidth := width - 8 // Match the viewport content width
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

		// Highlight each line
		for _, line := range target.Recipe {
			highlighted := m.Highlighter.HighlightLine(line, language)
			util.WriteString(&builder, highlighted+"\n")
		}

		// Show language badge for non-bash languages
		if language != "bash" && language != "" {
			langBadge := lipgloss.NewStyle().
				Foreground(TextSecondary).
				Render(language)
			util.WriteString(&builder, "\n"+langBadge+"\n")
		}

		// Graph view hint with icon
		util.WriteString(&builder, "\n")
		hintStyle := lipgloss.NewStyle().
			Foreground(TextMuted).
			Italic(true)
		util.WriteString(&builder, hintStyle.Render(IconInfo+" Press 'g' to view dependency graph")+"\n")
	} else {
		noRecipeStyle := lipgloss.NewStyle().
			Foreground(TextMuted).
			Italic(true)
		util.WriteString(&builder, noRecipeStyle.Render("(no recipe - meta target)")+"\n")
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

	return builder.String()
}

// renderRecipePreview renders the right column with recipe and safety info
// The viewport content is set in the Update function, this just renders it
func (m Model) renderRecipePreview(target *Target, width, height int) string {
	if target == nil {
		return renderEmptyPreview(width, height)
	}

	// Render viewport (content is set in Update function)
	viewportContent := m.RecipeViewport.View()

	// Calculate exact dimensions matching left column
	contentWidth := width - 8   // Match left column: 6 (padding) + 2 (border)
	contentHeight := height - 6 // Match left column: 4 (padding) + 2 (border)

	// Force viewport content to exact dimensions (same as left column)
	// This ensures both columns have identical heights
	viewportContent = lipgloss.Place(
		contentWidth,
		contentHeight,
		lipgloss.Left,
		lipgloss.Top,
		viewportContent,
	)

	// Overlay scroll percentage indicator at bottom-right if content is scrollable
	if m.RecipeViewport.TotalLineCount() > m.RecipeViewport.VisibleLineCount() {
		scrollPercent := int(m.RecipeViewport.ScrollPercent() * 100)

		// Create compact scroll indicator
		scrollIndicator := lipgloss.NewStyle().
			Foreground(TextMuted).
			Padding(0, 1).
			Render(fmt.Sprintf("%d%%", scrollPercent))

		// Place indicator overlay at bottom-right
		indicatorOverlay := lipgloss.Place(
			contentWidth,
			contentHeight,
			lipgloss.Right,
			lipgloss.Bottom,
			scrollIndicator,
		)

		// Combine content and indicator using JoinHorizontal at each line
		// This overlays the indicator on the last line
		contentLines := strings.Split(viewportContent, "\n")
		indicatorLines := strings.Split(indicatorOverlay, "\n")

		// Replace the last line with the indicator overlay
		if len(contentLines) == len(indicatorLines) && len(contentLines) > 0 {
			contentLines[len(contentLines)-1] = indicatorLines[len(indicatorLines)-1]
			viewportContent = strings.Join(contentLines, "\n")
		}
	}

	// Apply modern border with increased padding
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BorderColor).
		Padding(2, 3).
		Margin(0, 0)

	return containerStyle.Render(viewportContent)
}

// renderSafetyWarnings renders safety match information
func renderSafetyWarnings(matches []safety.MatchResult) string {
	var builder strings.Builder

	for i, match := range matches {
		if i > 0 {
			util.WriteString(&builder, "\n")
		}

		// Separator before each warning
		separator := lipgloss.NewStyle().
			Foreground(BorderColor).
			Render(strings.Repeat("─", 50))
		util.WriteString(&builder, separator+"\n\n")

		// Severity indicator and rule ID with modern icons
		var severityStr string
		var severityIcon string
		var severityColor lipgloss.AdaptiveColor

		switch match.Severity {
		case safety.SeverityCritical:
			severityIcon = "○" // Empty circle (red)
			severityStr = "Critical"
			severityColor = ErrorColor
		case safety.SeverityWarning:
			severityIcon = "○" // Empty circle (yellow)
			severityStr = "Warning"
			severityColor = WarningColor
		case safety.SeverityInfo:
			severityIcon = "○" // Empty circle (blue)
			severityStr = "Info"
			severityColor = SecondaryColor
		}

		icon := lipgloss.NewStyle().
			Foreground(severityColor).
			Bold(true).
			Render(severityIcon)

		// Use text without background (no badge) for all severity levels
		severityBadge := lipgloss.NewStyle().
			Foreground(severityColor).
			Bold(true).
			Render(severityStr)

		// Build box content with word wrapping
		// Max width for text (accounting for border and padding)
		maxWidth := 70
		var boxContent strings.Builder

		// Header (inside box now)
		header := icon + " " + severityBadge + " " + lipgloss.NewStyle().
			Foreground(TextSecondary).
			Render(match.Rule.ID)
		util.WriteString(&boxContent, header+"\n")

		// Matched line
		if match.MatchedLine != "" {
			util.WriteString(&boxContent, "\n")
			matchedLine := fmt.Sprintf("Matched: %s", match.MatchedLine)
			wrappedMatched := wordwrap.String(matchedLine, maxWidth)
			util.WriteString(&boxContent, wrappedMatched+"\n")
		}

		// Description
		if match.Rule.Description != "" {
			util.WriteString(&boxContent, "\n")
			wrappedDesc := wordwrap.String(match.Rule.Description, maxWidth)
			util.WriteString(&boxContent, wrappedDesc)
		}

		// Suggestion (inside the box now)
		if match.Rule.Suggestion != "" {
			util.WriteString(&boxContent, "\n\n")
			suggestionLine := IconInfo + " " + match.Rule.Suggestion
			wrappedSuggestion := wordwrap.String(suggestionLine, maxWidth)
			suggestionText := lipgloss.NewStyle().
				Foreground(TextMuted).
				Italic(true).
				Render(wrappedSuggestion)
			util.WriteString(&boxContent, suggestionText)
		}

		// Render box with border
		safetyBox := lipgloss.NewStyle().
			Foreground(TextSecondary).
			Border(lipgloss.NormalBorder()).
			BorderForeground(BorderColor).
			Padding(1, 2).
			Render(boxContent.String())

		util.WriteString(&builder, safetyBox+"\n")
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

	// Wrap in modern border
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BorderColor).
		Padding(2, 3)

	return borderStyle.Render(placedContent)
}

// renderStatusBar renders the bottom status bar with colored nuggets (lipgloss-style)
func (m Model) renderStatusBar() string {
	// Count stats - use maps to track unique targets by name
	totalTargets := 0
	dangerousTargets := make(map[string]bool)
	criticalTargets := make(map[string]bool)
	regressedTargets := make(map[string]bool)

	for _, item := range m.List.Items() {
		if target, ok := item.(Target); ok {
			totalTargets++
			if target.IsDangerous {
				if target.DangerLevel == safety.SeverityCritical {
					criticalTargets[target.Name] = true
				} else {
					dangerousTargets[target.Name] = true
				}
			}
			if target.PerfStats != nil && target.PerfStats.IsRegressed {
				regressedTargets[target.Name] = true
			}
		}
	}

	// Get unique counts
	dangerousCount := len(dangerousTargets)
	criticalCount := len(criticalTargets)
	regressedCount := len(regressedTargets)

	// Base status bar style - with background for entire bar
	statusBarStyle := lipgloss.NewStyle().
		Foreground(TextPrimary)

	// Colored nugget style (only for first item)
	coloredNuggetStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#000000"}).
		Background(PrimaryColor).
		Padding(0, 1).
		MarginRight(1)

	// Plain nugget style (inherits status bar background, just text)
	plainNuggetStyle := lipgloss.NewStyle().
		Foreground(TextSecondary).
		Padding(0, 1)

	// Yellow text style for regressed items
	yellowNuggetStyle := lipgloss.NewStyle().
		Foreground(WarningColor).
		Padding(0, 1)

	// Workspace path nugget (only colored one)
	workspacePath := m.getWorkspaceDisplayPath()
	pathNugget := coloredNuggetStyle.Render(workspacePath)

	var sections []string
	sections = append(sections, pathNugget)

	// Target count (plain text on status bar background)
	targetInfo := plainNuggetStyle.Render(fmt.Sprintf("%d targets", totalTargets))
	sections = append(sections, targetInfo)

	// Dangerous count - only if there are dangerous targets
	if dangerousCount > 0 {
		dangerIcon := lipgloss.NewStyle().Foreground(WarningColor).Render("○")
		dangerInfo := plainNuggetStyle.Render(fmt.Sprintf("%s %d dangerous", dangerIcon, dangerousCount))
		sections = append(sections, dangerInfo)
	}

	// Critical count - only if there are critical targets
	if criticalCount > 0 {
		criticalIcon := lipgloss.NewStyle().Foreground(ErrorColor).Render("○")
		criticalInfo := plainNuggetStyle.Render(fmt.Sprintf("%s %d critical", criticalIcon, criticalCount))
		sections = append(sections, criticalInfo)
	}

	// Regressed count with yellow text - only if there are regressed targets
	if regressedCount > 0 {
		regressedInfo := yellowNuggetStyle.Render(fmt.Sprintf("%d regressed", regressedCount))
		sections = append(sections, regressedInfo)
	}

	// Calculate width used by nuggets
	leftBar := lipgloss.JoinHorizontal(lipgloss.Top, sections...)
	leftWidth := lipgloss.Width(leftBar)

	// Right side: shortcuts
	var helpText string
	if item := m.List.SelectedItem(); item != nil {
		if target, ok := item.(Target); ok && target.IsDangerous {
			switch target.DangerLevel {
			case safety.SeverityCritical:
				criticalIcon := lipgloss.NewStyle().Foreground(ErrorColor).Render("○")
				helpText = criticalIcon + " Critical • enter: confirm • esc: cancel • q: quit"
			case safety.SeverityWarning:
				warningIcon := lipgloss.NewStyle().Foreground(WarningColor).Render("○")
				helpText = warningIcon + " Warning • enter: run • esc: cancel • q: quit"
			case safety.SeverityInfo:
				infoIcon := lipgloss.NewStyle().Foreground(SecondaryColor).Render("○")
				helpText = infoIcon + " Info • enter: run • esc: cancel • q: quit"
			}
		} else {
			helpText = formatKeyBindings(m.KeyBindings)
		}
	} else {
		helpText = formatKeyBindings(m.KeyBindings)
	}

	// Right section with help text
	right := lipgloss.NewStyle().
		Foreground(TextMuted).
		Padding(0, 1).
		Render(helpText)
	rightWidth := lipgloss.Width(right)

	// Middle section fills remaining space
	// Account for status bar horizontal padding (2 chars: 1 left + 1 right)
	middleWidth := max(m.Width-2-leftWidth-rightWidth, 1)
	middle := lipgloss.NewStyle().
		Width(middleWidth).
		Align(lipgloss.Left).
		Render("")

	// Combine all sections
	bar := lipgloss.JoinHorizontal(lipgloss.Top, leftBar, middle, right)

	return statusBarStyle.
		Width(m.Width).
		Padding(1, 1).
		Render(bar)
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
		Render(strings.Repeat("─", 50))
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
		Render(icon + " Performance Regression " + badge)
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
		Border(lipgloss.NormalBorder()).
		BorderForeground(BorderColor).
		Padding(1, 2).
		Render(statsContent)

	util.WriteString(&builder, statsBox+"\n\n")

	// Helpful hint
	hint := lipgloss.NewStyle().
		Foreground(TextMuted).
		Italic(true).
		Render(IconInfo + " This target recently got slower - investigate recent changes")

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
		Render(strings.Repeat("─", 50))
	util.WriteString(&builder, separator+"\n\n")

	// Header with icon
	icon := lipgloss.NewStyle().
		Foreground(SecondaryColor).
		Render(IconRecent)

	header := lipgloss.NewStyle().
		Foreground(TextSecondary).
		Bold(true).
		Render(icon + " Performance")
	util.WriteString(&builder, header+"\n\n")

	// Stats
	statsStyle := lipgloss.NewStyle().Foreground(TextSecondary)
	util.WriteString(&builder, statsStyle.Render(fmt.Sprintf("Last run: %s\n", formatDuration(stats.LastDuration))))
	util.WriteString(&builder, statsStyle.Render(fmt.Sprintf("Average:  %s (%d runs)\n", formatDuration(stats.AvgDuration), stats.ExecutionCount)))

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
		Foreground(BorderColor).
		Render(strings.Repeat("─", 50))
	util.WriteString(&builder, separator+"\n\n")

	// Header
	header := lipgloss.NewStyle().
		Foreground(SecondaryColor).
		Bold(true).
		Render("Variables Used")
	util.WriteString(&builder, header+"\n\n")

	// List variables (max 5, show "and N more")
	displayCount := min(len(vars), 5)
	varStyle := lipgloss.NewStyle().Foreground(TextPrimary)

	for i := 0; i < displayCount; i++ {
		util.WriteString(&builder, varStyle.Render(vars[i])+"\n")
	}

	if len(vars) > 5 {
		moreStyle := lipgloss.NewStyle().
			Foreground(TextMuted).
			Italic(true)
		util.WriteString(&builder, moreStyle.Render(fmt.Sprintf("... and %d more\n", len(vars)-5)))
	}

	// Hint to view all
	util.WriteString(&builder, "\n")
	hintStyle := lipgloss.NewStyle().
		Foreground(TextMuted).
		Italic(true)
	util.WriteString(&builder, hintStyle.Render(IconInfo+" Press 'v' to view all variables")+"\n")

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
