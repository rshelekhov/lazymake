package tui

import "github.com/charmbracelet/lipgloss"

// Color palette
var (
	PrimaryColor   = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	SecondaryColor = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	SuccessColor   = lipgloss.Color("42")
	ErrorColor     = lipgloss.AdaptiveColor{Light: "196", Dark: "196"}
	MutedColor     = lipgloss.AdaptiveColor{Light: "241", Dark: "241"}
	SeparatorColor = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	TextColor      = lipgloss.Color("252")
)

// Styles
var (
	TitleStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true).
			Padding(0, 0, 1, 0)

	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(PrimaryColor).
				Bold(true).
				PaddingLeft(1)

	NormalItemStyle = lipgloss.NewStyle().
			Foreground(TextColor).
			PaddingLeft(1)

	DescriptionStyle = lipgloss.NewStyle().
				Foreground(MutedColor).
				PaddingLeft(3)

	// DocDescriptionStyle is used for ## documented comments (industry standard)
	DocDescriptionStyle = lipgloss.NewStyle().
				Foreground(SecondaryColor).
				PaddingLeft(3)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ErrorColor).
			Bold(true).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ErrorColor).
			Padding(1, 2)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(SuccessColor).
			Bold(true)

	// SectionHeaderStyle is used for "RECENT" and "ALL TARGETS" headers
	SectionHeaderStyle = lipgloss.NewStyle().
				Foreground(MutedColor).
				Bold(true).
				PaddingTop(1).
				PaddingLeft(1)

	// SeparatorStyle is used for the line between sections
	SeparatorStyle = lipgloss.NewStyle().
			Foreground(SeparatorColor).
			PaddingLeft(1)

	// StatusBarStyle is used for the status bars in different views
	StatusBarStyle = lipgloss.NewStyle().
		// Border(lipgloss.RoundedBorder()).
		// BorderForeground(SecondaryColor).
		// Foreground(MutedColor).
		Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
		Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"}).
		Padding(0, 1)
)
