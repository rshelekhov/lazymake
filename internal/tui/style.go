package tui

import "github.com/charmbracelet/lipgloss"

// Color palette
var (
	PrimaryColor   = lipgloss.Color("205") // Pink
	SecondaryColor = lipgloss.Color("86")  // Cyan
	SuccessColor   = lipgloss.Color("42")  // Green
	ErrorColor     = lipgloss.Color("196") // Red
	MutedColor     = lipgloss.Color("241") // Gray
	TextColor      = lipgloss.Color("252") // Light gray
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

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ErrorColor).
			Bold(true).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ErrorColor).
			Padding(1, 2)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(SuccessColor).
			Bold(true)
)
