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
		Foreground(TextMuted).
		BorderForeground(PrimaryColor)

	// Normal items
	d.Styles.NormalTitle = d.Styles.NormalTitle.
		Foreground(TextPrimary)

	d.Styles.NormalDesc = d.Styles.NormalDesc.
		Foreground(TextMuted)

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

	isSelected := index == m.Index()
	icon, iconColor := d.getTargetIcon(target)
	desc := d.buildDescription(target)
	titleStyle, descStyle, titleColor := d.getStyles(target, isSelected)
	title := d.buildTitle(target, icon, iconColor, titleColor)

	var output strings.Builder
	output.WriteString(titleStyle.Render(title))
	output.WriteString("\n")

	d.renderDescription(desc, descStyle, m.Width(), &output)
	fmt.Fprint(w, output.String())
}

// getTargetIcon returns the icon and color for a target based on its status
func (d ItemDelegate) getTargetIcon(target Target) (string, lipgloss.AdaptiveColor) {
	switch {
	case target.IsDangerous && target.DangerLevel == safety.SeverityCritical:
		return IconDangerCritical, ErrorColor
	case target.IsDangerous && target.DangerLevel == safety.SeverityWarning:
		return IconDangerWarning, WarningColor
	case target.IsDangerous && target.DangerLevel == safety.SeverityInfo:
		return "○", SecondaryColor
	case target.PerfStats != nil && target.PerfStats.IsRegressed:
		return IconRegression, WarningColor
	case target.IsRecent:
		return IconRecent, SecondaryColor
	default:
		return "", lipgloss.AdaptiveColor{}
	}
}

// buildDescription creates the description text with performance badge if needed
func (d ItemDelegate) buildDescription(target Target) string {
	desc := target.Description
	if shouldShowDurationBadge(target) {
		badge := DurationBadge(target.PerfStats.LastDuration, target.PerfStats.IsRegressed)
		if desc != "" {
			return desc + " " + badge
		}
		return badge
	}
	return desc
}

// getStyles returns the appropriate styles based on selection state and target
func (d ItemDelegate) getStyles(target Target, isSelected bool) (titleStyle, descStyle lipgloss.Style, titleColor lipgloss.AdaptiveColor) {
	if isSelected {
		titleStyle = d.Styles.SelectedTitle
		titleColor = PrimaryColor
		descStyle = d.Styles.SelectedDesc.UnsetWidth()
	} else {
		titleStyle = d.Styles.NormalTitle
		titleColor = TextPrimary
		descStyle = d.Styles.NormalDesc.UnsetWidth()
	}

	// Use different description color for ## comments
	if target.CommentType == makefile.CommentDouble {
		descStyle = descStyle.Foreground(TextSecondary)
	}

	return titleStyle, descStyle, titleColor
}

// buildTitle creates the styled title with icon and target name
func (d ItemDelegate) buildTitle(target Target, icon string, iconColor, titleColor lipgloss.AdaptiveColor) string {
	var titleParts []string
	if icon != "" {
		iconStyled := lipgloss.NewStyle().Foreground(iconColor).Render(icon)
		titleParts = append(titleParts, iconStyled)
	}
	nameStyled := lipgloss.NewStyle().Foreground(titleColor).Render(target.Name)
	titleParts = append(titleParts, nameStyled)
	return strings.Join(titleParts, " ")
}

// renderDescription renders the description with text wrapping
func (d ItemDelegate) renderDescription(desc string, descStyle lipgloss.Style, listWidth int, output *strings.Builder) {
	if desc == "" {
		return
	}

	availableWidth := listWidth - 4 // Leave some margin
	if availableWidth < 20 {
		availableWidth = 20 // Minimum width
	}

	wrappedDesc := wordwrap.String(desc, availableWidth)
	descLines := strings.Split(wrappedDesc, "\n")

	for i, line := range descLines {
		if i > 0 {
			output.WriteString("\n")
		}
		output.WriteString(descStyle.Render(line))
	}
}

// Modern icon constants - more consistent than emojis across terminals
const (
	IconDangerCritical = "○" // Empty circle (red outline)
	IconDangerWarning  = "○" // Empty circle (yellow outline)
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
