package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"urgent/internal/auth"
	"urgent/internal/calendar"
	"urgent/internal/config"
	"urgent/internal/tui"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Unified calendar model stages.
type unifiedCalendarStage int

const (
	unifiedStageCheckingSetup unifiedCalendarStage = iota
	unifiedStageRunningSetup
	unifiedStageLoadingAccounts
	unifiedStageSelectingAccount
	unifiedStageLoadingCalendars
	unifiedStageSelectingCalendars
	unifiedStageSaving
	unifiedStageOAuth
	unifiedStageSuccess
	unifiedStageError
)

// unifiedCalendarModel manages account selection, calendar selection, and account management.
type unifiedCalendarModel struct {
	stage           unifiedCalendarStage
	store           auth.Store
	cfgMgr          config.ReadWriter
	accountEmails   []string
	accountSelector tui.AccountSelectorModel
	selectedEmail   string
	calendarManager calendarManagerModel
	cancelled       bool
	err             error
	width           int
	height          int

	// Modal state
	activeModal *tui.ModalModel
	modalEmail  string // For disconnect: which account

	// OAuth state
	oauthCtx    context.Context
	oauthCancel context.CancelFunc
	oauthURL    string

	// Spinner for loading states
	spinner spinner.Model
}

func initialUnifiedCalendarModel(accountEmails []string, store auth.Store, cfgMgr config.ReadWriter) unifiedCalendarModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(tui.PrimaryColor)

	return unifiedCalendarModel{
		stage:         unifiedStageCheckingSetup,
		store:         store,
		cfgMgr:        cfgMgr,
		accountEmails: accountEmails,
		spinner:       s,
	}
}

func (m unifiedCalendarModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.checkCredentialsCmd())
}

func (m unifiedCalendarModel) checkCredentialsCmd() tea.Cmd {
	return func() tea.Msg {
		_, _, err := m.store.GetOAuthCredentials()
		if err != nil {
			return tui.SetupRequiredMsg{}
		}
		return tui.CredentialsOKMsg{}
	}
}

func (m unifiedCalendarModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle modal overlay first (takes priority)
	if m.activeModal != nil {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			return m.handleModalKeyMsg(keyMsg)
		}
	}

	switch msg := msg.(type) {
	case spinner.TickMsg:
		return m.handleSpinnerTick(msg)

	case tea.WindowSizeMsg:
		return m.handleWindowSizeMsg(msg)

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case tui.SetupRequiredMsg:
		m.err = errors.New("OAuth credentials not configured.\n\nRun 'urgent setup' first to configure your OAuth credentials")
		m.stage = unifiedStageError
		return m, nil

	case tui.CredentialsOKMsg:
		m.stage = unifiedStageLoadingAccounts
		return m, m.loadAccountInfoCmd()

	case tui.AddAccountRequestMsg:
		return m.startOAuthFlow()

	case tui.DeleteAccountRequestMsg:
		m.modalEmail = msg.Email.String()
		modal := tui.NewConfirmModal("Remove from Keychain?", msg.Email.String())
		m.activeModal = &modal
		return m, nil

	case tui.OAuthURLMsg:
		m.oauthURL = msg.URL
		return m, nil

	case tui.OAuthCompleteMsg:
		return m.handleOAuthComplete(msg)

	case tui.AccountDeletedMsg:
		return m.handleAccountDeleted(msg)

	case accountInfoLoadedMsg:
		return m.handleAccountInfoLoaded(msg)

	case quitMsg:
		return m, tea.Quit

	default:
		return m.forwardToCalendarManager(msg)
	}
}

// handleModalKeyMsg processes key input when a modal is active.
func (m unifiedCalendarModel) handleModalKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	updated, cmd := m.activeModal.Update(msg)
	modal := updated.(tui.ModalModel)
	m.activeModal = &modal

	if m.activeModal.IsConfirmed() {
		return m.handleModalConfirm()
	}
	if m.activeModal.IsCancelled() {
		m.activeModal = nil
		m.modalEmail = ""
		return m, nil
	}
	return m, cmd
}

// handleSpinnerTick updates the spinner during loading states.
func (m unifiedCalendarModel) handleSpinnerTick(msg spinner.TickMsg) (tea.Model, tea.Cmd) {
	if m.stage == unifiedStageCheckingSetup || m.stage == unifiedStageLoadingAccounts || m.stage == unifiedStageOAuth {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

// handleWindowSizeMsg processes window resize events.
func (m unifiedCalendarModel) handleWindowSizeMsg(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height

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
}

// handleKeyMsg processes keyboard input based on current stage.
func (m unifiedCalendarModel) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global quit with ctrl+c
	if msg.String() == "ctrl+c" {
		if m.oauthCancel != nil {
			m.oauthCancel()
		}
		m.cancelled = true
		return m, tea.Quit
	}

	switch m.stage {
	case unifiedStageSelectingAccount:
		return m.handleAccountSelectorKeyMsg(msg)
	case unifiedStageOAuth:
		return m.handleOAuthKeyMsg(msg)
	case unifiedStageLoadingCalendars, unifiedStageSelectingCalendars, unifiedStageSaving:
		return m.handleCalendarManagerKeyMsg(msg)
	case unifiedStageError:
		if msg.String() == "q" || msg.String() == "esc" {
			return m, tea.Quit
		}
	}
	return m, nil
}

// handleAccountSelectorKeyMsg processes input in account selection stage.
func (m unifiedCalendarModel) handleAccountSelectorKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	result, cmd := m.accountSelector.Update(msg)
	m.accountSelector = result.(tui.AccountSelectorModel)

	if m.accountSelector.IsConfirmed() {
		return m.transitionToCalendarSelection()
	}
	if m.accountSelector.IsCancelled() {
		m.cancelled = true
		return m, tea.Quit
	}
	return m, cmd
}

// transitionToCalendarSelection moves from account selection to calendar selection.
func (m unifiedCalendarModel) transitionToCalendarSelection() (tea.Model, tea.Cmd) {
	m.selectedEmail = m.accountSelector.GetSelectedAccount()
	if m.selectedEmail == "" {
		m.err = errors.New("no account selected")
		m.stage = unifiedStageError
		return m, nil
	}

	m.calendarManager = initialCalendarManagerModel(m.selectedEmail, m.store, m.cfgMgr)
	m.stage = unifiedStageLoadingCalendars

	initCmd := m.calendarManager.Init()
	if m.width > 0 {
		sizeCmd := func() tea.Msg {
			return tea.WindowSizeMsg{Width: m.width, Height: m.height}
		}
		return m, tea.Batch(initCmd, sizeCmd)
	}
	return m, initCmd
}

// handleOAuthKeyMsg processes input during OAuth flow.
func (m unifiedCalendarModel) handleOAuthKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		if m.oauthCancel != nil {
			m.oauthCancel()
		}
		m.oauthURL = ""
		m.stage = unifiedStageLoadingAccounts
		return m, m.loadAccountInfoCmd()
	}
	return m, nil
}

// handleCalendarManagerKeyMsg processes input in calendar manager stages.
func (m unifiedCalendarModel) handleCalendarManagerKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Allow going back from calendar selection
	if msg.String() == "esc" && m.stage == unifiedStageSelectingCalendars {
		m.stage = unifiedStageLoadingAccounts
		return m, m.loadAccountInfoCmd()
	}

	result, cmd := m.calendarManager.Update(msg)
	m.calendarManager = result.(calendarManagerModel)

	return m.handleCalendarManagerTransitions(cmd)
}

// handleCalendarManagerTransitions checks for stage transitions from calendar manager.
func (m unifiedCalendarModel) handleCalendarManagerTransitions(cmd tea.Cmd) (tea.Model, tea.Cmd) {
	if m.calendarManager.IsSuccess() {
		// Go back to account list after successful calendar selection
		m.stage = unifiedStageLoadingAccounts
		return m, m.loadAccountInfoCmd()
	}
	if m.calendarManager.IsError() {
		m.err = m.calendarManager.GetError()
		m.stage = unifiedStageError
		return m, nil
	}
	if m.calendarManager.IsCancelled() {
		m.stage = unifiedStageLoadingAccounts
		return m, m.loadAccountInfoCmd()
	}
	m.syncCalendarManagerStage()
	return m, cmd
}

// handleAccountInfoLoaded processes loaded account information.
func (m unifiedCalendarModel) handleAccountInfoLoaded(msg accountInfoLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		m.stage = unifiedStageError
		return m, nil
	}

	m.accountSelector = tui.NewAccountSelectorModel(msg.accounts)
	m.stage = unifiedStageSelectingAccount

	if m.width > 0 {
		result, cmd := m.accountSelector.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		m.accountSelector = result.(tui.AccountSelectorModel)
		return m, cmd
	}
	return m, nil
}

// forwardToCalendarManager forwards messages to calendar manager when appropriate.
func (m unifiedCalendarModel) forwardToCalendarManager(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.stage {
	case unifiedStageLoadingCalendars, unifiedStageSelectingCalendars, unifiedStageSaving:
		result, cmd := m.calendarManager.Update(msg)
		m.calendarManager = result.(calendarManagerModel)
		return m.handleCalendarManagerTransitions(cmd)
	}
	return m, nil
}

// syncCalendarManagerStage synchronizes the unified model's stage with the calendar manager's stage.
func (m *unifiedCalendarModel) syncCalendarManagerStage() {
	switch {
	case m.calendarManager.IsLoading():
		m.stage = unifiedStageLoadingCalendars
	case m.calendarManager.IsSelecting():
		m.stage = unifiedStageSelectingCalendars
	case m.calendarManager.IsSaving():
		m.stage = unifiedStageSaving
	}
}

func (m unifiedCalendarModel) handleModalConfirm() (tea.Model, tea.Cmd) {
	email := m.modalEmail
	m.activeModal = nil
	m.modalEmail = ""

	// Delete the account
	return m, m.deleteAccountCmd(email)
}

func (m unifiedCalendarModel) deleteAccountCmd(email string) tea.Cmd {
	return func() tea.Msg {
		var errs []error

		// Delete token from Keychain
		if err := m.store.DeleteToken(email); err != nil {
			errs = append(errs, fmt.Errorf("keychain: %w", err))
		}

		// Delete account config (calendar settings)
		if err := m.cfgMgr.DeleteAccount(email); err != nil {
			errs = append(errs, fmt.Errorf("config: %w", err))
		}

		if len(errs) > 0 {
			return tui.AccountDeletedMsg{Email: email, Err: errors.Join(errs...)}
		}

		return tui.AccountDeletedMsg{Email: email, Err: nil}
	}
}

func (m unifiedCalendarModel) startOAuthFlow() (tea.Model, tea.Cmd) {
	m.oauthCtx, m.oauthCancel = context.WithCancel(context.Background())
	m.stage = unifiedStageOAuth
	m.oauthURL = ""
	urlChan := make(chan string, 1)

	authCmd := func() tea.Msg {
		clientID, clientSecret, err := m.store.GetOAuthCredentials()
		if err != nil {
			return tui.OAuthCompleteMsg{Err: err}
		}

		manager := auth.NewManager(m.store, clientID, clientSecret)

		result, err := manager.AuthenticateWithURLCallback(m.oauthCtx, func(url string) {
			select {
			case urlChan <- url:
			default:
			}
		})

		if err != nil {
			return tui.OAuthCompleteMsg{Err: err}
		}

		// Save token
		if err := m.store.SaveToken(result.Email, result.Token); err != nil {
			return tui.OAuthCompleteMsg{Err: fmt.Errorf("failed to save token: %w", err)}
		}

		return tui.OAuthCompleteMsg{Email: result.Email}
	}

	urlCmd := func() tea.Msg {
		select {
		case url := <-urlChan:
			return tui.OAuthURLMsg{URL: url}
		case <-m.oauthCtx.Done():
			return nil
		}
	}

	return m, tea.Batch(authCmd, urlCmd)
}

func (m unifiedCalendarModel) handleOAuthComplete(msg tui.OAuthCompleteMsg) (tea.Model, tea.Cmd) {
	m.oauthURL = ""
	// Cancel context to clean up any pending goroutines (e.g., URL channel reader)
	if m.oauthCancel != nil {
		m.oauthCancel()
		m.oauthCancel = nil
	}

	if msg.Err != nil {
		// Check if it was cancelled
		if errors.Is(msg.Err, context.Canceled) || msg.Err.Error() == "authentication cancelled" {
			m.stage = unifiedStageLoadingAccounts
			return m, m.loadAccountInfoCmd()
		}
		m.err = msg.Err
		m.stage = unifiedStageError
		return m, nil
	}

	// Account added successfully - add to list and go to calendar selection
	m.accountEmails = append(m.accountEmails, msg.Email)
	m.selectedEmail = msg.Email

	// Initialize calendar manager for the new account
	m.calendarManager = initialCalendarManagerModel(msg.Email, m.store, m.cfgMgr)
	m.stage = unifiedStageLoadingCalendars

	initCmd := m.calendarManager.Init()
	if m.width > 0 {
		sizeCmd := func() tea.Msg {
			return tea.WindowSizeMsg{Width: m.width, Height: m.height}
		}
		return m, tea.Batch(initCmd, sizeCmd)
	}

	return m, initCmd
}

func (m unifiedCalendarModel) handleAccountDeleted(msg tui.AccountDeletedMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		m.err = fmt.Errorf("failed to delete account: %w", msg.Err)
		m.stage = unifiedStageError
		return m, nil
	}

	// Remove from local list
	newEmails := make([]string, 0, len(m.accountEmails))
	for _, email := range m.accountEmails {
		if email != msg.Email {
			newEmails = append(newEmails, email)
		}
	}
	m.accountEmails = newEmails

	// Reload account list
	m.stage = unifiedStageLoadingAccounts
	return m, m.loadAccountInfoCmd()
}

// Message types for account loading.
type accountInfoLoadedMsg struct {
	accounts []tui.AccountInfo
	err      error
}

func (m unifiedCalendarModel) loadAccountInfoCmd() tea.Cmd {
	return func() tea.Msg {
		// Quickly load account info from config only (no API calls)
		accountInfos := make([]tui.AccountInfo, len(m.accountEmails))

		// Load config once outside the loop
		cfg, cfgErr := m.cfgMgr.Load()

		for i, email := range m.accountEmails {
			accountInfos[i] = tui.AccountInfo{
				Email: email,
			}

			// Use loaded config to get enabled calendars
			if cfgErr == nil && cfg != nil {
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
		return m.spinner.View() + " Initializing..."
	}

	var baseView string

	switch m.stage {
	case unifiedStageCheckingSetup:
		baseView = m.renderCheckingSetupScreen()

	case unifiedStageLoadingAccounts:
		baseView = m.renderLoadingAccountsScreen()

	case unifiedStageSelectingAccount:
		baseView = m.accountSelector.View()

	case unifiedStageOAuth:
		baseView = m.renderOAuthScreen()

	case unifiedStageLoadingCalendars, unifiedStageSelectingCalendars, unifiedStageSaving:
		baseView = m.calendarManager.View()

	case unifiedStageSuccess:
		baseView = m.calendarManager.View()

	case unifiedStageError:
		baseView = m.renderErrorScreen()

	default:
		baseView = ""
	}

	// Overlay modal if active
	if m.activeModal != nil {
		return m.activeModal.RenderOver(baseView, m.width, m.height)
	}

	return baseView
}

func (m unifiedCalendarModel) renderCheckingSetupScreen() string {
	return tui.RenderStatusScreen(m.width, m.height, "", "", m.spinner.View(), "Checking credentials...", tui.PrimaryColor, "", "")
}

func (m unifiedCalendarModel) renderLoadingAccountsScreen() string {
	return tui.RenderStatusScreen(m.width, m.height, "", "", m.spinner.View(), "Loading accounts...", tui.PrimaryColor, "", "")
}

func (m unifiedCalendarModel) renderOAuthScreen() string {
	topBar := tui.RenderTopBar(m.width, "", "Add Account")

	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(tui.PrimaryColor)

	b.WriteString(titleStyle.Render("Authenticate with Google"))
	b.WriteString("\n\n")

	if m.oauthURL == "" {
		b.WriteString(lipgloss.NewStyle().
			Foreground(tui.MutedColor).
			Render("Preparing authorization..."))
	} else {
		b.WriteString(lipgloss.NewStyle().
			Foreground(tui.SuccessColor).
			Render("✓ Browser opened"))
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().
			Foreground(tui.MutedColor).
			Render("Please complete authorization in your browser"))
		b.WriteString("\n\n")
		b.WriteString(lipgloss.NewStyle().
			Foreground(tui.MutedColor).
			Render("If browser didn't open, visit:"))
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().
			Foreground(tui.PrimaryColor).
			Render(m.oauthURL))
	}

	// Content at top with padding, not centered
	content := lipgloss.NewStyle().Padding(1, 2).Render(b.String())

	helpKeys := map[string]string{
		"esc": "cancel",
	}
	footer := tui.RenderFooter(m.width, tui.FormatHelpKeys(helpKeys), "")

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

// Calendar manager model stages.
type calendarManagerStage int

const (
	calendarManagerLoading calendarManagerStage = iota
	calendarManagerSelecting
	calendarManagerSaving
	calendarManagerSuccess
	calendarManagerError
)

// calendarManagerModel handles calendar selection for a single account.
type calendarManagerModel struct {
	stage            calendarManagerStage
	email            string
	store            auth.Store
	cfgMgr           config.ReadWriter
	calendars        []tui.CalendarItem
	calendarSelector tui.CalendarSelectorModel
	selectedCount    int
	cancelled        bool
	err              error
	width            int
	height           int

	// Spinner for loading states
	spinner spinner.Model
}

func initialCalendarManagerModel(email string, store auth.Store, cfgMgr config.ReadWriter) calendarManagerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(tui.PrimaryColor)

	return calendarManagerModel{
		stage:   calendarManagerLoading,
		email:   email,
		store:   store,
		cfgMgr:  cfgMgr,
		spinner: s,
	}
}

func (m calendarManagerModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.fetchCalendarsCmd())
}

func (m calendarManagerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		// Update spinner during loading/saving states
		if m.stage == calendarManagerLoading || m.stage == calendarManagerSaving {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

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
				return m, tea.Batch(m.spinner.Tick, m.saveConfigCmd())
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

// Message types for calendar operations.
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
		ctx := context.Background()

		// Create calendar client using injected store
		client, err := calendar.NewClient(ctx, m.email, m.store)
		if err != nil {
			return calendarsLoadedMsg{err: fmt.Errorf("failed to create calendar client: %w", err)}
		}

		// Fetch calendars
		cals, err := client.GetCalendars(ctx)
		if err != nil {
			return calendarsLoadedMsg{err: fmt.Errorf("failed to fetch calendars: %w", err)}
		}

		// Get currently enabled calendars from config
		enabledIDs, _ := m.cfgMgr.GetEnabledCalendarIDs(m.email)

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
		err := m.cfgMgr.SetEnabledCalendars(m.email, selections)
		if err != nil {
			return calendarConfigSavedMsg{err: fmt.Errorf("failed to save config: %w", err)}
		}

		return calendarConfigSavedMsg{count: len(selections), err: nil}
	}
}

func (m calendarManagerModel) View() string {
	if m.width == 0 {
		return m.spinner.View() + " Initializing..."
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
	return tui.RenderStatusScreen(m.width, m.height, "", breadcrumb, m.spinner.View(), "Loading calendars from Google...", tui.PrimaryColor, "", "")
}

func (m calendarManagerModel) renderSavingScreen() string {
	breadcrumb := fmt.Sprintf("Accounts › %s › Calendars", tui.TruncateWithEllipsis(m.email, 30))
	return tui.RenderStatusScreen(m.width, m.height, "", breadcrumb, m.spinner.View(), "Saving configuration...", tui.PrimaryColor, "", "")
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

// State query methods - allow parent to check state without accessing internal stage field.

// IsSuccess returns true when calendar selection completed successfully.
func (m calendarManagerModel) IsSuccess() bool {
	return m.stage == calendarManagerSuccess
}

// IsError returns true when an error occurred.
func (m calendarManagerModel) IsError() bool {
	return m.stage == calendarManagerError
}

// GetError returns the error if in error state.
func (m calendarManagerModel) GetError() error {
	return m.err
}

// IsCancelled returns true when user cancelled the operation.
func (m calendarManagerModel) IsCancelled() bool {
	return m.cancelled
}

// IsLoading returns true when loading calendars.
func (m calendarManagerModel) IsLoading() bool {
	return m.stage == calendarManagerLoading
}

// IsSelecting returns true when user is selecting calendars.
func (m calendarManagerModel) IsSelecting() bool {
	return m.stage == calendarManagerSelecting
}

// IsSaving returns true when saving configuration.
func (m calendarManagerModel) IsSaving() bool {
	return m.stage == calendarManagerSaving
}
