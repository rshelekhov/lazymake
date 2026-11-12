package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/rshelekhov/lazymake/internal/util"
	"github.com/rshelekhov/lazymake/internal/variables"
)

// renderVariablesView displays the full-screen variable inspector
func (m Model) renderVariablesView() string {
	if m.Width == 0 || m.Height == 0 {
		return "Loading variable inspector..."
	}

	// Render viewport content
	viewportContent := m.VariablesViewport.View()

	// Calculate dimensions
	statusBarHeight := 3
	availableHeight := m.Height - statusBarHeight
	contentWidth := m.Width - 8
	contentHeight := availableHeight - 6

	// Force viewport content to exact dimensions
	viewportContent = lipgloss.Place(
		contentWidth,
		contentHeight,
		lipgloss.Left,
		lipgloss.Top,
		viewportContent,
	)

	// Overlay scroll percentage indicator at bottom-right if content is scrollable
	if m.VariablesViewport.TotalLineCount() > m.VariablesViewport.VisibleLineCount() {
		scrollPercent := int(m.VariablesViewport.ScrollPercent() * 100)

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

		// Combine content and indicator
		contentLines := strings.Split(viewportContent, "\n")
		indicatorLines := strings.Split(indicatorOverlay, "\n")

		if len(contentLines) == len(indicatorLines) && len(contentLines) > 0 {
			contentLines[len(contentLines)-1] = indicatorLines[len(indicatorLines)-1]
			viewportContent = strings.Join(contentLines, "\n")
		}
	}

	// Wrap content in bordered container
	contentWidth = m.Width - 2 // Account for border (2)
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BorderColor).
		Padding(2, 3).
		Width(contentWidth)

	borderedContent := containerStyle.Render(viewportContent)

	// Render status bar
	statusBar := m.renderVariablesStatusBar()

	// Combine content and status bar
	return lipgloss.JoinVertical(
		lipgloss.Left,
		borderedContent,
		statusBar,
	)
}

// buildVariablesContent builds the main content for the variables view
func (m Model) buildVariablesContent() string {
	var builder strings.Builder

	// Title
	title := TitleStyle.Render("Variable Inspector")
	util.WriteString(&builder, title+"\n\n")

	totalVars := len(m.Variables)

	if totalVars == 0 {
		// No variables found
		emptyStyle := lipgloss.NewStyle().
			Foreground(TextMuted).
			Italic(true)
		util.WriteString(&builder, emptyStyle.Render("No variables found in Makefile")+"\n\n")
	} else {
		// Render all variables (no selection/navigation)
		for i, variable := range m.Variables {
			if i > 0 {
				util.WriteString(&builder, "\n") // Separator between variables
			}
			varBlock := renderVariableBlock(variable)
			util.WriteString(&builder, varBlock)
		}
	}

	return builder.String()
}

// renderVariablesStatusBar renders the status bar for the variables view
func (m Model) renderVariablesStatusBar() string {
	totalVars := len(m.Variables)
	usedVars := countUsedVariables(m.Variables)
	unusedVars := totalVars - usedVars

	sections := []string{coloredNugget(fmt.Sprintf("%d variables", totalVars))}

	if usedVars > 0 {
		sections = append(sections, plainNugget(fmt.Sprintf("%d used", usedVars)))
	}
	if unusedVars > 0 {
		sections = append(sections, plainNugget(fmt.Sprintf("%d unused", unusedVars)))
	}

	leftBar := lipgloss.JoinHorizontal(lipgloss.Top, sections...)

	helpText := "v/esc: return • q: quit"
	if m.VariablesViewport.TotalLineCount() > m.VariablesViewport.VisibleLineCount() {
		helpText = "↑/↓: scroll • " + helpText
	}

	return assembleStatusBar(m.Width, leftBar, helpText)
}

// renderVariableBlock renders a single variable's information
func renderVariableBlock(v variables.Variable) string {
	var builder strings.Builder

	// Simple styles without selection (no border, no padding)
	titleStyle := lipgloss.NewStyle().
		Foreground(TextPrimary).
		Bold(true)

	contentStyle := lipgloss.NewStyle().
		Foreground(TextSecondary)

	// Type badge
	typeBadge := ""
	if v.Type.Symbol() != "" {
		typeBadge = lipgloss.NewStyle().
			Foreground(TextSecondary).
			Render(v.Type.Symbol()) + " "
	}

	typeLabel := lipgloss.NewStyle().
		Foreground(TextMuted).
		Render(v.Type.String())

	// Variable name in white (like target names)
	nameStyled := lipgloss.NewStyle().
		Foreground(TextPrimary).
		Bold(true).
		Render(v.Name)
	varHeader := nameStyled + " " + typeBadge + typeLabel

	// Render title
	util.WriteString(&builder, titleStyle.Render(varHeader)+"\n")

	// Detail lines
	var details []string

	// Raw value
	if v.RawValue != "" {
		rawLine := fmt.Sprintf("Raw:      %s", truncateValue(v.RawValue, 80))
		details = append(details, contentStyle.Render(rawLine))
	}

	// Expanded value (only if different from raw)
	if v.ExpandedValue != "" && v.ExpandedValue != v.RawValue {
		expandedLine := fmt.Sprintf("Expanded: %s", truncateValue(v.ExpandedValue, 80))
		// Use success color for expanded value
		expandedStyle := contentStyle.Foreground(SuccessColor)
		details = append(details, expandedStyle.Render(expandedLine))
	}

	// Usage information
	usageCount := len(v.UsedByTargets)
	if usageCount > 0 {
		// Show first few targets, then "and N more"
		displayTargets := v.UsedByTargets
		moreCount := 0
		if usageCount > 3 {
			displayTargets = v.UsedByTargets[:3]
			moreCount = usageCount - 3
		}

		usageLine := fmt.Sprintf("Used by:  %s", strings.Join(displayTargets, ", "))
		if moreCount > 0 {
			usageLine += fmt.Sprintf(" (and %d more)", moreCount)
		}
		usageLine += fmt.Sprintf(" (%d target%s)", usageCount, pluralize(usageCount))

		details = append(details, contentStyle.Render(usageLine))
	} else {
		unusedStyle := contentStyle.Foreground(TextMuted).Italic(true)
		details = append(details, unusedStyle.Render("Not used by any target"))
	}

	// Render all detail lines
	for _, detail := range details {
		util.WriteString(&builder, detail+"\n")
	}

	return builder.String()
}

// countUsedVariables counts how many variables are used by at least one target
func countUsedVariables(vars []variables.Variable) int {
	count := 0
	for _, v := range vars {
		if len(v.UsedByTargets) > 0 {
			count++
		}
	}
	return count
}

// truncateValue truncates a value string if it exceeds maxLen
func truncateValue(value string, maxLen int) string {
	if len(value) <= maxLen {
		return value
	}
	return value[:maxLen-3] + "..."
}

// pluralize returns "s" if count != 1, otherwise ""
func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}
