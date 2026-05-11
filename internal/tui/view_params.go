package tui

import (
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/rshelekhov/lazymake/internal/executor"
	"github.com/rshelekhov/lazymake/internal/util"
)

// renderRunParamsView draws the interactive form for environment
// variables and make flags (issue #37). The layout mirrors the
// dependency-graph view: full-width container with a rounded border
// on top, status bar (workspace path nugget + shortcuts) on the
// bottom.
func (m Model) renderRunParamsView() string {
	if m.Width == 0 || m.Height == 0 {
		return "Loading run parameters..."
	}

	content := m.renderRunParamsContent(m.Width)
	statusBar := m.renderRunParamsStatusBar(m.Width)

	return lipgloss.JoinVertical(lipgloss.Left, content, statusBar)
}

// renderRunParamsContent renders the bordered form area.
func (m Model) renderRunParamsContent(width int) string {
	var builder strings.Builder

	// Title
	title := TitleStyle.Render("Run Parameters")
	util.WriteString(&builder, title+"\n\n")

	// Target name
	if m.ParamsTarget != nil {
		targetInfo := lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true).
			Render("Target: " + m.ParamsTarget.Name)
		util.WriteString(&builder, targetInfo+"\n")

		if m.ParamsTarget.Description != "" {
			desc := lipgloss.NewStyle().
				Foreground(TextMuted).
				Render(m.ParamsTarget.Description)
			util.WriteString(&builder, desc+"\n")
		}
	}
	util.WriteString(&builder, "\n")

	// Inputs
	util.WriteString(&builder, m.ParamsEnvInput.View()+"\n")
	util.WriteString(&builder, m.ParamsFlagsInput.View()+"\n\n")

	// Hint about syntax
	hint := lipgloss.NewStyle().
		Foreground(TextMuted).
		Italic(true).
		Render("Tip: KEY=value pairs in env, quote values with spaces: MSG=\"hi there\"")
	util.WriteString(&builder, hint)

	// Inline validation error, if any
	if m.ParamsError != "" {
		errLine := lipgloss.NewStyle().
			Foreground(ErrorColor).
			Bold(true).
			Render("✗ " + m.ParamsError)
		util.WriteString(&builder, "\n\n"+errLine)
	}

	// Container matches the dependency-graph view: rounded border,
	// full width minus the border itself.
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BorderColor).
		Padding(1, 2).
		Width(width - 2)
	return containerStyle.Render(builder.String())
}

// renderRunParamsStatusBar mirrors the dependency-graph status bar:
// workspace path nugget on the left, shortcuts on the right.
func (m Model) renderRunParamsStatusBar(width int) string {
	statusBarStyle := lipgloss.NewStyle().Foreground(TextPrimary)

	coloredNuggetStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#000000"}).
		Background(PrimaryColor).
		Padding(0, 1).
		MarginRight(1)

	workspacePath := m.getWorkspaceDisplayPath()
	pathNugget := coloredNuggetStyle.Render(workspacePath)

	leftBar := lipgloss.JoinHorizontal(lipgloss.Top, pathNugget)
	leftWidth := lipgloss.Width(leftBar)

	helpText := "enter: run • tab: switch field • esc: cancel • q: quit"
	right := lipgloss.NewStyle().
		Foreground(TextMuted).
		Padding(0, 1).
		Render(helpText)
	rightWidth := lipgloss.Width(right)

	// Account for status bar horizontal padding (1 left + 1 right = 2).
	middleWidth := max(width-2-leftWidth-rightWidth, 1)
	middle := lipgloss.NewStyle().
		Width(middleWidth).
		Align(lipgloss.Left).
		Render("")

	bar := lipgloss.JoinHorizontal(lipgloss.Top, leftBar, middle, right)

	return statusBarStyle.
		Width(width).
		Padding(1, 1).
		Render(bar)
}

// renderParamsLine formats env+flags as a single-line header for the
// output view. Returns "" when opts carries no params, so callers can
// omit the line entirely on plain runs.
//
// Example: `Params: ENV=prod TOKEN="ab cd"  Flags: -j4 -k`
func renderParamsLine(opts executor.ExecutionOptions) string {
	if len(opts.Env) == 0 && len(opts.Flags) == 0 {
		return ""
	}
	labelStyle := lipgloss.NewStyle().Foreground(TextSecondary).Bold(true)
	valueStyle := lipgloss.NewStyle().Foreground(TextMuted)

	var parts []string
	if len(opts.Env) > 0 {
		keys := make([]string, 0, len(opts.Env))
		for k := range opts.Env {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		pairs := make([]string, 0, len(keys))
		for _, k := range keys {
			v := opts.Env[k]
			if strings.ContainsAny(v, " \t") {
				v = `"` + v + `"`
			}
			pairs = append(pairs, k+"="+v)
		}
		parts = append(parts, labelStyle.Render("Env: ")+valueStyle.Render(strings.Join(pairs, " ")))
	}
	if len(opts.Flags) > 0 {
		parts = append(parts, labelStyle.Render("Flags: ")+valueStyle.Render(strings.Join(opts.Flags, " ")))
	}
	return strings.Join(parts, "   ")
}
