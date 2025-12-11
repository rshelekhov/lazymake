package tui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
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
	Recipe        []string               // Command lines to execute
	IsDangerous   bool                   // Whether target has dangerous commands
	DangerLevel   safety.Severity        // Highest severity level
	SafetyMatches []safety.MatchResult   // All matched safety rules
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

func (d ItemDelegate) Height() int { return 2 }

func (d ItemDelegate) Spacing() int { return 1 }

func (d ItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

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

	// Handle regular target
	target, ok := listItem.(Target)
	if !ok {
		return
	}

	// Build the target name with indicators (priority: danger > recent)
	var targetName string

	// Danger indicators take priority
	if target.IsDangerous {
		switch target.DangerLevel {
		case safety.SeverityCritical:
			targetName = "🚨 " + target.Name
		case safety.SeverityWarning:
			targetName = "⚠️  " + target.Name
		default:
			// SeverityInfo - no indicator
			targetName = target.Name
		}
	} else if target.IsRecent {
		targetName = "⏱  " + target.Name
	} else {
		targetName = target.Name
	}

	var str string
	if index == m.Index() {
		str = SelectedItemStyle.Render("▶ " + targetName)
	} else {
		str = NormalItemStyle.Render("  " + targetName)
	}

	if target.Description != "" {
		// Use different styles based on comment type
		// ## comments use cyan (industry standard for documentation)
		// # comments use gray (backward compatibility)
		descStyle := DescriptionStyle
		if target.CommentType == makefile.CommentDouble {
			descStyle = DocDescriptionStyle
		}
		str += "\n" + descStyle.Render(target.Description)
	}

	_, _ = fmt.Fprint(w, str)
}
