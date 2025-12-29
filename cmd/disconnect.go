package cmd

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"urgent/internal/auth"
	"urgent/internal/tui"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	disconnectAccount string
)

var disconnectCmd = &cobra.Command{
	Use:   "disconnect",
	Short: "Disconnect a Google account",
	Long: `Remove a Google account and delete its credentials from Keychain.

Examples:
  urgent disconnect                         # Interactive selection
  urgent disconnect --account user@gmail.com`,
	RunE: runDisconnect,
}

func init() {
	rootCmd.AddCommand(disconnectCmd)
	disconnectCmd.Flags().StringVarP(&disconnectAccount, "account", "a", "", "Account email to disconnect")
}

type disconnectStage int

const (
	disconnectSelecting disconnectStage = iota
	disconnectConfirming
	disconnectDeleting
	disconnectSuccess
	disconnectError
)

type disconnectModel struct {
	stage    disconnectStage
	accounts []string
	selected string
	list     list.Model
	err      error
}

type accountItem string

func (i accountItem) FilterValue() string { return string(i) }
func (i accountItem) Title() string       { return string(i) }
func (i accountItem) Description() string { return "" }

func initialDisconnectModel(accounts []string, preselected string) disconnectModel {
	items := make([]list.Item, len(accounts))
	for i, acc := range accounts {
		items[i] = accountItem(acc)
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select Account to Disconnect"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(tui.PrimaryColor).
		Bold(true).
		Padding(1, 0)

	m := disconnectModel{
		accounts: accounts,
		selected: preselected,
		list:     l,
	}

	if preselected != "" {
		m.stage = disconnectConfirming
	} else {
		m.stage = disconnectSelecting
	}

	return m
}

func (m disconnectModel) Init() tea.Cmd {
	if m.selected != "" {
		return m.deleteAccount()
	}

	return nil
}

func (m disconnectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			if m.stage == disconnectSuccess || m.stage == disconnectError {
				return m, tea.Quit
			}

			if m.stage == disconnectSelecting {
				return m, tea.Quit
			}

		case "enter":
			if m.stage == disconnectSelecting {
				if item, ok := m.list.SelectedItem().(accountItem); ok {
					m.selected = string(item)
					m.stage = disconnectConfirming

					return m, m.deleteAccount()
				}
			}

		case "y":
			if m.stage == disconnectConfirming {
				return m, m.deleteAccount()
			}

		case "n":
			if m.stage == disconnectConfirming {
				return m, tea.Quit
			}
		}

	case tea.WindowSizeMsg:
		h, v := lipgloss.NewStyle().GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)

	case deleteCompleteMsg:
		if msg.err != nil {
			m.stage = disconnectError
			m.err = msg.err
		} else {
			m.stage = disconnectSuccess
		}

		return m, tea.Quit
	}

	var cmd tea.Cmd
	if m.stage == disconnectSelecting {
		m.list, cmd = m.list.Update(msg)
	}

	return m, cmd
}

type deleteCompleteMsg struct {
	err error
}

func (m disconnectModel) deleteAccount() tea.Cmd {
	return func() tea.Msg {
		store := auth.NewKeychainStore()
		err := store.DeleteToken(m.selected)

		return deleteCompleteMsg{err: err}
	}
}

func (m disconnectModel) View() string {
	var b strings.Builder

	switch m.stage {
	case disconnectSelecting:
		return m.list.View()

	case disconnectConfirming:
		title := tui.TitleStyle().Render("Confirm Disconnect")
		b.WriteString(title + "\n\n")

		box := tui.BoxStyle().Render(
			fmt.Sprintf("Disconnect account: %s?\n\n"+
				"This will remove all stored credentials.", m.selected),
		)
		b.WriteString(box + "\n\n")

		help := tui.HelpStyle().Render("y = yes • n = no")
		b.WriteString(help)

	case disconnectDeleting:
		title := tui.TitleStyle().Render("Disconnecting...")
		b.WriteString(title + "\n")

	case disconnectSuccess:
		title := tui.SuccessStyle().Render("✓ Account Disconnected")
		b.WriteString(title + "\n\n")

		box := tui.BoxStyle().Render(
			fmt.Sprintf("Account %s has been removed.\n"+
				"Credentials deleted from Keychain.", m.selected),
		)
		b.WriteString(box + "\n")

	case disconnectError:
		title := tui.ErrorStyle().Render("✗ Disconnect Failed")
		b.WriteString(title + "\n\n")

		errorMsg := tui.BoxStyle().Render(
			fmt.Sprintf("Error: %v", m.err),
		)
		b.WriteString(errorMsg + "\n\n")

		help := tui.HelpStyle().Render("Press q to exit")
		b.WriteString(help)
	}

	return b.String()
}

func runDisconnect(_ *cobra.Command, _ []string) error {
	store := auth.NewKeychainStore()

	// Get all connected accounts
	accounts, err := store.ListAccounts()
	if err != nil {
		return fmt.Errorf("failed to list accounts: %w", err)
	}

	if len(accounts) == 0 {
		return errors.New("no accounts connected")
	}

	// Check if account flag was provided
	if disconnectAccount != "" {
		// Verify account exists
		found := slices.Contains(accounts, disconnectAccount)

		if !found {
			return fmt.Errorf("account %s not found", disconnectAccount)
		}
	}

	p := tea.NewProgram(initialDisconnectModel(accounts, disconnectAccount))

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("failed to run disconnect TUI: %w", err)
	}

	m := finalModel.(disconnectModel)
	if m.stage == disconnectError {
		return m.err
	}

	return nil
}
