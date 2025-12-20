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

	var builder strings.Builder

	// Title
	title := TitleStyle.Render("Variable Inspector")
	util.WriteString(&builder, title+"\n\n")

	// Stats line with badges
	totalVars := len(m.Variables)
	usedVars := countUsedVariables(m.Variables)
	unusedVars := totalVars - usedVars

	statsStyle := lipgloss.NewStyle().
		Foreground(TextSecondary).
		Padding(0, 2)

	totalBadge := Badge(fmt.Sprintf("%d", totalVars), TextPrimary, BackgroundSubtle)
	usedBadge := Badge(fmt.Sprintf("%d used", usedVars), lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"}, SuccessColor)
	unusedBadge := Badge(fmt.Sprintf("%d unused", unusedVars), TextSecondary, BackgroundSubtle)

	stats := totalBadge + "  " + usedBadge + "  " + unusedBadge
	util.WriteString(&builder, statsStyle.Render(stats)+"\n\n")

	if totalVars == 0 {
		// No variables found
		emptyStyle := lipgloss.NewStyle().
			Foreground(TextMuted).
			Italic(true).
			Padding(0, 2)
		util.WriteString(&builder, emptyStyle.Render("No variables found in Makefile")+"\n\n")
	} else {
		// Render each variable
		// Calculate how many variables can fit on screen
		// Each variable block takes approximately 5-6 lines
		// Leave space for title (2), stats (2), status bar (3) = 7 lines
		availableHeight := m.Height - 7
		varsPerScreen := max(1, availableHeight/6)

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

		// Show scroll indicator if needed
		if totalVars > varsPerScreen {
			scrollInfo := fmt.Sprintf("  Showing %d-%d of %d", startIdx+1, endIdx, totalVars)
			scrollStyle := lipgloss.NewStyle().
				Foreground(TextMuted).
				Italic(true)
			util.WriteString(&builder, "\n"+scrollStyle.Render(scrollInfo)+"\n")
		}
	}

	// Status bar with keyboard shortcuts
	statusBar := renderStatusBar(m.Width, fmt.Sprintf("%d variables", totalVars), "v/esc: return • ↑↓/jk: navigate • q: quit")
	util.WriteString(&builder, "\n"+statusBar)

	return builder.String()
}

// renderVariableBlock renders a single variable's information
func renderVariableBlock(v variables.Variable, selected bool) string {
	var builder strings.Builder

	// Variable name with type badge
	nameStyle := lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true)

	// Different style for selected variable
	if selected {
		nameStyle = nameStyle.
			Background(PrimaryColor).
			Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#000000"})
	}

	// Type badge using Badge component
	typeBadge := ""
	if v.Type.Symbol() != "" {
		typeBadge = Badge(v.Type.Symbol(), TextSecondary, BackgroundSubtle) + " "
	}

	typeLabel := lipgloss.NewStyle().
		Foreground(TextMuted).
		Render(v.Type.String())

	varHeader := fmt.Sprintf("  %s  %s%s",
		nameStyle.Render(v.Name),
		typeBadge,
		typeLabel,
	)
	util.WriteString(&builder, varHeader+"\n")

	// Indented content style
	contentStyle := lipgloss.NewStyle().
		Foreground(TextSecondary).
		Padding(0, 4)

	// Raw value
	if v.RawValue != "" {
		rawLine := fmt.Sprintf("Raw:      %s", truncateValue(v.RawValue, 80))
		util.WriteString(&builder, contentStyle.Render(rawLine)+"\n")
	}

	// Expanded value (only if different from raw)
	if v.ExpandedValue != "" && v.ExpandedValue != v.RawValue {
		expandedLine := fmt.Sprintf("Expanded: %s", truncateValue(v.ExpandedValue, 80))
		expandedStyle := lipgloss.NewStyle().
			Foreground(SuccessColor).
			Padding(0, 4)
		util.WriteString(&builder, expandedStyle.Render(expandedLine)+"\n")
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

		usageStyle := lipgloss.NewStyle().
			Foreground(SecondaryColor).
			Padding(0, 4)
		util.WriteString(&builder, usageStyle.Render(usageLine)+"\n")
	} else {
		unusedStyle := lipgloss.NewStyle().
			Foreground(TextMuted).
			Italic(true).
			Padding(0, 4)
		util.WriteString(&builder, unusedStyle.Render("Not used by any target")+"\n")
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
