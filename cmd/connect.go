// Package cmd provides the command-line interface for the urgent application.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"urgent/internal/auth"
	"urgent/internal/calendar"
	"urgent/internal/config"
	"urgent/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect a Google account",
	Long: `Authenticate with a Google account using OAuth2.

A browser window will open for you to authorize the application.
Once authorized, your credentials are stored securely in Keychain.
You'll then select which calendars to enable.`,
	RunE: runConnect,
}

func init() {
	rootCmd.AddCommand(connectCmd)
}

type connectStage int

const (
	connectInitializing connectStage = iota
	connectAuthenticating
	connectFetchingCalendars
	connectSelectingCalendars
	connectSavingConfig
	connectSuccess
	connectError
)

type connectModel struct {
	stage            connectStage
	email            string
	authURL          string
	calendars        []tui.CalendarItem
	calendarSelector tui.CalendarSelectorModel
	selectedCount    int
	err              error
	width            int
	height           int
	ctx              context.Context
	cancelAuth       context.CancelFunc
}

func initialConnectModel() connectModel {
	ctx, cancel := context.WithCancel(context.Background())

	return connectModel{
		stage:      connectInitializing,
		ctx:        ctx,
		cancelAuth: cancel,
	}
}

func (m connectModel) Init() tea.Cmd {
	return m.authenticate()
}

func (m connectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width

		m.height = msg.Height

		if m.stage == connectSelectingCalendars {
			result, cmd := m.calendarSelector.Update(msg)
			m.calendarSelector = result.(tui.CalendarSelectorModel)

			return m, cmd
		}

		return m, nil

	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			// Cancel authentication if in progress
			if m.stage == connectAuthenticating && m.cancelAuth != nil {
				m.cancelAuth()
			}

			if m.stage == connectSuccess || m.stage == connectError {
				return m, tea.Quit
			}
		}

		// Pass keys to calendar selector
		if m.stage == connectSelectingCalendars {
			result, cmd := m.calendarSelector.Update(msg)
			m.calendarSelector = result.(tui.CalendarSelectorModel)

			// Check if user confirmed or cancelled
			if m.calendarSelector.IsConfirmed() {
				m.stage = connectSavingConfig

				return m, m.saveConfig()
			}

			if m.calendarSelector.IsCancelled() {
				m.stage = connectError
				m.err = errors.New("calendar selection cancelled")

				return m, tea.Quit
			}

			return m, cmd
		}

	case authURLMsg:
		// Update the auth URL as soon as it's available
		m.authURL = msg.authURL
		m.stage = connectAuthenticating

		return m, nil

	case authCompleteMsg:
		// Store the authURL if we don't have it yet
		if msg.authURL != "" && m.authURL == "" {
			m.authURL = msg.authURL
		}

		if msg.err != nil {
			m.stage = connectError
			m.err = msg.err

			return m, tea.Quit
		}

		m.email = msg.email
		m.stage = connectFetchingCalendars

		return m, m.fetchCalendars()

	case calendarsMsg:
		if msg.err != nil {
			m.stage = connectError
			m.err = msg.err

			return m, tea.Quit
		}

		m.calendars = msg.calendars
		m.stage = connectSelectingCalendars
		// Initialize calendar selector with all calendars selected by default
		m.calendarSelector = tui.NewCalendarSelectorModel(m.email, msg.calendars, msg.allIDs)

		// Send window size to selector if we have it
		if m.width > 0 {
			result, cmd := m.calendarSelector.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
			m.calendarSelector = result.(tui.CalendarSelectorModel)

			return m, cmd
		}

		return m, nil

	case configSavedMsg:
		if msg.err != nil {
			m.stage = connectError
			m.err = msg.err

			return m, tea.Quit
		}

		m.selectedCount = msg.count
		m.stage = connectSuccess

		return m, tea.Quit
	}

	// Pass other messages to calendar selector if active
	if m.stage == connectSelectingCalendars {
		result, cmd := m.calendarSelector.Update(msg)
		m.calendarSelector = result.(tui.CalendarSelectorModel)

		return m, cmd
	}

	return m, nil
}

type authCompleteMsg struct {
	email   string
	authURL string
	err     error
}

type authURLMsg struct {
	authURL string
}

type calendarsMsg struct {
	calendars []tui.CalendarItem
	allIDs    []string
	err       error
}

type configSavedMsg struct {
	count int
	err   error
}

func (m connectModel) authenticate() tea.Cmd {
	// Create a channel to communicate between the auth goroutine and bubbletea
	urlChan := make(chan string, 1)

	// Start authentication in a goroutine
	authCmd := func() tea.Msg {
		// Get OAuth credentials
		clientID, clientSecret, err := auth.GetOAuthCredentials()
		if err != nil {
			return authCompleteMsg{err: err}
		}

		// Initialize auth manager
		store := auth.NewKeychainStore()
		manager := auth.NewManager(store, clientID, clientSecret)

		// Perform authentication with URL callback
		result, err := manager.AuthenticateWithURLCallback(m.ctx, func(url string) {
			// Send URL to channel so it can be displayed immediately
			select {
			case urlChan <- url:
			default:
			}
		})
		if err != nil {
			return authCompleteMsg{authURL: "", err: err}
		}

		// Save token
		if err := store.SaveToken(result.Email, result.Token); err != nil {
			return authCompleteMsg{authURL: result.AuthURL, err: fmt.Errorf("failed to save token: %w", err)}
		}

		return authCompleteMsg{email: result.Email, authURL: result.AuthURL, err: nil}
	}

	// Command to listen for URL from channel
	urlCmd := func() tea.Msg {
		url := <-urlChan

		return authURLMsg{authURL: url}
	}

	// Return both commands as a batch
	return tea.Batch(authCmd, urlCmd)
}

func (m connectModel) fetchCalendars() tea.Cmd {
	return func() tea.Msg {
		store := auth.NewKeychainStore()
		ctx := context.Background()

		// Create calendar client
		client, err := calendar.NewClient(ctx, m.email, store)
		if err != nil {
			return calendarsMsg{err: fmt.Errorf("failed to create calendar client: %w", err)}
		}

		// Fetch calendars
		cals, err := client.GetCalendars(ctx)
		if err != nil {
			return calendarsMsg{err: fmt.Errorf("failed to fetch calendars: %w", err)}
		}

		// Convert to TUI items
		items := make([]tui.CalendarItem, len(cals))

		allIDs := make([]string, len(cals))

		for i, cal := range cals {
			items[i] = tui.CalendarItem{
				ID:      cal.ID,
				Name:    cal.Summary,
				Desc:    cal.Description,
				Primary: cal.Primary,
			}
			allIDs[i] = cal.ID
		}

		return calendarsMsg{calendars: items, allIDs: allIDs, err: nil}
	}
}

func (m connectModel) saveConfig() tea.Cmd {
	return func() tea.Msg {
		cfgMgr := config.NewManager()

		// Get selected calendar IDs
		selectedIDs := m.calendarSelector.GetSelectedIDs()

		// Create calendar selections with names
		selections := make([]config.CalendarSelection, len(selectedIDs))

		calendarMap := make(map[string]tui.CalendarItem)
		for _, cal := range m.calendars {
			calendarMap[cal.ID] = cal
		}

		for i, id := range selectedIDs {
			cal := calendarMap[id]
			selections[i] = config.CalendarSelection{
				ID:   id,
				Name: cal.Name,
			}
		}

		// Save to config
		err := cfgMgr.SetEnabledCalendars(m.email, selections)
		if err != nil {
			return configSavedMsg{err: fmt.Errorf("failed to save config: %w", err)}
		}

		return configSavedMsg{count: len(selections), err: nil}
	}
}

func (m connectModel) View() string {
	switch m.stage {
	case connectInitializing, connectAuthenticating:
		return m.renderAuth()
	case connectFetchingCalendars:
		return m.renderLoading()
	case connectSelectingCalendars:
		return m.renderCalendarSelection()
	case connectSavingConfig:
		return m.renderSaving()
	case connectSuccess:
		return m.renderSuccess()
	case connectError:
		return m.renderError()
	}

	return ""
}

func (m connectModel) renderAuth() string {
	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(tui.PrimaryColor)

	b.WriteString(titleStyle.Render("🔐 Authenticate with Google"))
	b.WriteString("\n\n")

	if m.authURL == "" {
		b.WriteString(lipgloss.NewStyle().
			Foreground(tui.MutedColor).
			Render("  Preparing authorization..."))
	} else {
		b.WriteString(lipgloss.NewStyle().
			Foreground(tui.SuccessColor).
			Render("  ✓ Browser opened"))
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().
			Foreground(tui.MutedColor).
			Render("  Please complete authorization in your browser"))
		b.WriteString("\n\n")
		b.WriteString(lipgloss.NewStyle().
			Foreground(tui.MutedColor).
			Render("  If browser didn't open, visit:"))
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().
			Foreground(tui.PrimaryColor).
			Render("  " + m.authURL))
	}

	b.WriteString("\n\n")
	b.WriteString(lipgloss.NewStyle().
		Foreground(tui.MutedColor).
		Faint(true).
		Render("  Press Ctrl+C to cancel"))

	return b.String()
}

func (m connectModel) renderLoading() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(tui.PrimaryColor)

	b.WriteString(titleStyle.Render("📅 Loading Calendars"))
	b.WriteString("\n\n")

	b.WriteString(lipgloss.NewStyle().
		Foreground(tui.MutedColor).
		Render(fmt.Sprintf("  Fetching calendars for %s...", m.email)))

	return b.String()
}

func (m connectModel) renderCalendarSelection() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(tui.PrimaryColor)

	b.WriteString(titleStyle.Render("📅 Select Calendars · " + m.email))
	b.WriteString("\n\n")

	b.WriteString(m.calendarSelector.View())

	return b.String()
}

func (m connectModel) renderSaving() string {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(tui.PrimaryColor).
		Render("💾 Saving configuration...")
}

func (m connectModel) renderSuccess() string {
	var b strings.Builder

	// Success header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(tui.SuccessColor)

	b.WriteString(headerStyle.Render("✓ Connected Successfully"))
	b.WriteString("\n\n")

	// Account info
	labelStyle := lipgloss.NewStyle().
		Foreground(tui.MutedColor)

	valueStyle := lipgloss.NewStyle().
		Foreground(tui.PrimaryColor).
		Bold(true)

	b.WriteString(labelStyle.Render("  Account:    "))
	b.WriteString(valueStyle.Render(m.email))
	b.WriteString("\n")

	b.WriteString(labelStyle.Render("  Calendars:  "))
	b.WriteString(valueStyle.Render(fmt.Sprintf("%d selected", m.selectedCount)))
	b.WriteString("\n")

	b.WriteString(labelStyle.Render("  Config:     "))
	b.WriteString(lipgloss.NewStyle().
		Foreground(tui.MutedColor).
		Render(config.NewManager().GetConfigPath()))
	b.WriteString("\n\n")

	// Next steps
	b.WriteString(labelStyle.Render("  Try these commands:"))
	b.WriteString("\n\n")

	cmdStyle := lipgloss.NewStyle().
		Foreground(tui.PrimaryColor).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(tui.MutedColor)

	commands := []struct {
		cmd  string
		desc string
	}{
		{"urgent today", "View today's events"},
		{"urgent next --within 30", "Next event in 30 min"},
		{"urgent calendars", "Manage calendars"},
	}

	for _, c := range commands {
		b.WriteString("    ")
		b.WriteString(cmdStyle.Render(c.cmd))
		b.WriteString("  ")
		b.WriteString(descStyle.Render(c.desc))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Foreground(tui.MutedColor).
		Faint(true).
		Render("  Press q to exit"))

	return b.String()
}

func (m connectModel) renderError() string {
	var b strings.Builder

	// Error header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(tui.ErrorColor)

	b.WriteString(headerStyle.Render("✗ Connection Failed"))
	b.WriteString("\n\n")

	// Error message
	b.WriteString(lipgloss.NewStyle().
		Foreground(tui.ErrorColor).
		Render("  " + m.err.Error()))
	b.WriteString("\n\n")

	// Help section
	helpStyle := lipgloss.NewStyle().
		Foreground(tui.MutedColor)

	b.WriteString(helpStyle.Render("  Common fixes:"))
	b.WriteString("\n\n")

	items := []string{
		"• Run 'urgent setup' to configure OAuth credentials",
		"• Check your internet connection",
		"• Complete authorization in the browser",
	}

	for _, item := range items {
		b.WriteString(helpStyle.Render("    " + item))
		b.WriteString("\n")
	}

	if m.authURL != "" {
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("  OAuth URL:"))
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().
			Foreground(tui.PrimaryColor).
			Render("  " + m.authURL))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Foreground(tui.MutedColor).
		Faint(true).
		Render("  Press q to exit"))

	return b.String()
}

func runConnect(_ *cobra.Command, _ []string) error {
	p := tea.NewProgram(
		initialConnectModel(),
		tea.WithInput(os.Stdin),
		tea.WithOutput(os.Stderr),
	)

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("failed to run connect TUI: %w", err)
	}

	m := finalModel.(connectModel)
	if m.stage == connectError {
		return m.err
	}

	return nil
}
