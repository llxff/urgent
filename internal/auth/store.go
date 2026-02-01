package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/zalando/go-keyring"
	"golang.org/x/oauth2"
)

const (
	// ServiceName is the Keychain service name for user tokens.
	ServiceName = "com.urgent.cli"
	// OAuthServiceName is the Keychain service name for OAuth credentials.
	OAuthServiceName = "com.urgent.cli.oauth"
)

var (
	// ErrNotFound is returned when a token is not found in the store.
	ErrNotFound = errors.New("token not found")
	// ErrInvalidToken is returned when a token cannot be unmarshaled.
	ErrInvalidToken = errors.New("invalid token data")
)

// Store manages OAuth2 token persistence.
//
// All methods are safe for concurrent use.
type Store interface {
	// SaveToken stores an OAuth2 token for the given email address.
	SaveToken(email string, token *oauth2.Token) error

	// GetToken retrieves a stored token for the given email address.
	// Returns ErrNotFound if no token exists for the email.
	GetToken(email string) (*oauth2.Token, error)

	// DeleteToken removes a stored token for the given email address.
	DeleteToken(email string) error

	// ListAccounts returns all email addresses with stored tokens.
	ListAccounts() ([]string, error)

	// GetOAuthCredentials retrieves OAuth client ID and secret.
	// Returns an error if credentials are not configured.
	GetOAuthCredentials() (clientID, clientSecret string, err error)

	// SaveOAuthCredentials stores OAuth client ID and secret.
	SaveOAuthCredentials(clientID, clientSecret string) error
}

// KeychainStore implements Store using macOS Keychain.
type KeychainStore struct{}

// NewKeychainStore creates a new Keychain-based credential store.
func NewKeychainStore() *KeychainStore {
	return &KeychainStore{}
}

// Verify KeychainStore implements Store at compile time.
var _ Store = (*KeychainStore)(nil)

// SaveToken stores an OAuth2 token in the Keychain.
func (s *KeychainStore) SaveToken(email string, token *oauth2.Token) error {
	if email == "" {
		return errors.New("email cannot be empty")
	}

	if token == nil {
		return errors.New("token cannot be nil")
	}

	// Serialize token to JSON
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	// Store in Keychain
	err = keyring.Set(ServiceName, email, string(data))
	if err != nil {
		return fmt.Errorf("failed to save token to keychain: %w", err)
	}

	// Add to accounts list
	if err := s.addAccountToList(email); err != nil {
		return fmt.Errorf("failed to update accounts list: %w", err)
	}

	return nil
}

// GetToken retrieves a token from the Keychain.
func (s *KeychainStore) GetToken(email string) (*oauth2.Token, error) {
	if email == "" {
		return nil, errors.New("email cannot be empty")
	}

	// Retrieve from Keychain
	data, err := keyring.Get(ServiceName, email)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("failed to get token from keychain: %w", err)
	}

	// Deserialize token
	var token oauth2.Token

	err = json.Unmarshal([]byte(data), &token)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}

	return &token, nil
}

// DeleteToken removes a token from the Keychain.
func (s *KeychainStore) DeleteToken(email string) error {
	if email == "" {
		return errors.New("email cannot be empty")
	}

	err := keyring.Delete(ServiceName, email)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return ErrNotFound
		}

		return fmt.Errorf("failed to delete token from keychain: %w", err)
	}

	// Remove from accounts list
	if err := s.removeAccountFromList(email); err != nil {
		return fmt.Errorf("failed to update accounts list: %w", err)
	}

	return nil
}

// ListAccounts returns all accounts with stored tokens.
//
// Note: go-keyring doesn't provide a list operation, so we maintain
// a list of accounts in a separate Keychain entry.
func (s *KeychainStore) ListAccounts() ([]string, error) {
	// Get the accounts list from Keychain
	data, err := keyring.Get(ServiceName, "accounts-list")
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			// No accounts yet
			return []string{}, nil
		}

		return nil, fmt.Errorf("failed to get accounts list: %w", err)
	}

	// Deserialize account list
	var accounts []string

	err = json.Unmarshal([]byte(data), &accounts)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal accounts list: %w", err)
	}

	return accounts, nil
}

// addAccountToList adds an email to the accounts list.
func (s *KeychainStore) addAccountToList(email string) error {
	accounts, err := s.ListAccounts()
	if err != nil {
		return err
	}

	// Check if already in list
	if slices.Contains(accounts, email) {
		return nil // Already exists
	}

	// Add to list
	accounts = append(accounts, email)

	// Serialize and save
	data, err := json.Marshal(accounts)
	if err != nil {
		return fmt.Errorf("failed to marshal accounts list: %w", err)
	}

	err = keyring.Set(ServiceName, "accounts-list", string(data))
	if err != nil {
		return fmt.Errorf("failed to save accounts list: %w", err)
	}

	return nil
}

// removeAccountFromList removes an email from the accounts list.
func (s *KeychainStore) removeAccountFromList(email string) error {
	accounts, err := s.ListAccounts()
	if err != nil {
		return err
	}

	// Filter out the email
	filtered := make([]string, 0, len(accounts))

	for _, acc := range accounts {
		if acc != email {
			filtered = append(filtered, acc)
		}
	}

	// Serialize and save
	data, err := json.Marshal(filtered)
	if err != nil {
		return fmt.Errorf("failed to marshal accounts list: %w", err)
	}

	err = keyring.Set(ServiceName, "accounts-list", string(data))
	if err != nil {
		return fmt.Errorf("failed to save accounts list: %w", err)
	}

	return nil
}

// GetOAuthCredentials retrieves OAuth client ID and secret from Keychain.
func (s *KeychainStore) GetOAuthCredentials() (clientID, clientSecret string, err error) {
	return GetOAuthCredentials()
}

// SaveOAuthCredentials stores OAuth client ID and secret in Keychain.
func (s *KeychainStore) SaveOAuthCredentials(clientID, clientSecret string) error {
	return SaveOAuthCredentials(clientID, clientSecret)
}

// SaveOAuthCredentials stores OAuth client ID and secret in Keychain.
// Deprecated: Use KeychainStore.SaveOAuthCredentials instead.
func SaveOAuthCredentials(clientID, clientSecret string) error {
	if clientID == "" || clientSecret == "" {
		return errors.New("client ID and secret cannot be empty")
	}

	// Store client ID
	err := keyring.Set(OAuthServiceName, "client-id", clientID)
	if err != nil {
		return fmt.Errorf("failed to save client ID: %w", err)
	}

	// Store client secret
	err = keyring.Set(OAuthServiceName, "client-secret", clientSecret)
	if err != nil {
		return fmt.Errorf("failed to save client secret: %w", err)
	}

	return nil
}

// GetOAuthCredentials retrieves OAuth client ID and secret from Keychain.
func GetOAuthCredentials() (clientID, clientSecret string, err error) {
	// Get client ID
	clientID, err = keyring.Get(OAuthServiceName, "client-id")
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", "", errors.New("OAuth credentials not found. Run 'urgent setup' first")
		}

		return "", "", fmt.Errorf("failed to get client ID: %w", err)
	}

	// Get client secret
	clientSecret, err = keyring.Get(OAuthServiceName, "client-secret")
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", "", errors.New("OAuth credentials not found. Run 'urgent setup' first")
		}

		return "", "", fmt.Errorf("failed to get client secret: %w", err)
	}

	return clientID, clientSecret, nil
}
