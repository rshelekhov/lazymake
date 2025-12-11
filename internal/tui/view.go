package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/rshelekhov/lazymake/internal/graph"
	"github.com/rshelekhov/lazymake/internal/makefile"
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
	var builder strings.Builder

	title := TitleStyle.Render("Dependency Graph")
	writeString(&builder, title+"\n\n")

	if m.GraphTarget != "" {
		targetInfo := lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true).
			Render("Target: " + m.GraphTarget)
		writeString(&builder, targetInfo+"\n")
	}

	depthStr := "all levels"
	if m.GraphDepth >= 0 {
		depthStr = fmt.Sprintf("%d level(s)", m.GraphDepth+1)
	}
	depthInfo := lipgloss.NewStyle().
		Foreground(MutedColor).
		Render(fmt.Sprintf("Depth: %s", depthStr))

	writeString(&builder, depthInfo+"\n\n")

	var graphToRender *graph.Graph
	if m.GraphTarget != "" && m.Graph.Nodes[m.GraphTarget] != nil {
		// Show subgraph for specific target
		graphToRender = m.Graph.GetSubgraph(m.GraphTarget, m.GraphDepth)
	} else {
		// Show full graph
		graphToRender = m.Graph
	}

	renderer := graph.TreeRenderer{
		ShowOrder:    m.ShowOrder,
		ShowCritical: m.ShowCritical,
		ShowParallel: m.ShowParallel,
	}

	treeStr := graphToRender.RenderTree(renderer)
	writeString(&builder, treeStr+"\n")

	legend := graph.RenderLegend(m.ShowOrder, m.ShowCritical, m.ShowParallel)
	if legend != "" {
		legendStyled := lipgloss.NewStyle().
			Foreground(MutedColor).
			Render(legend)
		writeString(&builder, legendStyled+"\n")
	}

	writeString(&builder, "\n")
	controls := "g/esc = return • +/- = depth • o = order • c = critical • p = parallel • q = quit"
	controlsStyled := lipgloss.NewStyle().
		Foreground(MutedColor).
		Render(controls)
	writeString(&builder, controlsStyled)

	containerStyle := lipgloss.NewStyle().Padding(1, 2)
	return containerStyle.Render(builder.String())
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

func writeString(b *strings.Builder, s string) {
	_, _ = b.WriteString(s)
}
