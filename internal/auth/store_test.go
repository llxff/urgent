package auth

import (
	"errors"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestTestStore_SaveToken(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		token   *oauth2.Token
		wantErr bool
	}{
		{
			name:    "valid token",
			email:   "user@example.com",
			token:   &oauth2.Token{AccessToken: "abc123"},
			wantErr: false,
		},
		{
			name:    "empty email",
			email:   "",
			token:   &oauth2.Token{AccessToken: "abc123"},
			wantErr: true,
		},
		{
			name:    "nil token",
			email:   "user@example.com",
			token:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewTestStore()

			err := store.SaveToken(tt.email, tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTestStore_GetToken(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*TestStore)
		email     string
		wantToken bool
		wantErr   bool
	}{
		{
			name: "existing token",
			setup: func(s *TestStore) {
				_ = s.SaveToken("user@example.com", &oauth2.Token{AccessToken: "abc123"})
			},
			email:     "user@example.com",
			wantToken: true,
			wantErr:   false,
		},
		{
			name:      "non-existent token",
			setup:     func(_ *TestStore) {},
			email:     "nonexistent@example.com",
			wantToken: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewTestStore()
			tt.setup(store)

			token, err := store.GetToken(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetToken() error = %v, wantErr %v", err, tt.wantErr)
			}

			if (token != nil) != tt.wantToken {
				t.Errorf("GetToken() token = %v, wantToken %v", token != nil, tt.wantToken)
			}
		})
	}
}

func TestTestStore_DeleteToken(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*TestStore)
		email   string
		wantErr bool
	}{
		{
			name: "existing token",
			setup: func(s *TestStore) {
				_ = s.SaveToken("user@example.com", &oauth2.Token{AccessToken: "abc123"})
			},
			email:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "non-existent token",
			setup:   func(_ *TestStore) {},
			email:   "nonexistent@example.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewTestStore()
			tt.setup(store)

			err := store.DeleteToken(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteToken() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify token is actually deleted
			if !tt.wantErr {
				_, err := store.GetToken(tt.email)
				if !errors.Is(err, ErrNotFound) {
					t.Errorf("Expected token to be deleted")
				}
			}
		})
	}
}

func TestTestStore_ListAccounts(t *testing.T) {
	store := NewTestStore()

	// Initially empty
	accounts, err := store.ListAccounts()
	if err != nil {
		t.Fatalf("ListAccounts() error = %v", err)
	}

	if len(accounts) != 0 {
		t.Errorf("Expected 0 accounts, got %d", len(accounts))
	}

	// Add some accounts
	_ = store.SaveToken("user1@example.com", &oauth2.Token{AccessToken: "token1"})
	_ = store.SaveToken("user2@example.com", &oauth2.Token{AccessToken: "token2"})

	accounts, err = store.ListAccounts()
	if err != nil {
		t.Fatalf("ListAccounts() error = %v", err)
	}

	if len(accounts) != 2 {
		t.Errorf("Expected 2 accounts, got %d", len(accounts))
	}

	// Delete one
	_ = store.DeleteToken("user1@example.com")

	accounts, err = store.ListAccounts()
	if err != nil {
		t.Fatalf("ListAccounts() error = %v", err)
	}

	if len(accounts) != 1 {
		t.Errorf("Expected 1 account, got %d", len(accounts))
	}
}

func TestManager_RefreshToken(t *testing.T) {
	store := NewTestStore()
	manager := NewManager(store, "client-id", "client-secret")

	// This test verifies the manager structure, but actual refresh requires HTTP calls
	// which are better tested with httptest in integration tests
	if manager.store == nil {
		t.Error("Manager store should not be nil")
	}

	if manager.config == nil {
		t.Error("Manager config should not be nil")
	}
}

func TestTokenExpiry(t *testing.T) {
	store := NewTestStore()

	// Save token with expiry
	expiry := time.Now().Add(1 * time.Hour)
	token := &oauth2.Token{
		AccessToken:  "abc123",
		RefreshToken: "refresh123",
		Expiry:       expiry,
	}

	err := store.SaveToken("user@example.com", token)
	if err != nil {
		t.Fatalf("SaveToken() error = %v", err)
	}

	// Retrieve and verify
	retrieved, err := store.GetToken("user@example.com")
	if err != nil {
		t.Fatalf("GetToken() error = %v", err)
	}

	if retrieved.AccessToken != token.AccessToken {
		t.Errorf("Expected access token %s, got %s", token.AccessToken, retrieved.AccessToken)
	}

	if retrieved.RefreshToken != token.RefreshToken {
		t.Errorf("Expected refresh token %s, got %s", token.RefreshToken, retrieved.RefreshToken)
	}

	// Check if token would be considered valid (not expired)
	if retrieved.Expiry.Before(time.Now()) {
		t.Error("Token should not be expired")
	}
}
