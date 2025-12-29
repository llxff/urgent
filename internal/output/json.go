package output

import (
	"encoding/json"
	"fmt"

	"urgent/internal/calendar"
)

// JSONFormatter implements Formatter for JSON output.
type JSONFormatter struct{}

// NewJSONFormatter creates a new JSON formatter.
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

// FormatEvents formats events as JSON.
func (f *JSONFormatter) FormatEvents(events []*calendar.Event, filter string) (string, error) {
	filtered := FilterEvents(events, filter)

	eventsList := make([]EventJSON, 0, len(filtered))
	for _, event := range filtered {
		eventsList = append(eventsList, convertToEventJSON(event))
	}

	output := EventsOutput{
		Events: eventsList,
		Count:  len(eventsList),
		Filter: filter,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal events: %w", err)
	}

	return string(data), nil
}

// FormatNextEvent formats the next event as JSON.
func (f *JSONFormatter) FormatNextEvent(event *calendar.Event, minutesUntil int) (string, error) {
	output := NextEventOutput{
		HasEvent: event != nil,
	}

	if event != nil {
		output.MinutesUntil = minutesUntil
		output.Event = convertToEventJSON(event)
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal next event: %w", err)
	}

	return string(data), nil
}
