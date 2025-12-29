package config

import (
	"os"
	"path/filepath"
	"testing"
)

// CreateTempConfigDir creates a temporary directory for config tests.
// Returns the directory path and a cleanup function.
func CreateTempConfigDir(t *testing.T) (string, func()) {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "urgent-config-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	cleanup := func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("failed to remove temp dir: %v", err)
		}
	}

	return tempDir, cleanup
}

// WriteTestConfig writes a test configuration to a temporary file.
func WriteTestConfig(t *testing.T, tempDir string, cfg *Config) string {
	t.Helper()

	// Create a manager with custom config path
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create custom manager with this path
	m := &Manager{configPath: configPath}

	err := m.Save(cfg)
	if err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	return configPath
}

// NewTestConfig creates a configuration with sample data for testing.
func NewTestConfig() *Config {
	return &Config{
		Accounts: map[string]AccountConfig{
			"test@example.com": {
				EnabledCalendars: []CalendarSelection{
					{ID: "primary", Name: "test@example.com"},
					{ID: "work@example.com", Name: "Work Calendar"},
				},
			},
			"other@example.com": {
				EnabledCalendars: []CalendarSelection{
					{ID: "primary"},
					{ID: "shared@example.com", Name: "Shared"},
				},
			},
		},
	}
}

// NewManagerWithPath creates a manager with a custom config path for testing.
func NewManagerWithPath(path string) *Manager {
	return &Manager{configPath: path}
}
