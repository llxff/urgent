package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

const (
	// authTimeout is the maximum time to wait for OAuth callback.
	authTimeout = 5 * time.Minute
)

// Manager handles OAuth2 authentication flow.
type Manager struct {
	store         Store
	config        *oauth2.Config
	browserOpener func(string) error // For testing: allows mocking browser opening
}

// NewManager creates a new OAuth2 manager.
func NewManager(store Store, clientID, clientSecret string) *Manager {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  "", // Will be set dynamically
		Scopes: []string{
			calendar.CalendarReadonlyScope,             // View events on calendars
			calendar.CalendarCalendarlistReadonlyScope, // List calendars
			"https://www.googleapis.com/auth/userinfo.email",
		},
	}

	return &Manager{
		store:         store,
		config:        config,
		browserOpener: openBrowser, // Use real browser opener by default
	}
}

// AuthResult contains the result of a successful OAuth2 authentication.
type AuthResult struct {
	Token   *oauth2.Token // The OAuth2 access token
	Email   string        // The authenticated user's email address
	AuthURL string        // The authorization URL that was used
}

// Authenticate performs the OAuth2 authentication flow.
//
// It starts a local HTTP server on a random port, opens the browser for user
// consent, and waits for the callback with the authorization code.
//
// Returns the authentication result or an error.
func (m *Manager) Authenticate(ctx context.Context) (*AuthResult, error) {
	return m.AuthenticateWithURLCallback(ctx, nil)
}

// AuthenticateWithURLCallback performs OAuth2 authentication with an optional callback
// that receives the auth URL as soon as it's generated (before opening the browser).
// This is useful for displaying the URL in a TUI while waiting for auth to complete.
func (m *Manager) AuthenticateWithURLCallback(ctx context.Context, urlCallback func(string)) (*AuthResult, error) {
	// Start local HTTP server on random port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to start local server: %w", err)
	}
	defer func() {
		if err := listener.Close(); err != nil {
			// Listener close failed during cleanup - no action needed
			_ = err
		}
	}()

	port := listener.Addr().(*net.TCPAddr).Port
	redirectURL := fmt.Sprintf("http://localhost:%d/callback", port)

	// Update config with dynamic redirect URL
	m.config.RedirectURL = redirectURL

	// Generate random state for CSRF protection
	state, err := generateState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}

	// Create channel for auth result
	authResult := make(chan authResponse, 1)

	// Create HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		m.handleCallback(w, r, state, authResult)
	})

	server := &http.Server{Handler: mux}

	// Start server in background
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			// Server error during cleanup - no action needed
			_ = err
		}
	}()

	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			// Shutdown error during cleanup - no action needed
			_ = err
		}
	}()

	// Generate auth URL
	authURL := m.config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)

	// Call the URL callback if provided
	if urlCallback != nil {
		urlCallback(authURL)
	}

	// Open browser
	if err := m.browserOpener(authURL); err != nil {
		return nil, fmt.Errorf("failed to open browser: %w\nPlease visit: %s", err, authURL)
	}

	// Wait for callback or timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, authTimeout)
	defer cancel()

	select {
	case result := <-authResult:
		if result.err != nil {
			return nil, result.err
		}

		return &AuthResult{
			Token:   result.token,
			Email:   result.email,
			AuthURL: authURL,
		}, nil
	case <-timeoutCtx.Done():
		if errors.Is(ctx.Err(), context.Canceled) {
			return nil, errors.New("authentication cancelled")
		}

		return nil, errors.New("authentication timeout")
	}
}

type authResponse struct {
	token *oauth2.Token
	email string
	err   error
}

func (m *Manager) handleCallback(w http.ResponseWriter, r *http.Request, expectedState string, result chan<- authResponse) {
	// Verify state
	state := r.URL.Query().Get("state")
	if state != expectedState {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)

		result <- authResponse{err: errors.New("invalid state parameter (possible CSRF attack)")}

		return
	}

	// Get authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		errMsg := r.URL.Query().Get("error")
		if errMsg == "" {
			errMsg = "no authorization code received"
		}

		http.Error(w, "Authorization failed", http.StatusBadRequest)

		result <- authResponse{err: fmt.Errorf("authorization failed: %s", errMsg)}

		return
	}

	// Exchange code for token
	token, err := m.config.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange code", http.StatusInternalServerError)

		result <- authResponse{err: fmt.Errorf("failed to exchange code: %w", err)}

		return
	}

	// Get user email
	email, err := m.getUserEmail(token)
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)

		result <- authResponse{err: fmt.Errorf("failed to get user email: %w", err)}

		return
	}

	// Send success response to browser
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := fmt.Fprintf(w, `
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<title>Authentication Successful</title>
			<style>
				body { 
					font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; 
					display: flex; 
					align-items: center; 
					justify-content: center; 
					height: 100vh; 
					margin: 0; 
					background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
				}
				.container { 
					background: white; 
					padding: 48px; 
					border-radius: 16px; 
					box-shadow: 0 20px 60px rgba(0,0,0,0.3); 
					text-align: center;
					max-width: 480px;
				}
				h1 { 
					color: #4CAF50; 
					margin: 0 0 16px 0;
					font-size: 32px;
					font-weight: 600;
				}
				.checkmark {
					font-size: 64px;
					margin-bottom: 16px;
					animation: scaleIn 0.3s ease-out;
				}
				@keyframes scaleIn {
					from { transform: scale(0); }
					to { transform: scale(1); }
				}
				p { 
					color: #666; 
					margin: 8px 0;
					font-size: 16px;
					line-height: 1.6;
				}
				.next-step {
					background: #f5f5f5;
					padding: 20px;
					border-radius: 8px;
					margin-top: 24px;
					border-left: 4px solid #667eea;
				}
				.next-step strong {
					color: #333;
					font-size: 18px;
					display: block;
					margin-bottom: 8px;
				}
				.next-step p {
					margin: 0;
					color: #666;
				}
			</style>
		</head>
		<body>
			<div class="container">
				<div class="checkmark">✓</div>
				<h1>Authentication Successful!</h1>
				<p>Account: <strong>%s</strong></p>
				<div class="next-step">
					<strong>→ Return to your terminal</strong>
					<p>Select which calendars to sync</p>
				</div>
		</div>
	</body>
	</html>
	`, email); err != nil {
		// Failed to write response, but token exchange was successful
		result <- authResponse{token: token, email: email, err: nil}
		return
	}

	result <- authResponse{token: token, email: email, err: nil}
}

// RefreshToken refreshes an expired OAuth2 token.
func (m *Manager) RefreshToken(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error) {
	if token.RefreshToken == "" {
		return nil, errors.New("no refresh token available")
	}

	tokenSource := m.config.TokenSource(ctx, token)

	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	return newToken, nil
}

// getUserEmail retrieves the user's email address from the token.
func (m *Manager) getUserEmail(token *oauth2.Token) (string, error) {
	client := m.config.Client(context.Background(), token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return "", fmt.Errorf("failed to get user info: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Error closing response body - no action needed
			_ = err
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get user info: status %d", resp.StatusCode)
	}

	var userInfo struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return "", fmt.Errorf("failed to decode user info: %w", err)
	}

	return userInfo.Email, nil
}

// generateState generates a random state string for CSRF protection.
func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

// openBrowser opens the default browser to the given URL.
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}
