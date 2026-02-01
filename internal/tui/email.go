package tui

import (
	"regexp"
	"strings"
)

// Email represents a validated email address.
// Using a distinct type prevents accidental mixing with arbitrary strings.
type Email string

// Simple email validation pattern - checks for basic structure.
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// NewEmail creates an Email from a string.
// Returns the email even if invalid - use IsValid() to check.
func NewEmail(s string) Email {
	return Email(strings.TrimSpace(s))
}

// String returns the email as a string.
func (e Email) String() string {
	return string(e)
}

// IsValid returns true if the email has a valid format.
func (e Email) IsValid() bool {
	return emailRegex.MatchString(string(e))
}

// Domain returns the domain part of the email (after @).
func (e Email) Domain() string {
	parts := strings.SplitN(string(e), "@", 2)
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}

// IsEmpty returns true if the email is empty.
func (e Email) IsEmpty() bool {
	return string(e) == ""
}
