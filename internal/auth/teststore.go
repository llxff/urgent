package auth

import (
	"sync"

	"golang.org/x/oauth2"
)

// TestStore is an in-memory implementation of Store for testing.
//
// All methods are safe for concurrent use.
type TestStore struct {
	tokens       map[string]*oauth2.Token
	clientID     string
	clientSecret string
	mu           sync.RWMutex
}

// Verify TestStore implements Store at compile time.
var _ Store = (*TestStore)(nil)

// NewTestStore creates a new in-memory credential store for testing.
func NewTestStore() *TestStore {
	return &TestStore{
		tokens: make(map[string]*oauth2.Token),
	}
}

// SaveToken stores a token in memory.
func (s *TestStore) SaveToken(email string, token *oauth2.Token) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if email == "" {
		return ErrNotFound
	}

	if token == nil {
		return ErrInvalidToken
	}

	s.tokens[email] = token

	return nil
}

// GetToken retrieves a token from memory.
func (s *TestStore) GetToken(email string) (*oauth2.Token, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	token, ok := s.tokens[email]
	if !ok {
		return nil, ErrNotFound
	}

	return token, nil
}

// DeleteToken removes a token from memory.
func (s *TestStore) DeleteToken(email string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tokens[email]; !ok {
		return ErrNotFound
	}

	delete(s.tokens, email)

	return nil
}

// ListAccounts returns all email addresses with stored tokens.
func (s *TestStore) ListAccounts() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	accounts := make([]string, 0, len(s.tokens))
	for email := range s.tokens {
		accounts = append(accounts, email)
	}

	return accounts, nil
}

// GetOAuthCredentials retrieves OAuth credentials from memory.
func (s *TestStore) GetOAuthCredentials() (clientID, clientSecret string, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.clientID == "" || s.clientSecret == "" {
		return "", "", ErrNotFound
	}

	return s.clientID, s.clientSecret, nil
}

// SaveOAuthCredentials stores OAuth credentials in memory.
func (s *TestStore) SaveOAuthCredentials(clientID, clientSecret string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.clientID = clientID
	s.clientSecret = clientSecret

	return nil
}
