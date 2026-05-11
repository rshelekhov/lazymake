package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/rshelekhov/lazymake/internal/presets"
	"github.com/rshelekhov/lazymake/internal/util"
)

// renderRunPresetsView draws the saved-presets picker (issue #35).
// Layout mirrors the graph and params views: full-width content with
// a rounded border, status bar pinned to the bottom.
func (m Model) renderRunPresetsView() string {
	if m.Width == 0 || m.Height == 0 {
		return "Loading presets..."
	}

	content := m.renderPresetsContent(m.Width)
	statusBar := m.renderPresetsStatusBar(m.Width)
	return lipgloss.JoinVertical(lipgloss.Left, content, statusBar)
}

// renderPresetsContent renders the bordered list of presets.
func (m Model) renderPresetsContent(width int) string {
	var builder strings.Builder

	title := TitleStyle.Render("Saved Presets")
	util.WriteString(&builder, title+"\n\n")

	if m.PresetsTarget != nil {
		targetInfo := lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true).
			Render("Target: " + m.PresetsTarget.Name)
		util.WriteString(&builder, targetInfo+"\n\n")
	}

	if len(m.PresetsItems) == 0 {
		empty := lipgloss.NewStyle().
			Foreground(TextMuted).
			Italic(true).
			Render("No saved presets for this target yet.")
		hint := lipgloss.NewStyle().
			Foreground(TextMuted).
			Render("Press " + lipgloss.NewStyle().Bold(true).Render("e") +
				" to open the params form, then " +
				lipgloss.NewStyle().Bold(true).Render("Ctrl+S") +
				" to save the current values as a preset.")
		util.WriteString(&builder, empty+"\n\n"+hint+"\n")
	} else {
		util.WriteString(&builder, m.renderPresetsRows()+"\n")
	}

	if m.PresetsSubMode == presetsSubModeDeleteConfirm {
		confirm := lipgloss.NewStyle().
			Foreground(WarningColor).
			Bold(true).
			Render(fmt.Sprintf("Delete preset %q? [y/n]", m.PresetsPendingDelete))
		util.WriteString(&builder, "\n"+confirm)
	}

	if m.PresetsStatus != "" {
		ok := lipgloss.NewStyle().
			Foreground(SuccessColor).
			Bold(true).
			Render("✓ " + m.PresetsStatus)
		util.WriteString(&builder, "\n"+ok)
	}
	if m.PresetsError != "" {
		errLine := lipgloss.NewStyle().
			Foreground(ErrorColor).
			Bold(true).
			Render("✗ " + m.PresetsError)
		util.WriteString(&builder, "\n"+errLine)
	}

	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BorderColor).
		Padding(1, 2).
		Width(width - 2)
	return containerStyle.Render(builder.String())
}

// renderPresetsRows lays out one row per preset. The currently
// last-used preset is marked with ★ so users can pick it out without
// memorizing names.
func (m Model) renderPresetsRows() string {
	var builder strings.Builder

	cursorStyle := lipgloss.NewStyle().Foreground(PrimaryColor).Bold(true)
	nameStyle := lipgloss.NewStyle().Foreground(TextPrimary).Bold(true)
	mutedStyle := lipgloss.NewStyle().Foreground(TextMuted)
	starStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))

	maxNameLen := 0
	for _, p := range m.PresetsItems {
		if l := len(p.Name); l > maxNameLen {
			maxNameLen = l
		}
	}

	lastUsedName := ""
	if m.Presets != nil && m.PresetsTarget != nil {
		if lu, ok := m.Presets.LastUsed(m.MakefilePath, m.PresetsTarget.Name); ok {
			lastUsedName = lu.Name
		}
	}

	for i, p := range m.PresetsItems {
		// Cursor marker
		marker := "  "
		if i == m.PresetsCursor {
			marker = cursorStyle.Render("▶ ")
		}

		// ★ for the last-used preset, blank space otherwise to keep
		// the columns aligned.
		star := "  "
		if p.Name == lastUsedName {
			star = starStyle.Render("★ ")
		}

		// Name column (padded to align summaries).
		name := nameStyle.Render(p.Name)
		padding := strings.Repeat(" ", maxNameLen-len(p.Name)+2)

		// Summary: env (key count) + flags (joined).
		summary := summarizePreset(p)

		util.WriteString(&builder, marker+star+name+padding+mutedStyle.Render(summary)+"\n")
	}
	return builder.String()
}

// summarizePreset returns a one-line description of a preset's
// payload: K env vars, the joined flags, and a relative "updated"
// timestamp. Kept compact so the picker fits comfortably on 80-col
// terminals.
func summarizePreset(p presets.Preset) string {
	parts := []string{}
	if n := len(p.Env); n > 0 {
		parts = append(parts, fmt.Sprintf("env:%d", n))
	}
	if len(p.Flags) > 0 {
		parts = append(parts, "flags: "+strings.Join(p.Flags, " "))
	}
	if len(parts) == 0 {
		parts = append(parts, "no params")
	}
	if !p.UpdatedAt.IsZero() {
		parts = append(parts, "updated "+humanizeTime(p.UpdatedAt))
	}
	return strings.Join(parts, "  •  ")
}

// humanizeTime formats t relative to now, e.g. "2m ago", "3d ago",
// falling back to an ISO date for distant timestamps so the picker
// stays useful long-term.
func humanizeTime(t time.Time) string {
	delta := time.Since(t)
	switch {
	case delta < time.Minute:
		return "just now"
	case delta < time.Hour:
		return fmt.Sprintf("%dm ago", int(delta.Minutes()))
	case delta < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(delta.Hours()))
	case delta < 30*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(delta.Hours()/24))
	default:
		return t.Format("2006-01-02")
	}
}

// renderPresetsStatusBar mirrors the params/graph views: workspace
// nugget on the left, contextual help on the right.
func (m Model) renderPresetsStatusBar(width int) string {
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

	helpText := "enter: run • e: edit • d: delete • esc: cancel • q: quit"
	if m.PresetsSubMode == presetsSubModeDeleteConfirm {
		helpText = "y: confirm • n/esc: cancel • q: quit"
	}
	right := lipgloss.NewStyle().
		Foreground(TextMuted).
		Padding(0, 1).
		Render(helpText)
	rightWidth := lipgloss.Width(right)

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
