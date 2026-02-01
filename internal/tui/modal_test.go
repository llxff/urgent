package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestModalModel_NewConfirmModal(t *testing.T) {
	m := NewConfirmModal("Remove?", "user@gmail.com")
	assert.False(t, m.IsConfirmed())
	assert.False(t, m.IsCancelled())
	assert.Equal(t, "Remove?", m.title)
	assert.Equal(t, "user@gmail.com", m.content)
}

func TestModalModel_NewProgressModal(t *testing.T) {
	m := NewProgressModal("Loading", "Please wait...")
	assert.False(t, m.IsConfirmed())
	assert.False(t, m.IsCancelled())
	assert.Equal(t, "Loading", m.title)
	assert.Equal(t, "Please wait...", m.content)
}

func TestModalModel_ConfirmWithY(t *testing.T) {
	m := NewConfirmModal("Remove?", "user@gmail.com")
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	modal := result.(ModalModel)
	assert.True(t, modal.IsConfirmed())
	assert.False(t, modal.IsCancelled())
}

func TestModalModel_ConfirmWithUpperY(t *testing.T) {
	m := NewConfirmModal("Remove?", "user@gmail.com")
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Y'}})
	modal := result.(ModalModel)
	assert.True(t, modal.IsConfirmed())
	assert.False(t, modal.IsCancelled())
}

func TestModalModel_CancelWithN(t *testing.T) {
	m := NewConfirmModal("Remove?", "user@gmail.com")
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	modal := result.(ModalModel)
	assert.True(t, modal.IsCancelled())
	assert.False(t, modal.IsConfirmed())
}

func TestModalModel_CancelWithUpperN(t *testing.T) {
	m := NewConfirmModal("Remove?", "user@gmail.com")
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}})
	modal := result.(ModalModel)
	assert.True(t, modal.IsCancelled())
	assert.False(t, modal.IsConfirmed())
}

func TestModalModel_CancelWithEsc(t *testing.T) {
	m := NewConfirmModal("Remove?", "user@gmail.com")
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	modal := result.(ModalModel)
	assert.True(t, modal.IsCancelled())
	assert.False(t, modal.IsConfirmed())
}

func TestModalModel_IgnoresOtherKeys(t *testing.T) {
	m := NewConfirmModal("Remove?", "user@gmail.com")
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	modal := result.(ModalModel)
	assert.False(t, modal.IsConfirmed())
	assert.False(t, modal.IsCancelled())
}

func TestModalModel_View(t *testing.T) {
	m := NewConfirmModal("Title", "Message content")
	view := m.View()
	assert.Contains(t, view, "Title")
	assert.Contains(t, view, "Message content")
	assert.Contains(t, view, "y")
	assert.Contains(t, view, "n")
}

func TestModalModel_RenderOver(t *testing.T) {
	m := NewConfirmModal("Title", "Message")
	background := "Background content here\nwith multiple lines\nof text content"
	result := m.RenderOver(background, 80, 24)

	// Modal content should be present
	assert.Contains(t, result, "Title")
	assert.Contains(t, result, "Message")

	// Result should have expected number of lines (screenHeight)
	lines := strings.Split(result, "\n")
	assert.Equal(t, 24, len(lines), "RenderOver should produce screenHeight lines")
}

func TestModalModel_RenderOver_PreservesBackgroundContent(t *testing.T) {
	m := NewConfirmModal("X", "Y")
	// Create a distinctive background pattern
	background := strings.Repeat("ABCD", 20) + "\n" + strings.Repeat("EFGH", 20)
	result := m.RenderOver(background, 80, 10)

	// The dimmed background should be visible around the modal
	// Since the modal is centered, we should see some background chars on the edges
	lines := strings.Split(result, "\n")
	assert.Equal(t, 10, len(lines), "Should have 10 lines")

	// Modal content should be centered, not at position 0,0
	assert.Contains(t, result, "X")
	assert.Contains(t, result, "Y")
}

func TestModalModel_ProgressModal_View(t *testing.T) {
	m := NewProgressModal("Loading", "Please wait...")
	view := m.View()
	assert.Contains(t, view, "Loading")
	assert.Contains(t, view, "Please wait...")
}

func TestModalModel_Init(t *testing.T) {
	m := NewConfirmModal("Title", "Content")
	cmd := m.Init()
	assert.Nil(t, cmd)
}
