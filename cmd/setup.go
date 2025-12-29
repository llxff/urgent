package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"urgent/internal/auth"
	"urgent/internal/tui"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Store OAuth2 credentials in Keychain",
	Long: `Store OAuth2 client ID and secret securely in macOS Keychain.

This is a one-time setup. You need to:
1. Create a Google Cloud project
2. Enable Google Calendar API
3. Create OAuth2 Desktop credentials
4. Run this command to store the credentials

After setup, run 'urgent connect' to authorize a Google account.`,
	RunE: runSetup,
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

type setupStage int

const (
	setupWelcome setupStage = iota
	setupClientID
	setupClientSecret
	setupSaving
	setupSuccess
	setupError
)

type setupModel struct {
	stage             setupStage
	clientIDInput     textinput.Model
	clientSecretInput textinput.Model
	err               error
	saved             bool
}

func initialSetupModel() setupModel {
	// Client ID input
	clientIDInput := textinput.New()
	clientIDInput.Placeholder = "123456789.apps.googleusercontent.com"
	clientIDInput.Focus()
	clientIDInput.CharLimit = 200
	clientIDInput.Width = 60

	// Client Secret input
	clientSecretInput := textinput.New()
	clientSecretInput.Placeholder = "GOCSPX-..."
	clientSecretInput.EchoMode = textinput.EchoPassword
	clientSecretInput.EchoCharacter = '•'
	clientSecretInput.CharLimit = 200
	clientSecretInput.Width = 60

	return setupModel{
		stage:             setupWelcome,
		clientIDInput:     clientIDInput,
		clientSecretInput: clientSecretInput,
	}
}

func (m setupModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m setupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.stage == setupSuccess || m.stage == setupError {
				return m, tea.Quit
			}
		case "esc":
			return m, tea.Quit
		case "enter":
			return m.handleEnter()
		}

	case saveCompletedMsg:
		if msg.err != nil {
			m.stage = setupError
			m.err = msg.err
		} else {
			m.stage = setupSuccess
			m.saved = true
		}

		return m, tea.Quit
	}

	// Update active input
	var cmd tea.Cmd

	switch m.stage {
	case setupClientID:
		m.clientIDInput, cmd = m.clientIDInput.Update(msg)
	case setupClientSecret:
		m.clientSecretInput, cmd = m.clientSecretInput.Update(msg)
	}

	return m, cmd
}

func (m setupModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.stage {
	case setupWelcome:
		m.stage = setupClientID

		return m, nil
	case setupClientID:
		if strings.TrimSpace(m.clientIDInput.Value()) == "" {
			return m, nil
		}

		m.stage = setupClientSecret
		m.clientSecretInput.Focus()

		return m, nil
	case setupClientSecret:
		if strings.TrimSpace(m.clientSecretInput.Value()) == "" {
			return m, nil
		}

		m.stage = setupSaving

		return m, m.saveCredentials()
	}

	return m, nil
}

type saveCompletedMsg struct {
	err error
}

func (m setupModel) saveCredentials() tea.Cmd {
	return func() tea.Msg {
		clientID := strings.TrimSpace(m.clientIDInput.Value())
		clientSecret := strings.TrimSpace(m.clientSecretInput.Value())

		err := auth.SaveOAuthCredentials(clientID, clientSecret)

		return saveCompletedMsg{err: err}
	}
}

func (m setupModel) View() string {
	var b strings.Builder

	switch m.stage {
	case setupWelcome:
		title := tui.TitleStyle().Render("Google Calendar CLI - Initial Setup")
		b.WriteString(title + "\n\n")

		instructions := tui.BoxStyle().Render(
			"You need OAuth2 credentials from Google Cloud Console\n\n" +
				"Steps:\n" +
				"  1. Go to console.cloud.google.com\n" +
				"  2. Create or select a project\n" +
				"  3. Enable Google Calendar API\n" +
				"  4. Create OAuth2 Desktop credentials\n" +
				"  5. Configure redirect URI: http://localhost\n" +
				"  6. Enter Client ID and Secret below",
		)
		b.WriteString(instructions + "\n\n")

		help := tui.HelpStyle().Render("Press Enter to continue • Esc to cancel")
		b.WriteString(help)

	case setupClientID:
		title := tui.TitleStyle().Render("OAuth Client ID")
		b.WriteString(title + "\n\n")

		box := tui.BoxStyle().Render(
			"Paste your OAuth2 Client ID:\n\n" +
				m.clientIDInput.View(),
		)
		b.WriteString(box + "\n\n")

		help := tui.HelpStyle().Render("Enter to continue • Esc to cancel")
		b.WriteString(help)

	case setupClientSecret:
		title := tui.TitleStyle().Render("OAuth Client Secret")
		b.WriteString(title + "\n\n")

		box := tui.BoxStyle().Render(
			"Paste your OAuth2 Client Secret:\n\n" +
				m.clientSecretInput.View() + "\n\n" +
				tui.MutedStyle().Render("(hidden for security)"),
		)
		b.WriteString(box + "\n\n")

		help := tui.HelpStyle().Render("Enter to save • Esc to cancel")
		b.WriteString(help)

	case setupSaving:
		title := tui.TitleStyle().Render("Saving to Keychain...")
		b.WriteString(title + "\n")

	case setupSuccess:
		title := tui.SuccessStyle().Render("✓ Setup Complete!")
		b.WriteString(title + "\n\n")

		box := tui.BoxStyle().Render(
			"OAuth credentials saved to Keychain\n" +
				"Service: " + auth.OAuthServiceName + "\n\n" +
				"Next step: urgent connect",
		)
		b.WriteString(box + "\n")

	case setupError:
		title := tui.ErrorStyle().Render("✗ Setup Failed")
		b.WriteString(title + "\n\n")

		errorMsg := tui.BoxStyle().Render(
			"Error: " + m.err.Error(),
		)
		b.WriteString(errorMsg + "\n\n")

		help := tui.HelpStyle().Render("Press q to exit")
		b.WriteString(help)
	}

	return b.String()
}

func runSetup(_ *cobra.Command, _ []string) error {
	p := tea.NewProgram(initialSetupModel())

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("failed to run setup TUI: %w", err)
	}

	// Check if setup was successful
	m := finalModel.(setupModel)
	if m.stage == setupError {
		return m.err
	}

	if !m.saved {
		return errors.New("setup cancelled")
	}

	return nil
}
