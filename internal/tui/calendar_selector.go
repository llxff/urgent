package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// calendarSelectorKeyMap defines key bindings for the calendar selector.
type calendarSelectorKeyMap struct {
	Up        key.Binding
	Down      key.Binding
	Toggle    key.Binding
	SelectAll key.Binding
	SelectNone key.Binding
	Confirm   key.Binding
	Cancel    key.Binding
	Quit      key.Binding
}

// Layout constants for CalendarSelector.
const (
	// calendarReservedFrameHeight is the vertical space reserved for frame elements.
	// Breakdown: top bar (2) + summary (3) + footer (2) = 7 lines.
	calendarReservedFrameHeight = 7
	// calendarMinVisibleItems is the minimum number of calendar items to show.
	calendarMinVisibleItems = 5
)

// defaultCalendarSelectorKeys returns the default key bindings.
func defaultCalendarSelectorKeys() calendarSelectorKeyMap {
	return calendarSelectorKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Toggle: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "toggle"),
		),
		SelectAll: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "all"),
		),
		SelectNone: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "none"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "save"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

// ShortHelp returns key bindings for the short help view.
func (k calendarSelectorKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Confirm, k.Toggle, k.SelectAll, k.SelectNone, k.Up, k.Down, k.Cancel}
}

// FullHelp returns key bindings for the full help view.
func (k calendarSelectorKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},
		{k.Toggle, k.SelectAll, k.SelectNone},
		{k.Confirm, k.Cancel, k.Quit},
	}
}

// CalendarItem represents a calendar in the selection list.
type CalendarItem struct {
	ID       string
	Name     string
	Desc     string // Calendar description (field named Desc to avoid conflict with Description() method)
	Primary  bool
	Selected bool
}

// CalendarSelectorModel manages the calendar selection TUI with polished UI.
type CalendarSelectorModel struct {
	email                 string
	items                 []CalendarItem
	selected              map[string]bool
	cursor                int
	confirmed             bool
	cancelled             bool
	width                 int
	height                int
	showSavedConfirmation bool
	savedMessage          string
	keys                  calendarSelectorKeyMap
	help                  help.Model
}

// NewCalendarSelectorModel creates a new calendar selector with polished UI.
func NewCalendarSelectorModel(email string, calendars []CalendarItem, enabledIDs []string) CalendarSelectorModel {
	// Create selected map from enabled IDs
	selected := make(map[string]bool)
	for _, id := range enabledIDs {
		selected[id] = true
	}

	// Mark items as selected
	items := make([]CalendarItem, len(calendars))
	for i, cal := range calendars {
		cal.Selected = selected[cal.ID]
		items[i] = cal
	}

	h := help.New()
	h.ShowAll = false // Use short help by default

	return CalendarSelectorModel{
		email:    email,
		items:    items,
		selected: selected,
		cursor:   0,
		keys:     defaultCalendarSelectorKeys(),
		help:     h,
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
		// Update help model width
		m.help.Width = msg.Width
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			// q only works as quit if not in middle of action
			return m, tea.Quit

		case key.Matches(msg, m.keys.Cancel):
			m.cancelled = true
			return m, tea.Quit

		case key.Matches(msg, m.keys.Confirm):
			m.confirmed = true
			m.showSavedConfirmation = true
			m.savedMessage = "Saved"
			return m, tea.Quit

		case key.Matches(msg, m.keys.Up):
			m.moveCursorUp()
			return m, nil

		case key.Matches(msg, m.keys.Down):
			m.moveCursorDown()
			return m, nil

		case key.Matches(msg, m.keys.Toggle):
			m.toggleCurrent()
			return m, nil

		case key.Matches(msg, m.keys.SelectAll):
			m.selectAll()
			return m, nil

		case key.Matches(msg, m.keys.SelectNone):
			m.selectNone()
			return m, nil
		}
	}

	return m, nil
}

func (m *CalendarSelectorModel) moveCursorUp() {
	if m.cursor > 0 {
		m.cursor--
	}
}

func (m *CalendarSelectorModel) moveCursorDown() {
	if m.cursor < len(m.items)-1 {
		m.cursor++
	}
}

func (m *CalendarSelectorModel) toggleCurrent() {
	if m.cursor >= 0 && m.cursor < len(m.items) {
		item := &m.items[m.cursor]
		m.selected[item.ID] = !m.selected[item.ID]
		item.Selected = m.selected[item.ID]
	}
}

func (m *CalendarSelectorModel) selectAll() {
	for i := range m.items {
		m.selected[m.items[i].ID] = true
		m.items[i].Selected = true
	}
}

func (m *CalendarSelectorModel) selectNone() {
	for i := range m.items {
		m.selected[m.items[i].ID] = false
		m.items[i].Selected = false
	}
}

func (m CalendarSelectorModel) View() string {
	if m.width == 0 {
		return LoadingSpinner() + " Initializing..."
	}

	// Breadcrumb
	breadcrumb := fmt.Sprintf("Accounts › %s › Calendars", TruncateWithEllipsis(m.email, 30))

	// Top bar
	topBar := RenderTopBar(m.width, "Urgent Calendars", breadcrumb)

	// Content
	content := m.renderContent()

	// Footer with help from bubbles/help component
	status := ""
	if m.showSavedConfirmation {
		status = ConfirmationMessage(m.savedMessage)
	}

	footer := RenderFooter(m.width, m.help.View(m.keys), status)

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
	selectedCount := m.countSelected()
	summaryStyle := lipgloss.NewStyle().
		Foreground(MutedColor).
		MarginTop(1).
		MarginBottom(1)

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
			Render(fmt.Sprintf("   ... %d more", len(m.items)-visibleEnd))
		b.WriteString(indicator)
		b.WriteString("\n")
	}

	return b.String()
}

func (m CalendarSelectorModel) calculateVisibleRange() (int, int) {
	// Calculate how many items can fit
	availableHeight := m.height - calendarReservedFrameHeight
	if availableHeight < calendarMinVisibleItems {
		availableHeight = calendarMinVisibleItems
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
	if item.Selected {
		checkbox = "[✓]"
	}

	checkboxStyle := lipgloss.NewStyle()
	if item.Selected {
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

func (m CalendarSelectorModel) countSelected() int {
	count := 0
	for _, sel := range m.selected {
		if sel {
			count++
		}
	}
	return count
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
