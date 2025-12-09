package tui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rshelekhov/lazymake/internal/makefile"
)

// Target represents a Makefile target in the TUI
type Target struct {
	Name        string
	Description string
	CommentType makefile.CommentType
}

// Implement list.Item interface
func (t Target) FilterValue() string {
	return t.Name + " " + t.Description
}

// ItemDelegate renders list items
type ItemDelegate struct{}

func (d ItemDelegate) Height() int { return 2 }

func (d ItemDelegate) Spacing() int { return 1 }

func (d ItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d ItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	target, ok := listItem.(Target)
	if !ok {
		return
	}

	var str string
	if index == m.Index() {
		str = SelectedItemStyle.Render("▶ " + target.Name)
	} else {
		str = NormalItemStyle.Render("  " + target.Name)
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
