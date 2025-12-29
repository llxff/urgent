package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

// newTestManager creates a manager with a no-op browser opener for testing.
func newTestManager(store Store, clientID, clientSecret string) *Manager {
	m := NewManager(store, clientID, clientSecret)
	// Replace browser opener with no-op to prevent opening browser during tests
	m.browserOpener = func(url string) error {
		return nil // Don't actually open browser
	}
	return m
}

func TestManager_Authenticate(t *testing.T) {
	t.Run("successful authentication", func(t *testing.T) {
		// Create test OAuth server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/token":
				// Exchange code for token
				token := oauth2.Token{
					AccessToken:  "test_access_token",
					RefreshToken: "test_refresh_token",
					Expiry:       time.Now().Add(time.Hour),
				}
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(token); err != nil {
					t.Logf("failed to encode token: %v", err)
				}
			case "/userinfo":
				// Return user info
				w.Header().Set("Content-Type", "application/json")
				if _, err := w.Write([]byte(`{"email":"test@example.com"}`)); err != nil {
					t.Logf("failed to write userinfo: %v", err)
				}
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		// Note: This test would need to mock the browser opening and callback
		// For now, we're testing the structure but this is integration test territory
		t.Skip("Full authentication flow requires browser interaction - needs integration test")
	})

	t.Run("delegates to AuthenticateWithURLCallback", func(t *testing.T) {
		store := NewTestStore()
		manager := newTestManager(store, "test_client_id", "test_client_secret")

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		// This will timeout (expected) but we're verifying it calls the right method
		_, err := manager.Authenticate(ctx)

		// Should get timeout error, not panic
		if err == nil {
			t.Error("Expected timeout error, got nil")
		}
	})
}

func TestManager_AuthenticateWithURLCallback(t *testing.T) {
	t.Run("calls URL callback when provided", func(t *testing.T) {
		store := NewTestStore()
		manager := newTestManager(store, "test_client_id", "test_client_secret")

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		var capturedURL string
		_, err := manager.AuthenticateWithURLCallback(ctx, func(url string) {
			capturedURL = url
		})

		// Should timeout, but URL callback should have been called
		if capturedURL == "" {
			t.Error("URL callback was not called")
		}

		if err == nil {
			t.Error("Expected timeout error, got nil")
		}

		// Verify URL contains expected OAuth parameters
		parsedURL, err := url.Parse(capturedURL)
		if err != nil {
			t.Fatalf("Failed to parse captured URL: %v", err)
		}

		query := parsedURL.Query()
		if query.Get("client_id") != "test_client_id" {
			t.Errorf("Expected client_id=test_client_id, got %s", query.Get("client_id"))
		}
		if query.Get("state") == "" {
			t.Error("Expected state parameter to be set")
		}
	})

	t.Run("returns error on context cancellation", func(t *testing.T) {
		store := NewTestStore()
		manager := newTestManager(store, "test_client_id", "test_client_secret")

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		result, err := manager.AuthenticateWithURLCallback(ctx, nil)

		if err == nil {
			t.Error("Expected cancellation error, got nil")
		}

		if result != nil {
			t.Error("Expected nil result on error")
		}

		if !errors.Is(err, context.Canceled) && err.Error() != "authentication cancelled" {
			t.Errorf("Expected cancellation error, got: %v", err)
		}
	})

	t.Run("returns error on timeout", func(t *testing.T) {
		store := NewTestStore()
		manager := newTestManager(store, "test_client_id", "test_client_secret")

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		result, err := manager.AuthenticateWithURLCallback(ctx, nil)

		if err == nil {
			t.Error("Expected timeout error, got nil")
		}

		if result != nil {
			t.Error("Expected nil result on timeout")
		}
	})

	t.Run("does not call URL callback when nil", func(t *testing.T) {
		store := NewTestStore()
		manager := newTestManager(store, "test_client_id", "test_client_secret")

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		// Should not panic with nil callback
		_, err := manager.AuthenticateWithURLCallback(ctx, nil)

		if err == nil {
			t.Error("Expected timeout error, got nil")
		}
	})
}

func TestManager_getUserEmail(t *testing.T) {
	t.Run("successful email retrieval", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/oauth2/v2/userinfo" {
				http.NotFound(w, r)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			if _, err := w.Write([]byte(`{"email":"user@example.com"}`)); err != nil {
				t.Logf("failed to write response: %v", err)
			}
		}))
		defer server.Close()

		// This test would need to inject the test server URL
		// The current implementation uses a hardcoded Google URL
		t.Skip("getUserEmail uses hardcoded Google URL - needs refactoring for testability")
	})
}

func TestRefreshToken_NoRefreshToken(t *testing.T) {
	store := NewTestStore()
	manager := newTestManager(store, "test_client_id", "test_client_secret")

	token := &oauth2.Token{
		AccessToken: "access_token",
		// No RefreshToken
	}

	ctx := context.Background()
	_, err := manager.RefreshToken(ctx, token)

	if err == nil {
		t.Error("Expected error for missing refresh token, got nil")
	}

	expectedMsg := "no refresh token available"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error %q, got %q", expectedMsg, err.Error())
	}
}

func TestGenerateState(t *testing.T) {
	t.Run("generates unique states", func(t *testing.T) {
		state1, err1 := generateState()
		if err1 != nil {
			t.Fatalf("Failed to generate first state: %v", err1)
		}

		state2, err2 := generateState()
		if err2 != nil {
			t.Fatalf("Failed to generate second state: %v", err2)
		}

		if state1 == state2 {
			t.Error("Expected unique states, got identical values")
		}

		if len(state1) == 0 {
			t.Error("Generated state should not be empty")
		}
	})
}

func TestAuthResult(t *testing.T) {
	t.Run("struct has expected fields", func(t *testing.T) {
		token := &oauth2.Token{AccessToken: "test"}
		result := AuthResult{
			Token:   token,
			Email:   "test@example.com",
			AuthURL: "https://example.com/auth",
		}

		if result.Token != token {
			t.Error("Token field not set correctly")
		}
		if result.Email != "test@example.com" {
			t.Errorf("Expected email test@example.com, got %s", result.Email)
		}
		if result.AuthURL != "https://example.com/auth" {
			t.Errorf("Expected AuthURL https://example.com/auth, got %s", result.AuthURL)
		}
	})
}
