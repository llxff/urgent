package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

var (
	// PrimaryColor is used for interactive elements and highlights.
	PrimaryColor = lipgloss.AdaptiveColor{
		Light: "63", // Blue
		Dark:  "63",
	}

	// SuccessColor is used for successful operations.
	SuccessColor = lipgloss.AdaptiveColor{
		Light: "42", // Green
		Dark:  "42",
	}

	// ErrorColor is used for errors.
	ErrorColor = lipgloss.AdaptiveColor{
		Light: "160", // Red
		Dark:  "160",
	}

	// MutedColor is used for secondary text.
	MutedColor = lipgloss.AdaptiveColor{
		Light: "240", // Gray
		Dark:  "240",
	}
)

func init() {
	// Set the color profile based on terminal capabilities
	lipgloss.SetColorProfile(termenv.ColorProfile())
}

// TitleStyle returns a style for titles.
func TitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(PrimaryColor).
		MarginTop(1).
		MarginBottom(1)
}

// BoxStyle returns a style for bordered boxes.
func BoxStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(PrimaryColor).
		Padding(1, 2)
}

// ErrorStyle returns a style for error messages.
func ErrorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(ErrorColor).
		Bold(true).
		MarginTop(1)
}

// SuccessStyle returns a style for success messages.
func SuccessStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(SuccessColor).
		Bold(true).
		MarginTop(1)
}

// MutedStyle returns a style for muted text.
func MutedStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(MutedColor)
}

// HelpStyle returns a style for help text.
func HelpStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(MutedColor).
		MarginTop(1)
}
