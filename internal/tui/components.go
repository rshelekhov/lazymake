package tui

import (
	"time"

	"github.com/charmbracelet/lipgloss"
)

// ModernBorder creates a refined border for main containers
func ModernBorder(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BorderColor).
		Width(width).
		Padding(2, 3)
}

// AccentBorder creates a border for focused/important elements
func AccentBorder(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BorderAccent).
		Width(width).
		Padding(2, 3)
}

// SubtleBox creates a subtle box for secondary content
func SubtleBox(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(TextMuted).
		Foreground(TextSecondary).
		Width(width).
		Padding(0, 1)
}

// Badge creates a modern pill-style badge
func Badge(text string, fgColor, bgColor lipgloss.AdaptiveColor) string {
	return lipgloss.NewStyle().
		Foreground(fgColor).
		Background(bgColor).
		Padding(0, 1).
		Bold(true).
		Render(text)
}

// StatusPill creates a colored status indicator pill
func StatusPill(status string) string {
	var fg, bg lipgloss.AdaptiveColor

	switch status {
	case "success":
		fg = lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"}
		bg = SuccessColor
	case "error":
		fg = lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"}
		bg = ErrorColor
	case "warning":
		fg = lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#000000"}
		bg = WarningColor
	case "info":
		fg = lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#000000"}
		bg = PrimaryColor
	default:
		fg = TextPrimary
		bg = BackgroundSubtle
	}

	return Badge(status, fg, bg)
}

// DurationBadge creates a color-coded performance badge
func DurationBadge(d time.Duration, isRegressed bool) string {
	durationStr := formatDurationCompact(d)

	var style lipgloss.Style
	if isRegressed {
		// Red/orange background for regressed
		style = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"}).
			Background(WarningColor).
			Padding(0, 1).
			Bold(true)
	} else if d < time.Second {
		// Green background for fast
		style = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"}).
			Background(SuccessColor).
			Padding(0, 1)
	} else {
		// Blue background for normal
		style = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"}).
			Background(SecondaryColor).
			Padding(0, 1)
	}

	return style.Render(durationStr)
}

// formatDurationCompact formats a duration for display in badges (compact form)
func formatDurationCompact(d time.Duration) string {
	switch {
	case d < time.Second:
		return d.Round(time.Millisecond).String()
	case d < time.Minute:
		return d.Round(100 * time.Millisecond).String()
	default:
		return d.Round(time.Second).String()
	}
}
