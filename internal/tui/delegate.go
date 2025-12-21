package tui

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
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
	Recipe           []string             // Command lines to execute
	LanguageOverride string               // Manual language override for syntax highlighting
	IsDangerous      bool                 // Whether target has dangerous commands
	DangerLevel      safety.Severity      // Highest severity level
	SafetyMatches    []safety.MatchResult // All matched safety rules

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

// ItemDelegate renders list items using bubbles default delegate styling with our colors
type ItemDelegate struct {
	list.DefaultDelegate
}

// NewItemDelegate creates a new delegate with our custom styling
func NewItemDelegate() ItemDelegate {
	d := list.NewDefaultDelegate()

	// Apply our GitHub-inspired colors
	// Selected item (highlighted)
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		Foreground(PrimaryColor).
		BorderForeground(PrimaryColor)

	d.Styles.SelectedDesc = d.Styles.SelectedDesc.
		Foreground(SecondaryColor).
		BorderForeground(PrimaryColor)

	// Normal items
	d.Styles.NormalTitle = d.Styles.NormalTitle.
		Foreground(TextPrimary)

	d.Styles.NormalDesc = d.Styles.NormalDesc.
		Foreground(TextSecondary)

	// Dimmed (when filtering)
	d.Styles.DimmedTitle = d.Styles.DimmedTitle.
		Foreground(TextMuted)

	d.Styles.DimmedDesc = d.Styles.DimmedDesc.
		Foreground(TextMuted)

	// Filter match highlighting
	d.Styles.FilterMatch = d.Styles.FilterMatch.
		Foreground(WarningColor).
		Bold(true)

	return ItemDelegate{DefaultDelegate: d}
}

func (d ItemDelegate) Height() int  { return 2 } // Base height, may expand for wrapped text
func (d ItemDelegate) Spacing() int { return 1 }
func (d ItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return d.DefaultDelegate.Update(msg, m)
}

// targetItem is a wrapper to make Target compatible with DefaultDelegate
type targetItem struct {
	title string
	desc  string
}

func (i targetItem) Title() string       { return i.title }
func (i targetItem) Description() string { return i.desc }
func (i targetItem) FilterValue() string { return i.title + " " + i.desc }

func (d ItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
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

	// Handle regular target - custom rendering with text wrapping
	target, ok := listItem.(Target)
	if !ok {
		return
	}

	// Determine if this item is selected
	isSelected := index == m.Index()

	// Build title with icon prefix based on target status
	var icon string
	var iconColor lipgloss.AdaptiveColor

	switch {
	case target.IsDangerous && target.DangerLevel == safety.SeverityCritical:
		icon = IconDangerCritical
		iconColor = ErrorColor
	case target.IsDangerous && target.DangerLevel == safety.SeverityWarning:
		icon = IconDangerWarning
		iconColor = WarningColor
	case target.PerfStats != nil && target.PerfStats.IsRegressed:
		icon = IconRegression
		iconColor = WarningColor
	case target.IsRecent:
		icon = IconRecent
		iconColor = SecondaryColor
	}

	// Build title with icon
	var titleParts []string
	if icon != "" {
		iconStyled := lipgloss.NewStyle().Foreground(iconColor).Render(icon)
		titleParts = append(titleParts, iconStyled)
	}
	titleParts = append(titleParts, target.Name)
	title := strings.Join(titleParts, " ")

	// Build description with badge if needed
	desc := target.Description
	if shouldShowDurationBadge(target) {
		badge := DurationBadge(target.PerfStats.LastDuration, target.PerfStats.IsRegressed)
		if desc != "" {
			desc = desc + " " + badge
		} else {
			desc = badge
		}
	}

	// Select appropriate styles based on selection state
	var titleStyle, descStyle lipgloss.Style
	if isSelected {
		titleStyle = d.Styles.SelectedTitle
		descStyle = d.Styles.SelectedDesc.UnsetWidth() // Remove fixed width to prevent padding

		// Use different description color for ## comments
		if target.CommentType == makefile.CommentDouble {
			descStyle = descStyle.Foreground(SecondaryColor)
		}
	} else {
		titleStyle = d.Styles.NormalTitle
		descStyle = d.Styles.NormalDesc.UnsetWidth() // Remove fixed width to prevent padding

		// Use different description color for ## comments
		if target.CommentType == makefile.CommentDouble {
			descStyle = descStyle.Foreground(SecondaryColor)
		}
	}

	// Render title
	var output strings.Builder
	output.WriteString(titleStyle.Render(title))
	output.WriteString("\n")

	// Render description with wrapping
	if desc != "" {
		// Calculate available width for wrapping
		// Account for list padding and any indentation
		availableWidth := m.Width() - 4 // Leave some margin
		if availableWidth < 20 {
			availableWidth = 20 // Minimum width
		}

		// Wrap the description text
		wrappedDesc := wordwrap.String(desc, availableWidth)
		descLines := strings.Split(wrappedDesc, "\n")

		// Render each line with the description style
		for i, line := range descLines {
			if i > 0 {
				output.WriteString("\n")
			}
			output.WriteString(descStyle.Render(line))
		}
	}

	fmt.Fprint(w, output.String())
}

// Modern icon constants - more consistent than emojis across terminals
const (
	IconDangerCritical = "●" // Filled circle
	IconDangerWarning  = "○" // Empty circle
	IconRegression     = "↑" // Up arrow (performance up = bad)
	IconRecent         = "◆" // Diamond
	IconFavorite       = "★" // Star
	IconSuccess        = "✓" // Check
	IconError          = "✗" // X mark
	IconInfo           = "ℹ" // Info
	IconArrowRight     = "▸" // Right arrow (selected)
)

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
