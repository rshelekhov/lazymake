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

	// Title with modern icon
	icon := lipgloss.NewStyle().
		Foreground(ErrorColor).
		Bold(true).
		Render(IconDangerCritical)

	titleBadge := StatusPill("error")

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(TextPrimary).
		Render(icon + " DANGEROUS COMMAND " + titleBadge)
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

			// Rule header with modern icon and badge
			var severityStr string
			var severityIcon string
			var severityColor lipgloss.AdaptiveColor

			switch match.Severity {
			case safety.SeverityCritical:
				severityIcon = IconDangerCritical
				severityStr = "critical"
				severityColor = ErrorColor
			case safety.SeverityWarning:
				severityIcon = IconDangerWarning
				severityStr = "warning"
				severityColor = WarningColor
			case safety.SeverityInfo:
				severityIcon = IconInfo
				severityStr = "info"
				severityColor = SecondaryColor
			}

			icon := lipgloss.NewStyle().
				Foreground(severityColor).
				Bold(true).
				Render(severityIcon)

			badge := StatusPill(severityStr)

			ruleHeader := icon + " " + badge + " " +
				lipgloss.NewStyle().
					Foreground(TextSecondary).
					Render(match.Rule.ID)
			util.WriteString(&builder, ruleHeader+"\n")

			// Matched command
			if match.MatchedLine != "" {
				matchedStyle := lipgloss.NewStyle().
					Foreground(TextMuted).
					Render("Command: " + match.MatchedLine)
				util.WriteString(&builder, matchedStyle+"\n")
			}

			// Description
			if match.Rule.Description != "" {
				descStyle := lipgloss.NewStyle().
					Foreground(TextSecondary)
				util.WriteString(&builder, "\n"+descStyle.Render(match.Rule.Description)+"\n")
			}

			// Suggestion
			if match.Rule.Suggestion != "" {
				suggestionStyle := lipgloss.NewStyle().
					Foreground(SecondaryColor).
					Italic(true)
				util.WriteString(&builder, "\n"+suggestionStyle.Render(IconInfo+" "+match.Rule.Suggestion)+"\n")
			}
		}
	}

	util.WriteString(&builder, "\n")

	// Actions
	actionsStyle := lipgloss.NewStyle().
		Foreground(TextSecondary).
		Align(lipgloss.Center)

	enterAction := lipgloss.NewStyle().
		Foreground(ErrorColor).
		Bold(true).
		Render("[Enter]")

	escAction := lipgloss.NewStyle().
		Foreground(SuccessColor).
		Bold(true).
		Render("[Esc]")

	actions := actionsStyle.Render(
		enterAction + " Continue Anyway     " + escAction + " Cancel (Recommended)",
	)
	util.WriteString(&builder, actions)

	// Calculate dialog dimensions
	contentWidth := min(80, m.Width-10)
	contentHeight := 0 // Auto-height

	// Wrap in prominent border with subtle background
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(ErrorColor).
		Background(BackgroundSubtle).
		Padding(3, 4).
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
