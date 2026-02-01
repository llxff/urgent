package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"

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
	Short: "Manage calendar selection and accounts",
	Long: `Manage Google Calendar accounts and calendar selection.

This command provides a unified interface for:
  - Adding new Google accounts (press 'a')
  - Removing accounts (press 'd')
  - Selecting which calendars to enable for each account

Examples:
  urgent calendars                     # Interactive account/calendar management
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

	// List mode requires accounts
	if calendarsList {
		if len(accounts) == 0 {
			return errors.New("no accounts connected")
		}
		return listCalendars(accounts, store)
	}

	// Selection mode for specific account
	if calendarsAccount != "" {
		if len(accounts) == 0 {
			return errors.New("no accounts connected")
		}
		found := slices.Contains(accounts, calendarsAccount)
		if !found {
			return fmt.Errorf("account %s not found. Connected accounts: %v", calendarsAccount, accounts)
		}
		return selectCalendarsForAccount(calendarsAccount, store)
	}

	// Interactive mode - show unified TUI (handles empty accounts too)
	return showAccountSelector(accounts, store)
}

func showAccountSelector(accountEmails []string, store auth.Store) error {
	cfgMgr := config.NewManager()
	// Use a unified calendar manager that handles both account selection and calendar selection
	p := tea.NewProgram(
		initialUnifiedCalendarModel(accountEmails, store, cfgMgr),
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

	cfgMgr := config.NewManager()
	p := tea.NewProgram(
		initialCalendarManagerModel(email, store, cfgMgr),
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
