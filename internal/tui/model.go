package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/rshelekhov/lazymake/internal/graph"
	"github.com/rshelekhov/lazymake/internal/makefile"
)

type AppState int

const (
	StateList AppState = iota
	StateExecuting
	StateOutput
	StateHelp
	StateGraph
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

	// Graph state
	Graph        *graph.Graph
	GraphTarget  string // Selected target for graph view
	GraphDepth   int    // -1 = unlimited, 0 = direct deps only, etc.
	ShowOrder    bool   // Show execution order numbers
	ShowCritical bool   // Show critical path markers
	ShowParallel bool   // Show parallel markers

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

	depGraph := graph.BuildGraph(targets)

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

	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("g"),
				key.WithHelp("g", "graph view"),
			),
			key.NewBinding(
				key.WithKeys("?"),
				key.WithHelp("?", "toggle help"),
			),
		}
	}

	return Model{
		List:         l,
		State:        StateList,
		Targets:      tuiTargets,
		Graph:        depGraph,
		GraphDepth:   -1,
		ShowOrder:    true,
		ShowCritical: true,
		ShowParallel: true,
	}
}
