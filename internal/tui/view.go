package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/rshelekhov/lazymake/internal/graph"
	"github.com/rshelekhov/lazymake/internal/makefile"
	"github.com/rshelekhov/lazymake/internal/util"
)

func (m Model) View() string {
	if m.Err != nil {
		return m.renderErrorView()
	}

	switch m.State {
	case StateExecuting:
		return "\n  ⏳ Executing: make " + m.ExecutingTarget + "\n\n  Please wait...\n"
	case StateHelp:
		return m.renderHelpView()
	case StateGraph:
		return m.renderGraphView()
	case StateOutput:
		return m.renderOutputView()
	case StateConfirmDangerous:
		return m.renderConfirmView()
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

	errorStyle := lipgloss.NewStyle().
		Foreground(ErrorColor).
		Bold(true).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ErrorColor).
		Padding(1, 2).
		Width(innerWidth)

	return "\n" + errorStyle.Render("Error: "+m.Err.Error()) + "\n\n  Press q to quit\n"
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
		Foreground(MutedColor).
		Render("Available targets:\n")
	helpContent += desc + "\n"

	// List all targets with descriptions
	if len(m.Targets) == 0 {
		helpContent += lipgloss.NewStyle().
			Foreground(MutedColor).
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
	helpContent += "  " + lipgloss.NewStyle().Foreground(SecondaryColor).Render("Cyan") + " = ## documented target (recommended)\n"
	helpContent += "  " + lipgloss.NewStyle().Foreground(MutedColor).Render("Gray") + " = # regular comment\n"

	// Footer with keyboard shortcuts
	footer := lipgloss.NewStyle().
		Foreground(MutedColor).
		Render("\nPress ? to toggle help • g to view dependency graph • esc to return • q to quit")
	helpContent += footer

	// Wrap in a container with padding (no width constraint to avoid layout issues)
	containerStyle := lipgloss.NewStyle().
		Padding(1, 2)

	return containerStyle.Render(helpContent)
}

// renderGraphView displays the dependency graph for the selected target
func (m Model) renderGraphView() string {
	if m.Width == 0 || m.Height == 0 {
		// Fallback for zero dimensions
		return "Loading graph..."
	}

	graphContent := m.renderGraphContent(m.Width)

	legend := m.renderGraphLegend(m.Width)

	statusBar := m.renderGraphStatusBar(m.Width)

	sections := []string{graphContent}
	if legend != "" {
		sections = append(sections, legend)
	}
	sections = append(sections, statusBar)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
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
		Foreground(MutedColor).
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
	}

	treeStr := graphToRender.RenderTree(renderer)
	util.WriteString(&builder, treeStr)

	// Apply border (matching main view pattern)
	// Set width to full terminal width minus border
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(SecondaryColor).
		Padding(1, 2).
		Width(width - 2) // Account for border (2)

	return containerStyle.Render(builder.String())
}

// renderGraphLegend renders an enhanced legend with color-coded symbols
func (m Model) renderGraphLegend(width int) string {
	if !m.ShowOrder && !m.ShowCritical && !m.ShowParallel {
		return "" // Skip if no annotations enabled
	}

	var builder strings.Builder

	// Build legend items with colors
	var items []string

	if m.ShowOrder {
		item := lipgloss.NewStyle().
			Foreground(SecondaryColor).
			Bold(true).
			Render("[N]") +
			"  " +
			lipgloss.NewStyle().
				Foreground(TextColor).
				Render("Execution Order")
		items = append(items, item)
	}

	if m.ShowCritical {
		item := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700")). // Gold
			Bold(true).
			Render(" ★ ") +
			"  " +
			lipgloss.NewStyle().
				Foreground(TextColor).
				Render("Critical Path")
		items = append(items, item)
	}

	if m.ShowParallel {
		item := lipgloss.NewStyle().
			Foreground(SuccessColor). // Green
			Bold(true).
			Render(" ∥ ") +
			"  " +
			lipgloss.NewStyle().
				Foreground(TextColor).
				Render("Parallel Execution")
		items = append(items, item)
	}

	// Join items
	for _, item := range items {
		util.WriteString(&builder, item+"\n")
	}

	// Apply border
	// Set width to full terminal width minus border
	legendStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(SecondaryColor).
		Padding(1, 2).
		Width(width - 2) // Account for border (2)

	return legendStyle.Render(builder.String())
}

// renderGraphStatusBar renders the keyboard controls in status bar format
func (m Model) renderGraphStatusBar(width int) string {
	// Left side: graph stats
	leftContent := ""
	if m.GraphTarget != "" {
		leftContent = fmt.Sprintf("Target: %s", m.GraphTarget)
	} else if m.Graph != nil {
		nodeCount := len(m.Graph.Nodes)
		leftContent = fmt.Sprintf("%d nodes", nodeCount)
	}

	// Right side: keyboard shortcuts
	rightContent := "g/esc: return • +/-: depth • o: order • c: critical • p: parallel • q: quit"

	// Calculate widths (border=2 + padding=2)
	contentWidth := width - 4
	leftWidth := len(leftContent) + 2
	rightWidth := contentWidth - leftWidth

	// Apply styles
	leftStyle := lipgloss.NewStyle().Foreground(MutedColor)
	rightStyle := lipgloss.NewStyle().Foreground(MutedColor).Align(lipgloss.Right)

	left := leftStyle.Width(leftWidth).Render(leftContent)
	right := rightStyle.Width(rightWidth).Render(rightContent)

	content := left + right

	// Place content (single line height)
	placedContent := lipgloss.Place(
		contentWidth,
		1,
		lipgloss.Left,
		lipgloss.Center,
		content,
	)

	// Wrap in border (matching view_list.go pattern)
	statusBarStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(SecondaryColor).
		Foreground(MutedColor).
		Padding(0, 1)

	return statusBarStyle.Render(placedContent)
}

// renderOutputView displays output of the executed target
func (m Model) renderOutputView() string {
	var header string
	if m.ExecutionError != nil {
		header = ErrorStyle.Render("❌ Failed: make " + m.ExecutingTarget)
	} else {
		header = SuccessStyle.Render("✓ Success: make " + m.ExecutingTarget)
	}

	contentWidth := getContentWidth(m.Width)
	// Calculate inner width for text: subtract border (2) + padding (4) = 6
	innerWidth := max(contentWidth-6, 20)

	viewportStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(SecondaryColor).
		Padding(1, 2).
		Width(innerWidth)

	footer := lipgloss.NewStyle().
		Foreground(MutedColor).
		Render("\nPress esc to return • q to quit")

	return "\n" + header + "\n\n" + viewportStyle.Render(m.Viewport.View()) + footer
}

// getContentWidth calculates responsive width for content blocks
// Uses 90% of terminal width with min/max constraints
func getContentWidth(terminalWidth int) int {
	width := min(max(int(float64(terminalWidth)*0.9), 40), 120)
	return width
}
