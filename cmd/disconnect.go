package cmd

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"urgent/internal/auth"
	"urgent/internal/tui"

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
	cursor   int
	width    int
	height   int
	err      error
}

func initialDisconnectModel(accounts []string, preselected string) disconnectModel {
	stage := disconnectSelecting
	cursor := 0

	// If account is preselected, find its index and go to confirming
	if preselected != "" {
		for i, acc := range accounts {
			if acc == preselected {
				cursor = i
				break
			}
		}
		stage = disconnectConfirming
	}

	return disconnectModel{
		stage:    stage,
		accounts: accounts,
		selected: preselected,
		cursor:   cursor,
	}
}

func (m disconnectModel) Init() tea.Cmd {
	if m.selected != "" && m.stage == disconnectConfirming {
		m.stage = disconnectDeleting
		return m.deleteAccount()
	}

	return nil
}

func (m disconnectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch m.stage {
		case disconnectSelecting:
			switch msg.String() {
			case "q", "esc":
				return m, tea.Quit
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.accounts)-1 {
					m.cursor++
				}
			case "enter":
				m.selected = m.accounts[m.cursor]
				m.stage = disconnectConfirming
				return m, nil
			}

		case disconnectConfirming:
			switch msg.String() {
			case "y", "enter":
				m.stage = disconnectDeleting
				return m, m.deleteAccount()
			case "n", "esc":
				// Go back to selection
				m.selected = ""
				m.stage = disconnectSelecting
				return m, nil
			}

		case disconnectSuccess, disconnectError:
			if msg.String() == "q" || msg.String() == "enter" {
				return m, tea.Quit
			}
		}

	case deleteCompleteMsg:
		if msg.err != nil {
			m.stage = disconnectError
			m.err = msg.err
		} else {
			m.stage = disconnectSuccess
		}
		return m, nil
	}

	return m, nil
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
	if m.width == 0 {
		return tui.LoadingSpinner() + " Initializing..."
	}

	switch m.stage {
	case disconnectSelecting:
		return m.renderSelectionScreen()
	case disconnectConfirming:
		return m.renderConfirmationScreen()
	case disconnectDeleting:
		return m.renderDeletingScreen()
	case disconnectSuccess:
		return m.renderSuccessScreen()
	case disconnectError:
		return m.renderErrorScreen()
	}

	return ""
}

func (m disconnectModel) renderSelectionScreen() string {
	topBar := tui.RenderTopBar(m.width, "", "Disconnect")

	// Content
	var b strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		MarginBottom(1).
		MarginTop(1).
		MarginLeft(2)
	b.WriteString(headerStyle.Render("Select an account to disconnect:"))
	b.WriteString("\n\n")

	// Account list
	for i, acc := range m.accounts {
		item := m.renderAccountItem(acc, i == m.cursor)
		b.WriteString(item)
		b.WriteString("\n")
	}

	content := b.String()

	// Footer
	helpKeys := map[string]string{
		"enter": "select",
		"↑↓":    "navigate",
		"esc":   "cancel",
	}
	footer := tui.RenderFooter(m.width, tui.FormatHelpKeys(helpKeys), "")

	frame := tui.NewFrame(m.width, m.height)
	return frame.Render(topBar, content, footer)
}

func (m disconnectModel) renderAccountItem(email string, isFocused bool) string {
	width := m.width - 8

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

	cursor := "  "
	if isFocused {
		cursor = "❯ "
	}

	return itemStyle.Render(cursor + email)
}

func (m disconnectModel) renderConfirmationScreen() string {
	topBar := tui.RenderTopBar(m.width, "", "Disconnect › Confirm")

	var b strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		MarginBottom(1).
		MarginTop(1).
		MarginLeft(2)
	b.WriteString(headerStyle.Render("Confirm Disconnect"))
	b.WriteString("\n\n")

	// Warning box
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.ErrorColor).
		Padding(1, 2).
		MarginLeft(2).
		Width(m.width - 8)

	message := fmt.Sprintf("Disconnect account: %s?\n\nThis will remove all stored credentials from Keychain.", m.selected)
	b.WriteString(boxStyle.Render(message))

	content := b.String()

	helpKeys := map[string]string{
		"y":   "yes",
		"n":   "no",
		"esc": "back",
	}
	footer := tui.RenderFooter(m.width, tui.FormatHelpKeys(helpKeys), "")

	frame := tui.NewFrame(m.width, m.height)
	return frame.Render(topBar, content, footer)
}

func (m disconnectModel) renderDeletingScreen() string {
	topBar := tui.RenderTopBar(m.width, "", "Disconnect")

	content := tui.CenterVertical(
		tui.StatusIndicator(tui.LoadingSpinner(), "Disconnecting account...", tui.PrimaryColor),
		m.height-4,
	)

	footer := tui.RenderFooter(m.width, "", "")

	frame := tui.NewFrame(m.width, m.height)
	return frame.Render(topBar, content, footer)
}

func (m disconnectModel) renderSuccessScreen() string {
	topBar := tui.RenderTopBar(m.width, "", "Disconnect")

	var b strings.Builder

	// Status with margin
	statusStyle := lipgloss.NewStyle().MarginTop(1).MarginLeft(2)
	b.WriteString(statusStyle.Render(tui.StatusIndicator("✓", "Account Disconnected", tui.SuccessColor)))
	b.WriteString("\n\n")

	summaryStyle := lipgloss.NewStyle().Foreground(tui.MutedColor).MarginLeft(2)
	b.WriteString(summaryStyle.Render(fmt.Sprintf(
		"Account: %s\nCredentials deleted from Keychain",
		m.selected,
	)))

	content := b.String()

	helpKeys := map[string]string{
		"q": "quit",
	}
	footer := tui.RenderFooter(m.width, tui.FormatHelpKeys(helpKeys), "")

	frame := tui.NewFrame(m.width, m.height)
	return frame.Render(topBar, content, footer)
}

func (m disconnectModel) renderErrorScreen() string {
	topBar := tui.RenderTopBar(m.width, "", "Disconnect")

	var b strings.Builder

	// Status with margin
	statusStyle := lipgloss.NewStyle().MarginTop(1).MarginLeft(2)
	b.WriteString(statusStyle.Render(tui.StatusIndicator("✗", "Disconnect Failed", tui.ErrorColor)))
	b.WriteString("\n\n")

	errorStyle := lipgloss.NewStyle().Foreground(tui.ErrorColor).MarginLeft(2)
	b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))

	content := b.String()

	helpKeys := map[string]string{
		"q": "quit",
	}
	footer := tui.RenderFooter(m.width, tui.FormatHelpKeys(helpKeys), "")

	frame := tui.NewFrame(m.width, m.height)
	return frame.Render(topBar, content, footer)
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

	p := tea.NewProgram(
		initialDisconnectModel(accounts, disconnectAccount),
		tea.WithInput(os.Stdin),
		tea.WithOutput(os.Stderr),
		tea.WithAltScreen(),
	)

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
