package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/rshelekhov/lazymake/internal/makefile"
)

type AppState int

const (
	StateList AppState = iota
	StateExecuting
	StateOutput
	StateHelp
)

type Model struct {
	// UI Components
	List     list.Model
	Viewport viewport.Model

	// State
	State           AppState
	ExecutingTarget string
	Output          string
	ExecutionError  error
	Targets         []Target // Store targets for help view

	// Dimensions
	Width  int
	Height int

	Err error
}

func NewModel(makefilePath string) Model {
	targets, err := makefile.Parse(makefilePath)
	if err != nil {
		return Model{Err: err}
	}

	// Convert targets to TUI format
	tuiTargets := make([]Target, len(targets))
	items := make([]list.Item, len(targets))
	for i, t := range targets {
		tuiTarget := Target{
			Name:        t.Name,
			Description: t.Description,
			CommentType: t.CommentType,
		}
		tuiTargets[i] = tuiTarget
		items[i] = tuiTarget
	}

	delegate := ItemDelegate{}
	l := list.New(items, delegate, 0, 0)
	l.Title = "Makefile Targets"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = TitleStyle

	return Model{
		List:    l,
		State:   StateList,
		Targets: tuiTargets,
	}
}
