package tui

import (
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/rshelekhov/lazymake/config"
	"github.com/rshelekhov/lazymake/internal/export"
	"github.com/rshelekhov/lazymake/internal/graph"
	"github.com/rshelekhov/lazymake/internal/highlight"
	"github.com/rshelekhov/lazymake/internal/history"
	"github.com/rshelekhov/lazymake/internal/makefile"
	"github.com/rshelekhov/lazymake/internal/safety"
	"github.com/rshelekhov/lazymake/internal/shell"
	"github.com/rshelekhov/lazymake/internal/variables"
	"github.com/rshelekhov/lazymake/internal/workspace"
)

type AppState int

const (
	StateList AppState = iota
	StateExecuting
	StateOutput
	StateHelp
	StateGraph
	StateConfirmDangerous
	StateVariables
	StateWorkspace
)

type Model struct {
	// UI Components
	List     list.Model
	Viewport viewport.Model
	Progress progress.Model
	Spinner  spinner.Model

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

	// Variable inspector state
	Variables         []variables.Variable
	VariableListIndex int

	// History state
	History       *history.History
	MakefilePath  string   // Absolute path to current Makefile
	RecentTargets []Target // Cached recent targets for current Makefile

	// Confirmation state
	PendingTarget *Target // Target awaiting dangerous command confirmation

	// Execution timing
	ExecutionStartTime time.Time
	ExecutionElapsed   time.Duration

	// Export and shell integration
	Exporter         *export.Exporter
	ShellIntegration *shell.Integration

	// Workspace management
	WorkspaceManager *workspace.Manager
	WorkspaceList    list.Model // For workspace picker UI

	// Syntax highlighting
	Highlighter *highlight.Highlighter

	// Key bindings for status bar display
	KeyBindings []key.Binding

	// Dimensions
	Width  int
	Height int

	Err error
}

// loadAndParseMakefile parses the makefile and related data
func loadAndParseMakefile(makefilePath string) ([]makefile.Target, *graph.Graph, []variables.Variable, error) {
	targets, err := makefile.Parse(makefilePath)
	if err != nil {
		return nil, nil, nil, err
	}

	depGraph := graph.BuildGraph(targets)

	// Parse and analyze variables
	vars, err := variables.ParseVariables(makefilePath)
	if err != nil {
		// Graceful degradation: continue without variables
		vars = []variables.Variable{}
	} else {
		// Expand variables using make
		_ = variables.ExpandVariables(makefilePath, vars)
		// Analyze usage across targets
		variables.AnalyzeUsage(vars, targets)
	}

	return targets, depGraph, vars, nil
}

// convertAndEnrichWithSafety converts makefile targets to TUI targets and adds safety checks
func convertAndEnrichWithSafety(targets []makefile.Target) []Target {
	// Load safety configuration
	safetyConfig, err := safety.LoadConfig()
	if err != nil {
		safetyConfig = safety.DefaultConfig()
	}

	var safetyResults map[string]*safety.CheckResult
	if safetyConfig.Enabled {
		checker, err := safety.NewChecker(safetyConfig)
		if err == nil {
			safetyResults = checker.CheckAllTargets(targets)
		}
	}

	// Convert targets to TUI format
	tuiTargets := make([]Target, len(targets))
	for i, t := range targets {
		tuiTargets[i] = Target{
			Name:        t.Name,
			Description: t.Description,
			CommentType: t.CommentType,
			Recipe:      t.Recipe,
		}

		// Populate safety fields if target was flagged
		if safetyResults != nil {
			if result, found := safetyResults[t.Name]; found {
				tuiTargets[i].IsDangerous = result.IsDangerous
				tuiTargets[i].DangerLevel = result.DangerLevel
				tuiTargets[i].SafetyMatches = result.Matches
			}
		}
	}

	return tuiTargets
}

// enrichWithHistory loads history and enriches targets with performance data
// Returns the list of recent targets and the history object
func enrichWithHistory(tuiTargets []Target, absPath string) ([]Target, *history.History) {
	// Load history
	hist, err := history.Load()
	if err != nil {
		hist = &history.History{Entries: make(map[string][]history.Entry)}
	}

	// Filter valid targets from history
	targetNames := extractTargetNames(tuiTargets)
	hist.FilterValid(absPath, targetNames)

	// Enrich targets with performance stats
	enrichTargetsWithPerformance(hist, absPath, tuiTargets)

	// Get recent entries and build recent targets list
	recentEntries := hist.GetRecent(absPath)
	return buildRecentTargets(recentEntries, tuiTargets), hist
}

// buildItemsList creates the list items for display
func buildItemsList(tuiTargets, recentTargets []Target) []list.Item {
	items := make([]list.Item, 0, len(tuiTargets)+len(recentTargets)+3)

	if len(recentTargets) > 0 {
		items = append(items, HeaderTarget{Label: "RECENT"})
		for _, t := range recentTargets {
			items = append(items, t)
		}
		items = append(items, SeparatorTarget{})
	}

	items = append(items, HeaderTarget{Label: "ALL TARGETS"})
	for _, t := range tuiTargets {
		items = append(items, t)
	}

	return items
}

func NewModel(cfg *config.Config) Model {
	// Parse makefile and load data
	targets, depGraph, vars, err := loadAndParseMakefile(cfg.MakefilePath)
	if err != nil {
		return Model{Err: err}
	}

	// Convert to TUI targets and enrich with safety checks
	tuiTargets := convertAndEnrichWithSafety(targets)

	// Get absolute path for history lookups
	absPath, err := filepath.Abs(cfg.MakefilePath)
	if err != nil {
		absPath = cfg.MakefilePath
	}

	// Enrich with history and performance data
	recentTargets, hist := enrichWithHistory(tuiTargets, absPath)

	// Build items list for display
	items := buildItemsList(tuiTargets, recentTargets)

	// Define key bindings for both list and status bar display
	keyBindings := []key.Binding{
		key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "run"),
		),
		key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "dependency graph"),
		),
		key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "quit"),
		),
	}

	delegate := NewItemDelegate()
	l := list.New(items, delegate, 0, 0)
	l.Title = "Makefile Targets"
	l.SetShowStatusBar(false) // Disabled - we use custom status bar
	l.SetShowHelp(false)      // Disabled - help text shown in custom status bar
	l.SetFilteringEnabled(true)
	l.Styles.Title = TitleStyle

	// Customize filter prompt to be shorter and prevent truncation
	l.FilterInput.Prompt = "/ "
	l.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(SecondaryColor)

	l.AdditionalShortHelpKeys = func() []key.Binding {
		return keyBindings
	}

	// Position cursor on first actual target (skip headers)
	// Find the first Target item
	for i, item := range items {
		if _, ok := item.(Target); ok {
			l.Select(i)
			break
		}
	}

	// Initialize export if enabled
	var exporter *export.Exporter
	if cfg.Export != nil && cfg.Export.Enabled {
		exporter, _ = export.NewExporter(cfg.Export)
	}

	// Initialize shell integration if enabled
	var shellInteg *shell.Integration
	if cfg.ShellIntegration != nil && cfg.ShellIntegration.Enabled {
		shellInteg, _ = shell.NewIntegration(cfg.ShellIntegration)
	}

	// Initialize syntax highlighter
	highlighter := highlight.NewHighlighter()

	// Initialize modern progress bar
	prog := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)

	// Initialize spinner with modern dot style
	spin := spinner.New()
	spin.Spinner = spinner.Dot
	spin.Style = lipgloss.NewStyle().Foreground(PrimaryColor)

	return Model{
		List:              l,
		Progress:          prog,
		Spinner:           spin,
		State:             StateList,
		Targets:           tuiTargets,
		Graph:             depGraph,
		GraphDepth:        -1,
		ShowOrder:         true,
		ShowCritical:      true,
		ShowParallel:      true,
		Variables:         vars,
		VariableListIndex: 0,
		History:           hist,
		MakefilePath:      absPath,
		RecentTargets:     recentTargets,
		Exporter:          exporter,
		ShellIntegration:  shellInteg,
		Highlighter:       highlighter,
		KeyBindings:       keyBindings,
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

// enrichTargetsWithPerformance populates PerfStats for all targets
func enrichTargetsWithPerformance(hist *history.History, makefilePath string, targets []Target) {
	for i := range targets {
		targets[i].PerfStats = hist.GetPerformanceStats(makefilePath, targets[i].Name)
	}
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
			// Create a copy and mark as recent (preserves all fields including safety & performance)
			recent := t
			recent.IsRecent = true
			recentTargets = append(recentTargets, recent)
		}
	}

	return recentTargets
}

// rebuildListItems reconstructs the list items from current targets and recent targets
func rebuildListItems(recentTargets, allTargets []Target) []list.Item {
	items := make([]list.Item, 0, len(allTargets)+len(recentTargets)+3)

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
	for _, t := range allTargets {
		items = append(items, t)
	}

	return items
}

// SwitchWorkspace reinitializes the model with a new Makefile path
// This performs a full reinitialization to ensure clean state
func (m Model) SwitchWorkspace(newMakefilePath string, cfg *config.Config) Model {
	// Create new config with updated Makefile path
	newCfg := *cfg
	newCfg.MakefilePath = newMakefilePath

	// Create fresh model with new Makefile
	newModel := NewModel(&newCfg)

	// Preserve UI state
	newModel.Width = m.Width
	newModel.Height = m.Height
	newModel.WorkspaceManager = m.WorkspaceManager

	// Record workspace access
	if m.WorkspaceManager != nil {
		m.WorkspaceManager.RecordAccess(newMakefilePath)
		_ = m.WorkspaceManager.Save() // Async, ignore errors (non-critical)
	}

	return newModel
}
