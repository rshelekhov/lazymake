package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/rshelekhov/lazymake/internal/history"
	"github.com/rshelekhov/lazymake/internal/safety"
	"github.com/rshelekhov/lazymake/internal/util"
)

// renderConfirmRerunView renders the full-screen confirmation dialog
// shown before the rerun-last shortcut (issue #36) starts execution.
//
// It mirrors the structural conventions of renderConfirmView (title,
// target, body, actions) but tells the user *what is going to run*
// rather than *that the run is dangerous*. When the rerun target is
// flagged as critical-dangerous we also embed the safety details
// inside this dialog so the user only confirms once — see the issue
// thread for why we don't stack StateConfirmRerun and
// StateConfirmDangerous.
func (m Model) renderConfirmRerunView() string {
	if m.PendingTarget == nil {
		return "Error: No pending target"
	}
	target := m.PendingTarget

	var builder strings.Builder

	title := lipgloss.NewStyle().
		Foreground(TextSecondary).
		Bold(true).
		Render("RERUN LAST EXECUTION")
	util.WriteString(&builder, title+"\n\n")

	targetLine := lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true).
		Render("Target: " + target.Name)
	util.WriteString(&builder, targetLine+"\n\n")

	// Surface the previous timestamp/outcome whenever we have it. This
	// is informational only — the env/flags block below is the part
	// that really answers "what is about to run?".
	if m.History != nil {
		if rec, ok := m.History.LastExecution(m.MakefilePath, target.Name); ok {
			util.WriteString(&builder, renderRerunMetaLine(rec)+"\n\n")
		}
	}

	// The plan block always renders, even for the no-params case, so
	// the user always sees a concrete command line and never has to
	// guess whether the shortcut "did anything".
	planBox := renderRerunPlanBox(target.Name, m.CurrentRunOpts.Env, m.CurrentRunOpts.Flags)
	util.WriteString(&builder, planBox+"\n")

	// Danger details: if the target is critical-dangerous we promote
	// the rerun confirm into a combined modal. SafetyMatches come
	// straight from the safety checker so wording stays consistent
	// with renderConfirmView.
	if target.IsDangerous && target.DangerLevel == safety.SeverityCritical && len(target.SafetyMatches) > 0 {
		for _, match := range target.SafetyMatches {
			util.WriteString(&builder, "\n"+renderRerunDangerBox(match)+"\n")
		}
	}

	util.WriteString(&builder, "\n")

	actionsStyle := lipgloss.NewStyle().Foreground(TextSecondary).Align(lipgloss.Left)
	enterAction := lipgloss.NewStyle().Foreground(SuccessColor).Bold(true).Render("[Enter]")
	escAction := lipgloss.NewStyle().Foreground(TextSecondary).Bold(true).Render("[Esc]")
	util.WriteString(&builder, actionsStyle.Render(enterAction+" Rerun     "+escAction+" Cancel"))

	// Dimensions / centering mirror renderConfirmView so the dialog
	// looks like a sibling of the danger modal rather than an outlier.
	contentWidth := min(80, m.Width-10)

	borderColor := BorderColor
	if target.IsDangerous && target.DangerLevel == safety.SeverityCritical {
		borderColor = ErrorColor
	}

	dialog := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(borderColor).
		Padding(2, 4).
		Width(contentWidth).
		Align(lipgloss.Left).
		Render(builder.String())

	verticalPadding := max((m.Height-strings.Count(dialog, "\n"))/2, 0)
	paddingStyle := lipgloss.NewStyle().
		PaddingTop(verticalPadding).
		PaddingLeft((m.Width - contentWidth) / 2)

	return paddingStyle.Render(dialog)
}

// renderRerunMetaLine produces the "Last run: 2h ago • succeeded"
// header. Outcome wording matches the rest of the TUI (success view
// uses SuccessColor, failure uses ErrorColor) so users carry their
// existing color associations from the output view into this modal.
func renderRerunMetaLine(rec history.ExecutionRecord) string {
	when := humanizeTime(rec.Timestamp)
	outcome := "succeeded"
	outcomeColor := SuccessColor
	if !rec.Success {
		outcome = "failed"
		outcomeColor = ErrorColor
	}
	whenStr := lipgloss.NewStyle().Foreground(TextSecondary).Render("Last run: " + when)
	outcomeStr := lipgloss.NewStyle().Foreground(outcomeColor).Bold(true).Render(outcome)
	sep := lipgloss.NewStyle().Foreground(TextMuted).Render(" • ")
	return whenStr + sep + outcomeStr
}

// renderRerunPlanBox renders the boxed command preview shown inside
// the confirm dialog. Building the equivalent `make` invocation makes
// the rerun action explicit: users see the exact env vars and flags
// before they press Enter.
func renderRerunPlanBox(targetName string, env map[string]string, flags []string) string {
	maxWidth := 60

	var content strings.Builder
	header := lipgloss.NewStyle().Foreground(TextSecondary).Bold(true).Render("Plan")
	util.WriteString(&content, header+"\n\n")

	cmd := buildRerunCommandLine(targetName, env, flags)
	cmdLine := lipgloss.NewStyle().Foreground(TextPrimary).Render(wordwrap.String(cmd, maxWidth))
	util.WriteString(&content, cmdLine)

	if len(env) > 0 {
		util.WriteString(&content, "\n\n")
		envLabel := lipgloss.NewStyle().Foreground(TextSecondary).Render("env:")
		util.WriteString(&content, envLabel+" "+formatEnvForDisplay(env))
	}
	if len(flags) > 0 {
		util.WriteString(&content, "\n")
		flagsLabel := lipgloss.NewStyle().Foreground(TextSecondary).Render("flags:")
		util.WriteString(&content, flagsLabel+" "+strings.Join(flags, " "))
	}
	if len(env) == 0 && len(flags) == 0 {
		util.WriteString(&content, "\n\n")
		hint := lipgloss.NewStyle().Foreground(TextMuted).Italic(true).
			Render(IconInfo + " Previous run used no extra env vars or make flags.")
		util.WriteString(&content, hint)
	}

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(BorderColor).
		Padding(1, 2).
		Render(content.String())
}

// buildRerunCommandLine reconstructs the equivalent shell invocation,
// e.g. `FOO=bar make build -j4`. Keys are sorted to keep output
// stable for snapshot-style tests and to avoid the visual jitter that
// Go's map iteration order would otherwise cause across renders.
func buildRerunCommandLine(targetName string, env map[string]string, flags []string) string {
	var parts []string
	if len(env) > 0 {
		keys := make([]string, 0, len(env))
		for k := range env {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			parts = append(parts, fmt.Sprintf("%s=%s", k, env[k]))
		}
	}
	parts = append(parts, "make", targetName)
	parts = append(parts, flags...)
	return strings.Join(parts, " ")
}

// formatEnvForDisplay renders the env map as "KEY=VAL KEY2=VAL2"
// with sorted keys. Distinct from formatEnvRaw because the display
// form intentionally drops the quoting machinery — we're not asking
// the user to round-trip this string through the parser.
func formatEnvForDisplay(env map[string]string) string {
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, env[k]))
	}
	return strings.Join(parts, " ")
}

// renderRerunDangerBox renders the per-safety-match warning embedded
// inside the rerun-confirm dialog when the target is critical. The
// styling is intentionally a softer echo of renderConfirmView's full
// danger modal — we are NOT replacing the safety dialog wholesale,
// just surfacing enough so the user can decide to back out.
func renderRerunDangerBox(match safety.MatchResult) string {
	maxWidth := 60
	var content strings.Builder

	icon := lipgloss.NewStyle().Foreground(ErrorColor).Bold(true).Render(IconDangerCritical)
	badge := lipgloss.NewStyle().Foreground(ErrorColor).Bold(true).Render("critical")
	ruleID := lipgloss.NewStyle().Foreground(TextSecondary).Render(match.Rule.ID)
	header := icon + " " + badge + " " + ruleID
	util.WriteString(&content, header+"\n")

	if match.MatchedLine != "" {
		util.WriteString(&content, "\n")
		util.WriteString(&content, wordwrap.String("Command: "+match.MatchedLine, maxWidth)+"\n")
	}
	if match.Rule.Description != "" {
		util.WriteString(&content, "\n")
		util.WriteString(&content, wordwrap.String(match.Rule.Description, maxWidth))
	}

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(ErrorColor).
		Padding(1, 2).
		Render(content.String())
}
