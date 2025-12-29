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
	nextWithin int
)

var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "Show the next upcoming event",
	Long: `Show the next event if it starts within the specified number of minutes.

Useful for cron jobs and status bar integrations.

Examples:
  urgent next --within 30        # Show next event within 30 minutes
  urgent next --within 10 -o json`,
	RunE: runNext,
}

func init() {
	rootCmd.AddCommand(nextCmd)
	nextCmd.Flags().IntVarP(&nextWithin, "within", "w", 60, "Show event if within N minutes")
}

func runNext(_ *cobra.Command, _ []string) error {
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

	// Calculate time range
	now := time.Now()
	endTime := now.Add(time.Duration(nextWithin) * time.Minute)

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
		events, err := client.GetEvents(ctx, now, endTime, enabledCalendarIDs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to fetch events for %s: %v\n", email, err)

			continue
		}

		allEvents = append(allEvents, events...)
	}

	// Find the next event (earliest start time)
	var nextEvent *calendar.Event

	for _, event := range allEvents {
		if event.Start.After(now) {
			if nextEvent == nil || event.Start.Before(nextEvent.Start) {
				nextEvent = event
			}
		}
	}

	// Calculate minutes until event
	minutesUntil := 0
	if nextEvent != nil {
		minutesUntil = int(time.Until(nextEvent.Start).Minutes())
	}

	// Determine output format
	outputFormat := GetOutputFormat()

	var formatter output.Formatter
	if outputFormat == "json" {
		formatter = output.NewJSONFormatter()
	} else {
		formatter = output.NewTableFormatter()
	}

	result, err := formatter.FormatNextEvent(nextEvent, minutesUntil)
	if err != nil {
		return fmt.Errorf("failed to format next event: %w", err)
	}

	fmt.Print(result)

	// Exit with code 1 if no event (useful for scripts)
	if nextEvent == nil {
		os.Exit(1)
	}

	return nil
}
