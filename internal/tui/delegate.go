package tui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// Target represents a Makefile target
type Target struct {
	Name        string
	Description string
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
		str += "\n" + DescriptionStyle.Render(target.Description)
	}

	// TODO: fix unhandled error warning
	fmt.Fprint(w, str)
}
