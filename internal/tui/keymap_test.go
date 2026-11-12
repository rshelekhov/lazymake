package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	"github.com/rshelekhov/lazymake/internal/safety"
)

// newListWithTargets builds a list.Model populated with the given
// targets. Used so the test can drive getHelpText through different
// "has targets?" and "what's selected?" combinations.
func newListWithTargets(targets []Target) list.Model {
	items := make([]list.Item, 0, len(targets))
	for _, t := range targets {
		items = append(items, t)
	}
	return list.New(items, NewItemDelegate(), 80, 20)
}

func TestGetHelpText_SlimDefault(t *testing.T) {
	m := Model{
		List: newListWithTargets([]Target{{Name: "build"}}),
	}

	got := m.getHelpText()
	want := "enter: run • /: filter • ?: keys • q: quit"
	if got != want {
		t.Errorf("slim default mismatch:\n got: %q\nwant: %q", got, want)
	}
}

func TestGetHelpText_FilterMode(t *testing.T) {
	m := Model{
		List:        newListWithTargets([]Target{{Name: "build"}}),
		IsFiltering: true,
	}

	got := m.getHelpText()
	want := "enter: apply • esc: cancel"
	if got != want {
		t.Errorf("filter-mode mismatch:\n got: %q\nwant: %q", got, want)
	}
}

func TestGetHelpText_EmptyListDropsEnter(t *testing.T) {
	// No Target items in the list — the bar must not advertise
	// `enter: run` because there is nothing to run.
	m := Model{
		List: newListWithTargets(nil),
	}

	got := m.getHelpText()
	if strings.Contains(got, "enter") {
		t.Errorf("empty-list help text should not mention enter; got %q", got)
	}
	for _, expected := range []string{"/: filter", "?: keys", "q: quit"} {
		if !strings.Contains(got, expected) {
			t.Errorf("expected %q in empty-list help text, got %q", expected, got)
		}
	}
}

func TestGetHelpText_DangerousVariants(t *testing.T) {
	cases := []struct {
		name     string
		severity safety.Severity
		want     string // substring that must appear
	}{
		{"critical", safety.SeverityCritical, "Critical"},
		{"warning", safety.SeverityWarning, "Warning"},
		{"info", safety.SeverityInfo, "Info"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			target := Target{
				Name:        "deploy",
				IsDangerous: true,
				DangerLevel: c.severity,
			}
			m := Model{List: newListWithTargets([]Target{target})}

			got := m.getHelpText()
			if !strings.Contains(got, c.want) {
				t.Errorf("expected dangerous-%s help text to contain %q, got %q",
					c.name, c.want, got)
			}
			if !strings.Contains(got, "esc: cancel") {
				t.Errorf("expected dangerous-%s help text to offer esc: cancel, got %q",
					c.name, got)
			}
		})
	}
}

func TestGetHelpText_FilterTakesPrecedenceOverDangerous(t *testing.T) {
	// While the filter input is active, dangerous-target hints must
	// not appear — keys other than enter/esc are intercepted.
	target := Target{Name: "drop-db", IsDangerous: true, DangerLevel: safety.SeverityCritical}
	m := Model{
		List:        newListWithTargets([]Target{target}),
		IsFiltering: true,
	}

	got := m.getHelpText()
	if strings.Contains(got, "Critical") {
		t.Errorf("filter mode should suppress dangerous hint; got %q", got)
	}
	if got != "enter: apply • esc: cancel" {
		t.Errorf("expected filter-mode help text, got %q", got)
	}
}

func TestRenderKeymapView_ContainsAllSectionsAndDescriptions(t *testing.T) {
	m := Model{
		Width:  100,
		Height: 30,
		Targets: []Target{
			{Name: "monsterbuild"},
			{Name: "monstertest"},
		},
	}

	out := m.renderKeymapView()

	// Section headers
	for _, section := range []string{"Run", "Navigation", "Other"} {
		if !strings.Contains(out, section) {
			t.Errorf("expected section %q in keymap view, got %q", section, out)
		}
	}

	// Descriptions — these are specific enough that their presence
	// proves the corresponding key entry was rendered.
	for _, desc := range []string{
		"run selected target",
		"run with params",
		"saved presets",
		"rerun last preset",
		"rerun last execution",
		"filter targets",
		"dependency graph",
		"variable inspector",
		"workspace picker",
		"cheat sheet",
		"quit",
	} {
		if !strings.Contains(out, desc) {
			t.Errorf("expected description %q in keymap view, got %q", desc, out)
		}
	}

	// "ctrl+r" key label appears verbatim — sanity check that
	// multi-character key tokens are emitted.
	if !strings.Contains(out, "ctrl+r") {
		t.Errorf("expected key token %q in keymap view, got %q", "ctrl+r", out)
	}

	// Target listing was removed — the keymap view must not leak
	// target names from the model. We use deliberately unusual names
	// that cannot appear inside any description string above.
	for _, name := range []string{"monsterbuild", "monstertest"} {
		if strings.Contains(out, name) {
			t.Errorf("keymap view should not contain target name %q; got %q", name, out)
		}
	}
}
