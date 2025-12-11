package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/rshelekhov/lazymake/internal/safety"
	"github.com/rshelekhov/lazymake/internal/util"
)

// renderConfirmView renders the dangerous command confirmation dialog
func (m Model) renderConfirmView() string {
	if m.PendingTarget == nil {
		return "Error: No pending target"
	}

	target := m.PendingTarget

	var builder strings.Builder

	// Title with danger emoji
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(ErrorColor).
		Render("🚨 DANGEROUS COMMAND WARNING")
	util.WriteString(&builder, title+"\n\n")

	// Target name
	targetLine := lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true).
		Render("Target: " + target.Name)
	util.WriteString(&builder, targetLine+"\n\n")

	// Show all safety matches
	if len(target.SafetyMatches) > 0 {
		for i, match := range target.SafetyMatches {
			if i > 0 {
				util.WriteString(&builder, "\n")
			}

			// Rule header with severity
			var severityStr string
			var severityColor lipgloss.Color

			switch match.Severity {
			case safety.SeverityCritical:
				severityStr = "CRITICAL"
				severityColor = ErrorColor
			case safety.SeverityWarning:
				severityStr = "WARNING"
				severityColor = lipgloss.Color("#FFA500")
			case safety.SeverityInfo:
				severityStr = "INFO"
				severityColor = SecondaryColor
			}

			ruleHeader := lipgloss.NewStyle().
				Foreground(severityColor).
				Bold(true).
				Render(severityStr + ": " + match.Rule.ID)
			util.WriteString(&builder, ruleHeader+"\n")

			// Matched command
			if match.MatchedLine != "" {
				matchedStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("#666666")).
					Render("Command: " + match.MatchedLine)
				util.WriteString(&builder, matchedStyle+"\n")
			}

			// Description
			if match.Rule.Description != "" {
				descStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("#CCCCCC"))
				util.WriteString(&builder, "\n"+descStyle.Render(match.Rule.Description)+"\n")
			}

			// Suggestion
			if match.Rule.Suggestion != "" {
				suggestionStyle := lipgloss.NewStyle().
					Foreground(SecondaryColor).
					Italic(true)
				util.WriteString(&builder, "\n"+suggestionStyle.Render("💡 "+match.Rule.Suggestion)+"\n")
			}
		}
	}

	util.WriteString(&builder, "\n")

	// Actions
	actionsStyle := lipgloss.NewStyle().
		Foreground(MutedColor).
		Align(lipgloss.Center)

	enterAction := lipgloss.NewStyle().
		Foreground(ErrorColor).
		Bold(true).
		Render("[Enter]")

	escAction := lipgloss.NewStyle().
		Foreground(SecondaryColor).
		Bold(true).
		Render("[Esc]")

	actions := actionsStyle.Render(
		enterAction + " Continue Anyway     " + escAction + " Cancel",
	)
	util.WriteString(&builder, actions)

	// Calculate dialog dimensions
	contentWidth := min(80, m.Width-10)
	contentHeight := 0 // Auto-height

	// Wrap in prominent border
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(ErrorColor).
		Padding(2, 4).
		Width(contentWidth).
		Height(contentHeight).
		Align(lipgloss.Center)

	dialog := dialogStyle.Render(builder.String())

	// Center the dialog on screen
	verticalPadding := max((m.Height-strings.Count(dialog, "\n"))/2, 0)
	paddingStyle := lipgloss.NewStyle().
		PaddingTop(verticalPadding).
		PaddingLeft((m.Width - contentWidth) / 2)

	return paddingStyle.Render(dialog)
}
