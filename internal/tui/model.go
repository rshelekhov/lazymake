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

	items := make([]list.Item, len(targets))
	for i, t := range targets {
		items[i] = Target{
			Name:        t.Name,
			Description: t.Description,
		}
	}

	delegate := ItemDelegate{}
	l := list.New(items, delegate, 0, 0)
	l.Title = "Makefile Targets"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = TitleStyle

	return Model{
		List:  l,
		State: StateList,
	}
}
