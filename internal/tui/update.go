package tui

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rshelekhov/lazymake/internal/executor"
	"github.com/rshelekhov/lazymake/internal/export"
	"github.com/rshelekhov/lazymake/internal/safety"
)

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.Err != nil {
		return m.updateError(msg)
	}

	switch m.State {
	case StateList:
		return m.updateList(msg)
	case StateOutput:
		return m.updateOutput(msg)
	case StateHelp:
		return m.updateHelp(msg)
	case StateGraph:
		return m.updateGraph(msg)
	case StateConfirmDangerous:
		return m.updateConfirmDangerous(msg)
	case StateExecuting:
		return m.updateExecuting(msg)
	case StateVariables:
		return m.updateVariables(msg)
	case StateWorkspace:
		return m.updateWorkspace(msg)
	default:
		return m, nil
	}
}

func (m Model) updateError(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}
	return m, nil
}

func (m Model) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	case tea.WindowSizeMsg:
		return m.handleWindowResize(msg), nil
	}

	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

// handleKeyPress processes keyboard input in list view
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "?":
		m.State = StateHelp
		return m, nil
	case "v":
		m.State = StateVariables
		m.initVariablesViewport()
		return m, nil
	case "w":
		m.State = StateWorkspace
		m.initWorkspacePicker()
		return m, nil
	case "g":
		return m.handleGraphView()
	case "enter":
		return m.handleTargetSelection()
	case "down", "j":
		return m.handleNavigateDown()
	case "up", "k":
		return m.handleNavigateUp()
	case "ctrl+d":
		m.RecipeViewport.HalfPageDown()
		return m, nil
	case "ctrl+u":
		m.RecipeViewport.HalfPageUp()
		return m, nil
	}

	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

// handleGraphView switches to graph view for selected target
func (m Model) handleGraphView() (tea.Model, tea.Cmd) {
	if selected := m.List.SelectedItem(); selected != nil {
		if target, ok := selected.(Target); ok {
			m.State = StateGraph
			m.GraphTarget = target.Name
			return m, nil
		}
	}
	return m, nil
}

// handleTargetSelection executes or confirms the selected target
func (m Model) handleTargetSelection() (tea.Model, tea.Cmd) {
	selected := m.List.SelectedItem()
	target, ok := selected.(Target)
	if !ok {
		return m, nil
	}

	// Check if target is critical and requires confirmation
	if target.IsDangerous && target.DangerLevel == safety.SeverityCritical {
		targetCopy := target
		m.PendingTarget = &targetCopy
		m.State = StateConfirmDangerous
		return m, nil
	}

	// Safe or non-critical target - execute immediately
	m.History.RecordExecution(m.MakefilePath, target.Name)
	_ = m.History.Save()

	// Refresh recent targets for next render
	recentEntries := m.History.GetRecent(m.MakefilePath)
	m.RecentTargets = buildRecentTargets(recentEntries, m.Targets)

	m.State = StateExecuting
	m.ExecutingTarget = target.Name
	m.ExecutionStartTime = time.Now()
	m.ExecutionElapsed = 0

	return m, tea.Batch(
		executeTarget(target.Name),
		tickTimer(),
		m.Spinner.Tick,
	)
}

// handleNavigateDown navigates to next target and updates recipe view
func (m Model) handleNavigateDown() (tea.Model, tea.Cmd) {
	m = navigateToNextTarget(m, true)
	m = updateRecipeViewportContent(m)
	return m, nil
}

// handleNavigateUp navigates to previous target and updates recipe view
func (m Model) handleNavigateUp() (tea.Model, tea.Cmd) {
	m = navigateToNextTarget(m, false)
	m = updateRecipeViewportContent(m)
	return m, nil
}

// handleWindowResize updates dimensions and layout when window size changes
func (m Model) handleWindowResize(msg tea.WindowSizeMsg) Model {
	m.Width = msg.Width
	m.Height = msg.Height

	// Calculate list size for left column (30% of width)
	leftWidth := max(int(float64(msg.Width)*0.30), 30)
	listWidth := leftWidth - 2  // Account for border
	listHeight := msg.Height - 5 // Account for status bar and border

	m.List.SetSize(listWidth, listHeight)

	// Calculate recipe viewport dimensions
	rightWidth := m.calculateRightWidth(msg.Width)
	availableHeight := msg.Height - 3

	m.initRecipeViewport(rightWidth, availableHeight)
	m.updateRecipeViewportForSelection(rightWidth)

	return m
}

// calculateRightWidth calculates the width for the right column (recipe view)
func (m Model) calculateRightWidth(totalWidth int) int {
	leftWidthPercent := 0.35
	minLeftWidth := 35
	calcLeftWidth := int(float64(totalWidth) * leftWidthPercent)

	if calcLeftWidth < minLeftWidth && totalWidth >= minLeftWidth*2 {
		calcLeftWidth = minLeftWidth
	} else if calcLeftWidth < minLeftWidth {
		calcLeftWidth = int(float64(totalWidth) * leftWidthPercent)
	}
	if calcLeftWidth < 10 {
		calcLeftWidth = 10
	}

	return max(totalWidth-calcLeftWidth-1, 10)
}

// updateRecipeViewportForSelection updates recipe content for selected target
func (m Model) updateRecipeViewportForSelection(rightWidth int) {
	if selectedItem := m.List.SelectedItem(); selectedItem != nil {
		if target, ok := selectedItem.(Target); ok {
			content := m.buildRecipeContent(&target, rightWidth)
			m.RecipeViewport.SetContent(content)
			m.RecipeViewport.GotoTop()
		}
	}
}

// navigateToNextTarget moves the cursor to the next/previous target, skipping separators and headers
func navigateToNextTarget(m Model, down bool) Model {
	items := m.List.Items()
	currentIndex := m.List.Index()

	if down {
		// Search downward for next target
		for i := currentIndex + 1; i < len(items); i++ {
			if _, ok := items[i].(Target); ok {
				m.List.Select(i)
				return m
			}
		}
	} else {
		// Search upward for previous target
		for i := currentIndex - 1; i >= 0; i-- {
			if _, ok := items[i].(Target); ok {
				m.List.Select(i)
				return m
			}
		}
	}

	return m
}

// updateRecipeViewportContent updates the recipe viewport content for the currently selected target
func updateRecipeViewportContent(m Model) Model {
	// Calculate right column width (matching renderListView logic)
	leftWidthPercent := 0.35
	minLeftWidth := 35
	calcLeftWidth := int(float64(m.Width) * leftWidthPercent)
	if calcLeftWidth < minLeftWidth && m.Width >= minLeftWidth*2 {
		calcLeftWidth = minLeftWidth
	} else if calcLeftWidth < minLeftWidth {
		calcLeftWidth = int(float64(m.Width) * leftWidthPercent)
	}
	if calcLeftWidth < 10 {
		calcLeftWidth = 10
	}

	// Estimate right width
	rightWidth := max(m.Width-calcLeftWidth-1, 10)

	// Update viewport content for currently selected target
	if selectedItem := m.List.SelectedItem(); selectedItem != nil {
		if target, ok := selectedItem.(Target); ok {
			content := m.buildRecipeContent(&target, rightWidth)
			m.RecipeViewport.SetContent(content)
			m.RecipeViewport.GotoTop() // Auto-scroll to top on new selection
		}
	}

	return m
}

func (m Model) updateOutput(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.State = StateList
			return m, nil
		}
		var cmd tea.Cmd
		m.Viewport, cmd = m.Viewport.Update(msg)
		return m, cmd

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.resizeViewport()
	}

	return m, nil
}

func (m Model) updateHelp(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc", "?":
			m.State = StateList
		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}
	return m, nil
}

// updateGraph handles the graph view state
func (m Model) updateGraph(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "esc", "g":
			// Return to list view and clear graph target
			m.State = StateList
			m.GraphTarget = ""
			return m, nil

		case "+", "=":
			// Increase depth (show more levels)
			if m.GraphDepth == -1 {
				// Already unlimited, do nothing
			} else {
				m.GraphDepth++
			}

		case "-", "_":
			// Decrease depth (show fewer levels)
			if m.GraphDepth == -1 {
				m.GraphDepth = 5 // Start with 5 levels when coming from unlimited
			} else if m.GraphDepth > 0 {
				m.GraphDepth--
			}

		case "o", "O":
			// Toggle order display
			m.ShowOrder = !m.ShowOrder

		case "c", "C":
			// Toggle critical path display
			m.ShowCritical = !m.ShowCritical

		case "p", "P":
			// Toggle parallel display
			m.ShowParallel = !m.ShowParallel
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}

	return m, nil
}

func (m Model) updateExecuting(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case timerTickMsg:
		// Update elapsed time
		if m.State == StateExecuting {
			m.ExecutionElapsed = time.Since(m.ExecutionStartTime)
			// Update spinner
			m.Spinner, cmd = m.Spinner.Update(msg)
			return m, tea.Batch(tickTimer(), cmd) // Continue ticking and spinning
		}
		return m, nil // Stop if not executing

	case spinner.TickMsg:
		// Handle spinner animation ticks
		if m.State == StateExecuting {
			m.Spinner, cmd = m.Spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case executeFinishedMsg:
		// Timer stops automatically (state changes from StateExecuting)

		// Record execution with timing data
		success := msg.result.Err == nil
		m.History.RecordExecutionWithTiming(m.MakefilePath, m.ExecutingTarget, msg.result.Duration, success)
		_ = m.History.Save() // Async, ignore errors

		// Export execution result (async, non-blocking)
		if m.Exporter != nil {
			go func() {
				record := export.NewExecutionRecord(
					m.MakefilePath,
					m.ExecutingTarget,
					msg.result,
				)
				if err := m.Exporter.Export(record); err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "Export failed: %v\n", err)
				}
			}()
		}

		// Shell integration (async, non-blocking)
		if m.ShellIntegration != nil {
			go func() {
				if err := m.ShellIntegration.RecordExecution(m.ExecutingTarget); err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "Shell integration failed: %v\n", err)
				}
			}()
		}

		// Refresh performance stats for all targets
		enrichTargetsWithPerformance(m.History, m.MakefilePath, m.Targets)

		// Refresh recent targets to show updated timing
		recentEntries := m.History.GetRecent(m.MakefilePath)
		m.RecentTargets = buildRecentTargets(recentEntries, m.Targets)

		// Rebuild and update list items to reflect new performance stats
		updatedItems := rebuildListItems(m.RecentTargets, m.Targets)
		m.List.SetItems(updatedItems)

		m.State = StateOutput
		m.Output = msg.result.Output
		m.ExecutionError = msg.result.Err
		m.initViewport(msg.result.Output)

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}
	return m, nil
}

func (m *Model) initViewport(content string) {
	vw, vh := computeViewportSize(m.Width, m.Height)
	m.Viewport = viewport.New(vw, vh)
	m.Viewport.SetContent(content)
	m.Viewport.Style = lipgloss.NewStyle()
}

func (m *Model) resizeViewport() {
	vw, vh := computeViewportSize(m.Width, m.Height)
	m.Viewport.Width = vw
	m.Viewport.Height = vh
}

// initRecipeViewport initializes the recipe preview viewport with given dimensions
func (m *Model) initRecipeViewport(width, height int) {
	// Calculate content dimensions using CORRECT values (matching left column)
	contentWidth := width - 8   // 6 (padding) + 2 (border) = 8
	contentHeight := height - 6 // 4 (padding) + 2 (border) = 6

	m.RecipeViewport = viewport.New(contentWidth, contentHeight)
	m.RecipeViewport.Style = lipgloss.NewStyle()

	// Start at top (will auto-scroll on selection change)
	m.RecipeViewport.YPosition = 0
}

func (m *Model) initVariablesViewport() {
	// Calculate available height (subtract status bar height)
	statusBarHeight := 3 // Border + padding
	availableHeight := m.Height - statusBarHeight

	// Calculate content dimensions (account for border and padding)
	contentWidth := m.Width - 8   // 6 (padding) + 2 (border) = 8
	contentHeight := availableHeight - 6 // 4 (padding) + 2 (border) = 6

	m.VariablesViewport = viewport.New(contentWidth, contentHeight)
	m.VariablesViewport.Style = lipgloss.NewStyle()

	// Set content
	content := m.buildVariablesContent()
	m.VariablesViewport.SetContent(content)

	// Start at top
	m.VariablesViewport.YPosition = 0
}

func computeViewportSize(winWidth, winHeight int) (int, int) {
	width := getContentWidth(winWidth)

	viewportWidth := width - 6 // 2 border + 4 padding

	// Account for UI elements in renderOutputView:
	// - Leading newline: 1 line
	// - Header: 1 line
	// - Potential regression alert: 0-2 lines
	// - Double newline: 2 lines
	// - Viewport border: 2 lines (top + bottom)
	// - Viewport padding: 2 lines (top + bottom)
	// - Footer: 2 lines
	// Total overhead: 10-12 lines
	// Use 14 to include safety margin for text wrapping
	viewportHeight := winHeight - 14

	if viewportWidth < 20 {
		viewportWidth = 20
	}
	if viewportHeight < 3 {
		viewportHeight = 3
	}

	return viewportWidth, viewportHeight
}

// updateConfirmDangerous handles the dangerous command confirmation dialog
func (m Model) updateConfirmDangerous(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "esc":
			// Cancel confirmation, return to list
			m.State = StateList
			m.PendingTarget = nil
			return m, nil

		case "enter":
			// Proceed with execution of dangerous target
			if m.PendingTarget != nil {
				target := *m.PendingTarget

				// Record execution in history
				m.History.RecordExecution(m.MakefilePath, target.Name)
				_ = m.History.Save()

				// Refresh recent targets
				recentEntries := m.History.GetRecent(m.MakefilePath)
				m.RecentTargets = buildRecentTargets(recentEntries, m.Targets)

				// Clear pending target and start execution
				m.PendingTarget = nil
				m.State = StateExecuting
				m.ExecutingTarget = target.Name
				m.ExecutionStartTime = time.Now()
				m.ExecutionElapsed = 0
				return m, tea.Batch(
					executeTarget(target.Name),
					tickTimer(),   // Start timer
					m.Spinner.Tick, // Start spinner animation
				)
			}
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}

	return m, nil
}

// Custom message for execution results
type executeFinishedMsg struct {
	result executor.Result
}

// Custom message for timer ticks
type timerTickMsg struct{}

func executeTarget(target string) tea.Cmd {
	return func() tea.Msg {
		result := executor.Execute(target)
		return executeFinishedMsg{result: result}
	}
}

func tickTimer() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return timerTickMsg{}
	})
}

// updateVariables handles the variable inspector view state
func (m Model) updateVariables(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "esc", "v":
			// Return to list view
			m.State = StateList
			return m, nil
		}
		// Pass other keys to viewport for scrolling
		var cmd tea.Cmd
		m.VariablesViewport, cmd = m.VariablesViewport.Update(msg)
		return m, cmd

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.initVariablesViewport() // Reinitialize viewport with new dimensions
	}

	return m, nil
}
