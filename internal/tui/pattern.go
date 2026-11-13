package tui

import (
	"errors"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rshelekhov/lazymake/internal/util"
)

// Pattern rules (`build-%`, `%.o`) are listed like any other target, but
// `make build-%` is not a valid invocation: Make only expands `%` when it
// matches a concrete goal such as `build-foo`. Before running a pattern
// target, or opening params/presets for it, we ask for the stem,
// substitute it, and continue with the resulting concrete target
// through the normal flow (safety checks, params, history).

// stemAction is what happens after the stem is entered.
type stemAction int

const (
	stemActionRun     stemAction = iota // Enter from the list: run it
	stemActionParams                    // "e": open the params form
	stemActionPresets                   // "p": open the presets picker
)

// resolvePatternName substitutes stem for the first `%` in pattern, the
// same way Make derives a goal from a pattern rule's target.
func resolvePatternName(pattern, stem string) string {
	return strings.Replace(pattern, "%", stem, 1)
}

// validateStem rejects stems that would produce something other than a
// single concrete goal on the make command line.
func validateStem(stem string) error {
	switch {
	case stem == "":
		return errors.New("stem must not be empty")
	case strings.ContainsAny(stem, " \t"):
		return errors.New("stem must not contain whitespace")
	case strings.Contains(stem, "%"):
		return errors.New("stem must not contain %")
	}
	return nil
}

// openStemPrompt switches to the stem prompt for a pattern target,
// pre-filling the stem used last time for the same pattern.
func (m Model) openStemPrompt(target Target, next stemAction) Model {
	input := textinput.New()
	input.Placeholder = "e.g. foo"
	input.Prompt = "STEM  > "
	input.CharLimit = 256
	input.Width = stemInputWidth(m.Width)
	if prev, ok := m.LastStems[target.Name]; ok {
		input.SetValue(prev)
	}
	input.Focus()

	targetCopy := target
	m.StemInput = input
	m.StemTarget = &targetCopy
	m.StemNext = next
	m.StemError = ""
	m.State = StatePatternStem
	return m
}

func stemInputWidth(termWidth int) int {
	// Same chrome budget as the params form: border + padding + prompt.
	if termWidth > 14 {
		return termWidth - 14
	}
	return 60
}

// updatePatternStem handles input while the stem prompt is open.
func (m Model) updatePatternStem(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			return m.cancelStemPrompt(), nil
		case tea.KeyEnter:
			return m.submitStemPrompt()
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.StemInput.Width = stemInputWidth(m.Width)
		return m, nil
	}

	var cmd tea.Cmd
	m.StemInput, cmd = m.StemInput.Update(msg)
	m.StemError = ""
	return m, cmd
}

// cancelStemPrompt returns to the list without side effects.
func (m Model) cancelStemPrompt() Model {
	m.State = StateList
	m.StemTarget = nil
	m.StemError = ""
	return m
}

// submitStemPrompt validates the stem, builds the concrete target and
// hands it to the flow the prompt was opened for.
func (m Model) submitStemPrompt() (tea.Model, tea.Cmd) {
	if m.StemTarget == nil {
		return m.cancelStemPrompt(), nil
	}

	stem := strings.TrimSpace(m.StemInput.Value())
	if err := validateStem(stem); err != nil {
		m.StemError = err.Error()
		return m, nil
	}

	pattern := *m.StemTarget
	if m.LastStems == nil {
		m.LastStems = make(map[string]string)
	}
	m.LastStems[pattern.Name] = stem

	concrete := pattern
	concrete.Name = resolvePatternName(pattern.Name, stem)
	concrete.IsPatternRule = false

	m.StemTarget = nil
	m.StemError = ""
	m.State = StateList

	switch m.StemNext {
	case stemActionParams:
		m.initParamsForm(&concrete)
		m.State = StateRunParams
		return m, textinput.Blink
	case stemActionPresets:
		return m.openPresetsPicker(&concrete, StateList), nil
	default:
		return m.runTarget(concrete)
	}
}

// renderPatternStemView draws the stem prompt in the same bordered
// layout as the params form, with a live preview of the make goal.
func (m Model) renderPatternStemView() string {
	if m.Width == 0 || m.Height == 0 {
		return "Loading..."
	}

	var builder strings.Builder
	util.WriteString(&builder, TitleStyle.Render("Pattern Target")+"\n\n")

	if m.StemTarget != nil {
		targetInfo := lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true).
			Render("Target: " + m.StemTarget.Name)
		util.WriteString(&builder, targetInfo+"\n")
		if m.StemTarget.Description != "" {
			desc := lipgloss.NewStyle().Foreground(TextMuted).Render(m.StemTarget.Description)
			util.WriteString(&builder, desc+"\n")
		}
	}
	util.WriteString(&builder, "\n"+m.StemInput.View()+"\n\n")

	hint := "Enter the value that replaces % in the target name"
	if m.StemTarget != nil {
		if stem := strings.TrimSpace(m.StemInput.Value()); validateStem(stem) == nil {
			hint = "Runs: make " + resolvePatternName(m.StemTarget.Name, stem)
		}
	}
	util.WriteString(&builder, lipgloss.NewStyle().Foreground(TextMuted).Italic(true).Render(hint))

	if m.StemError != "" {
		errLine := lipgloss.NewStyle().
			Foreground(ErrorColor).
			Bold(true).
			Render("✗ " + m.StemError)
		util.WriteString(&builder, "\n\n"+errLine)
	}

	content := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BorderColor).
		Padding(1, 2).
		Width(m.Width - 2).
		Render(builder.String())

	helpText := "enter: run • esc: cancel"
	if m.StemNext != stemActionRun {
		helpText = "enter: continue • esc: cancel"
	}
	statusBar := assembleStatusBar(m.Width, coloredNugget(m.getWorkspaceDisplayPath()), helpText)

	return lipgloss.JoinVertical(lipgloss.Left, content, statusBar)
}
