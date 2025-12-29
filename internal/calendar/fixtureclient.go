package calendar

import (
	"context"
	"encoding/json"
	"os"
	"time"
)

// FixtureClient is a test implementation of CalendarProvider using JSON fixtures.
type FixtureClient struct {
	events    []*Event
	calendars []*CalendarInfo
}

// NewFixtureClient creates a new fixture-based calendar client.
func NewFixtureClient(events []*Event, calendars []*CalendarInfo) *FixtureClient {
	return &FixtureClient{
		events:    events,
		calendars: calendars,
	}
}

// NewFixtureClientFromFile loads fixtures from a JSON file.
func NewFixtureClientFromFile(filename string) (*FixtureClient, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var fixtures struct {
		Events    []*Event        `json:"events"`
		Calendars []*CalendarInfo `json:"calendars"`
	}

	if err := json.Unmarshal(data, &fixtures); err != nil {
		return nil, err
	}

	return &FixtureClient{
		events:    fixtures.Events,
		calendars: fixtures.Calendars,
	}, nil
}

// GetEvents returns filtered events from fixtures.
// If calendarIDs is nil or empty, returns all events.
// If calendarIDs is provided, only returns events from those calendars.
func (f *FixtureClient) GetEvents(ctx context.Context, timeMin, timeMax time.Time, calendarIDs []string) ([]*Event, error) {
	var filtered []*Event

	// Create calendar ID set for quick lookup
	calendarIDSet := make(map[string]bool)
	for _, id := range calendarIDs {
		calendarIDSet[id] = true
	}

	filterByCalendar := len(calendarIDs) > 0

	for _, event := range f.events {
		// Check time range
		if (event.Start.Equal(timeMin) || event.Start.After(timeMin)) &&
			(event.Start.Equal(timeMax) || event.Start.Before(timeMax)) {
			// Check calendar filter
			if !filterByCalendar || calendarIDSet[event.CalendarID] {
				filtered = append(filtered, event)
			}
		}
	}

	return filtered, nil
}

// GetCalendars returns calendars from fixtures.
func (f *FixtureClient) GetCalendars(ctx context.Context) ([]*CalendarInfo, error) {
	return f.calendars, nil
}
