package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"urgent/internal/auth"
	"urgent/internal/calendar"
	"urgent/internal/config"
	"urgent/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	calendarsAccount string
	calendarsList    bool
)

var calendarsCmd = &cobra.Command{
	Use:   "calendars",
	Short: "Manage calendar selection",
	Long: `Select which calendars to enable for each account.

Examples:
  urgent calendars                     # Select account, then calendars
  urgent calendars --account user@example.com  # Select calendars for account
  urgent calendars --list              # List all calendars and their status`,
	RunE: runCalendars,
}

func init() {
	rootCmd.AddCommand(calendarsCmd)
	calendarsCmd.Flags().StringVarP(&calendarsAccount, "account", "a", "", "Specify account email")
	calendarsCmd.Flags().BoolVarP(&calendarsList, "list", "l", false, "List all calendars and their enabled/disabled status")
}

func runCalendars(_ *cobra.Command, _ []string) error {
	store := auth.NewKeychainStore()

	// Get all connected accounts
	accounts, err := store.ListAccounts()
	if err != nil {
		return fmt.Errorf("failed to list accounts: %w", err)
	}

	if len(accounts) == 0 {
		return errors.New("no accounts connected. Run 'urgent connect' first")
	}

	// List mode
	if calendarsList {
		return listCalendars(accounts, store)
	}

	// Selection mode
	if calendarsAccount != "" {
		// Use specified account
		found := slices.Contains(accounts, calendarsAccount)

		if !found {
			return fmt.Errorf("account %s not found. Connected accounts: %v", calendarsAccount, accounts)
		}

		return selectCalendarsForAccount(calendarsAccount, store)
	}

	// If only one account, use it directly
	if len(accounts) == 1 {
		return selectCalendarsForAccount(accounts[0], store)
	}

	// Multiple accounts - show polished selection TUI
	return showAccountSelector(accounts, store)
}

func showAccountSelector(accountEmails []string, store auth.Store) error {
	// Use a unified calendar manager that handles both account selection and calendar selection
	p := tea.NewProgram(
		initialUnifiedCalendarModel(accountEmails, store),
		tea.WithInput(os.Stdin),
		tea.WithOutput(os.Stderr),
		tea.WithAltScreen(), // Use alternate screen buffer
	)

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("failed to run calendar manager: %w", err)
	}

	m := finalModel.(unifiedCalendarModel)
	if m.cancelled {
		return errors.New("cancelled")
	}

	if m.err != nil {
		return m.err
	}

	return nil
}

// Unified calendar model - manages both account selection and calendar selection in one program.
type unifiedCalendarStage int

const (
	unifiedStageLoadingAccounts unifiedCalendarStage = iota
	unifiedStageSelectingAccount
	unifiedStageLoadingCalendars
	unifiedStageSelectingCalendars
	unifiedStageSaving
	unifiedStageSuccess
	unifiedStageError
)

type unifiedCalendarModel struct {
	stage           unifiedCalendarStage
	store           auth.Store
	accountEmails   []string
	accountSelector tui.AccountSelectorModel
	selectedEmail   string
	calendarManager calendarManagerModel
	cancelled       bool
	err             error
	width           int
	height          int
}

func initialUnifiedCalendarModel(accountEmails []string, store auth.Store) unifiedCalendarModel {
	return unifiedCalendarModel{
		stage:         unifiedStageLoadingAccounts,
		store:         store,
		accountEmails: accountEmails,
	}
}

func (m unifiedCalendarModel) Init() tea.Cmd {
	return m.loadAccountInfoCmd()
}

func (m unifiedCalendarModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Forward to active sub-component
		switch m.stage {
		case unifiedStageSelectingAccount:
			result, cmd := m.accountSelector.Update(msg)
			m.accountSelector = result.(tui.AccountSelectorModel)
			return m, cmd
		case unifiedStageSelectingCalendars, unifiedStageLoadingCalendars, unifiedStageSaving:
			result, cmd := m.calendarManager.Update(msg)
			m.calendarManager = result.(calendarManagerModel)
			return m, cmd
		}

		return m, nil

	case tea.KeyMsg:
		// Global quit keys
		if msg.String() == "ctrl+c" {
			m.cancelled = true
			return m, tea.Quit
		}

		// Forward to active component
		switch m.stage {
		case unifiedStageSelectingAccount:
			result, cmd := m.accountSelector.Update(msg)
			m.accountSelector = result.(tui.AccountSelectorModel)

			if m.accountSelector.IsConfirmed() {
				// Transition to calendar selection
				m.selectedEmail = m.accountSelector.GetSelectedAccount()
				if m.selectedEmail == "" {
					m.err = errors.New("no account selected")
					m.stage = unifiedStageError
					return m, nil
				}

				// Initialize calendar manager
				m.calendarManager = initialCalendarManagerModel(m.selectedEmail, m.store)
				m.stage = unifiedStageLoadingCalendars

				// Initialize and send window size if we have it
				initCmd := m.calendarManager.Init()
				if m.width > 0 {
					sizeCmd := func() tea.Msg {
						return tea.WindowSizeMsg{Width: m.width, Height: m.height}
					}
					return m, tea.Batch(initCmd, sizeCmd)
				}

				return m, initCmd
			}

			if m.accountSelector.IsCancelled() {
				m.cancelled = true
				return m, tea.Quit
			}

			return m, cmd

		case unifiedStageLoadingCalendars, unifiedStageSelectingCalendars, unifiedStageSaving:
			// Check if user wants to go back from calendar selection
			if msg.String() == "esc" && m.stage == unifiedStageSelectingCalendars {
				// Go back to account selection
				m.stage = unifiedStageLoadingAccounts
				return m, m.loadAccountInfoCmd()
			}

			result, cmd := m.calendarManager.Update(msg)
			m.calendarManager = result.(calendarManagerModel)

			// Handle calendar manager state transitions
			if m.calendarManager.stage == calendarManagerSuccess {
				m.stage = unifiedStageSuccess
				// Quit immediately after success
				return m, tea.Quit
			}

			if m.calendarManager.stage == calendarManagerError {
				m.err = m.calendarManager.err
				m.stage = unifiedStageError
				return m, nil
			}

			if m.calendarManager.cancelled {
				// Go back to account selection
				m.stage = unifiedStageLoadingAccounts
				return m, m.loadAccountInfoCmd()
			}

			// Sync stage
			switch m.calendarManager.stage {
			case calendarManagerLoading:
				m.stage = unifiedStageLoadingCalendars
			case calendarManagerSelecting:
				m.stage = unifiedStageSelectingCalendars
			case calendarManagerSaving:
				m.stage = unifiedStageSaving
			}

			return m, cmd

		case unifiedStageError:
			if msg.String() == "q" {
				return m, tea.Quit
			}
		}

	case accountInfoLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.stage = unifiedStageError
			return m, nil
		}

		m.accountSelector = tui.NewAccountSelectorModel(msg.accounts)
		m.stage = unifiedStageSelectingAccount

		// Send window size if we have it
		if m.width > 0 {
			result, cmd := m.accountSelector.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
			m.accountSelector = result.(tui.AccountSelectorModel)
			return m, cmd
		}

		return m, nil

	case quitMsg:
		return m, tea.Quit

	default:
		// Forward all other messages (including calendarsLoadedMsg, calendarConfigSavedMsg, etc)
		// to the appropriate active component
		switch m.stage {
		case unifiedStageLoadingCalendars, unifiedStageSelectingCalendars, unifiedStageSaving:
			result, cmd := m.calendarManager.Update(msg)
			m.calendarManager = result.(calendarManagerModel)

			// Handle state transitions
			if m.calendarManager.stage == calendarManagerSuccess {
				m.stage = unifiedStageSuccess
				// Quit immediately after success
				return m, tea.Quit
			}

			if m.calendarManager.stage == calendarManagerError {
				m.err = m.calendarManager.err
				m.stage = unifiedStageError
				return m, nil
			}

			// Sync stage
			switch m.calendarManager.stage {
			case calendarManagerLoading:
				m.stage = unifiedStageLoadingCalendars
			case calendarManagerSelecting:
				m.stage = unifiedStageSelectingCalendars
			case calendarManagerSaving:
				m.stage = unifiedStageSaving
			}

			return m, cmd
		}
	}

	return m, nil
}

type accountInfoLoadedMsg struct {
	accounts []tui.AccountInfo
	err      error
}

func (m unifiedCalendarModel) loadAccountInfoCmd() tea.Cmd {
	return func() tea.Msg {
		// Quickly load account info from config only (no API calls)
		accountInfos := make([]tui.AccountInfo, len(m.accountEmails))
		cfgMgr := config.NewManager()

		for i, email := range m.accountEmails {
			accountInfos[i] = tui.AccountInfo{
				Email: email,
			}

			// Load config to get enabled calendars
			cfg, err := cfgMgr.Load()
			if err == nil && cfg != nil {
				if accountCfg, exists := cfg.Accounts[email]; exists {
					accountInfos[i].EnabledCount = len(accountCfg.EnabledCalendars)

					// Get calendar names
					names := make([]string, 0, len(accountCfg.EnabledCalendars))
					for _, cal := range accountCfg.EnabledCalendars {
						if cal.Name != "" {
							names = append(names, cal.Name)
						} else {
							names = append(names, cal.ID) // Fallback to ID if no name
						}
					}
					accountInfos[i].EnabledCalendars = names
				}
			}
			// We don't know total count without API call, so leave it at 0
			// The UI will show "X enabled" instead of "X of Y"
		}

		return accountInfoLoadedMsg{accounts: accountInfos, err: nil}
	}
}

func (m unifiedCalendarModel) View() string {
	if m.width == 0 {
		return tui.LoadingSpinner() + " Initializing..."
	}

	switch m.stage {
	case unifiedStageLoadingAccounts:
		return m.renderLoadingAccountsScreen()

	case unifiedStageSelectingAccount:
		return m.accountSelector.View()

	case unifiedStageLoadingCalendars, unifiedStageSelectingCalendars, unifiedStageSaving:
		return m.calendarManager.View()

	case unifiedStageSuccess:
		return m.calendarManager.View()

	case unifiedStageError:
		return m.renderErrorScreen()
	}

	return ""
}

func (m unifiedCalendarModel) renderLoadingAccountsScreen() string {
	topBar := tui.RenderTopBar(m.width, "", "Accounts")

	content := tui.CenterVertical(
		tui.StatusIndicator(tui.LoadingSpinner(), "Loading accounts...", tui.PrimaryColor),
		m.height-4,
	)

	footer := tui.RenderFooter(m.width, "", "")

	frame := tui.NewFrame(m.width, m.height)
	return frame.Render(topBar, content, footer)
}

func (m unifiedCalendarModel) renderErrorScreen() string {
	topBar := tui.RenderTopBar(m.width, "", "")

	var b strings.Builder
	b.WriteString(tui.StatusIndicator("✗", "Error", tui.ErrorColor))
	b.WriteString("\n\n")

	errorStyle := lipgloss.NewStyle().Foreground(tui.ErrorColor)
	b.WriteString(errorStyle.Render(fmt.Sprintf("%v", m.err)))

	content := tui.CenterVertical(b.String(), m.height-4)

	helpKeys := map[string]string{
		"q": "exit",
	}
	footer := tui.RenderFooter(m.width, tui.FormatHelpKeys(helpKeys), "")

	frame := tui.NewFrame(m.width, m.height)
	return frame.Render(topBar, content, footer)
}

func listCalendars(accounts []string, store auth.Store) error {
	cfgMgr := config.NewManager()
	ctx := context.Background()

	for i, email := range accounts {
		if i > 0 {
			fmt.Println()
		}

		fmt.Printf("%s:\n", lipgloss.NewStyle().Bold(true).Foreground(tui.PrimaryColor).Render(email))

		// Get enabled calendars from config
		enabledIDs, _ := cfgMgr.GetEnabledCalendarIDs(email)

		enabledSet := make(map[string]bool)

		for _, id := range enabledIDs {
			enabledSet[id] = true
		}

		// Fetch calendars from Google
		client, err := calendar.NewClient(ctx, email, store)
		if err != nil {
			fmt.Printf("  Error: failed to create client: %v\n", err)

			continue
		}

		cals, err := client.GetCalendars(ctx)
		if err != nil {
			fmt.Printf("  Error: failed to fetch calendars: %v\n", err)

			continue
		}

		// Display each calendar
		for _, cal := range cals {
			enabled := len(enabledIDs) == 0 || enabledSet[cal.ID] // Default to enabled if no config
			var marker string
			status := ""

			if enabled {
				marker = lipgloss.NewStyle().Foreground(tui.SuccessColor).Render("✓ ")
			} else {
				marker = "  "
				status = lipgloss.NewStyle().Foreground(tui.MutedColor).Render(" [disabled]")
			}

			name := cal.Summary
			if cal.Primary {
				name += " (primary)"
			}

			fmt.Printf("%s%s%s\n", marker, name, status)
		}
	}

	return nil
}

func selectCalendarsForAccount(email string, store auth.Store) error {
	// Pre-flight check: verify we can access credentials BEFORE starting TUI
	// This prevents hanging inside the TUI if Keychain prompts for permission
	ctx := context.Background()

	_, err := calendar.NewClient(ctx, email, store)
	if err != nil {
		return fmt.Errorf("failed to create calendar client: %w\n\nHint: Check if Keychain is prompting for permission", err)
	}

	p := tea.NewProgram(
		initialCalendarManagerModel(email, store),
		tea.WithInput(os.Stdin),
		tea.WithOutput(os.Stderr),
		tea.WithAltScreen(), // Use alternate screen buffer
	)

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("failed to run calendar selector: %w", err)
	}

	m := finalModel.(calendarManagerModel)
	if m.stage == calendarManagerError {
		return m.err
	}

	if m.cancelled {
		return errors.New("cancelled")
	}

	return nil
}

type calendarManagerStage int

const (
	calendarManagerLoading calendarManagerStage = iota
	calendarManagerSelecting
	calendarManagerSaving
	calendarManagerSuccess
	calendarManagerError
)

type calendarManagerModel struct {
	stage            calendarManagerStage
	email            string
	store            auth.Store
	calendars        []tui.CalendarItem
	calendarSelector tui.CalendarSelectorModel
	selectedCount    int
	cancelled        bool
	err              error
	width            int
	height           int
}

func initialCalendarManagerModel(email string, store auth.Store) calendarManagerModel {
	return calendarManagerModel{
		stage: calendarManagerLoading,
		email: email,
		store: store,
	}
}

func (m calendarManagerModel) Init() tea.Cmd {
	return m.fetchCalendarsCmd()
}

func (m calendarManagerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if m.stage == calendarManagerSelecting {
			result, cmd := m.calendarSelector.Update(msg)
			m.calendarSelector = result.(tui.CalendarSelectorModel)

			return m, cmd
		}

		return m, nil

	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			if m.stage == calendarManagerSuccess || m.stage == calendarManagerError {
				return m, tea.Quit
			}
		}

		if m.stage == calendarManagerSelecting {
			result, cmd := m.calendarSelector.Update(msg)
			m.calendarSelector = result.(tui.CalendarSelectorModel)

			if m.calendarSelector.IsConfirmed() {
				m.stage = calendarManagerSaving

				return m, m.saveConfigCmd()
			}

			if m.calendarSelector.IsCancelled() {
				m.cancelled = true

				return m, tea.Quit
			}

			return m, cmd
		}

	case calendarsLoadedMsg:
		if msg.err != nil {
			m.stage = calendarManagerError
			m.err = msg.err

			return m, nil
		}

		m.calendars = msg.calendars
		m.stage = calendarManagerSelecting
		m.calendarSelector = tui.NewCalendarSelectorModel(m.email, msg.calendars, msg.enabledIDs)

		// Send window size to selector if we have it
		if m.width > 0 {
			result, cmd := m.calendarSelector.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
			m.calendarSelector = result.(tui.CalendarSelectorModel)

			return m, cmd
		}

		return m, nil

	case calendarConfigSavedMsg:
		if msg.err != nil {
			m.stage = calendarManagerError
			m.err = msg.err

			return m, nil
		}

		m.selectedCount = msg.count
		m.stage = calendarManagerSuccess

		// Quit immediately after success
		return m, tea.Quit
	case quitMsg:
		return m, tea.Quit
	}

	if m.stage == calendarManagerSelecting {
		result, cmd := m.calendarSelector.Update(msg)
		m.calendarSelector = result.(tui.CalendarSelectorModel)

		return m, cmd
	}

	return m, nil
}

type calendarsLoadedMsg struct {
	calendars  []tui.CalendarItem
	enabledIDs []string
	err        error
}

type calendarConfigSavedMsg struct {
	count int
	err   error
}

type quitMsg struct{}

func (m calendarManagerModel) fetchCalendarsCmd() tea.Cmd {
	return func() tea.Msg {
		store := auth.NewKeychainStore()
		ctx := context.Background()

		// Create calendar client
		client, err := calendar.NewClient(ctx, m.email, store)
		if err != nil {
			return calendarsLoadedMsg{err: fmt.Errorf("failed to create calendar client: %w", err)}
		}

		// Fetch calendars
		cals, err := client.GetCalendars(ctx)
		if err != nil {
			return calendarsLoadedMsg{err: fmt.Errorf("failed to fetch calendars: %w", err)}
		}

		cfgMgr := config.NewManager()
		// Get currently enabled calendars from config
		enabledIDs, _ := cfgMgr.GetEnabledCalendarIDs(m.email)

		// Convert to TUI items
		items := make([]tui.CalendarItem, len(cals))
		for i, cal := range cals {
			items[i] = tui.CalendarItem{
				ID:      cal.ID,
				Name:    cal.Summary,
				Desc:    cal.Description,
				Primary: cal.Primary,
			}
		}

		// If no config exists, default to all calendars enabled
		if len(enabledIDs) == 0 {
			enabledIDs = make([]string, len(cals))
			for i, cal := range cals {
				enabledIDs[i] = cal.ID
			}
		}

		return calendarsLoadedMsg{calendars: items, enabledIDs: enabledIDs, err: nil}
	}
}

func (m calendarManagerModel) saveConfigCmd() tea.Cmd {
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
			return calendarConfigSavedMsg{err: fmt.Errorf("failed to save config: %w", err)}
		}

		return calendarConfigSavedMsg{count: len(selections), err: nil}
	}
}

func (m calendarManagerModel) View() string {
	if m.width == 0 {
		return tui.LoadingSpinner() + " Initializing..."
	}

	switch m.stage {
	case calendarManagerLoading:
		return m.renderLoadingScreen()

	case calendarManagerSelecting:
		return m.calendarSelector.View()

	case calendarManagerSaving:
		return m.renderSavingScreen()

	case calendarManagerSuccess:
		return m.renderSuccessScreen()

	case calendarManagerError:
		return m.renderErrorScreen()
	}

	return ""
}

func (m calendarManagerModel) renderLoadingScreen() string {
	breadcrumb := fmt.Sprintf("Accounts › %s › Calendars", tui.TruncateWithEllipsis(m.email, 30))
	topBar := tui.RenderTopBar(m.width, "", breadcrumb)

	content := tui.CenterVertical(
		tui.StatusIndicator(tui.LoadingSpinner(), "Loading calendars from Google...", tui.PrimaryColor),
		m.height-4,
	)

	footer := tui.RenderFooter(m.width, "", "")

	frame := tui.NewFrame(m.width, m.height)
	return frame.Render(topBar, content, footer)
}

func (m calendarManagerModel) renderSavingScreen() string {
	breadcrumb := fmt.Sprintf("Accounts › %s › Calendars", tui.TruncateWithEllipsis(m.email, 30))
	topBar := tui.RenderTopBar(m.width, "", breadcrumb)

	content := tui.CenterVertical(
		tui.StatusIndicator(tui.LoadingSpinner(), "Saving configuration...", tui.PrimaryColor),
		m.height-4,
	)

	footer := tui.RenderFooter(m.width, "", "")

	frame := tui.NewFrame(m.width, m.height)
	return frame.Render(topBar, content, footer)
}

func (m calendarManagerModel) renderSuccessScreen() string {
	breadcrumb := fmt.Sprintf("Accounts › %s › Calendars", tui.TruncateWithEllipsis(m.email, 30))
	topBar := tui.RenderTopBar(m.width, "", breadcrumb)

	var b strings.Builder
	b.WriteString(tui.StatusIndicator("✓", "Calendar selection saved!", tui.SuccessColor))
	b.WriteString("\n\n")

	summaryStyle := lipgloss.NewStyle().Foreground(tui.MutedColor)
	b.WriteString(summaryStyle.Render(fmt.Sprintf(
		"Account: %s\nCalendars: %d selected",
		m.email,
		m.selectedCount,
	)))

	content := tui.CenterVertical(b.String(), m.height-4)

	helpKeys := map[string]string{}
	footer := tui.RenderFooter(m.width, tui.FormatHelpKeys(helpKeys), "Closing...")

	frame := tui.NewFrame(m.width, m.height)
	return frame.Render(topBar, content, footer)
}

func (m calendarManagerModel) renderErrorScreen() string {
	breadcrumb := fmt.Sprintf("Accounts › %s › Calendars", tui.TruncateWithEllipsis(m.email, 30))
	topBar := tui.RenderTopBar(m.width, "", breadcrumb)

	var b strings.Builder
	b.WriteString(tui.StatusIndicator("✗", "Error", tui.ErrorColor))
	b.WriteString("\n\n")

	errorStyle := lipgloss.NewStyle().Foreground(tui.ErrorColor)
	b.WriteString(errorStyle.Render(fmt.Sprintf("%v", m.err)))

	content := tui.CenterVertical(b.String(), m.height-4)

	helpKeys := map[string]string{
		"q": "exit",
	}
	footer := tui.RenderFooter(m.width, tui.FormatHelpKeys(helpKeys), "")

	frame := tui.NewFrame(m.width, m.height)
	return frame.Render(topBar, content, footer)
}
