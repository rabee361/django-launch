package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Brand Colors
	primaryColor   = lipgloss.Color("#00D787") // Green / Mint
	secondaryColor = lipgloss.Color("#00AFFF") // Sky Blue
	accentColor    = lipgloss.Color("#FF79C6") // Pink
	mutedColor     = lipgloss.Color("#6272A4") // Gray
	errorColor     = lipgloss.Color("#FF5555") // Red
	darkBg         = lipgloss.Color("#1E1E2E")

	// Styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			MarginBottom(1)

	stepBadgeStyle = lipgloss.NewStyle().
			Bold(true).
			Background(secondaryColor).
			Foreground(darkBg).
			Padding(0, 1).
			MarginRight(1)

	stepHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF"))

	selectedItemStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(primaryColor)

	unselectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#E0E0E0"))

	focusedPromptStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true)

	checkedStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	uncheckedStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			MarginTop(1)

	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(mutedColor).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)

	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor)

	errorAlertStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(errorColor)
)
