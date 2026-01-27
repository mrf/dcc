package ui

import "github.com/charmbracelet/lipgloss"

// Color constants matching the Rust version
var (
	ColorGreen   = lipgloss.Color("#00ff00")
	ColorYellow  = lipgloss.Color("#ffff00")
	ColorOrange  = lipgloss.Color("#ff8c00")
	ColorRed     = lipgloss.Color("#ff0000")
	ColorCyan    = lipgloss.Color("#00ffff")
	ColorMagenta = lipgloss.Color("#ff00ff")
	ColorBlue    = lipgloss.Color("#0087ff")
	ColorWhite   = lipgloss.Color("#ffffff")
	ColorGray    = lipgloss.Color("#808080")
	ColorDarkGray= lipgloss.Color("#404040")
)

// Panel styles
var (
	// Base panel style
	PanelStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			Padding(0, 1)

	// Selected panel has bold border
	SelectedPanelStyle = PanelStyle.
				BorderStyle(lipgloss.ThickBorder())

	// Loading panel has dim border
	LoadingPanelStyle = PanelStyle.
				BorderForeground(ColorDarkGray)
)

// Text styles
var (
	BoldStyle   = lipgloss.NewStyle().Bold(true)
	ItalicStyle = lipgloss.NewStyle().Italic(true)
	DimStyle    = lipgloss.NewStyle().Faint(true)
)

// Status indicator styles
var (
	StatusGreen   = lipgloss.NewStyle().Foreground(ColorGreen)
	StatusYellow  = lipgloss.NewStyle().Foreground(ColorYellow)
	StatusOrange  = lipgloss.NewStyle().Foreground(ColorOrange)
	StatusRed     = lipgloss.NewStyle().Foreground(ColorRed)
	StatusCyan    = lipgloss.NewStyle().Foreground(ColorCyan)
	StatusMagenta = lipgloss.NewStyle().Foreground(ColorMagenta)
	StatusBlue    = lipgloss.NewStyle().Foreground(ColorBlue)
	StatusGray    = lipgloss.NewStyle().Foreground(ColorGray)
)

// Status bar style
var StatusBarStyle = lipgloss.NewStyle().
	Faint(true).
	Padding(0, 1)

// Title styles
var TitleStyle = lipgloss.NewStyle().Bold(true)

// GetPanelStyle returns the appropriate panel style based on selection and loading state
func GetPanelStyle(selected, loading bool, borderColor lipgloss.Color) lipgloss.Style {
	style := PanelStyle.BorderForeground(borderColor)

	if loading {
		style = style.BorderForeground(ColorDarkGray)
	}

	if selected {
		style = style.BorderStyle(lipgloss.ThickBorder())
	}

	return style
}

// AgeColor returns the color for a PR based on its age in days
func AgeColor(days int64) lipgloss.Color {
	switch {
	case days <= 1:
		return ColorGreen
	case days <= 4:
		return ColorYellow
	case days <= 6:
		return ColorOrange
	default:
		return ColorRed
	}
}

// StashAgeColor returns the color for a stash based on its age in days
func StashAgeColor(days int64) lipgloss.Color {
	switch {
	case days < 7:
		return ColorGray
	case days < 30:
		return ColorYellow
	default:
		return ColorRed
	}
}

// PortColor returns the color for a port based on its type
func PortColor(port uint16) lipgloss.Color {
	switch {
	case port == 80 || port == 443 || port == 8080 || port == 8443:
		return ColorGreen // Web servers
	case port == 5432 || port == 3306 || port == 27017 || port == 6379:
		return ColorYellow // Databases
	case port >= 3000 && port <= 3999:
		return ColorCyan // Dev servers
	default:
		return ColorWhite
	}
}

// MeetingStatusColor returns the color based on minutes until next meeting
func MeetingStatusColor(minutes int64, inMeeting bool) lipgloss.Color {
	if inMeeting {
		return ColorBlue
	}
	switch {
	case minutes > 60:
		return ColorGreen
	case minutes > 30:
		return ColorYellow
	case minutes > 10:
		return ColorOrange
	default:
		return ColorRed
	}
}

// GitStatusColor returns the color based on the change type
func GitStatusColor(staged, modified, untracked int) lipgloss.Color {
	if staged > 0 {
		return ColorGreen
	}
	if modified > 0 {
		return ColorYellow
	}
	return ColorCyan
}

// Truncate truncates a string to maxLen, adding "..." if needed
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
