package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rshelekhov/lazymake/internal/executor"
	"github.com/rshelekhov/lazymake/internal/export"
	"github.com/rshelekhov/lazymake/internal/safety"
	"github.com/rshelekhov/lazymake/internal/shell"
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
	// Handle custom filtering
	if m.IsFiltering {
		return m.handleFilteringKeys(msg)
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "/":
		m.IsFiltering = true
		m.FilterInput = ""
		return m, nil
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
	case "ctrl+d":
		m.RecipeViewport.HalfPageDown()
		return m, nil
	case "ctrl+u":
		m.RecipeViewport.HalfPageUp()
		return m, nil
	case "down", "j":
		m = navigateToTarget(m, true)
		m = updateRecipeViewportContent(m)
		return m, nil
	case "up", "k":
		m = navigateToTarget(m, false)
		m = updateRecipeViewportContent(m)
		return m, nil
	}

	// Delegate all other keys to the list component
	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)

	// After list update, ensure cursor is on a Target (not header/separator)
	m = ensureCursorOnTarget(m)
	m = updateRecipeViewportContent(m)

	return m, cmd
}

// navigateToTarget moves to next/previous Target, skipping headers and separators
func navigateToTarget(m Model, down bool) Model {
	items := m.List.Items()
	currentIndex := m.List.Index()

	if down {
		// Navigate down
		for i := currentIndex + 1; i < len(items); i++ {
			if _, ok := items[i].(Target); ok {
				m.List.Select(i)
				return m
			}
		}
	} else {
		// Navigate up
		for i := currentIndex - 1; i >= 0; i-- {
			if _, ok := items[i].(Target); ok {
				m.List.Select(i)
				return m
			}
		}
	}

	return m
}

// handleFilteringKeys handles key input when filtering mode is active
func (m Model) handleFilteringKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		// Exit filtering mode and reset
		m.IsFiltering = false
		m.FilterInput = ""
		m = applyCustomFilter(m)
		m = ensureCursorOnTarget(m)
		m = updateRecipeViewportContent(m)
		return m, nil

	case tea.KeyBackspace:
		if len(m.FilterInput) > 0 {
			m.FilterInput = m.FilterInput[:len(m.FilterInput)-1]
			m = applyCustomFilter(m)
			m = ensureCursorOnTarget(m)
			m = updateRecipeViewportContent(m)
		}
		return m, nil

	case tea.KeyRunes:
		// Check if it's "/" and filter is empty - close filter
		if string(msg.Runes) == "/" && m.FilterInput == "" {
			m.IsFiltering = false
			m = applyCustomFilter(m)
			m = ensureCursorOnTarget(m)
			m = updateRecipeViewportContent(m)
			return m, nil
		}
		// Add typed character to filter
		m.FilterInput += string(msg.Runes)
		m = applyCustomFilter(m)
		m = ensureCursorOnTarget(m)
		m = updateRecipeViewportContent(m)
		return m, nil

	case tea.KeyEnter:
		// Keep filtering active, just select target
		return m.handleTargetSelection()
	}

	// Allow navigation while filtering - delegate to list
	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)
	m = ensureCursorOnTarget(m)
	m = updateRecipeViewportContent(m)
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

	// Reset streaming output and initialize viewport
	m.StreamingOutput = &strings.Builder{}
	m.initExecutingViewport()

	return m, tea.Batch(
		executeTargetStreaming(target.Name, m.MakefilePath),
		tickTimer(),
		m.Spinner.Tick,
	)
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

// applyCustomFilter filters targets based on current FilterInput and updates the list
func applyCustomFilter(m Model) Model {
	if m.FilterInput == "" {
		// No filter, show all with headers
		items := buildItemsList(m.AllTargets, m.RecentTargets)
		m.List.SetItems(items)
		return m
	}

	// Fuzzy filter targets
	var filteredTargets []Target
	for _, target := range m.AllTargets {
		filterValue := target.Name + " " + target.Description
		if fuzzyMatch(m.FilterInput, filterValue) {
			filteredTargets = append(filteredTargets, target)
		}
	}

	// Build filtered items WITHOUT headers (clean list during search)
	var items []list.Item
	for _, t := range filteredTargets {
		items = append(items, t)
	}
	m.List.SetItems(items)

	// Move cursor to first item
	if len(items) > 0 {
		m.List.Select(0)
	}

	return m
}

// fuzzyMatch performs simple case-insensitive substring matching
func fuzzyMatch(pattern, text string) bool {
	pattern = strings.ToLower(pattern)
	text = strings.ToLower(text)
	return strings.Contains(text, pattern)
}

// ensureCursorOnTarget checks if cursor is on a Target, if not moves to nearest Target
// This is used after list updates that might leave cursor on header/separator
func ensureCursorOnTarget(m Model) Model {
	selectedItem := m.List.SelectedItem()
	if selectedItem == nil {
		return m
	}

	// Check if current selection is a Target
	if _, ok := selectedItem.(Target); ok {
		return m // Already on a target
	}

	// Current item is Header/Separator, find nearest Target
	items := m.List.Items()
	currentIndex := m.List.Index()

	// Try forward first
	for i := currentIndex + 1; i < len(items); i++ {
		if _, ok := items[i].(Target); ok {
			m.List.Select(i)
			return m
		}
	}

	// Try backward
	for i := currentIndex - 1; i >= 0; i-- {
		if _, ok := items[i].(Target); ok {
			m.List.Select(i)
			return m
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
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleExecutingKeyPress(msg)

	case timerTickMsg:
		return m.handleTimerTick(msg)

	case spinner.TickMsg:
		return m.handleSpinnerTick(msg)

	case streamStartedMsg:
		m.OutputChunks = msg.chunks
		m.CancelExecution = msg.cancel
		return m, waitForChunk(m.OutputChunks)

	case outputChunkMsg:
		return m.handleOutputChunk(msg)

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		if m.State == StateExecuting {
			m.initExecutingViewport()
			m.ExecutingViewport.SetContent(m.StreamingOutput.String())
			m.ExecutingViewport.GotoBottom()
		}
	}
	return m, nil
}

// handleExecutingKeyPress handles keyboard input during execution
func (m Model) handleExecutingKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		if m.CancelExecution != nil {
			m.CancelExecution()
		}
		return m.handleExecutionComplete(fmt.Errorf("execution canceled"))
	case "j", "down":
		m.ExecutingViewport.ScrollDown(1)
	case "k", "up":
		m.ExecutingViewport.ScrollUp(1)
	case "ctrl+d":
		m.ExecutingViewport.HalfPageDown()
	case "ctrl+u":
		m.ExecutingViewport.HalfPageUp()
	case "G":
		m.ExecutingViewport.GotoBottom()
	case "g":
		m.ExecutingViewport.GotoTop()
	}
	return m, nil
}

// handleTimerTick updates elapsed time during execution
func (m Model) handleTimerTick(msg timerTickMsg) (tea.Model, tea.Cmd) {
	if m.State != StateExecuting {
		return m, nil
	}
	m.ExecutionElapsed = time.Since(m.ExecutionStartTime)
	var cmd tea.Cmd
	m.Spinner, cmd = m.Spinner.Update(msg)
	return m, tea.Batch(tickTimer(), cmd)
}

// handleSpinnerTick handles spinner animation ticks
func (m Model) handleSpinnerTick(msg spinner.TickMsg) (tea.Model, tea.Cmd) {
	if m.State != StateExecuting {
		return m, nil
	}
	var cmd tea.Cmd
	m.Spinner, cmd = m.Spinner.Update(msg)
	return m, cmd
}

// handleOutputChunk processes a chunk of streaming output
func (m Model) handleOutputChunk(msg outputChunkMsg) (tea.Model, tea.Cmd) {
	if msg.done {
		return m.handleExecutionComplete(msg.err)
	}
	m.StreamingOutput.WriteString(msg.chunk)
	m.ExecutingViewport.SetContent(m.StreamingOutput.String())
	m.ExecutingViewport.GotoBottom()
	return m, waitForChunk(m.OutputChunks)
}

// handleExecutionComplete transitions to output state after streaming execution
func (m Model) handleExecutionComplete(err error) (tea.Model, tea.Cmd) {
	// Calculate execution duration
	duration := time.Since(m.ExecutionStartTime)

	// Record execution with timing data
	success := err == nil
	m.History.RecordExecutionWithTiming(m.MakefilePath, m.ExecutingTarget, duration, success)
	_ = m.History.Save() // Async, ignore errors

	// Build result for export
	result := executor.Result{
		Output:    m.StreamingOutput.String(),
		Err:       err,
		Duration:  duration,
		StartTime: m.ExecutionStartTime,
		EndTime:   time.Now(),
	}

	// Extract exit code from error
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = -1
		}
	}

	// Export execution result (async, non-blocking)
	if m.Exporter != nil {
		go func() {
			record := export.NewExecutionRecord(
				m.MakefilePath,
				m.ExecutingTarget,
				result,
			)
			if err := m.Exporter.Export(record); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Export failed: %v\n", err)
			}
		}()
	}

	// Shell integration (async, non-blocking)
	if m.ShellIntegration != nil {
		go func() {
			if err := m.ShellIntegration.RecordExecution(shell.ExecutionInfo{
			Target:       m.ExecutingTarget,
			MakefilePath: m.MakefilePath,
		}); err != nil {
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

	// Transition to output view
	m.State = StateOutput
	m.Output = m.StreamingOutput.String()
	m.ExecutionError = err
	m.initViewport(m.Output)

	// Clean up
	m.CancelExecution = nil
	m.OutputChunks = nil

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

// initExecutingViewport initializes viewport for streaming output
func (m *Model) initExecutingViewport() {
	// Calculate available height - must account for ALL UI elements:
	// - Title + 2 newlines: 3 lines
	// - Progress bar + time display: 2 lines
	// - Newline + separator + newline: 3 lines
	// - "Output:" + 2 newlines: 3 lines
	// - Container border (top + bottom): 2 lines
	// - Container padding (top + bottom): 2 lines
	// - Status bar: 2 lines
	// Total overhead: ~17 lines, use 18 for safety
	availableHeight := m.Height - 18
	if availableHeight < 3 {
		availableHeight = 3
	}

	// Calculate content width to match container inner width
	// Container: width-2, border: 2, padding: 4 (2 left + 2 right)
	// Inner content width: (width-2) - 2 - 4 = width - 8
	contentWidth := m.Width - 8
	if contentWidth < 20 {
		contentWidth = 20
	}

	m.ExecutingViewport = viewport.New(contentWidth, availableHeight)
	m.ExecutingViewport.Style = lipgloss.NewStyle()
	m.ExecutingViewport.YPosition = 0
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

				// Reset streaming output and initialize viewport
				m.StreamingOutput = &strings.Builder{}
				m.initExecutingViewport()

				return m, tea.Batch(
					executeTargetStreaming(target.Name, m.MakefilePath),
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

// Custom message for timer ticks
type timerTickMsg struct{}

// streamStartedMsg indicates streaming has begun
type streamStartedMsg struct {
	chunks <-chan executor.OutputChunk
	cancel func()
}

// outputChunkMsg delivers a chunk of output during streaming
type outputChunkMsg struct {
	chunk string
	done  bool
	err   error
}

func tickTimer() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return timerTickMsg{}
	})
}

// executeTargetStreaming starts streaming execution
func executeTargetStreaming(target, makefilePath string) tea.Cmd {
	return func() tea.Msg {
		chunks, cancel := executor.ExecuteStreaming(target, makefilePath)
		return streamStartedMsg{chunks: chunks, cancel: cancel}
	}
}

// waitForChunk waits for next output chunk from channel
func waitForChunk(chunks <-chan executor.OutputChunk) tea.Cmd {
	return func() tea.Msg {
		chunk, ok := <-chunks
		if !ok {
			// Channel closed
			return outputChunkMsg{done: true}
		}
		return outputChunkMsg{
			chunk: chunk.Data,
			done:  chunk.Done,
			err:   chunk.Err,
		}
	}
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
