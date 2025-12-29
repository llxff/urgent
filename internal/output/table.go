package output

import (
	"fmt"
	"strings"

	"urgent/internal/calendar"
)

// TableFormatter implements Formatter for table output.
type TableFormatter struct{}

// NewTableFormatter creates a new table formatter.
func NewTableFormatter() *TableFormatter {
	return &TableFormatter{}
}

// FormatEvents formats events as a simple table.
func (t *TableFormatter) FormatEvents(events []*calendar.Event, filter string) (string, error) {
	filtered := FilterEvents(events, filter)

	if len(filtered) == 0 {
		return "No events found.\n", nil
	}

	var b strings.Builder

	b.WriteString(fmt.Sprintf("Found %d event(s):\n\n", len(filtered)))

	for _, event := range filtered {
		startTime := event.Start.Format("3:04 PM")
		endTime := event.End.Format("3:04 PM")

		b.WriteString(fmt.Sprintf("• %s\n", event.Summary))
		b.WriteString(fmt.Sprintf("  %s - %s", startTime, endTime))

		if event.Location != "" {
			b.WriteString(" | " + event.Location)
		}

		b.WriteString(fmt.Sprintf(" | %s\n", event.CalendarName))
		b.WriteString("\n")
	}

	return b.String(), nil
}

// FormatNextEvent formats the next event as simple text.
func (t *TableFormatter) FormatNextEvent(event *calendar.Event, minutesUntil int) (string, error) {
	if event == nil {
		return "No upcoming events.\n", nil
	}

	startTime := event.Start.Format("3:04 PM")
	endTime := event.End.Format("3:04 PM")

	return fmt.Sprintf("%s in %d minutes (%s - %s)\n",
		event.Summary, minutesUntil, startTime, endTime), nil
}
