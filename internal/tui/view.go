package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/rshelekhov/lazymake/internal/graph"
	"github.com/rshelekhov/lazymake/internal/history"
	"github.com/rshelekhov/lazymake/internal/makefile"
	"github.com/rshelekhov/lazymake/internal/util"
)

func (m Model) View() string {
	if m.Err != nil {
		return m.renderErrorView()
	}

	switch m.State {
	case StateExecuting:
		return m.renderExecutingView()
	case StateHelp:
		return m.renderHelpView()
	case StateGraph:
		return m.renderGraphView()
	case StateOutput:
		return m.renderOutputView()
	case StateConfirmDangerous:
		return m.renderConfirmView()
	case StateVariables:
		return m.renderVariablesView()
	case StateWorkspace:
		return m.renderWorkspaceView()
	case StateList:
		return m.renderListView()
	default:
		return lipgloss.NewStyle().Margin(1, 2).Render(m.List.View())
	}
}

// renderErrorView displays error message
func (m Model) renderErrorView() string {
	contentWidth := getContentWidth(m.Width)
	// Calculate inner width for text: subtract border (2) + padding (4) = 6
	innerWidth := max(contentWidth-6, 20)

	// errorStyle := lipgloss.NewStyle().
	// 	Foreground(ErrorColor).
	// 	Bold(true).
	// 	Border(lipgloss.RoundedBorder()).
	// 	BorderForeground(ErrorColor).
	// 	Padding(1, 2).
	// 	Width(innerWidth)
	//
	// return "\n" + errorStyle.Render("Error: "+m.Err.Error()) + "\n\n  Press q to quit\n"

	errorMsg := lipgloss.NewStyle().
		Foreground(ErrorColor).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ErrorColor).
		Padding(1, 2).
		Width(innerWidth).
		Render("Error: " + m.Err.Error())

	content := "\n" + errorMsg + "\n\n  Press q to quit\n"

	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(SecondaryColor).
		Padding(1, 2).
		Width(contentWidth - 2)

	return containerStyle.Render(content)
}

// renderHelpView displays all targets with their descriptions
func (m Model) renderHelpView() string {
	// Build the help content
	var helpContent string

	// Title
	title := TitleStyle.Render("Makefile Targets - Help")
	helpContent += title + "\n\n"

	// Description
	desc := lipgloss.NewStyle().
		Foreground(TextMuted).
		Render("Available targets:\n")
	helpContent += desc + "\n"

	// List all targets with descriptions
	if len(m.Targets) == 0 {
		helpContent += lipgloss.NewStyle().
			Foreground(TextMuted).
			Render("  No targets found\n")
	} else {
		// Find the longest target name for alignment
		maxNameLen := 0
		for _, target := range m.Targets {
			if len(target.Name) > maxNameLen {
				maxNameLen = len(target.Name)
			}
		}

		// Render each target with aligned descriptions
		for _, target := range m.Targets {
			// Target name with color
			targetName := lipgloss.NewStyle().
				Foreground(PrimaryColor).
				Bold(true).
				Render(target.Name)

			// Calculate padding to align descriptions
			padding := maxNameLen - len(target.Name) + 2 // +2 for spacing

			targetLine := "  " + targetName

			// Add description with appropriate style and alignment
			if target.Description != "" {
				var descStyle lipgloss.Style
				if target.CommentType == makefile.CommentDouble {
					descStyle = DocDescriptionStyle
				} else {
					descStyle = DescriptionStyle
				}
				// Add padding spaces before description
				for range padding {
					targetLine += " "
				}
				targetLine += descStyle.Render(target.Description)
			}

			helpContent += targetLine + "\n"
		}
	}

	// Legend - use plain formatting to avoid lipgloss layout issues
	helpContent += "\n"
	helpContent += "Legend:\n"
	helpContent += "  " + lipgloss.NewStyle().Foreground(TextSecondary).Render("Gray") + " = ## documented target (recommended)\n"
	helpContent += "  " + lipgloss.NewStyle().Foreground(TextMuted).Render("Gray Dark") + " = # regular comment\n"

	// Wrap in a container with padding (no width constraint to avoid layout issues)
	containerStyle := lipgloss.NewStyle().
		Padding(1, 2)

	content := containerStyle.Render(helpContent)

	// Status bar with lipgloss-style nuggets
	statusBarStyle := lipgloss.NewStyle().
		Foreground(TextSecondary).
		Background(BackgroundSubtle)

	helpText := "?: toggle help • g: graph • esc: return • q: quit"
	rightStyle := lipgloss.NewStyle().
		Foreground(TextMuted).
		Padding(0, 1).
		Align(lipgloss.Right)

	bar := statusBarStyle.
		Width(m.Width).
		Render(rightStyle.Render(helpText))

	return content + "\n" + bar
}

// renderGraphView displays the dependency graph for the selected target
func (m Model) renderGraphView() string {
	if m.Width == 0 || m.Height == 0 {
		// Fallback for zero dimensions
		return "Loading graph..."
	}

	graphContent := m.renderGraphContent(m.Width)

	statusBar := m.renderGraphStatusBar(m.Width)

	return lipgloss.JoinVertical(lipgloss.Left, graphContent, statusBar)
}

// renderGraphContent renders the main graph content with border
func (m Model) renderGraphContent(width int) string {
	var builder strings.Builder

	// Title
	title := TitleStyle.Render("Dependency Graph")
	util.WriteString(&builder, title+"\n\n")

	// Target info (if specific target selected)
	if m.GraphTarget != "" {
		targetInfo := lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true).
			Render("Target: " + m.GraphTarget)
		util.WriteString(&builder, targetInfo+"\n")
	}

	// Depth info
	depthStr := "all levels"
	if m.GraphDepth >= 0 {
		depthStr = fmt.Sprintf("%d level(s)", m.GraphDepth+1)
	}
	depthInfo := lipgloss.NewStyle().
		Foreground(TextMuted).
		Render(fmt.Sprintf("Depth: %s", depthStr))
	util.WriteString(&builder, depthInfo+"\n\n")

	// Render tree
	var graphToRender *graph.Graph
	if m.GraphTarget != "" && m.Graph.Nodes[m.GraphTarget] != nil {
		graphToRender = m.Graph.GetSubgraph(m.GraphTarget, m.GraphDepth)
	} else {
		graphToRender = m.Graph
	}

	renderer := graph.TreeRenderer{
		ShowOrder:    m.ShowOrder,
		ShowCritical: m.ShowCritical,
		ShowParallel: m.ShowParallel,
		// Add color formatting functions
		FormatOrder: func(s string) string {
			return lipgloss.NewStyle().Foreground(SecondaryColor).Bold(true).Render(s)
		},
		FormatCritical: func(s string) string {
			return lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")).Bold(true).Render(s)
		},
		FormatParallel: func(s string) string {
			return lipgloss.NewStyle().Foreground(SuccessColor).Bold(true).Render(s)
		},
	}

	treeStr := graphToRender.RenderTree(renderer)
	util.WriteString(&builder, treeStr)

	// Add legend if any annotations are enabled
	if m.ShowOrder || m.ShowCritical || m.ShowParallel {
		// Separator line
		separator := lipgloss.NewStyle().
			Foreground(BorderColor).
			Render(strings.Repeat("─", width-8)) // Account for border and padding
		util.WriteString(&builder, "\n"+separator+"\n\n")

		// Build legend items
		if m.ShowOrder {
			item := lipgloss.NewStyle().
				Foreground(SecondaryColor).
				Bold(true).
				Render("[N]") +
				"  " +
				lipgloss.NewStyle().
					Foreground(TextPrimary).
					Render("Execution Order")
			util.WriteString(&builder, item+"\n")
		}

		if m.ShowCritical {
			item := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFD700")). // Gold
				Bold(true).
				Render(" ★ ") +
				"  " +
				lipgloss.NewStyle().
					Foreground(TextPrimary).
					Render("Critical Path")
			util.WriteString(&builder, item+"\n")
		}

		if m.ShowParallel {
			item := lipgloss.NewStyle().
				Foreground(SuccessColor). // Green
				Bold(true).
				Render("|| ") +
				"  " +
				lipgloss.NewStyle().
					Foreground(TextPrimary).
					Render("Parallel Execution")
			util.WriteString(&builder, item)
		}
	}

	// Apply border (matching main view pattern)
	// Set width to full terminal width minus border
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BorderColor).
		Padding(1, 2).
		Width(width - 2) // Account for border (2)

	return containerStyle.Render(builder.String())
}

// renderGraphStatusBar renders the keyboard controls in status bar format
func (m Model) renderGraphStatusBar(width int) string {
	// Base status bar style - with background for entire bar
	statusBarStyle := lipgloss.NewStyle().
		Foreground(TextPrimary)

	// Colored nugget style (only for first item)
	coloredNuggetStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#000000"}).
		Background(PrimaryColor).
		Padding(0, 1).
		MarginRight(1)

	// Workspace path nugget (only colored one)
	workspacePath := m.getWorkspaceDisplayPath()
	pathNugget := coloredNuggetStyle.Render(workspacePath)

	var sections []string
	sections = append(sections, pathNugget)

	// Calculate width used by nuggets
	leftBar := lipgloss.JoinHorizontal(lipgloss.Top, sections...)
	leftWidth := lipgloss.Width(leftBar)

	// Right side: shortcuts
	helpText := "g/esc: return • +/-: depth • o: order • c: critical • p: parallel • q: quit"

	// Right section with help text
	right := lipgloss.NewStyle().
		Foreground(TextMuted).
		Padding(0, 1).
		Render(helpText)
	rightWidth := lipgloss.Width(right)

	// Middle section fills remaining space
	// Account for status bar horizontal padding (2 chars: 1 left + 1 right)
	middleWidth := max(width-2-leftWidth-rightWidth, 1)
	middle := lipgloss.NewStyle().
		Width(middleWidth).
		Align(lipgloss.Left).
		Render("")

	// Combine all sections
	bar := lipgloss.JoinHorizontal(lipgloss.Top, leftBar, middle, right)

	return statusBarStyle.
		Width(width).
		Padding(1, 1).
		Render(bar)
}

// renderOutputView displays output of the executed target
func (m Model) renderOutputView() string {
	var builder strings.Builder

	// Header inside the box
	var header string
	if m.ExecutionError != nil {
		header = ErrorStyle.Render("❌ Failed: make " + m.ExecutingTarget)
	} else {
		header = SuccessStyle.Render("✓ Success: make " + m.ExecutingTarget)
	}
	util.WriteString(&builder, header+"\n")

	// Check for performance regression
	for _, target := range m.Targets {
		if target.Name == m.ExecutingTarget && target.PerfStats != nil && target.PerfStats.IsRegressed {
			stats := target.PerfStats
			change := int(((float64(stats.LastDuration) - float64(stats.AvgDuration)) / float64(stats.AvgDuration)) * 100)

			alertStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("220")).
				Bold(true)

			regressionAlert := alertStyle.Render(fmt.Sprintf("⚠️  This run took %s (%d%% slower than usual avg %s)",
				formatDuration(stats.LastDuration),
				change,
				formatDuration(stats.AvgDuration)))
			util.WriteString(&builder, "\n"+regressionAlert+"\n")
			break
		}
	}

	// Add viewport content
	util.WriteString(&builder, "\n"+m.Viewport.View())

	// contentWidth := getContentWidth(m.Width)
	// Calculate inner width for text: subtract border (2) + padding (4) = 6
	// innerWidth := max(contentWidth-6, 20)

	contentWidth := m.Width - 2

	viewportStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(SecondaryColor).
		Padding(1, 2).
		// Width(innerWidth)
		Width(contentWidth)

	// Status bar with lipgloss-style nuggets
	statusBarStyle := lipgloss.NewStyle().
		Foreground(TextSecondary).
		Background(BackgroundSubtle)

	helpText := "esc: return • q: quit"
	rightStyle := lipgloss.NewStyle().
		Foreground(TextMuted).
		Padding(0, 1).
		Align(lipgloss.Right)

	statusBar := statusBarStyle.
		Width(m.Width).
		Render(rightStyle.Render(helpText))

	return "\n" + viewportStyle.Render(builder.String()) + "\n" + statusBar
}

// getContentWidth calculates responsive width for content blocks
// Uses 90% of terminal width with min/max constraints
func getContentWidth(terminalWidth int) int {
	width := min(max(int(float64(terminalWidth)*0.9), 40), 120)
	return width
}

// renderExecutingView renders the execution screen with real-time timer
func (m Model) renderExecutingView() string {
	elapsed := m.ExecutionElapsed
	width := m.Width

	// Get performance stats for the executing target
	var stats *history.PerformanceStats
	for _, target := range m.Targets {
		if target.Name == m.ExecutingTarget {
			stats = target.PerfStats
			break
		}
	}

	var builder strings.Builder

	// Title with spinner
	title := lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true).
		Render(m.Spinner.View() + " Executing: make " + m.ExecutingTarget)
	util.WriteString(&builder, "\n"+title+"\n\n")

	// Progress bar (if we have avg duration to estimate)
	if stats != nil && stats.AvgDuration > 0 {
		// Calculate progress percentage
		progress := float64(elapsed) / float64(stats.AvgDuration)
		if progress > 1.0 {
			progress = 1.0
		}

		// Render progress bar
		progressBar := m.Progress.ViewAs(progress)

		// Time display
		timeStyle := lipgloss.NewStyle().
			Foreground(TextSecondary)
		timeDisplay := timeStyle.Render(
			fmt.Sprintf("  %s / ~%s avg",
				formatDuration(elapsed),
				formatDuration(stats.AvgDuration)))

		util.WriteString(&builder, "  "+progressBar+"\n")
		util.WriteString(&builder, "  "+progressBar+"\n")
		util.WriteString(&builder, timeDisplay+"\n\n")
	} else {
		// Simple elapsed time
		timeStyle := lipgloss.NewStyle().
			Foreground(TextSecondary).
			Render("  Elapsed: " + formatDuration(elapsed))
		util.WriteString(&builder, timeStyle+"\n\n")
	}

	// Wait message
	waitMsg := lipgloss.NewStyle().
		Foreground(TextMuted).
		Italic(true).
		Render("  Please wait...")
	util.WriteString(&builder, waitMsg+"\n")

	// Modern container with accent border and subtle background
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BorderAccent).
		Background(BackgroundSubtle).
		Padding(2, 3).
		Width(width - 4)

	return containerStyle.Render(builder.String())
}
