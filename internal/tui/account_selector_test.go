package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestAccountSelector_NewModel(t *testing.T) {
	accounts := []AccountInfo{
		{Email: "user@gmail.com", EnabledCount: 3, TotalCount: 5},
		{Email: "work@example.com", EnabledCount: 1, TotalCount: 2},
	}
	m := NewAccountSelectorModel(accounts)

	assert.Len(t, m.accounts, 2)
	assert.Equal(t, 0, m.cursor)
	// Cursor position determines focus, not a separate field
	assert.Equal(t, "user@gmail.com", m.GetSelectedAccount())
}

func TestAccountSelector_Navigation(t *testing.T) {
	accounts := []AccountInfo{
		{Email: "user1@gmail.com"},
		{Email: "user2@gmail.com"},
		{Email: "user3@gmail.com"},
	}
	m := NewAccountSelectorModel(accounts)

	// Move down
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = result.(AccountSelectorModel)
	assert.Equal(t, 1, m.cursor)
	assert.Equal(t, "user2@gmail.com", m.GetSelectedAccount())

	// Move down again
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = result.(AccountSelectorModel)
	assert.Equal(t, 2, m.cursor)

	// Try to move past end
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = result.(AccountSelectorModel)
	assert.Equal(t, 2, m.cursor)

	// Move up
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = result.(AccountSelectorModel)
	assert.Equal(t, 1, m.cursor)
}

func TestAccountSelector_Confirm(t *testing.T) {
	accounts := []AccountInfo{{Email: "user@gmail.com"}}
	m := NewAccountSelectorModel(accounts)

	assert.False(t, m.IsConfirmed())

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = result.(AccountSelectorModel)

	assert.True(t, m.IsConfirmed())
	assert.False(t, m.IsCancelled())
}

func TestAccountSelector_Cancel(t *testing.T) {
	accounts := []AccountInfo{{Email: "user@gmail.com"}}
	m := NewAccountSelectorModel(accounts)

	assert.False(t, m.IsCancelled())

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = result.(AccountSelectorModel)

	assert.True(t, m.IsCancelled())
	assert.False(t, m.IsConfirmed())
}

func TestAccountSelector_GetSelectedAccount(t *testing.T) {
	accounts := []AccountInfo{
		{Email: "user1@gmail.com"},
		{Email: "user2@gmail.com"},
	}
	m := NewAccountSelectorModel(accounts)

	assert.Equal(t, "user1@gmail.com", m.GetSelectedAccount())

	// Move to second account
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = result.(AccountSelectorModel)

	assert.Equal(t, "user2@gmail.com", m.GetSelectedAccount())
}

func TestAccountSelector_PressA_ReturnsAddAccountMsg(t *testing.T) {
	accounts := []AccountInfo{{Email: "user@gmail.com"}}
	m := NewAccountSelectorModel(accounts)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

	assert.NotNil(t, cmd, "Expected command from 'a' key press")
	msg := cmd()
	_, ok := msg.(AddAccountRequestMsg)
	assert.True(t, ok, "Expected AddAccountRequestMsg")
}

func TestAccountSelector_PressD_ReturnsDeleteAccountMsg(t *testing.T) {
	accounts := []AccountInfo{{Email: "user@gmail.com"}}
	m := NewAccountSelectorModel(accounts)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})

	assert.NotNil(t, cmd, "Expected command from 'd' key press")
	msg := cmd().(DeleteAccountRequestMsg)
	assert.Equal(t, "user@gmail.com", msg.Email.String())
}

func TestAccountSelector_PressD_SelectedAccount(t *testing.T) {
	accounts := []AccountInfo{
		{Email: "user1@gmail.com"},
		{Email: "user2@gmail.com"},
	}
	m := NewAccountSelectorModel(accounts)

	// Move to second account
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = result.(AccountSelectorModel)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})

	msg := cmd().(DeleteAccountRequestMsg)
	assert.Equal(t, "user2@gmail.com", msg.Email.String(), "Should delete the selected account")
}

func TestAccountSelector_EmptyState_ShowsAddHint(t *testing.T) {
	m := NewAccountSelectorModel([]AccountInfo{})

	// Set dimensions for view to render properly
	result, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = result.(AccountSelectorModel)

	view := m.View()
	assert.Contains(t, view, "Press 'a' to add", "Empty state should show add hint")
}

func TestAccountSelector_NoRedundantHeader(t *testing.T) {
	accounts := []AccountInfo{{Email: "user@gmail.com"}}
	m := NewAccountSelectorModel(accounts)

	// Set dimensions for view to render properly
	result, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = result.(AccountSelectorModel)

	view := m.View()
	assert.NotContains(t, view, "Select an account to manage", "Should not contain redundant header")
}

func TestAccountSelector_EmptyState_NoAccounts(t *testing.T) {
	m := NewAccountSelectorModel([]AccountInfo{})

	assert.Equal(t, "", m.GetSelectedAccount())
}

func TestAccountSelector_FooterShowsKeys(t *testing.T) {
	accounts := []AccountInfo{{Email: "user@gmail.com"}}
	m := NewAccountSelectorModel(accounts)

	// Set dimensions for view to render properly
	result, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = result.(AccountSelectorModel)

	view := m.View()
	assert.Contains(t, view, "add", "Footer should show 'a add'")
	assert.Contains(t, view, "delete", "Footer should show 'd delete'")
}
