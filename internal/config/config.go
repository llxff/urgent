package config

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration.
type Config struct {
	Accounts map[string]AccountConfig `yaml:"accounts"`
}

// AccountConfig holds calendar preferences for a single account.
type AccountConfig struct {
	EnabledCalendars []CalendarSelection `yaml:"enabled_calendars"`
}

// CalendarSelection represents a calendar that can be enabled/disabled.
// Supports two formats:
// - Simple string: "primary"
// - Object with ID and optional name: {id: "primary", name: "My Calendar"}.
type CalendarSelection struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name,omitempty"` // Optional, for human readability only
}

// UnmarshalYAML implements custom YAML unmarshaling to support both string and object formats.
func (cs *CalendarSelection) UnmarshalYAML(node *yaml.Node) error {
	// Try to unmarshal as a string first
	var str string

	err := node.Decode(&str)
	if err == nil {
		cs.ID = str
		cs.Name = ""

		return nil
	}

	// If that fails, try to unmarshal as an object
	type rawCalendarSelection CalendarSelection

	var raw rawCalendarSelection
	if err = node.Decode(&raw); err != nil {
		return fmt.Errorf("calendar selection must be either a string or object with 'id' field: %w", err)
	}

	if raw.ID == "" {
		return errors.New("calendar selection object must have non-empty 'id' field")
	}

	cs.ID = raw.ID
	cs.Name = raw.Name

	return nil
}

// MarshalYAML implements custom YAML marshaling to always output object format when name is present.
func (cs CalendarSelection) MarshalYAML() (any, error) {
	// If name is present, marshal as object
	if cs.Name != "" {
		return &struct {
			ID   string `yaml:"id"`
			Name string `yaml:"name"`
		}{
			ID:   cs.ID,
			Name: cs.Name,
		}, nil
	}

	// Otherwise, marshal as simple string
	return cs.ID, nil
}

// NewConfig creates a new empty configuration.
func NewConfig() *Config {
	return &Config{
		Accounts: make(map[string]AccountConfig),
	}
}
