package output

import (
	"time"
	"urgent/internal/calendar"
)

// OutputFormat represents the output format type.
type OutputFormat string

const (
	// FormatTable is the default table format.
	FormatTable OutputFormat = "table"
	// FormatJSON is the JSON format.
	FormatJSON OutputFormat = "json"
)

// Formatter defines the interface for output formatting.
type Formatter interface {
	FormatEvents(events []*calendar.Event, filter string) (string, error)
	FormatNextEvent(event *calendar.Event, minutesUntil int) (string, error)
}

// EventJSON represents an event in JSON output.
type EventJSON struct {
	Summary  string `json:"summary"`
	Start    string `json:"start"`
	End      string `json:"end"`
	Location string `json:"location,omitempty"`
	Calendar string `json:"calendar"`
	Account  string `json:"account"`
	ColorHex string `json:"colorHex"`
	Status   string `json:"status"`
}

// EventsOutput represents the JSON output for events list.
type EventsOutput struct {
	Events []EventJSON `json:"events"`
	Count  int         `json:"count"`
	Filter string      `json:"filter"`
}

// NextEventOutput represents the JSON output for next event.
type NextEventOutput struct {
	HasEvent     bool      `json:"hasEvent"`
	MinutesUntil int       `json:"minutesUntil,omitempty"`
	Event        EventJSON `json:"event,omitempty"`
}

// FilterEvents filters events based on the filter type.
func FilterEvents(events []*calendar.Event, filter string) []*calendar.Event {
	if filter != "remaining" {
		return events
	}

	now := time.Now()

	var remaining []*calendar.Event

	for _, event := range events {
		if event.End.After(now) {
			remaining = append(remaining, event)
		}
	}

	return remaining
}

// convertToEventJSON converts a calendar event to JSON format.
func convertToEventJSON(event *calendar.Event) EventJSON {
	return EventJSON{
		Summary:  event.Summary,
		Start:    event.Start.Format(time.RFC3339),
		End:      event.End.Format(time.RFC3339),
		Location: event.Location,
		Calendar: event.CalendarName,
		Account:  event.Account,
		ColorHex: event.ColorHex,
		Status:   event.Status,
	}
}
