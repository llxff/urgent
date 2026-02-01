package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
	"urgent/internal/auth"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Store OAuth2 credentials in Keychain",
	Long: `Store OAuth2 client ID and secret securely in macOS Keychain.

This is a one-time setup. You need to:
1. Create a Google Cloud project
2. Enable Google Calendar API
3. Create OAuth2 Desktop credentials
4. Run this command to store the credentials

After setup, run 'urgent calendars' to add a Google account.`,
	RunE: runSetup,
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func runSetup(_ *cobra.Command, _ []string) error {
	return runSetupWithReader(os.Stdin)
}

func runSetupWithReader(reader *os.File) error {
	bufReader := bufio.NewReader(reader)

	fmt.Println("OAuth Credentials Setup")
	fmt.Println()
	fmt.Println("Steps to get credentials:")
	fmt.Println("  1. Go to console.cloud.google.com")
	fmt.Println("  2. Create or select a project")
	fmt.Println("  3. Enable Google Calendar API")
	fmt.Println("  4. Create OAuth2 Desktop credentials")
	fmt.Println("  5. Configure redirect URI: http://localhost")
	fmt.Println()

	// Read Client ID
	fmt.Print("Client ID: ")
	clientID, err := bufReader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read client ID: %w", err)
	}
	clientID = strings.TrimSpace(clientID)

	if clientID == "" {
		return errors.New("client ID is required")
	}

	// Read Client Secret (with masking if possible)
	fmt.Print("Secret: ")
	var secret string
	fd := int(reader.Fd())
	if term.IsTerminal(fd) {
		// Terminal input - mask the secret
		secretBytes, err := term.ReadPassword(fd)
		if err != nil {
			return fmt.Errorf("failed to read secret: %w", err)
		}
		secret = string(secretBytes)
		fmt.Println() // Print newline after hidden input
	} else {
		// Non-terminal input (e.g., pipe) - read normally
		secret, err = bufReader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read secret: %w", err)
		}
	}
	secret = strings.TrimSpace(secret)

	if secret == "" {
		return errors.New("secret is required")
	}

	// Save to keychain
	err = auth.SaveOAuthCredentials(clientID, secret)
	if err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	fmt.Println()
	fmt.Println("✓ Saved to Keychain")
	fmt.Println()
	fmt.Println("Next step: urgent calendars")

	return nil
}
