package output

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"urgent/internal/calendar"
)

func TestJSONFormatter_FormatEvents(t *testing.T) {
	now := time.Now()
	events := []*calendar.Event{
		{
			Summary:      "Meeting",
			Start:        now,
			End:          now.Add(1 * time.Hour),
			CalendarName: "Work",
			Account:      "test@example.com",
			ColorHex:     "ff0000",
			Status:       "confirmed",
		},
	}

	formatter := NewJSONFormatter()

	output, err := formatter.FormatEvents(events, "all")
	if err != nil {
		t.Fatalf("FormatEvents() error = %v", err)
	}

	if !strings.Contains(output, "Meeting") {
		t.Error("Output should contain event summary")
	}

	if !strings.Contains(output, "\"count\": 1") {
		t.Error("Output should contain count")
	}
}

func TestJSONFormatter_FormatNextEvent(t *testing.T) {
	now := time.Now()
	event := &calendar.Event{
		Summary:      "Meeting",
		Start:        now.Add(15 * time.Minute),
		End:          now.Add(1 * time.Hour),
		CalendarName: "Work",
	}

	formatter := NewJSONFormatter()

	output, err := formatter.FormatNextEvent(event, 15)
	if err != nil {
		t.Fatalf("FormatNextEvent() error = %v", err)
	}

	if !strings.Contains(output, "\"hasEvent\": true") {
		t.Error("Output should indicate event exists")
	}

	if !strings.Contains(output, "\"minutesUntil\": 15") {
		t.Error("Output should contain minutesUntil at top level")
	}

	if !strings.Contains(output, "Meeting") {
		t.Error("Output should contain event summary")
	}

	// Verify minutesUntil is NOT in the event object
	var result NextEventOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to unmarshal output: %v", err)
	}

	if result.MinutesUntil != 15 {
		t.Errorf("Expected minutesUntil = 15, got %d", result.MinutesUntil)
	}
}

func TestTableFormatter_FormatEvents(t *testing.T) {
	now := time.Now()
	events := []*calendar.Event{
		{
			Summary:      "Meeting",
			Start:        now,
			End:          now.Add(1 * time.Hour),
			CalendarName: "Work",
		},
	}

	formatter := NewTableFormatter()

	output, err := formatter.FormatEvents(events, "all")
	if err != nil {
		t.Fatalf("FormatEvents() error = %v", err)
	}

	if !strings.Contains(output, "Meeting") {
		t.Error("Output should contain event summary")
	}
}

func TestFilterEvents(t *testing.T) {
	now := time.Now()
	events := []*calendar.Event{
		{
			ID:    "1",
			Start: now.Add(-2 * time.Hour),
			End:   now.Add(-1 * time.Hour), // Past event
		},
		{
			ID:    "2",
			Start: now.Add(1 * time.Hour),
			End:   now.Add(2 * time.Hour), // Future event
		},
	}

	// Test "all" filter
	result := FilterEvents(events, "all")
	if len(result) != 2 {
		t.Errorf("Expected 2 events with 'all' filter, got %d", len(result))
	}

	// Test "remaining" filter
	result = FilterEvents(events, "remaining")
	if len(result) != 1 {
		t.Errorf("Expected 1 event with 'remaining' filter, got %d", len(result))
	}
}
