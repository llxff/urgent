package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestCalendarSelection_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name      string
		yaml      string
		want      CalendarSelection
		wantError bool
	}{
		{
			name: "simple string format",
			yaml: "primary",
			want: CalendarSelection{ID: "primary", Name: ""},
		},
		{
			name: "object format with name",
			yaml: "id: primary\nname: My Calendar",
			want: CalendarSelection{ID: "primary", Name: "My Calendar"},
		},
		{
			name: "object format without name",
			yaml: "id: work@example.com",
			want: CalendarSelection{ID: "work@example.com", Name: ""},
		},
		{
			name:      "object format missing id",
			yaml:      "name: Test",
			wantError: true,
		},
		{
			name:      "invalid format",
			yaml:      "- item1\n- item2",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got CalendarSelection

			err := yaml.Unmarshal([]byte(tt.yaml), &got)

			if tt.wantError {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.ID, got.ID)
			assert.Equal(t, tt.want.Name, got.Name)
		})
	}
}

func TestCalendarSelection_MarshalYAML(t *testing.T) {
	tests := []struct {
		name string
		cs   CalendarSelection
		want string
	}{
		{
			name: "with name - marshal as object",
			cs:   CalendarSelection{ID: "primary", Name: "My Calendar"},
			want: "id: primary\nname: My Calendar\n",
		},
		{
			name: "without name - marshal as string",
			cs:   CalendarSelection{ID: "primary", Name: ""},
			want: "primary\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := yaml.Marshal(tt.cs)
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(data))
		})
	}
}

func TestConfig_MarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name string
		yaml string
		want Config
	}{
		{
			name: "mixed format - strings and objects",
			yaml: `accounts:
  user@example.com:
    enabled_calendars:
      - primary
      - id: work@example.com
        name: Work Calendar
`,
			want: Config{
				Accounts: map[string]AccountConfig{
					"user@example.com": {
						EnabledCalendars: []CalendarSelection{
							{ID: "primary", Name: ""},
							{ID: "work@example.com", Name: "Work Calendar"},
						},
					},
				},
			},
		},
		{
			name: "all strings format",
			yaml: `accounts:
  user@example.com:
    enabled_calendars:
      - primary
      - work@example.com
`,
			want: Config{
				Accounts: map[string]AccountConfig{
					"user@example.com": {
						EnabledCalendars: []CalendarSelection{
							{ID: "primary", Name: ""},
							{ID: "work@example.com", Name: ""},
						},
					},
				},
			},
		},
		{
			name: "all objects format",
			yaml: `accounts:
  user@example.com:
    enabled_calendars:
      - id: primary
        name: user@example.com
      - id: work@example.com
        name: Work Calendar
`,
			want: Config{
				Accounts: map[string]AccountConfig{
					"user@example.com": {
						EnabledCalendars: []CalendarSelection{
							{ID: "primary", Name: "user@example.com"},
							{ID: "work@example.com", Name: "Work Calendar"},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Config

			err := yaml.Unmarshal([]byte(tt.yaml), &got)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestManager_Load(t *testing.T) {
	t.Run("file doesn't exist returns empty config", func(t *testing.T) {
		tempDir, cleanup := CreateTempConfigDir(t)
		defer cleanup()

		m := NewManagerWithPath(filepath.Join(tempDir, "nonexistent.yaml"))
		cfg, err := m.Load()

		require.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.NotNil(t, cfg.Accounts)
		assert.Empty(t, cfg.Accounts)
	})

	t.Run("loads valid config", func(t *testing.T) {
		tempDir, cleanup := CreateTempConfigDir(t)
		defer cleanup()

		expected := NewTestConfig()
		configPath := WriteTestConfig(t, tempDir, expected)

		m := NewManagerWithPath(configPath)
		cfg, err := m.Load()

		require.NoError(t, err)
		assert.Equal(t, expected, cfg)
	})

	t.Run("corrupted YAML returns error", func(t *testing.T) {
		tempDir, cleanup := CreateTempConfigDir(t)
		defer cleanup()

		configPath := filepath.Join(tempDir, "config.yaml")
		err := os.WriteFile(configPath, []byte("invalid: yaml: content: :::"), 0644)
		require.NoError(t, err)

		m := NewManagerWithPath(configPath)
		_, err = m.Load()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse config file")
	})

	t.Run("empty file returns empty config", func(t *testing.T) {
		tempDir, cleanup := CreateTempConfigDir(t)
		defer cleanup()

		configPath := filepath.Join(tempDir, "config.yaml")
		err := os.WriteFile(configPath, []byte(""), 0644)
		require.NoError(t, err)

		m := NewManagerWithPath(configPath)
		cfg, err := m.Load()

		require.NoError(t, err)
		assert.NotNil(t, cfg)
	})
}

func TestManager_Save(t *testing.T) {
	t.Run("saves config successfully", func(t *testing.T) {
		tempDir, cleanup := CreateTempConfigDir(t)
		defer cleanup()

		cfg := NewTestConfig()
		configPath := filepath.Join(tempDir, "config.yaml")

		m := NewManagerWithPath(configPath)
		err := m.Save(cfg)

		require.NoError(t, err)

		// Verify file exists
		_, err = os.Stat(configPath)
		assert.NoError(t, err)

		// Load and verify content
		loaded, err := m.Load()
		require.NoError(t, err)
		assert.Equal(t, cfg, loaded)
	})

	t.Run("creates config directory if missing", func(t *testing.T) {
		tempDir, cleanup := CreateTempConfigDir(t)
		defer cleanup()

		configPath := filepath.Join(tempDir, "nested", "dir", "config.yaml")

		m := NewManagerWithPath(configPath)
		err := m.Save(NewTestConfig())

		require.NoError(t, err)
		_, err = os.Stat(configPath)
		assert.NoError(t, err)
	})

	t.Run("atomic write - temp file cleaned on error", func(t *testing.T) {
		tempDir, cleanup := CreateTempConfigDir(t)
		defer cleanup()

		// Create a directory where the config file should be
		// This will cause the rename to fail
		configPath := filepath.Join(tempDir, "config.yaml")
		err := os.MkdirAll(configPath, 0755)
		require.NoError(t, err)

		m := NewManagerWithPath(configPath)
		err = m.Save(NewTestConfig())

		assert.Error(t, err)

		// Verify temp file was cleaned up
		tempPath := configPath + ".tmp"
		_, err = os.Stat(tempPath)
		assert.True(t, os.IsNotExist(err), "temp file should be cleaned up")
	})
}

func TestManager_GetEnabledCalendarIDs(t *testing.T) {
	t.Run("returns calendar IDs for existing account", func(t *testing.T) {
		tempDir, cleanup := CreateTempConfigDir(t)
		defer cleanup()

		cfg := NewTestConfig()
		configPath := WriteTestConfig(t, tempDir, cfg)

		m := NewManagerWithPath(configPath)
		ids, err := m.GetEnabledCalendarIDs("test@example.com")

		require.NoError(t, err)
		assert.Equal(t, []string{"primary", "work@example.com"}, ids)
	})

	t.Run("returns empty slice for non-existent account", func(t *testing.T) {
		tempDir, cleanup := CreateTempConfigDir(t)
		defer cleanup()

		cfg := NewTestConfig()
		configPath := WriteTestConfig(t, tempDir, cfg)

		m := NewManagerWithPath(configPath)
		ids, err := m.GetEnabledCalendarIDs("nonexistent@example.com")

		require.NoError(t, err)
		assert.Empty(t, ids)
	})

	t.Run("ignores name field - only returns IDs", func(t *testing.T) {
		tempDir, cleanup := CreateTempConfigDir(t)
		defer cleanup()

		cfg := &Config{
			Accounts: map[string]AccountConfig{
				"test@example.com": {
					EnabledCalendars: []CalendarSelection{
						{ID: "cal1", Name: "Calendar 1"},
						{ID: "cal2", Name: "Calendar 2"},
						{ID: "cal3", Name: ""},
					},
				},
			},
		}
		configPath := WriteTestConfig(t, tempDir, cfg)

		m := NewManagerWithPath(configPath)
		ids, err := m.GetEnabledCalendarIDs("test@example.com")

		require.NoError(t, err)
		assert.Equal(t, []string{"cal1", "cal2", "cal3"}, ids)
	})
}

func TestManager_SetEnabledCalendars(t *testing.T) {
	t.Run("creates new account config", func(t *testing.T) {
		tempDir, cleanup := CreateTempConfigDir(t)
		defer cleanup()

		configPath := filepath.Join(tempDir, "config.yaml")
		m := NewManagerWithPath(configPath)

		selections := []CalendarSelection{
			{ID: "primary", Name: "Primary"},
			{ID: "work", Name: "Work"},
		}

		err := m.SetEnabledCalendars("new@example.com", selections)
		require.NoError(t, err)

		// Verify it was saved
		ids, err := m.GetEnabledCalendarIDs("new@example.com")
		require.NoError(t, err)
		assert.Equal(t, []string{"primary", "work"}, ids)
	})

	t.Run("updates existing account config", func(t *testing.T) {
		tempDir, cleanup := CreateTempConfigDir(t)
		defer cleanup()

		cfg := NewTestConfig()
		configPath := WriteTestConfig(t, tempDir, cfg)

		m := NewManagerWithPath(configPath)

		// Update with new selections
		newSelections := []CalendarSelection{
			{ID: "calendar1", Name: "New Calendar"},
		}

		err := m.SetEnabledCalendars("test@example.com", newSelections)
		require.NoError(t, err)

		// Verify update
		ids, err := m.GetEnabledCalendarIDs("test@example.com")
		require.NoError(t, err)
		assert.Equal(t, []string{"calendar1"}, ids)
	})
}

func TestManager_DeleteAccount(t *testing.T) {
	t.Run("deletes existing account", func(t *testing.T) {
		tempDir, cleanup := CreateTempConfigDir(t)
		defer cleanup()

		cfg := NewTestConfig()
		configPath := WriteTestConfig(t, tempDir, cfg)

		m := NewManagerWithPath(configPath)

		err := m.DeleteAccount("test@example.com")
		require.NoError(t, err)

		// Verify it's deleted
		ids, err := m.GetEnabledCalendarIDs("test@example.com")
		require.NoError(t, err)
		assert.Empty(t, ids)

		// Verify other account still exists
		ids, err = m.GetEnabledCalendarIDs("other@example.com")
		require.NoError(t, err)
		assert.NotEmpty(t, ids)
	})

	t.Run("deleting non-existent account doesn't error", func(t *testing.T) {
		tempDir, cleanup := CreateTempConfigDir(t)
		defer cleanup()

		configPath := filepath.Join(tempDir, "config.yaml")
		m := NewManagerWithPath(configPath)

		err := m.DeleteAccount("nonexistent@example.com")
		assert.NoError(t, err)
	})
}

func TestManager_GetConfigPath(t *testing.T) {
	m := NewManagerWithPath("/custom/path/config.yaml")
	assert.Equal(t, "/custom/path/config.yaml", m.GetConfigPath())
}

func TestGetConfigPath(t *testing.T) {
	t.Run("uses XDG_CONFIG_HOME when set", func(t *testing.T) {
		oldXDG := os.Getenv("XDG_CONFIG_HOME")
		defer func() {
			if err := os.Setenv("XDG_CONFIG_HOME", oldXDG); err != nil {
				t.Logf("failed to restore XDG_CONFIG_HOME: %v", err)
			}
		}()

		if err := os.Setenv("XDG_CONFIG_HOME", "/custom/xdg"); err != nil {
			t.Fatalf("failed to set XDG_CONFIG_HOME: %v", err)
		}

		path := getConfigPath()

		assert.Equal(t, "/custom/xdg/urgent/config.yaml", path)
	})

	t.Run("uses ~/.config when XDG_CONFIG_HOME not set", func(t *testing.T) {
		oldXDG := os.Getenv("XDG_CONFIG_HOME")
		defer func() {
			if err := os.Setenv("XDG_CONFIG_HOME", oldXDG); err != nil {
				t.Logf("failed to restore XDG_CONFIG_HOME: %v", err)
			}
		}()

		if err := os.Setenv("XDG_CONFIG_HOME", ""); err != nil {
			t.Fatalf("failed to unset XDG_CONFIG_HOME: %v", err)
		}

		path := getConfigPath()

		home, _ := os.UserHomeDir()
		expected := filepath.Join(home, ".config", "urgent", "config.yaml")
		assert.Equal(t, expected, path)
	})
}

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()

	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.Accounts)
	assert.Empty(t, cfg.Accounts)
}
