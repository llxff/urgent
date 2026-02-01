package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCommand(t *testing.T) {
	// Test that root command is properly configured
	if rootCmd.Use != "urgent" {
		t.Errorf("Expected command name 'urgent', got '%s'", rootCmd.Use)
	}

	if rootCmd.Short == "" {
		t.Error("Root command should have a short description")
	}

	if rootCmd.Long == "" {
		t.Error("Root command should have a long description")
	}
}

func TestRootCommandFlags(t *testing.T) {
	// Test that global flags are registered
	flag := rootCmd.PersistentFlags().Lookup("output")
	if flag == nil {
		t.Fatal("Expected --output flag to be registered")
	}

	if flag.Shorthand != "o" {
		t.Errorf("Expected shorthand 'o', got '%s'", flag.Shorthand)
	}

	if flag.DefValue != "table" {
		t.Errorf("Expected default value 'table', got '%s'", flag.DefValue)
	}
}

func TestGetOutputFormat(t *testing.T) {
	tests := []struct {
		name     string
		setValue string
		want     string
	}{
		{
			name:     "default value",
			setValue: "table",
			want:     "table",
		},
		{
			name:     "json format",
			setValue: "json",
			want:     "json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the value
			outputFormat = tt.setValue

			// Get the value
			got := GetOutputFormat()
			if got != tt.want {
				t.Errorf("GetOutputFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSubcommandsRegistered(t *testing.T) {
	// Test that all expected subcommands are registered
	expectedCommands := []string{
		"setup",
		"calendars",
		"today",
		"next",
	}

	commands := rootCmd.Commands()

	commandNames := make(map[string]bool)

	for _, cmd := range commands {
		commandNames[cmd.Name()] = true
	}

	for _, expected := range expectedCommands {
		if !commandNames[expected] {
			t.Errorf("Expected command '%s' to be registered", expected)
		}
	}

	// Note: completion and help commands are added by Cobra and may appear in output
	// but not always in Commands() during testing
	t.Logf("Found %d commands: %v", len(commands), func() []string {
		names := make([]string, 0, len(commands))
		for _, cmd := range commands {
			names = append(names, cmd.Name())
		}

		return names
	}())
}

func TestCommandDescriptions(t *testing.T) {
	tests := []struct {
		name        string
		commandName string
	}{
		{"setup", "setup"},
		{"calendars", "calendars"},
		{"today", "today"},
		{"next", "next"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, cmd := range rootCmd.Commands() {
				if cmd.Name() == tt.commandName {
					if cmd.Short == "" {
						t.Errorf("Command '%s' should have a short description", tt.commandName)
					}

					if cmd.Long == "" {
						t.Errorf("Command '%s' should have a long description", tt.commandName)
					}

					return
				}
			}

			t.Errorf("Command '%s' not found", tt.commandName)
		})
	}
}

func TestExecute(t *testing.T) {
	// Test that Execute function exists and returns without error for help
	// We can't fully test Execute without mocking, but we can verify it exists
	// and has the correct signature by calling it with --help
	rootCmd.SetArgs([]string{"--help"})

	// This should not panic
	err := Execute()
	if err != nil {
		t.Errorf("Execute() with --help should not return error, got %v", err)
	}
}

func TestTodayCommandFlags(t *testing.T) {
	var cmd *cobra.Command

	for _, c := range rootCmd.Commands() {
		if c.Name() == "today" {
			cmd = c

			break
		}
	}

	if cmd == nil {
		t.Fatal("today command not found")
	}

	// Test --remaining flag
	flag := cmd.Flags().Lookup("remaining")
	if flag == nil {
		t.Error("Expected --remaining flag on today command")

		return
	}

	if flag.Shorthand != "r" {
		t.Errorf("Expected shorthand 'r', got '%s'", flag.Shorthand)
	}
}

func TestNextCommandFlags(t *testing.T) {
	var cmd *cobra.Command

	for _, c := range rootCmd.Commands() {
		if c.Name() == "next" {
			cmd = c

			break
		}
	}

	if cmd == nil {
		t.Fatal("next command not found")
	}

	// Test --within flag
	flag := cmd.Flags().Lookup("within")
	if flag == nil {
		t.Error("Expected --within flag on next command")

		return
	}

	if flag.Shorthand != "w" {
		t.Errorf("Expected shorthand 'w', got '%s'", flag.Shorthand)
	}

	if flag.DefValue != "60" {
		t.Errorf("Expected default value '60', got '%s'", flag.DefValue)
	}
}

func TestCalendarsCommandFlags(t *testing.T) {
	var cmd *cobra.Command

	for _, c := range rootCmd.Commands() {
		if c.Name() == "calendars" {
			cmd = c

			break
		}
	}

	if cmd == nil {
		t.Fatal("calendars command not found")
	}

	// Test --account flag
	flag := cmd.Flags().Lookup("account")
	if flag == nil {
		t.Error("Expected --account flag on calendars command")

		return
	}

	if flag.Shorthand != "a" {
		t.Errorf("Expected shorthand 'a', got '%s'", flag.Shorthand)
	}

	// Test --list flag
	listFlag := cmd.Flags().Lookup("list")
	if listFlag == nil {
		t.Error("Expected --list flag on calendars command")

		return
	}

	if listFlag.Shorthand != "l" {
		t.Errorf("Expected shorthand 'l', got '%s'", listFlag.Shorthand)
	}
}

func TestCommandUsage(t *testing.T) {
	// Test that all commands have proper usage strings
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "help" || cmd.Name() == "completion" {
			continue // Skip built-in commands
		}

		t.Run(cmd.Name(), func(t *testing.T) {
			if cmd.Use == "" {
				t.Errorf("Command '%s' should have Use field set", cmd.Name())
			}

			if cmd.Short == "" {
				t.Errorf("Command '%s' should have Short description", cmd.Name())
			}

			if cmd.Long == "" {
				t.Errorf("Command '%s' should have Long description", cmd.Name())
			}
		})
	}
}

func TestSilenceSettings(t *testing.T) {
	// Test that SilenceUsage and SilenceErrors are set appropriately
	if !rootCmd.SilenceUsage {
		t.Error("Root command should have SilenceUsage = true")
	}

	if !rootCmd.SilenceErrors {
		t.Error("Root command should have SilenceErrors = true")
	}
}

func TestCommandExamples(t *testing.T) {
	// Test that data commands have examples
	commandsWithExamples := []string{"today", "next"}

	for _, cmdName := range commandsWithExamples {
		t.Run(cmdName, func(t *testing.T) {
			for _, cmd := range rootCmd.Commands() {
				if cmd.Name() == cmdName {
					// The examples are in Long description, just check it's not empty
					if cmd.Long == "" {
						t.Errorf("Command '%s' should have examples in Long description", cmdName)
					}

					return
				}
			}

			t.Errorf("Command '%s' not found", cmdName)
		})
	}
}
