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

	// Calculate available space
	// Status bar takes 3 lines (border 2 + content 1)
	statusBarHeight := 3
	availableHeight := m.Height - statusBarHeight

	// Build main content
	content := m.buildVariablesContent(availableHeight)

	// Wrap content in bordered container
	contentWidth := m.Width - 2 // Account for border (2)
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BorderColor).
		Padding(2, 3).
		Width(contentWidth)

	borderedContent := containerStyle.Render(content)

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
func (m Model) buildVariablesContent(availableHeight int) string {
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
		// Render each variable
		// Calculate how many variables can fit on screen
		// Each variable block takes approximately 5-6 lines
		// Leave space for title (2), stats (2), border/padding (6) = 10 lines
		contentHeight := availableHeight - 10
		varsPerScreen := max(1, contentHeight/6)

		// Calculate scroll offset to keep selected variable visible
		startIdx := 0
		if m.VariableListIndex >= varsPerScreen {
			startIdx = m.VariableListIndex - varsPerScreen + 1
		}

		endIdx := min(startIdx+varsPerScreen, totalVars)

		for i := startIdx; i < endIdx; i++ {
			variable := m.Variables[i]
			selected := i == m.VariableListIndex
			varBlock := renderVariableBlock(variable, selected)
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

	// Base status bar style
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

	// Total variables nugget (only colored one)
	totalNugget := coloredNuggetStyle.Render(fmt.Sprintf("%d variables", totalVars))

	var sections []string
	sections = append(sections, totalNugget)

	// Used variables (plain text on status bar background)
	if usedVars > 0 {
		usedInfo := plainNuggetStyle.Render(fmt.Sprintf("%d used", usedVars))
		sections = append(sections, usedInfo)
	}

	// Unused variables (plain text on status bar background)
	if unusedVars > 0 {
		unusedInfo := plainNuggetStyle.Render(fmt.Sprintf("%d unused", unusedVars))
		sections = append(sections, unusedInfo)
	}

	leftBar := lipgloss.JoinHorizontal(lipgloss.Top, sections...)
	leftWidth := lipgloss.Width(leftBar)

	// Help text on the right
	helpText := "v/esc: return • ↑↓/jk: navigate • q: quit"

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

// renderVariableBlock renders a single variable's information
func renderVariableBlock(v variables.Variable, selected bool) string {
	var builder strings.Builder

	// Select appropriate styles based on selection state
	var titleStyle, contentStyle lipgloss.Style

	if selected {
		// Selected item with vertical border (matching target list)
		titleStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderLeft(true).
			BorderForeground(PrimaryColor).
			PaddingLeft(1)

		contentStyle = lipgloss.NewStyle().
			Foreground(TextSecondary).
			BorderStyle(lipgloss.NormalBorder()).
			BorderLeft(true).
			BorderForeground(PrimaryColor).
			PaddingLeft(1)
	} else {
		// Normal item without border
		titleStyle = lipgloss.NewStyle().
			Foreground(TextPrimary).
			Bold(true).
			PaddingLeft(2)

		contentStyle = lipgloss.NewStyle().
			Foreground(TextSecondary).
			PaddingLeft(2)
	}

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

	// Build title parts separately to preserve individual colors
	// Use TextPrimary (white) for variable name, like target names in main view
	var nameColor lipgloss.AdaptiveColor
	if selected {
		nameColor = PrimaryColor
	} else {
		nameColor = TextPrimary
	}
	nameStyled := lipgloss.NewStyle().Foreground(nameColor).Bold(true).Render(v.Name)
	varHeader := nameStyled + "  " + typeBadge + typeLabel

	// Apply title style (border will be added if selected)
	util.WriteString(&builder, titleStyle.Render(varHeader)+"\n")

	// Detail lines with matching border
	var details []string

	// Raw value
	if v.RawValue != "" {
		rawLine := fmt.Sprintf("Raw:      %s", truncateValue(v.RawValue, 80))
		details = append(details, contentStyle.Render(rawLine))
	}

	// Expanded value (only if different from raw)
	if v.ExpandedValue != "" && v.ExpandedValue != v.RawValue {
		expandedLine := fmt.Sprintf("Expanded: %s", truncateValue(v.ExpandedValue, 80))
		// Create style with success color but same border as content
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

		// Keep the same gray color as other details (TextSecondary from contentStyle)
		details = append(details, contentStyle.Render(usageLine))
	} else {
		unusedStyle := contentStyle.Foreground(TextMuted).Italic(true)
		details = append(details, unusedStyle.Render("Not used by any target"))
	}

	// Render all detail lines
	for _, detail := range details {
		util.WriteString(&builder, detail+"\n")
	}

	util.WriteString(&builder, "\n")

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
