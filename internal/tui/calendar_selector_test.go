package tui

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestNewCalendarSelectorModel(t *testing.T) {
	calendars := []CalendarItem{
		{ID: "cal1", Name: "Calendar 1", Primary: true},
		{ID: "cal2", Name: "Calendar 2"},
		{ID: "cal3", Name: "Calendar 3"},
	}
	enabledIDs := []string{"cal1", "cal3"}

	m := NewCalendarSelectorModel("test@example.com", calendars, enabledIDs)

	assert.Len(t, m.items, 3)
	assert.True(t, m.items[0].Selected, "cal1 should be selected")
	assert.False(t, m.items[1].Selected, "cal2 should not be selected")
	assert.True(t, m.items[2].Selected, "cal3 should be selected")
	assert.True(t, m.selected["cal1"])
	assert.False(t, m.selected["cal2"])
	assert.True(t, m.selected["cal3"])
}

func TestCalendarSelectorModel_ToggleSelection(t *testing.T) {
	calendars := []CalendarItem{
		{ID: "cal1", Name: "Calendar 1"},
		{ID: "cal2", Name: "Calendar 2"},
	}
	m := NewCalendarSelectorModel("test@example.com", calendars, []string{"cal1"})

	// Simulate window size
	result, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = result.(CalendarSelectorModel)

	// Toggle cal1 (currently selected)
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = result.(CalendarSelectorModel)
	assert.False(t, m.selected["cal1"], "cal1 should be deselected after toggle")

	// Toggle cal1 again
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = result.(CalendarSelectorModel)
	assert.True(t, m.selected["cal1"], "cal1 should be selected after second toggle")
}

func TestCalendarSelectorModel_SelectAll(t *testing.T) {
	calendars := []CalendarItem{
		{ID: "cal1", Name: "Calendar 1"},
		{ID: "cal2", Name: "Calendar 2"},
		{ID: "cal3", Name: "Calendar 3"},
	}
	m := NewCalendarSelectorModel("test@example.com", calendars, []string{})

	// Simulate window size
	result, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = result.(CalendarSelectorModel)

	// Press 'a' to select all
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = result.(CalendarSelectorModel)

	assert.True(t, m.selected["cal1"])
	assert.True(t, m.selected["cal2"])
	assert.True(t, m.selected["cal3"])
	assert.True(t, m.items[0].Selected)
	assert.True(t, m.items[1].Selected)
	assert.True(t, m.items[2].Selected)
}

func TestCalendarSelectorModel_SelectNone(t *testing.T) {
	calendars := []CalendarItem{
		{ID: "cal1", Name: "Calendar 1"},
		{ID: "cal2", Name: "Calendar 2"},
		{ID: "cal3", Name: "Calendar 3"},
	}
	m := NewCalendarSelectorModel("test@example.com", calendars, []string{"cal1", "cal2", "cal3"})

	// Simulate window size
	result, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = result.(CalendarSelectorModel)

	// Press 'n' to select none
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = result.(CalendarSelectorModel)

	assert.False(t, m.selected["cal1"])
	assert.False(t, m.selected["cal2"])
	assert.False(t, m.selected["cal3"])
	assert.False(t, m.items[0].Selected)
	assert.False(t, m.items[1].Selected)
	assert.False(t, m.items[2].Selected)
}

func TestCalendarSelectorModel_Confirm(t *testing.T) {
	calendars := []CalendarItem{
		{ID: "cal1", Name: "Calendar 1"},
	}
	m := NewCalendarSelectorModel("test@example.com", calendars, []string{"cal1"})

	// Simulate window size
	result, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = result.(CalendarSelectorModel)

	assert.False(t, m.IsConfirmed())

	// Press enter to confirm
	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = result.(CalendarSelectorModel)

	assert.True(t, m.IsConfirmed())
	assert.False(t, m.IsCancelled())
	assert.NotNil(t, cmd) // Should have quit command
}

func TestCalendarSelectorModel_Cancel(t *testing.T) {
	calendars := []CalendarItem{
		{ID: "cal1", Name: "Calendar 1"},
	}
	m := NewCalendarSelectorModel("test@example.com", calendars, []string{})

	// Simulate window size
	result, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = result.(CalendarSelectorModel)

	assert.False(t, m.IsCancelled())

	// Press esc to cancel
	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = result.(CalendarSelectorModel)

	assert.True(t, m.IsCancelled())
	assert.False(t, m.IsConfirmed())
	assert.NotNil(t, cmd) // Should have quit command
}

func TestCalendarSelectorModel_GetSelectedIDs(t *testing.T) {
	calendars := []CalendarItem{
		{ID: "cal1", Name: "Calendar 1"},
		{ID: "cal2", Name: "Calendar 2"},
		{ID: "cal3", Name: "Calendar 3"},
	}
	m := NewCalendarSelectorModel("test@example.com", calendars, []string{"cal1", "cal3"})

	ids := m.GetSelectedIDs()

	assert.Len(t, ids, 2)
	assert.Contains(t, ids, "cal1")
	assert.Contains(t, ids, "cal3")
	assert.NotContains(t, ids, "cal2")
}

func TestCalendarSelectorModel_View(t *testing.T) {
	calendars := []CalendarItem{
		{ID: "cal1", Name: "Calendar 1", Desc: "First calendar"},
	}
	m := NewCalendarSelectorModel("test@example.com", calendars, []string{"cal1"})

	// Before window size
	view := m.View()
	assert.Contains(t, view, "Initializing")

	// After window size
	result, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = result.(CalendarSelectorModel)
	view = m.View()
	assert.Contains(t, view, "Enabled: 1 / 1")
}

func TestCalendarSelectorModel_EmptyCalendars(t *testing.T) {
	m := NewCalendarSelectorModel("test@example.com", []CalendarItem{}, []string{})

	assert.Empty(t, m.items)
	assert.Empty(t, m.selected)

	ids := m.GetSelectedIDs()
	assert.Empty(t, ids)
}

func TestCalendarSelectorModel_ManyCalendars(t *testing.T) {
	// Test with 20 calendars
	calendars := make([]CalendarItem, 20)
	for i := range 20 {
		calendars[i] = CalendarItem{
			ID:   fmt.Sprintf("cal%d", i),
			Name: fmt.Sprintf("Calendar %d", i),
		}
	}

	m := NewCalendarSelectorModel("test@example.com", calendars, []string{})
	assert.Len(t, m.items, 20)

	// Select all
	result, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = result.(CalendarSelectorModel)
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = result.(CalendarSelectorModel)

	ids := m.GetSelectedIDs()
	assert.Len(t, ids, 20)
}

func TestCalendarSelectorModel_Navigation(t *testing.T) {
	calendars := []CalendarItem{
		{ID: "cal1", Name: "Calendar 1"},
		{ID: "cal2", Name: "Calendar 2"},
		{ID: "cal3", Name: "Calendar 3"},
	}
	m := NewCalendarSelectorModel("test@example.com", calendars, []string{})

	// Simulate window size
	result, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = result.(CalendarSelectorModel)

	// Initial cursor at 0
	assert.Equal(t, 0, m.cursor)

	// Move down
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = result.(CalendarSelectorModel)
	assert.Equal(t, 1, m.cursor)

	// Move down again
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = result.(CalendarSelectorModel)
	assert.Equal(t, 2, m.cursor)

	// Try to move down past end
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = result.(CalendarSelectorModel)
	assert.Equal(t, 2, m.cursor) // Should stay at 2

	// Move up
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = result.(CalendarSelectorModel)
	assert.Equal(t, 1, m.cursor)

	// Move up again
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = result.(CalendarSelectorModel)
	assert.Equal(t, 0, m.cursor)

	// Try to move up past start
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = result.(CalendarSelectorModel)
	assert.Equal(t, 0, m.cursor) // Should stay at 0
}
