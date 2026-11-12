package tui

import "github.com/charmbracelet/lipgloss"

// Status-bar primitives shared across the list, graph, output, params,
// presets, variables and workspace views. Each view composes its left
// side from these nuggets (workspace path, counts, badges) and passes
// the result, together with a context-specific help string, to
// assembleStatusBar, which handles the right-aligned hint and the
// filler in between.

// nuggetWhiteFg is the foreground used by colored nuggets (workspace
// path, DRY-RUN badge). White-on-light may look thin against the
// primary background, so we use the adaptive pair black-in-light /
// white-in-dark.
var nuggetWhiteFg = lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#000000"}

// coloredNugget renders a primary-accent nugget (used for the
// workspace path on most views).
func coloredNugget(text string) string {
	return lipgloss.NewStyle().
		Foreground(nuggetWhiteFg).
		Background(PrimaryColor).
		Padding(0, 1).
		MarginRight(1).
		Render(text)
}

// plainNugget renders a muted text-on-bar nugget (used for counts like
// "12 targets" or "3 used").
func plainNugget(text string) string {
	return lipgloss.NewStyle().
		Foreground(TextSecondary).
		Padding(0, 1).
		Render(text)
}

// warnNugget renders a yellow/warning-toned plain nugget (used for
// "N regressed" counts that should catch the eye without becoming a
// full background block).
func warnNugget(text string) string {
	return lipgloss.NewStyle().
		Foreground(WarningColor).
		Padding(0, 1).
		Render(text)
}

// dryRunNugget renders the DRY-RUN badge that sits at the very left
// of the main status bar when m.DryRun is on. Same structure as the
// colored nugget but with a warning background so it reads as a state
// flag, not a label.
func dryRunNugget(text string) string {
	return lipgloss.NewStyle().
		Foreground(nuggetWhiteFg).
		Background(WarningColor).
		Padding(0, 1).
		MarginRight(1).
		Render(text)
}

// assembleStatusBar combines a pre-rendered left side with a
// right-aligned help text, filling the gap so the bar always spans
// the full terminal width. Callers pass an empty leftRendered when
// they only need the help hint (e.g. the output view).
//
// The 2-cell horizontal padding budget (1 left + 1 right) is reserved
// from the middle filler so the bar's outer padding(1,1) does not
// push content past the width.
func assembleStatusBar(width int, leftRendered, helpText string) string {
	leftWidth := lipgloss.Width(leftRendered)

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

	bar := lipgloss.JoinHorizontal(lipgloss.Top, leftRendered, middle, right)

	return lipgloss.NewStyle().
		Foreground(TextPrimary).
		Width(width).
		Padding(1, 1).
		Render(bar)
}
