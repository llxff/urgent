package calendar

import (
	"context"
	"testing"
	"time"

	"google.golang.org/api/calendar/v3"
)

func TestFixtureClient_GetEvents(t *testing.T) {
	now := time.Now()
	events := []*Event{
		{
			ID:         "1",
			Summary:    "Morning Meeting",
			Start:      now.Add(1 * time.Hour),
			End:        now.Add(2 * time.Hour),
			CalendarID: "cal1",
		},
		{
			ID:         "2",
			Summary:    "Lunch",
			Start:      now.Add(4 * time.Hour),
			End:        now.Add(5 * time.Hour),
			CalendarID: "cal2",
		},
		{
			ID:         "3",
			Summary:    "Afternoon Meeting",
			Start:      now.Add(6 * time.Hour),
			End:        now.Add(7 * time.Hour),
			CalendarID: "cal1",
		},
	}

	client := NewFixtureClient(events, nil)

	tests := []struct {
		name        string
		timeMin     time.Time
		timeMax     time.Time
		calendarIDs []string
		wantCount   int
	}{
		{
			name:        "all events, no filter",
			timeMin:     now,
			timeMax:     now.Add(24 * time.Hour),
			calendarIDs: nil,
			wantCount:   3,
		},
		{
			name:        "all events, empty filter",
			timeMin:     now,
			timeMax:     now.Add(24 * time.Hour),
			calendarIDs: []string{},
			wantCount:   3,
		},
		{
			name:        "filter by cal1",
			timeMin:     now,
			timeMax:     now.Add(24 * time.Hour),
			calendarIDs: []string{"cal1"},
			wantCount:   2,
		},
		{
			name:        "filter by cal2",
			timeMin:     now,
			timeMax:     now.Add(24 * time.Hour),
			calendarIDs: []string{"cal2"},
			wantCount:   1,
		},
		{
			name:        "filter by multiple calendars",
			timeMin:     now,
			timeMax:     now.Add(24 * time.Hour),
			calendarIDs: []string{"cal1", "cal2"},
			wantCount:   3,
		},
		{
			name:        "filter by non-existent calendar",
			timeMin:     now,
			timeMax:     now.Add(24 * time.Hour),
			calendarIDs: []string{"nonexistent"},
			wantCount:   0,
		},
		{
			name:        "time range with filter",
			timeMin:     now,
			timeMax:     now.Add(3 * time.Hour),
			calendarIDs: []string{"cal1"},
			wantCount:   1,
		},
		{
			name:        "no events in range with filter",
			timeMin:     now.Add(-24 * time.Hour),
			timeMax:     now,
			calendarIDs: []string{"cal1"},
			wantCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.GetEvents(context.Background(), tt.timeMin, tt.timeMax, tt.calendarIDs)
			if err != nil {
				t.Fatalf("GetEvents returned error: %v", err)
			}

			if len(result) != tt.wantCount {
				t.Errorf("got %d events, want %d", len(result), tt.wantCount)
			}
		})
	}
}

func TestFixtureClient_GetCalendars(t *testing.T) {
	calendars := []*CalendarInfo{
		{ID: "1", Summary: "Work", BackgroundColor: "#a4bdfc"},
		{ID: "2", Summary: "Personal", BackgroundColor: "#7ae7bf"},
	}

	client := NewFixtureClient(nil, calendars)

	result, err := client.GetCalendars(context.Background())
	if err != nil {
		t.Fatalf("GetCalendars() error = %v", err)
	}

	if len(result) != len(calendars) {
		t.Errorf("Expected %d calendars, got %d", len(calendars), len(result))
	}
}

func TestConvertEvent_DateTime(t *testing.T) {
	// Create a mock client just for testing convertEvent
	client := &Client{email: "test@example.com"}

	calInfo := &CalendarInfo{
		ID:              "cal1",
		Summary:         "Test Calendar",
		BackgroundColor: "#a4bdfc",
	}

	now := time.Now()
	startTime := now.Add(1 * time.Hour).Format(time.RFC3339)
	endTime := now.Add(2 * time.Hour).Format(time.RFC3339)

	googleEvent := &calendar.Event{
		Id:          "event1",
		Summary:     "Test Event",
		Description: "Test Description",
		Location:    "Test Location",
		Status:      "confirmed",
		Start:       &calendar.EventDateTime{DateTime: startTime},
		End:         &calendar.EventDateTime{DateTime: endTime},
	}

	event := client.convertEvent(googleEvent, calInfo)

	// Verify all fields
	if event.ID != "event1" {
		t.Errorf("Expected ID 'event1', got '%s'", event.ID)
	}

	if event.Summary != "Test Event" {
		t.Errorf("Expected Summary 'Test Event', got '%s'", event.Summary)
	}

	if event.Description != "Test Description" {
		t.Errorf("Expected Description 'Test Description', got '%s'", event.Description)
	}

	if event.Location != "Test Location" {
		t.Errorf("Expected Location 'Test Location', got '%s'", event.Location)
	}

	if event.CalendarID != "cal1" {
		t.Errorf("Expected CalendarID 'cal1', got '%s'", event.CalendarID)
	}

	if event.CalendarName != "Test Calendar" {
		t.Errorf("Expected CalendarName 'Test Calendar', got '%s'", event.CalendarName)
	}

	if event.Account != "test@example.com" {
		t.Errorf("Expected Account 'test@example.com', got '%s'", event.Account)
	}

	if event.Status != "confirmed" {
		t.Errorf("Expected Status 'confirmed', got '%s'", event.Status)
	}

	if event.AllDay {
		t.Error("Expected AllDay to be false for DateTime event")
	}
	// Color should use calendar's BackgroundColor
	if event.ColorHex != "a4bdfc" {
		t.Errorf("Expected ColorHex 'a4bdfc', got '%s'", event.ColorHex)
	}
}

func TestConvertEvent_AllDay(t *testing.T) {
	client := &Client{email: "test@example.com"}

	calInfo := &CalendarInfo{
		ID:              "cal1",
		Summary:         "Test Calendar",
		BackgroundColor: "#fbd75b",
	}

	googleEvent := &calendar.Event{
		Id:      "event2",
		Summary: "All Day Event",
		Start:   &calendar.EventDateTime{Date: "2025-12-25"},
		End:     &calendar.EventDateTime{Date: "2025-12-26"},
	}

	event := client.convertEvent(googleEvent, calInfo)

	if !event.AllDay {
		t.Error("Expected AllDay to be true for Date event")
	}

	if event.Start.IsZero() {
		t.Error("Start time should be parsed")
	}

	if event.End.IsZero() {
		t.Error("End time should be parsed")
	}
}

func TestConvertEvent_ColorPriority(t *testing.T) {
	client := &Client{email: "test@example.com"}

	tests := []struct {
		name               string
		calBackgroundColor string
		expectedColor      string
	}{
		{
			name:               "calendar has background color",
			calBackgroundColor: "#a4bdfc",
			expectedColor:      "a4bdfc",
		},
		{
			name:               "default when no color",
			calBackgroundColor: "",
			expectedColor:      "4285f4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calInfo := &CalendarInfo{
				ID:              "cal1",
				Summary:         "Test Calendar",
				BackgroundColor: tt.calBackgroundColor,
			}

			googleEvent := &calendar.Event{
				Id:      "event1",
				Summary: "Test",
				Start:   &calendar.EventDateTime{DateTime: time.Now().Format(time.RFC3339)},
				End:     &calendar.EventDateTime{DateTime: time.Now().Add(1 * time.Hour).Format(time.RFC3339)},
			}

			event := client.convertEvent(googleEvent, calInfo)

			if event.ColorHex != tt.expectedColor {
				t.Errorf("Expected color '%s', got '%s'", tt.expectedColor, event.ColorHex)
			}
		})
	}
}

func TestConvertEvent_WithGoogleAPIFixtures(t *testing.T) {
	// Test with realistic Google Calendar API response structures
	client := &Client{email: "user@example.com"}

	tests := []struct {
		name         string
		googleEvent  *calendar.Event
		calInfo      *CalendarInfo
		wantSummary  string
		wantAllDay   bool
		wantColorHex string
	}{
		{
			name: "typical meeting event",
			googleEvent: &calendar.Event{
				Kind:    "calendar#event",
				Id:      "abc123xyz",
				Summary: "Team Standup",
				Start:   &calendar.EventDateTime{DateTime: "2025-12-25T09:00:00-08:00", TimeZone: "America/Los_Angeles"},
				End:     &calendar.EventDateTime{DateTime: "2025-12-25T09:30:00-08:00", TimeZone: "America/Los_Angeles"},
				Status:  "confirmed",
			},
			calInfo: &CalendarInfo{
				ID:              "primary",
				Summary:         "Work Calendar",
				BackgroundColor: "#5484ed",
			},
			wantSummary:  "Team Standup",
			wantAllDay:   false,
			wantColorHex: "5484ed",
		},
		{
			name: "all-day event",
			googleEvent: &calendar.Event{
				Kind:    "calendar#event",
				Id:      "holiday123",
				Summary: "New Year's Day",
				Start:   &calendar.EventDateTime{Date: "2025-01-01"},
				End:     &calendar.EventDateTime{Date: "2025-01-02"},
				Status:  "confirmed",
			},
			calInfo: &CalendarInfo{
				ID:              "holidays",
				Summary:         "US Holidays",
				BackgroundColor: "#dc2127",
			},
			wantSummary:  "New Year's Day",
			wantAllDay:   true,
			wantColorHex: "dc2127",
		},
		{
			name: "event with location and description",
			googleEvent: &calendar.Event{
				Id:          "meeting456",
				Summary:     "Client Meeting",
				Description: "Discuss Q1 roadmap",
				Location:    "Conference Room B",
				Start:       &calendar.EventDateTime{DateTime: "2025-12-25T14:00:00Z"},
				End:         &calendar.EventDateTime{DateTime: "2025-12-25T15:00:00Z"},
				Status:      "confirmed",
			},
			calInfo: &CalendarInfo{
				ID:              "work",
				Summary:         "Work",
				BackgroundColor: "#a4bdfc",
			},
			wantSummary:  "Client Meeting",
			wantAllDay:   false,
			wantColorHex: "a4bdfc",
		},
		{
			name: "event without optional fields",
			googleEvent: &calendar.Event{
				Id:      "simple",
				Summary: "Quick Sync",
				Start:   &calendar.EventDateTime{DateTime: "2025-12-25T10:00:00Z"},
				End:     &calendar.EventDateTime{DateTime: "2025-12-25T10:15:00Z"},
			},
			calInfo: &CalendarInfo{
				ID:      "personal",
				Summary: "Personal",
			},
			wantSummary:  "Quick Sync",
			wantAllDay:   false,
			wantColorHex: "4285f4", // Default color when none specified
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := client.convertEvent(tt.googleEvent, tt.calInfo)

			if event.Summary != tt.wantSummary {
				t.Errorf("Summary: got %q, want %q", event.Summary, tt.wantSummary)
			}

			if event.AllDay != tt.wantAllDay {
				t.Errorf("AllDay: got %v, want %v", event.AllDay, tt.wantAllDay)
			}

			if event.ColorHex != tt.wantColorHex {
				t.Errorf("ColorHex: got %q, want %q", event.ColorHex, tt.wantColorHex)
			}

			if event.Account != "user@example.com" {
				t.Errorf("Account: got %q, want %q", event.Account, "user@example.com")
			}

			if event.CalendarID != tt.calInfo.ID {
				t.Errorf("CalendarID: got %q, want %q", event.CalendarID, tt.calInfo.ID)
			}

			if event.CalendarName != tt.calInfo.Summary {
				t.Errorf("CalendarName: got %q, want %q", event.CalendarName, tt.calInfo.Summary)
			}
		})
	}
}

func TestConvertEvent_EdgeCases(t *testing.T) {
	client := &Client{email: "test@example.com"}
	calInfo := &CalendarInfo{ID: "cal1", Summary: "Calendar"}

	tests := []struct {
		name        string
		googleEvent *calendar.Event
		checkFunc   func(*testing.T, *Event)
	}{
		{
			name: "nil start/end times",
			googleEvent: &calendar.Event{
				Id:      "evt1",
				Summary: "Event",
				Start:   &calendar.EventDateTime{},
				End:     &calendar.EventDateTime{},
			},
			checkFunc: func(t *testing.T, e *Event) {
				// Should handle gracefully without panicking
				if e.Start.IsZero() {
					t.Log("Start time is zero (expected)")
				}
			},
		},
		{
			name: "empty summary",
			googleEvent: &calendar.Event{
				Id:      "evt2",
				Summary: "",
				Start:   &calendar.EventDateTime{DateTime: time.Now().Format(time.RFC3339)},
				End:     &calendar.EventDateTime{DateTime: time.Now().Add(1 * time.Hour).Format(time.RFC3339)},
			},
			checkFunc: func(t *testing.T, e *Event) {
				if e.Summary != "" {
					t.Errorf("Expected empty summary, got %q", e.Summary)
				}
			},
		},
		{
			name: "very long fields",
			googleEvent: &calendar.Event{
				Id:          "evt3",
				Summary:     string(make([]byte, 1000)), // Very long summary
				Description: string(make([]byte, 5000)), // Very long description
				Location:    string(make([]byte, 500)),  // Very long location
				Start:       &calendar.EventDateTime{DateTime: time.Now().Format(time.RFC3339)},
				End:         &calendar.EventDateTime{DateTime: time.Now().Add(1 * time.Hour).Format(time.RFC3339)},
			},
			checkFunc: func(t *testing.T, e *Event) {
				// Should handle long strings without issues
				if len(e.Summary) != 1000 {
					t.Errorf("Summary length: got %d, want 1000", len(e.Summary))
				}

				if len(e.Description) != 5000 {
					t.Errorf("Description length: got %d, want 5000", len(e.Description))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := client.convertEvent(tt.googleEvent, calInfo)
			tt.checkFunc(t, event)
		})
	}
}

func TestEvent_Structure(t *testing.T) {
	// Test that Event structure is properly defined
	now := time.Now()
	event := &Event{
		ID:           "test-id",
		Summary:      "Test Event",
		Description:  "Description",
		Location:     "Location",
		Start:        now,
		End:          now.Add(1 * time.Hour),
		CalendarID:   "cal-id",
		CalendarName: "Calendar Name",
		ColorHex:     "ff0000",
		Account:      "test@example.com",
		Status:       "confirmed",
		AllDay:       false,
	}

	if event.ID == "" {
		t.Error("Event ID should not be empty")
	}

	if event.Start.IsZero() {
		t.Error("Event Start should not be zero")
	}

	if event.End.Before(event.Start) {
		t.Error("Event End should be after Start")
	}
}

func TestCalendarInfo_Structure(t *testing.T) {
	// Test that CalendarInfo structure is properly defined
	cal := &CalendarInfo{
		ID:              "cal-1",
		Summary:         "My Calendar",
		BackgroundColor: "#4285f4",
		Primary:         true,
	}

	if cal.ID == "" {
		t.Error("CalendarInfo ID should not be empty")
	}

	if cal.Summary == "" {
		t.Error("CalendarInfo Summary should not be empty")
	}

	if cal.BackgroundColor == "" {
		t.Error("CalendarInfo BackgroundColor should not be empty")
	}
}

func TestFixtureClient_EmptyData(t *testing.T) {
	// Test fixture client with no data
	client := NewFixtureClient(nil, nil)

	events, err := client.GetEvents(context.Background(), time.Now(), time.Now().Add(24*time.Hour), nil)
	if err != nil {
		t.Fatalf("GetEvents() error = %v", err)
	}

	if len(events) != 0 {
		t.Errorf("Expected 0 events from empty client, got %d", len(events))
	}

	calendars, err := client.GetCalendars(context.Background())
	if err != nil {
		t.Fatalf("GetCalendars() error = %v", err)
	}

	if len(calendars) != 0 {
		t.Errorf("Expected 0 calendars from empty client, got %d", len(calendars))
	}
}

func TestFixtureClient_EdgeCases(t *testing.T) {
	now := time.Now()

	// Event exactly at boundary
	events := []*Event{
		{
			ID:      "boundary",
			Summary: "Boundary Event",
			Start:   now,
			End:     now.Add(1 * time.Hour),
		},
	}

	client := NewFixtureClient(events, nil)

	// Event at exact start time should be included
	result, err := client.GetEvents(context.Background(), now, now.Add(24*time.Hour), nil)
	if err != nil {
		t.Fatalf("GetEvents() error = %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected boundary event to be included, got %d events", len(result))
	}
}
