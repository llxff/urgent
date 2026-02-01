package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ModalModel is a centered overlay modal component.
type ModalModel struct {
	title      string
	content    string
	helpText   string
	confirmed  bool
	cancelled  bool
	isProgress bool // Progress modals don't respond to y/n
}

// NewConfirmModal creates a confirmation modal with y/n options.
func NewConfirmModal(title, message string) ModalModel {
	return ModalModel{
		title:      title,
		content:    message,
		helpText:   "y yes  •  n no",
		isProgress: false,
	}
}

// NewProgressModal creates a progress/info modal without interactive options.
func NewProgressModal(title, message string) ModalModel {
	return ModalModel{
		title:      title,
		content:    message,
		helpText:   "",
		isProgress: true,
	}
}

// Init implements tea.Model.
func (m ModalModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m ModalModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Progress modals don't respond to input
	if m.isProgress {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			m.confirmed = true
			return m, nil
		case "n", "N", "esc":
			m.cancelled = true
			return m, nil
		}
	}

	return m, nil
}

// View implements tea.Model.
func (m ModalModel) View() string {
	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(PrimaryColor).
		MarginBottom(1)
	b.WriteString(titleStyle.Render(m.title))
	b.WriteString("\n\n")

	// Content
	contentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "240", Dark: "250"})
	b.WriteString(contentStyle.Render(m.content))

	// Help text
	if m.helpText != "" {
		b.WriteString("\n\n")
		helpStyle := lipgloss.NewStyle().
			Foreground(MutedColor)
		b.WriteString(helpStyle.Render(m.helpText))
	}

	// Box style
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(PrimaryColor).
		Padding(1, 3)

	return boxStyle.Render(b.String())
}

// RenderOver renders the modal centered over a dimmed background.
func (m ModalModel) RenderOver(background string, screenWidth, screenHeight int) string {
	// Get modal content
	modalView := m.View()

	// Simply center the modal on screen, ignoring background
	// This avoids ANSI escape code issues with string splicing
	return lipgloss.Place(
		screenWidth,
		screenHeight,
		lipgloss.Center,
		lipgloss.Center,
		modalView,
	)
}

// truncateToWidth returns the prefix of a string up to the given width.
func truncateToWidth(s string, width int) string {
	if width <= 0 {
		return ""
	}
	result := ""
	currentWidth := 0
	for _, r := range s {
		runeWidth := lipgloss.Width(string(r))
		if currentWidth+runeWidth > width {
			break
		}
		result += string(r)
		currentWidth += runeWidth
	}
	// Pad with spaces if needed
	for currentWidth < width {
		result += " "
		currentWidth++
	}
	return result
}

// substringFromWidth returns the suffix of a string starting from the given width.
func substringFromWidth(s string, startWidth int) string {
	currentWidth := 0
	for i, r := range s {
		if currentWidth >= startWidth {
			return s[i:]
		}
		currentWidth += lipgloss.Width(string(r))
	}
	return ""
}

// IsConfirmed returns whether the user confirmed the modal.
func (m ModalModel) IsConfirmed() bool {
	return m.confirmed
}

// IsCancelled returns whether the user cancelled the modal.
func (m ModalModel) IsCancelled() bool {
	return m.cancelled
}

// dimBackground applies a dimming effect to the background content.
func dimBackground(content string) string {
	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"})

	lines := strings.Split(content, "\n")
	var dimmed strings.Builder

	for i, line := range lines {
		dimmed.WriteString(dimStyle.Render(line))
		if i < len(lines)-1 {
			dimmed.WriteString("\n")
		}
	}

	return dimmed.String()
}
