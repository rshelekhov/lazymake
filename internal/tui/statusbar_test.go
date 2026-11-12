package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestNuggetHelpersIncludeText(t *testing.T) {
	cases := []struct {
		name string
		fn   func(string) string
		text string
	}{
		{"colored", coloredNugget, "demo/path"},
		{"plain", plainNugget, "12 targets"},
		{"warn", warnNugget, "3 regressed"},
		{"dryRun", dryRunNugget, "DRY-RUN"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out := c.fn(c.text)
			if !strings.Contains(out, c.text) {
				t.Errorf("expected nugget output to contain %q, got %q", c.text, out)
			}
			if lipgloss.Width(out) == 0 {
				t.Errorf("expected non-zero width for %q nugget, got 0", c.name)
			}
		})
	}
}

func TestAssembleStatusBarRespectsWidth(t *testing.T) {
	cases := []struct {
		name  string
		width int
		left  string
		help  string
	}{
		{"narrow with left", 60, coloredNugget("workspace"), "enter: run • q: quit"},
		{"narrow no left", 60, "", "esc: return • q: quit"},
		{"wide with multiple nuggets", 140, coloredNugget("workspace") + plainNugget("12 targets"), "enter: run • /: filter • ?: keys • q: quit"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out := assembleStatusBar(c.width, c.left, c.help)
			if got := lipgloss.Width(out); got != c.width {
				t.Errorf("expected width %d, got %d for %q", c.width, got, c.name)
			}
			if !strings.Contains(out, c.help) {
				t.Errorf("expected help text %q to be present in bar", c.help)
			}
		})
	}
}

// At very narrow widths the left nugget and the help text cannot both
// fit, and lipgloss will truncate. We don't try to assert what
// survives — only that the rendered bar still has the requested width
// (i.e. assembleStatusBar does not blow up or produce a wider bar
// than its caller asked for).
func TestAssembleStatusBarClampsAtTinyWidth(t *testing.T) {
	out := assembleStatusBar(20, coloredNugget("workspace"), "q: quit")
	if got := lipgloss.Width(out); got != 20 {
		t.Errorf("expected width 20 at tiny size, got %d", got)
	}
}

func TestAssembleStatusBarHelpIsRightAligned(t *testing.T) {
	// Verify the help text appears physically to the right of the left
	// section: its byte index in the rendered string must come after
	// any non-empty left content.
	left := coloredNugget("workspace")
	help := "enter: run • ?: keys"
	out := assembleStatusBar(100, left, help)

	leftIdx := strings.Index(out, "workspace")
	helpIdx := strings.Index(out, "enter: run")
	if leftIdx < 0 || helpIdx < 0 {
		t.Fatalf("expected both workspace nugget and help text to be present; out=%q", out)
	}
	if helpIdx <= leftIdx {
		t.Errorf("expected help text to render after left nugget; leftIdx=%d helpIdx=%d", leftIdx, helpIdx)
	}
}
