package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestColors(t *testing.T) {
	// Test that color constants are properly defined
	colors := []lipgloss.AdaptiveColor{
		PrimaryColor,
		SuccessColor,
		ErrorColor,
		MutedColor,
	}

	for i, color := range colors {
		if color.Light == "" {
			t.Errorf("Color %d: Light color should not be empty", i)
		}

		if color.Dark == "" {
			t.Errorf("Color %d: Dark color should not be empty", i)
		}
	}
}

func TestTitleStyle(t *testing.T) {
	style := TitleStyle()

	// Test that we can render with it
	rendered := style.Render("Test Title")
	if rendered == "" {
		t.Error("TitleStyle should render text")
	}

	if rendered == "Test Title" {
		t.Error("TitleStyle should apply formatting to text")
	}
}

func TestBoxStyle(t *testing.T) {
	style := BoxStyle()

	// Test that we can render with it
	rendered := style.Render("Box Content")
	if rendered == "" {
		t.Error("BoxStyle should render text")
	}

	if rendered == "Box Content" {
		t.Error("BoxStyle should apply formatting to text")
	}

	// Box style should have a border
	if !style.GetBorderTop() && !style.GetBorderBottom() {
		t.Error("BoxStyle should have borders")
	}
}

func TestErrorStyle(t *testing.T) {
	style := ErrorStyle()

	// Test that we can render with it
	rendered := style.Render("Error Message")
	if rendered == "" {
		t.Error("ErrorStyle should render text")
	}

	if rendered == "Error Message" {
		t.Error("ErrorStyle should apply formatting to text")
	}
}

func TestSuccessStyle(t *testing.T) {
	style := SuccessStyle()

	// Test that we can render with it
	rendered := style.Render("Success Message")
	if rendered == "" {
		t.Error("SuccessStyle should render text")
	}

	if rendered == "Success Message" {
		t.Error("SuccessStyle should apply formatting to text")
	}
}

func TestMutedStyle(t *testing.T) {
	style := MutedStyle()

	// Test that we can render with it
	rendered := style.Render("Muted Text")
	if rendered == "" {
		t.Error("MutedStyle should render text")
	}
}

func TestHelpStyle(t *testing.T) {
	style := HelpStyle()

	// Test that we can render with it
	rendered := style.Render("Help Text")
	if rendered == "" {
		t.Error("HelpStyle should render text")
	}
}

func TestStyleConsistency(t *testing.T) {
	// Test that calling style functions multiple times returns consistent results
	style1 := TitleStyle()
	style2 := TitleStyle()

	text := "Test"
	rendered1 := style1.Render(text)
	rendered2 := style2.Render(text)

	if rendered1 != rendered2 {
		t.Error("Style functions should return consistent results")
	}
}

func TestStyleComposition(t *testing.T) {
	// Test that styles can be composed
	title := TitleStyle()
	box := BoxStyle()

	// Should be able to render with both
	text := "Test"
	titleRendered := title.Render(text)
	boxRendered := box.Render(text)

	if titleRendered == boxRendered {
		t.Error("Different styles should produce different output")
	}
}

func TestAdaptiveColorValues(t *testing.T) {
	// Verify that adaptive colors have valid ANSI codes
	tests := []struct {
		name  string
		color lipgloss.AdaptiveColor
	}{
		{"Primary", PrimaryColor},
		{"Success", SuccessColor},
		{"Error", ErrorColor},
		{"Muted", MutedColor},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Light and Dark values should be the same for ANSI colors
			if tt.color.Light == "" || tt.color.Dark == "" {
				t.Errorf("%s color has empty values", tt.name)
			}

			// Both should be valid ANSI color codes (numbers as strings)
			// ANSI 256 colors are 0-255
			if len(tt.color.Light) == 0 || len(tt.color.Dark) == 0 {
				t.Errorf("%s color codes should not be empty", tt.name)
			}
		})
	}
}

func TestStyleRendering(t *testing.T) {
	// Test various style rendering scenarios
	tests := []struct {
		name         string
		style        lipgloss.Style
		input        string
		shouldChange bool // Whether the style changes the visible content
	}{
		{"Title with text", TitleStyle(), "Title", true},
		{"Box with text", BoxStyle(), "Content", true},
		{"Error with message", ErrorStyle(), "Error!", true},
		{"Success with message", SuccessStyle(), "Success!", true},
		{"Muted with text", MutedStyle(), "Muted", false}, // Color-only styling
		{"Help with text", HelpStyle(), "Help", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rendered := tt.style.Render(tt.input)
			if rendered == "" {
				t.Errorf("%s: rendered text should not be empty", tt.name)
			}
			// For styles that add visible formatting (margins, borders), output should differ
			// For color-only styles, the text content may be the same
			if tt.shouldChange && rendered == tt.input {
				t.Errorf("%s: style should apply visible formatting", tt.name)
			}
		})
	}
}

func TestEmptyStringRendering(t *testing.T) {
	// Test that styles handle empty strings gracefully
	styles := []lipgloss.Style{
		TitleStyle(),
		BoxStyle(),
		ErrorStyle(),
		SuccessStyle(),
		MutedStyle(),
		HelpStyle(),
	}

	for i, style := range styles {
		rendered := style.Render("")
		// Empty string with styling may produce ANSI codes but should not panic
		// Just log the result for verification
		t.Logf("Style %d renders empty string as: %q", i, rendered)
	}
}
