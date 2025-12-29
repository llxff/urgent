package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AccountInfo holds information about a calendar account.
type AccountInfo struct {
	Email            string
	EnabledCount     int
	TotalCount       int
	EnabledCalendars []string // Names of enabled calendars
}

// AccountSelectorModel is a polished account selection screen.
type AccountSelectorModel struct {
	accounts      []AccountInfo
	cursor        int
	width         int
	height        int
	confirmed     bool
	cancelled     bool
	showInfoPanel bool
}

// NewAccountSelectorModel creates a new account selector.
func NewAccountSelectorModel(accounts []AccountInfo) AccountSelectorModel {
	return AccountSelectorModel{
		accounts:      accounts,
		cursor:        0,
		showInfoPanel: true,
	}
}

func (m AccountSelectorModel) Init() tea.Cmd {
	return nil
}

func (m AccountSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Show info panel on wide screens
		m.showInfoPanel = msg.Width > 100
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c"))):
			m.cancelled = true
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			m.cancelled = true
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			m.confirmed = true
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			if m.cursor < len(m.accounts)-1 {
				m.cursor++
			}
			return m, nil
		}
	}

	return m, nil
}

func (m AccountSelectorModel) View() string {
	if m.width == 0 {
		return LoadingSpinner() + " Initializing..."
	}

	// Top bar
	topBar := RenderTopBar(m.width, "", "Accounts")

	// Content
	content := m.renderContent()

	// Footer
	helpKeys := map[string]string{
		"enter": "select",
		"↑↓":    "navigate",
		"esc":   "cancel",
	}
	footer := RenderFooter(m.width, FormatHelpKeys(helpKeys), "")

	frame := NewFrame(m.width, m.height)
	return frame.Render(topBar, content, footer)
}

func (m AccountSelectorModel) renderContent() string {
	// Calculate layout - use ~60/40 split on wide screens
	mainWidth := m.width
	var sideWidth int
	if m.showInfoPanel {
		// Allocate 40% to side panel, but min 35 and max 50
		sideWidth = m.width * 40 / 100
		if sideWidth < 35 {
			sideWidth = 35
		}
		if sideWidth > 50 {
			sideWidth = 50
		}
		mainWidth = m.width - sideWidth - 4 // Leave space for padding and separation
	}

	// Render account list
	accountList := m.renderAccountList(mainWidth)

	// Render info panel if space allows
	if m.showInfoPanel && m.cursor < len(m.accounts) {
		infoPanel := m.renderInfoPanel(sideWidth)
		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			accountList,
			strings.Repeat(" ", 2),
			infoPanel,
		)
	}

	return accountList
}

func (m AccountSelectorModel) renderAccountList(width int) string {
	if len(m.accounts) == 0 {
		return EmptyState(
			"No Accounts Connected",
			"You haven't connected any Google accounts yet.",
			"Run 'urgent connect' to add an account",
		)
	}

	var b strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		MarginBottom(1).
		MarginTop(1).
		MarginLeft(2) // Add consistent left padding
	b.WriteString(headerStyle.Render("Select an account to manage:"))
	b.WriteString("\n\n")

	// Account items
	for i, acc := range m.accounts {
		item := m.renderAccountItem(acc, i == m.cursor, width-4)
		b.WriteString(item)
		b.WriteString("\n")
	}

	return b.String()
}

func (m AccountSelectorModel) renderAccountItem(acc AccountInfo, isFocused bool, width int) string {
	// Styles
	var itemStyle lipgloss.Style
	if isFocused {
		itemStyle = lipgloss.NewStyle().
			Background(lipgloss.AdaptiveColor{Light: "254", Dark: "235"}).
			Foreground(lipgloss.AdaptiveColor{Light: "16", Dark: "255"}).
			Padding(0, 2).
			Width(width).
			Bold(true)
	} else {
		itemStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Width(width)
	}

	mutedStyle := lipgloss.NewStyle().
		Foreground(MutedColor)

	// Build content
	var content strings.Builder

	// Cursor indicator
	cursor := "  "
	if isFocused {
		cursor = "❯ "
	}

	// Email (main text)
	email := acc.Email

	// Calendar count (secondary text)
	countText := ""
	if acc.TotalCount > 0 {
		// Show "X of Y" if we have total count
		countText = fmt.Sprintf("  %s", FormatCount(acc.EnabledCount, acc.TotalCount))
		if !isFocused {
			countText = mutedStyle.Render(countText)
		}
	} else if acc.EnabledCount > 0 {
		// Show just "X enabled" if we don't have total count
		countText = fmt.Sprintf("  %d enabled", acc.EnabledCount)
		if !isFocused {
			countText = mutedStyle.Render(countText)
		}
	}

	content.WriteString(cursor + email + countText)

	return itemStyle.Render(content.String())
}

func (m AccountSelectorModel) renderInfoPanel(width int) string {
	if m.cursor >= len(m.accounts) {
		return ""
	}

	acc := m.accounts[m.cursor]

	var items []string

	mutedStyle := lipgloss.NewStyle().Foreground(MutedColor)

	// Show enabled calendar names
	if len(acc.EnabledCalendars) > 0 {
		for _, cal := range acc.EnabledCalendars {
			truncated := TruncateWithEllipsis(cal, width-4)
			items = append(items, mutedStyle.Render("• ")+truncated)
		}
	} else {
		items = append(items, mutedStyle.Render("None configured"))
	}

	return InfoPanel("Enabled Calendars", items)
}

// GetSelectedAccount returns the email of the selected account.
func (m AccountSelectorModel) GetSelectedAccount() string {
	if m.cursor < len(m.accounts) {
		return m.accounts[m.cursor].Email
	}
	return ""
}

// IsConfirmed returns whether the selection was confirmed.
func (m AccountSelectorModel) IsConfirmed() bool {
	return m.confirmed
}

// IsCancelled returns whether the selection was cancelled.
func (m AccountSelectorModel) IsCancelled() bool {
	return m.cancelled
}
