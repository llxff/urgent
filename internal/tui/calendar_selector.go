package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CalendarItem represents a calendar in the selection list.
type CalendarItem struct {
	ID      string
	Name    string
	Desc    string // Calendar description (field named Desc to avoid conflict with Description() method)
	Primary bool
}

// CalendarSelectorModel manages the calendar selection TUI with polished UI.
type CalendarSelectorModel struct {
	email     string
	items     []CalendarItem
	selected  map[string]bool
	cursor    int
	confirmed bool
	cancelled bool
	width     int
	height    int
}

// NewCalendarSelectorModel creates a new calendar selector with polished UI.
func NewCalendarSelectorModel(email string, calendars []CalendarItem, enabledIDs []string) CalendarSelectorModel {
	// Create selected map from enabled IDs
	selected := make(map[string]bool)
	for _, id := range enabledIDs {
		selected[id] = true
	}

	return CalendarSelectorModel{
		email:    email,
		items:    calendars,
		selected: selected,
		cursor:   0,
	}
}

func (m CalendarSelectorModel) Init() tea.Cmd {
	return nil
}

func (m CalendarSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c"))):
			// q only works as quit if not in middle of action
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
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys(" "))):
			m.toggleCurrent()
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("a"))):
			m.selectAll()
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("n"))):
			m.selectNone()
			return m, nil
		}
	}

	return m, nil
}

func (m *CalendarSelectorModel) toggleCurrent() {
	if m.cursor >= 0 && m.cursor < len(m.items) {
		item := &m.items[m.cursor]
		m.selected[item.ID] = !m.selected[item.ID]
	}
}

func (m *CalendarSelectorModel) selectAll() {
	for i := range m.items {
		m.selected[m.items[i].ID] = true
	}
}

func (m *CalendarSelectorModel) selectNone() {
	for i := range m.items {
		m.selected[m.items[i].ID] = false
	}
}

func (m CalendarSelectorModel) View() string {
	if m.width == 0 {
		return LoadingSpinner() + " Initializing..."
	}

	// Breadcrumb
	breadcrumb := fmt.Sprintf("Accounts › %s › Calendars", TruncateWithEllipsis(m.email, 30))

	// Top bar
	topBar := RenderTopBar(m.width, "", breadcrumb)

	// Content
	content := m.renderContent()

	// Footer
	helpKeys := map[string]string{
		"enter": "save",
		"space": "toggle",
		"a":     "all",
		"n":     "none",
		"↑↓":    "navigate",
		"esc":   "back",
	}

	footer := RenderFooter(m.width, FormatHelpKeys(helpKeys), "")

	frame := NewFrame(m.width, m.height)
	return frame.Render(topBar, content, footer)
}

func (m CalendarSelectorModel) renderContent() string {
	if len(m.items) == 0 {
		return EmptyState(
			"No Calendars Found",
			"This account doesn't have any calendars.",
			"",
		)
	}

	var b strings.Builder

	// Summary header
	selectedCount := 0
	for _, sel := range m.selected {
		if sel {
			selectedCount++
		}
	}
	summaryStyle := lipgloss.NewStyle().
		Foreground(MutedColor).
		MarginTop(1).
		MarginBottom(1).
		MarginLeft(2) // Add consistent left padding

	b.WriteString(summaryStyle.Render(fmt.Sprintf(
		"Account: %s  •  Enabled: %d / %d",
		m.email,
		selectedCount,
		len(m.items),
	)))
	b.WriteString("\n\n")

	// Calendar list
	visibleStart, visibleEnd := m.calculateVisibleRange()

	for i := visibleStart; i < visibleEnd && i < len(m.items); i++ {
		item := m.renderCalendarItem(m.items[i], i == m.cursor, m.width-4)
		b.WriteString(item)
		b.WriteString("\n")
	}

	// Scroll indicator if needed
	if visibleEnd < len(m.items) {
		indicator := lipgloss.NewStyle().
			Foreground(MutedColor).
			MarginLeft(2). // Add consistent left padding
			Render(fmt.Sprintf("... %d more", len(m.items)-visibleEnd))
		b.WriteString(indicator)
		b.WriteString("\n")
	}

	return b.String()
}

func (m CalendarSelectorModel) calculateVisibleRange() (int, int) {
	// Calculate how many items can fit
	// Rough estimate: top bar (2) + summary (3) + footer (2) = 7 lines reserved
	availableHeight := m.height - 7
	if availableHeight < 5 {
		availableHeight = 5
	}

	maxVisible := availableHeight

	// Center cursor in viewport if possible
	halfVisible := maxVisible / 2
	start := m.cursor - halfVisible
	if start < 0 {
		start = 0
	}

	end := start + maxVisible
	if end > len(m.items) {
		end = len(m.items)
		start = end - maxVisible
		if start < 0 {
			start = 0
		}
	}

	return start, end
}

func (m CalendarSelectorModel) renderCalendarItem(item CalendarItem, isFocused bool, width int) string {
	// Checkbox
	checkbox := "[ ]"
	if m.selected[item.ID] {
		checkbox = "[✓]"
	}

	checkboxStyle := lipgloss.NewStyle()
	if m.selected[item.ID] {
		checkboxStyle = checkboxStyle.Foreground(SuccessColor).Bold(true)
	} else {
		checkboxStyle = checkboxStyle.Foreground(MutedColor)
	}

	// Cursor indicator
	cursor := "  "
	if isFocused {
		cursor = "❯ "
	}

	// Name with primary marker
	name := item.Name
	if item.Primary {
		primaryStyle := lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Italic(true)
		name = name + " " + primaryStyle.Render("(primary)")
	}

	// Build line
	content := cursor + checkboxStyle.Render(checkbox) + " " + name

	// Style the whole line
	var lineStyle lipgloss.Style
	if isFocused {
		lineStyle = lipgloss.NewStyle().
			Background(lipgloss.AdaptiveColor{Light: "254", Dark: "235"}).
			Foreground(lipgloss.AdaptiveColor{Light: "16", Dark: "255"}).
			Width(width).
			Padding(0, 1).
			Bold(true)
	} else {
		lineStyle = lipgloss.NewStyle().
			Width(width).
			Padding(0, 1)
	}

	return lineStyle.Render(content)
}

// GetSelectedIDs returns the IDs of selected calendars.
func (m CalendarSelectorModel) GetSelectedIDs() []string {
	var ids []string
	for id, selected := range m.selected {
		if selected {
			ids = append(ids, id)
		}
	}
	return ids
}

// IsConfirmed returns whether the user confirmed the selection.
func (m CalendarSelectorModel) IsConfirmed() bool {
	return m.confirmed
}

// IsCancelled returns whether the user cancelled the selection.
func (m CalendarSelectorModel) IsCancelled() bool {
	return m.cancelled
}
