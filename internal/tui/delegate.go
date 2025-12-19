package tui

import (
	"fmt"
	"io"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rshelekhov/lazymake/internal/history"
	"github.com/rshelekhov/lazymake/internal/makefile"
	"github.com/rshelekhov/lazymake/internal/safety"
)

// Target represents a Makefile target in the TUI
type Target struct {
	Name        string
	Description string
	CommentType makefile.CommentType
	IsRecent    bool // Marks targets that appear in recent history

	// Recipe and safety fields
	Recipe        []string             // Command lines to execute
	IsDangerous   bool                 // Whether target has dangerous commands
	DangerLevel   safety.Severity      // Highest severity level
	SafetyMatches []safety.MatchResult // All matched safety rules

	// Performance fields
	PerfStats *history.PerformanceStats // nil if no data
}

// Implement list.Item interface
func (t Target) FilterValue() string {
	return t.Name + " " + t.Description
}

// SeparatorTarget renders a horizontal line between sections
type SeparatorTarget struct{}

func (s SeparatorTarget) FilterValue() string { return "" }

// HeaderTarget renders a section header (e.g., "RECENT", "ALL TARGETS")
type HeaderTarget struct {
	Label string
}

func (h HeaderTarget) FilterValue() string { return "" }

// ItemDelegate renders list items
type ItemDelegate struct{}

func (d ItemDelegate) Height() int { return 3 } // Increased to handle text wrapping

func (d ItemDelegate) Spacing() int { return 1 }

func (d ItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d ItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	// Get available width for wrapping (subtract padding and margins)
	availableWidth := m.Width() - 4 // Account for padding and list margins

	// Handle separator
	if _, ok := listItem.(SeparatorTarget); ok {
		separator := SeparatorStyle.Render("─────────────────────────────────────")
		_, _ = fmt.Fprint(w, separator)
		return
	}

	// Handle section header
	if header, ok := listItem.(HeaderTarget); ok {
		headerStr := SectionHeaderStyle.Render(header.Label)
		_, _ = fmt.Fprint(w, headerStr)
		return
	}

	// Handle regular target
	target, ok := listItem.(Target)
	if !ok {
		return
	}

	// Build the target name with indicators (priority: danger > regression > recent)
	var targetName string

	// Danger indicators take priority
	switch {
	case target.IsDangerous && target.DangerLevel == safety.SeverityCritical:
		targetName = "🚨 " + target.Name
	case target.IsDangerous && target.DangerLevel == safety.SeverityWarning:
		targetName = "⚠️  " + target.Name
	case target.PerfStats != nil && target.PerfStats.IsRegressed:
		targetName = "📈 " + target.Name // Regression indicator
	case target.IsRecent:
		targetName = "⏱  " + target.Name // Recent indicator
	default:
		targetName = target.Name
	}

	var str string
	if index == m.Index() {
		// Wrap text to available width using a fresh style with Width set
		wrappedStyle := lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true).
			PaddingLeft(1).
			Width(availableWidth)
		str = wrappedStyle.Render("▶ " + targetName)
	} else {
		wrappedStyle := lipgloss.NewStyle().
			Foreground(TextColor).
			PaddingLeft(1).
			Width(availableWidth)
		str = wrappedStyle.Render("  " + targetName)
	}

	if target.Description != "" {
		// Use different styles based on comment type
		// ## comments use cyan (industry standard for documentation)
		// # comments use gray (backward compatibility)
		var descColor lipgloss.AdaptiveColor
		if target.CommentType == makefile.CommentDouble {
			descColor = SecondaryColor // Cyan for documented comments
		} else {
			descColor = MutedColor // Gray for regular comments
		}

		// Wrap description to available width
		wrappedDescStyle := lipgloss.NewStyle().
			Foreground(descColor).
			PaddingLeft(3).
			Width(availableWidth)
		str += "\n" + wrappedDescStyle.Render(target.Description)
	}

	// Add duration badge on line 3 if appropriate
	if shouldShowDurationBadge(target) {
		durationStr := formatDuration(target.PerfStats.LastDuration)
		color := getDurationColor(target.PerfStats)
		badge := lipgloss.NewStyle().
			Foreground(color).
			Align(lipgloss.Right).
			PaddingLeft(3).
			Width(availableWidth).
			Render(durationStr)
		str += "\n" + badge
	}

	_, _ = fmt.Fprint(w, str)
}

// shouldShowDurationBadge returns true if we should show duration badge for this target
func shouldShowDurationBadge(target Target) bool {
	if target.PerfStats == nil {
		return false
	}
	// Show if: regressed or recent (users judge "slow" by seeing duration)
	return target.PerfStats.IsRegressed || target.IsRecent
}

// formatDuration formats a duration for display
func formatDuration(d time.Duration) string {
	switch {
	case d < time.Second:
		return fmt.Sprintf("%dms", d.Milliseconds())
	case d < time.Minute:
		return fmt.Sprintf("%.1fs", d.Seconds())
	default:
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	}
}

// getDurationColor returns the color for a duration badge
func getDurationColor(stats *history.PerformanceStats) lipgloss.Color {
	switch {
	case stats.IsRegressed:
		return lipgloss.Color("220") // Orange (warning)
	case stats.AvgDuration < time.Second:
		return lipgloss.Color("42") // Green (fast)
	default:
		return lipgloss.Color("86") // Cyan (normal)
	}
}
