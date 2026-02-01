package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// accountSelectorKeyMap defines key bindings for the account selector.
type accountSelectorKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Select key.Binding
	Add    key.Binding
	Delete key.Binding
	Quit   key.Binding
	Cancel key.Binding
}

// defaultAccountSelectorKeys returns the default key bindings.
func defaultAccountSelectorKeys() accountSelectorKeyMap {
	return accountSelectorKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Add: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
	}
}

// ShortHelp returns key bindings for the short help view.
func (k accountSelectorKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Select, k.Add, k.Delete, k.Up, k.Down, k.Cancel}
}

// FullHelp returns key bindings for the full help view.
func (k accountSelectorKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},
		{k.Select, k.Add, k.Delete},
		{k.Cancel, k.Quit},
	}
}

// Layout constants for AccountSelector.
const (
	// MinWidthForInfoPanel is the minimum screen width to show the side info panel.
	MinWidthForInfoPanel = 100
	// InfoPanelWidthPercent is the percentage of screen width for the info panel.
	InfoPanelWidthPercent = 40
	// MinInfoPanelWidth is the minimum width of the info panel in characters.
	MinInfoPanelWidth = 35
	// MaxInfoPanelWidth is the maximum width of the info panel in characters.
	MaxInfoPanelWidth = 50
)

// AccountInfo holds information about a calendar account.
type AccountInfo struct {
	Email            string
	EnabledCount     int
	TotalCount       int
	EnabledCalendars []string // Names of enabled calendars
	IsSelected       bool
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
	keys          accountSelectorKeyMap
	help          help.Model
}

// NewAccountSelectorModel creates a new account selector.
func NewAccountSelectorModel(accounts []AccountInfo) AccountSelectorModel {
	// Make a copy to avoid mutating the caller's slice
	accountsCopy := make([]AccountInfo, len(accounts))
	copy(accountsCopy, accounts)

	h := help.New()
	h.ShowAll = false // Use short help by default

	return AccountSelectorModel{
		accounts:      accountsCopy,
		cursor:        0,
		showInfoPanel: true,
		keys:          defaultAccountSelectorKeys(),
		help:          h,
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
		m.showInfoPanel = msg.Width > MinWidthForInfoPanel
		// Update help model width
		m.help.Width = msg.Width
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.cancelled = true
			return m, tea.Quit

		case key.Matches(msg, m.keys.Cancel):
			m.cancelled = true
			return m, tea.Quit

		case key.Matches(msg, m.keys.Select):
			m.confirmed = true
			return m, tea.Quit

		case key.Matches(msg, m.keys.Up):
			m.moveCursorUp()
			return m, nil

		case key.Matches(msg, m.keys.Down):
			m.moveCursorDown()
			return m, nil

		case key.Matches(msg, m.keys.Add):
			return m, func() tea.Msg { return AddAccountRequestMsg{} }

		case key.Matches(msg, m.keys.Delete):
			if len(m.accounts) > 0 {
				email := m.accounts[m.cursor].Email
				return m, func() tea.Msg { return DeleteAccountRequestMsg{Email: NewEmail(email)} }
			}
			return m, nil
		}
	}

	return m, nil
}

func (m *AccountSelectorModel) moveCursorUp() {
	if m.cursor > 0 {
		m.cursor--
	}
}

func (m *AccountSelectorModel) moveCursorDown() {
	if m.cursor < len(m.accounts)-1 {
		m.cursor++
	}
}

func (m AccountSelectorModel) View() string {
	if m.width == 0 {
		return LoadingSpinner() + " Initializing..."
	}

	// Top bar
	topBar := RenderTopBar(m.width, "", "")

	// Content
	content := m.renderContent()

	// Footer with help from bubbles/help component
	footer := RenderFooter(m.width, m.help.View(m.keys), "")

	frame := NewFrame(m.width, m.height)
	return frame.Render(topBar, content, footer)
}

func (m AccountSelectorModel) renderContent() string {
	// Calculate layout for side panel
	var sideWidth int
	if m.showInfoPanel {
		// Allocate percentage to side panel, with min/max bounds
		sideWidth = m.width * InfoPanelWidthPercent / 100
		if sideWidth < MinInfoPanelWidth {
			sideWidth = MinInfoPanelWidth
		}
		if sideWidth > MaxInfoPanelWidth {
			sideWidth = MaxInfoPanelWidth
		}
	}

	// Render account list with reasonable width
	listWidth := 50 // Fixed width for account list
	accountList := m.renderAccountList(listWidth)

	// Render info panel if space allows
	if m.showInfoPanel && m.cursor < len(m.accounts) {
		infoPanel := m.renderInfoPanel(sideWidth)
		// Join list and panel - both aligned at top
		return lipgloss.JoinHorizontal(lipgloss.Top, accountList, "  ", infoPanel)
	}

	return accountList
}

func (m AccountSelectorModel) renderAccountList(width int) string {
	if len(m.accounts) == 0 {
		return EmptyState(
			"No Accounts Connected",
			"You haven't connected any Google accounts yet.",
			"Press 'a' to add your first account",
		)
	}

	var b strings.Builder

	// Account items (no redundant header)
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
