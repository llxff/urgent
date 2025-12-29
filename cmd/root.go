package cmd

import (
	"github.com/spf13/cobra"
)

var (
	// outputFormat is the global output format flag.
	outputFormat string
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "urgent",
	Short: "A beautiful TUI-based Google Calendar CLI",
	Long: `urgent is a command-line interface for Google Calendar that supports multiple accounts and automation-friendly JSON output.

Features:
  - Multiple Google account support
  - Secure credential storage in macOS Keychain
  - JSON output for automation and scripting
  - OAuth2 authentication with automatic token refresh`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags available to all commands
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format: table, json (only for data commands)")
}

// GetOutputFormat returns the current output format.
func GetOutputFormat() string {
	return outputFormat
}
