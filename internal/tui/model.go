package tui

import (
	"path/filepath"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/rshelekhov/lazymake/internal/graph"
	"github.com/rshelekhov/lazymake/internal/history"
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

	// History state
	History       *history.History
	MakefilePath  string   // Absolute path to current Makefile
	RecentTargets []Target // Cached recent targets for current Makefile

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
	for i, t := range targets {
		tuiTargets[i] = Target{
			Name:        t.Name,
			Description: t.Description,
			CommentType: t.CommentType,
		}
	}

	// Load history
	hist, err := history.Load()
	if err != nil {
		// Graceful degradation: continue with empty history
		hist = &history.History{Entries: make(map[string][]history.Entry)}
	}

	// Get absolute Makefile path
	absPath, err := filepath.Abs(makefilePath)
	if err != nil {
		absPath = makefilePath // Fallback to original path
	}

	// Filter valid targets from history
	targetNames := extractTargetNames(tuiTargets)
	hist.FilterValid(absPath, targetNames)

	// Get recent entries and build recent targets list
	recentEntries := hist.GetRecent(absPath)
	recentTargets := buildRecentTargets(recentEntries, tuiTargets)

	// Build items list with recent section
	items := make([]list.Item, 0, len(tuiTargets)+len(recentTargets)+3)

	if len(recentTargets) > 0 {
		// Add recent section
		items = append(items, HeaderTarget{Label: "RECENT"})
		for _, t := range recentTargets {
			items = append(items, t)
		}
		items = append(items, SeparatorTarget{})
	}

	// Add all targets section
	items = append(items, HeaderTarget{Label: "ALL TARGETS"})
	for _, t := range tuiTargets {
		items = append(items, t)
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

	// Position cursor on first actual target (skip headers)
	// Find the first Target item
	for i, item := range items {
		if _, ok := item.(Target); ok {
			l.Select(i)
			break
		}
	}

	return Model{
		List:          l,
		State:         StateList,
		Targets:       tuiTargets,
		Graph:         depGraph,
		GraphDepth:    -1,
		ShowOrder:     true,
		ShowCritical:  true,
		ShowParallel:  true,
		History:       hist,
		MakefilePath:  absPath,
		RecentTargets: recentTargets,
	}
}

// extractTargetNames extracts just the names from a slice of targets
func extractTargetNames(targets []Target) []string {
	names := make([]string, len(targets))
	for i, t := range targets {
		names[i] = t.Name
	}
	return names
}

// buildRecentTargets creates TUI targets from history entries
func buildRecentTargets(entries []history.Entry, allTargets []Target) []Target {
	if len(entries) == 0 {
		return nil
	}

	// Build a map of target name -> Target for quick lookup
	targetMap := make(map[string]Target)
	for _, t := range allTargets {
		targetMap[t.Name] = t
	}

	// Build recent targets list, preserving history order
	recentTargets := make([]Target, 0, len(entries))
	for _, entry := range entries {
		if t, ok := targetMap[entry.Name]; ok {
			// Create a copy and mark as recent
			recent := t
			recent.IsRecent = true
			recentTargets = append(recentTargets, recent)
		}
	}

	return recentTargets
}
