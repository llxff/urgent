package calendar

import (
	"context"
	"fmt"
	"strings"
	"time"

	"urgent/internal/auth"

	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// Event represents a calendar event with additional metadata.
type Event struct {
	ID           string
	Summary      string
	Description  string
	Location     string
	Start        time.Time
	End          time.Time
	CalendarID   string
	CalendarName string
	ColorHex     string
	Account      string
	Status       string
	AllDay       bool
}

// CalendarInfo represents metadata about a calendar.
type CalendarInfo struct {
	ID              string
	Summary         string
	Description     string
	BackgroundColor string // Actual hex color from API (e.g., "#4285f4")
	Primary         bool
	Selected        bool // Indicates if user has access
}

// CalendarProvider defines the interface for calendar operations.
type CalendarProvider interface {
	GetEvents(ctx context.Context, timeMin, timeMax time.Time, calendarIDs []string) ([]*Event, error)
	GetCalendars(ctx context.Context) ([]*CalendarInfo, error)
}

// Client is a Google Calendar API client.
type Client struct {
	service *calendar.Service
	email   string
	store   auth.Store
}

// NewClient creates a new Calendar API client.
//
// It retrieves the user's token from the store and creates an authenticated
// Google Calendar service.
func NewClient(ctx context.Context, email string, store auth.Store) (*Client, error) {
	// Get token from store
	token, err := store.GetToken(email)
	if err != nil {
		return nil, fmt.Errorf("failed to get token for %s: %w", email, err)
	}

	// Get OAuth credentials
	clientID, clientSecret, err := auth.GetOAuthCredentials()
	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth credentials: %w", err)
	}

	// Create OAuth config
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}

	// Create token source with automatic refresh
	tokenSource := config.TokenSource(ctx, token)

	// Create Calendar service
	service, err := calendar.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar service: %w", err)
	}

	return &Client{
		service: service,
		email:   email,
		store:   store,
	}, nil
}

// GetEvents retrieves events within the specified time range.
// If calendarIDs is nil or empty, fetches from all calendars.
// If calendarIDs is provided, only fetches from those specific calendars.
func (c *Client) GetEvents(ctx context.Context, timeMin, timeMax time.Time, calendarIDs []string) ([]*Event, error) {
	// Get calendar list first to fetch calendar names and colors
	calendars, err := c.GetCalendars(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get calendars: %w", err)
	}

	// Create map for quick lookup
	calendarMap := make(map[string]*CalendarInfo)
	for _, cal := range calendars {
		calendarMap[cal.ID] = cal
	}

	// Determine which calendars to query
	var targetCalendars []*CalendarInfo
	if len(calendarIDs) == 0 {
		// Fetch from all calendars
		targetCalendars = calendars
	} else {
		// Fetch only from specified calendars
		for _, id := range calendarIDs {
			if cal, ok := calendarMap[id]; ok {
				targetCalendars = append(targetCalendars, cal)
			}
		}
	}

	var allEvents []*Event

	// Query events from target calendars
	for _, cal := range targetCalendars {
		events, err := c.service.Events.List(cal.ID).
			Context(ctx).
			TimeMin(timeMin.Format(time.RFC3339)).
			TimeMax(timeMax.Format(time.RFC3339)).
			SingleEvents(true).
			OrderBy("startTime").
			Do()
		if err != nil {
			// Log error but continue with other calendars
			fmt.Printf("Warning: failed to fetch events from calendar %s: %v\n", cal.Summary, err)

			continue
		}

		for _, item := range events.Items {
			event := c.convertEvent(item, cal)
			allEvents = append(allEvents, event)
		}
	}

	return allEvents, nil
}

// GetCalendars retrieves the user's calendar list.
func (c *Client) GetCalendars(ctx context.Context) ([]*CalendarInfo, error) {
	// Add timeout to prevent hanging forever
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	calendarList, err := c.service.CalendarList.List().Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list calendars: %w", err)
	}

	var calendars []*CalendarInfo
	for _, item := range calendarList.Items {
		calendars = append(calendars, &CalendarInfo{
			ID:              item.Id,
			Summary:         item.Summary,
			Description:     item.Description,
			BackgroundColor: item.BackgroundColor,
			Primary:         item.Primary,
			Selected:        item.Selected,
		})
	}

	return calendars, nil
}

// convertEvent converts a Google Calendar event to our Event type.
func (c *Client) convertEvent(item *calendar.Event, cal *CalendarInfo) *Event {
	event := &Event{
		ID:           item.Id,
		Summary:      item.Summary,
		Description:  item.Description,
		Location:     item.Location,
		CalendarID:   cal.ID,
		CalendarName: cal.Summary,
		Account:      c.email,
		Status:       item.Status,
	}

	// Parse start time
	if item.Start.DateTime != "" {
		start, _ := time.Parse(time.RFC3339, item.Start.DateTime)
		event.Start = start
		event.AllDay = false
	} else if item.Start.Date != "" {
		start, _ := time.Parse("2006-01-02", item.Start.Date)
		event.Start = start
		event.AllDay = true
	}

	// Parse end time
	if item.End.DateTime != "" {
		end, _ := time.Parse(time.RFC3339, item.End.DateTime)
		event.End = end
	} else if item.End.Date != "" {
		end, _ := time.Parse("2006-01-02", item.End.Date)
		event.End = end
	}

	// Set color from calendar's background color (remove # prefix if present)
	if cal.BackgroundColor != "" {
		event.ColorHex = strings.TrimPrefix(cal.BackgroundColor, "#")
	} else {
		// Default to Google Calendar blue if no color is set
		event.ColorHex = "4285f4"
	}

	return event
}
