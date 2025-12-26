package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
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

	// Title without padding (like "RECENT" or "ALL TARGETS" but without left padding)
	title := lipgloss.NewStyle().
		Foreground(TextSecondary).
		Bold(true).
		Render("DANGEROUS COMMAND")
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

			// Use text without background (no badge) for severity level
			severityBadge := lipgloss.NewStyle().
				Foreground(severityColor).
				Bold(true).
				Render(severityStr)

			// Build box content with word wrapping
			// Max width for text (accounting for dialog border/padding + box border/padding)
			maxWidth := 60
			var boxContent strings.Builder

			// Header (inside box now)
			header := icon + " " + severityBadge + " " +
				lipgloss.NewStyle().
					Foreground(TextSecondary).
					Render(match.Rule.ID)
			util.WriteString(&boxContent, header+"\n")

			// Matched command
			if match.MatchedLine != "" {
				util.WriteString(&boxContent, "\n")
				matchedLine := fmt.Sprintf("Command: %s", match.MatchedLine)
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
	}

	util.WriteString(&builder, "\n")

	// Actions
	actionsStyle := lipgloss.NewStyle().
		Foreground(TextSecondary).
		Align(lipgloss.Left)

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

	// Wrap in prominent border without background
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(ErrorColor).
		Padding(2, 4).
		Width(contentWidth).
		Height(contentHeight).
		Align(lipgloss.Left)

	dialog := dialogStyle.Render(builder.String())

	// Center the dialog on screen
	verticalPadding := max((m.Height-strings.Count(dialog, "\n"))/2, 0)
	paddingStyle := lipgloss.NewStyle().
		PaddingTop(verticalPadding).
		PaddingLeft((m.Width - contentWidth) / 2)

	return paddingStyle.Render(dialog)
}
