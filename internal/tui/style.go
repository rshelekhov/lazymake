package tui

import "github.com/charmbracelet/lipgloss"

// Modern "Slate" color palette - GitHub-inspired, minimalist design
var (
	// Primary accent - vibrant blue (trust, action, focus)
	PrimaryColor = lipgloss.AdaptiveColor{Light: "#0969DA", Dark: "#58A6FF"}

	// Secondary accent - teal/cyan (success, info, highlights)
	SecondaryColor = lipgloss.AdaptiveColor{Light: "#0550AE", Dark: "#79C0FF"}

	// Success - green
	SuccessColor = lipgloss.AdaptiveColor{Light: "#1F883D", Dark: "#3FB950"}

	// Error - red
	ErrorColor = lipgloss.AdaptiveColor{Light: "#CF222E", Dark: "#F85149"}

	// Warning - amber
	WarningColor = lipgloss.AdaptiveColor{Light: "#9A6700", Dark: "#D29922"}

	// Text hierarchy
	TextPrimary = lipgloss.AdaptiveColor{Light: "#1F2328", Dark: "#E6EDF3"}
	TextSecondary = lipgloss.AdaptiveColor{Light: "#656D76", Dark: "#8B949E"}
	TextMuted = lipgloss.AdaptiveColor{Light: "#8C959F", Dark: "#6E7681"}

	// Legacy alias for backward compatibility
	MutedColor = TextMuted
	TextColor = TextPrimary

	// Borders and backgrounds
	BorderColor = lipgloss.AdaptiveColor{Light: "#D0D7DE", Dark: "#30363D"}
	BorderAccent = lipgloss.AdaptiveColor{Light: "#0969DA", Dark: "#58A6FF"}
	BackgroundSubtle = lipgloss.AdaptiveColor{Light: "#F6F8FA", Dark: "#161B22"}
	ShadowColor = lipgloss.AdaptiveColor{Light: "#D0D7DE", Dark: "#010409"}

	// Legacy alias for backward compatibility
	SeparatorColor = BorderColor
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
				Foreground(TextMuted).
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
				Foreground(TextSecondary).
				Bold(true).
				PaddingTop(1).
				PaddingLeft(1)

	// SeparatorStyle is used for the line between sections
	SeparatorStyle = lipgloss.NewStyle().
			Foreground(BorderColor).
			PaddingLeft(1)

	// StatusBarStyle is used for the status bars in different views
	StatusBarStyle = lipgloss.NewStyle().
		Foreground(TextPrimary).
		Background(BackgroundSubtle).
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(BorderColor).
		Padding(1, 2)
)
