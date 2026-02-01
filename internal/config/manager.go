package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Reader provides read-only access to configuration.
type Reader interface {
	Load() (*Config, error)
	GetEnabledCalendarIDs(email string) ([]string, error)
}

// Writer provides write access to configuration.
type Writer interface {
	SetEnabledCalendars(email string, selections []CalendarSelection) error
	DeleteAccount(email string) error
}

// ReadWriter combines Reader and Writer interfaces.
type ReadWriter interface {
	Reader
	Writer
	Save(cfg *Config) error
}

// Manager handles configuration persistence with XDG Base Directory support.
// It implements the ReadWriter interface.
type Manager struct {
	configPath string
}

// Verify Manager implements ReadWriter at compile time.
var _ ReadWriter = (*Manager)(nil)

// NewManager creates a new config manager.
// It determines the config file path following XDG Base Directory specification.
func NewManager() *Manager {
	return &Manager{
		configPath: getConfigPath(),
	}
}

// getConfigPath returns the config file path following XDG Base Directory spec.
// Priority:
// 1. $XDG_CONFIG_HOME/urgent/config.yaml
// 2. ~/.config/urgent/config.yaml
// 3. ~/.urgent/config.yaml (fallback).
func getConfigPath() string {
	// Check XDG_CONFIG_HOME first
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, "urgent", "config.yaml")
	}

	// Use ~/.config/urgent as default
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".config", "urgent", "config.yaml")
	}

	// Fallback to ~/.urgent (should rarely happen)
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".urgent", "config.yaml")
	}

	// Last resort fallback
	return "urgent-config.yaml"
}

// Load reads the configuration from disk.
// Returns an empty config if the file doesn't exist.
func (m *Manager) Load() (*Config, error) {
	// Check if file exists
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		return NewConfig(), nil
	}

	// Read file
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Initialize map if nil
	if cfg.Accounts == nil {
		cfg.Accounts = make(map[string]AccountConfig)
	}

	return &cfg, nil
}

// Save writes the configuration to disk using atomic write.
// The file is written to a temporary location first, then renamed to prevent corruption.
func (m *Manager) Save(cfg *Config) error {
	// Ensure config directory exists
	configDir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to temporary file first (atomic write pattern)
	tempPath := m.configPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary config file: %w", err)
	}

	// Rename to actual path (atomic on POSIX systems)
	if err := os.Rename(tempPath, m.configPath); err != nil {
		if removeErr := os.Remove(tempPath); removeErr != nil {
			// Failed to clean up temp file - no action needed
			_ = removeErr
		}

		return fmt.Errorf("failed to save config file: %w", err)
	}

	return nil
}

// GetEnabledCalendarIDs returns the enabled calendar IDs for a given account.
// Returns an empty slice if the account is not found.
// The name field is ignored - only IDs are returned.
func (m *Manager) GetEnabledCalendarIDs(email string) ([]string, error) {
	cfg, err := m.Load()
	if err != nil {
		return nil, err
	}

	accountCfg, ok := cfg.Accounts[email]
	if !ok {
		return []string{}, nil
	}

	ids := make([]string, 0, len(accountCfg.EnabledCalendars))
	for _, cal := range accountCfg.EnabledCalendars {
		ids = append(ids, cal.ID)
	}

	return ids, nil
}

// SetEnabledCalendars updates the enabled calendars for a given account.
func (m *Manager) SetEnabledCalendars(email string, selections []CalendarSelection) error {
	cfg, err := m.Load()
	if err != nil {
		return err
	}

	// Update or create account config
	cfg.Accounts[email] = AccountConfig{
		EnabledCalendars: selections,
	}

	return m.Save(cfg)
}

// DeleteAccount removes an account's configuration.
func (m *Manager) DeleteAccount(email string) error {
	cfg, err := m.Load()
	if err != nil {
		return err
	}

	delete(cfg.Accounts, email)

	return m.Save(cfg)
}

// GetConfigPath returns the path to the config file.
func (m *Manager) GetConfigPath() string {
	return m.configPath
}
