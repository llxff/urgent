package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"urgent/internal/auth"
	"urgent/internal/calendar"
	"urgent/internal/config"
	"urgent/internal/output"
)

var (
	todayRemaining bool
)

var todayCmd = &cobra.Command{
	Use:   "today",
	Short: "Show today's calendar events",
	Long: `Display all calendar events for today from all connected accounts.

Examples:
  urgent today                    # All events
  urgent today --remaining        # Only remaining events
  urgent today -o json            # JSON output
  urgent today --remaining -o json`,
	RunE: runToday,
}

func init() {
	rootCmd.AddCommand(todayCmd)
	todayCmd.Flags().BoolVarP(&todayRemaining, "remaining", "r", false, "Show only remaining events")
}

func runToday(_ *cobra.Command, _ []string) error {
	store := auth.NewKeychainStore()
	cfgMgr := config.NewManager()

	// Get all connected accounts
	accounts, err := store.ListAccounts()
	if err != nil {
		return fmt.Errorf("failed to list accounts: %w", err)
	}

	if len(accounts) == 0 {
		return errors.New("no accounts connected. Run 'urgent connect' first")
	}

	// Calculate today's time range
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Fetch events from all accounts
	var allEvents []*calendar.Event

	ctx := context.Background()

	for _, email := range accounts {
		client, err := calendar.NewClient(ctx, email, store)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to create client for %s: %v\n", email, err)

			continue
		}

		// Get enabled calendar IDs from config
		enabledCalendarIDs, err := cfgMgr.GetEnabledCalendarIDs(email)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load config for %s: %v\n", email, err)
			// Continue with all calendars
			enabledCalendarIDs = nil
		}

		// Fetch events (nil = all calendars, backward compatible)
		events, err := client.GetEvents(ctx, startOfDay, endOfDay, enabledCalendarIDs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to fetch events for %s: %v\n", email, err)

			continue
		}

		allEvents = append(allEvents, events...)
	}

	// Determine output format
	outputFormat := GetOutputFormat()

	filter := "all"
	if todayRemaining {
		filter = "remaining"
	}

	// Format and output
	var formatter output.Formatter
	if outputFormat == "json" {
		formatter = output.NewJSONFormatter()
	} else {
		formatter = output.NewTableFormatter()
	}

	result, err := formatter.FormatEvents(allEvents, filter)
	if err != nil {
		return fmt.Errorf("failed to format events: %w", err)
	}

	fmt.Print(result)

	// Exit with code 1 if no events (useful for scripts)
	filtered := output.FilterEvents(allEvents, filter)
	if len(filtered) == 0 {
		os.Exit(1)
	}

	return nil
}
