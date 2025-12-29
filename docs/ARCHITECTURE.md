# Architecture

## Overview

`urgent` is built with clean architecture principles, separating concerns into layers with clear boundaries and dependencies flowing inward.

## Layers

1. **Commands** (`cmd/`) - User-facing CLI interface
2. **Business Logic** (`internal/`) - Core functionality  
3. **External Services** - Google API, Keychain (abstracted via interfaces)

## Design Principles

### 1. Interface-Based Design

All external dependencies are accessed through interfaces, enabling:
- Easy testing with in-memory implementations
- Flexibility to swap implementations
- Clear contracts between modules

### 2. Dependency Injection

Dependencies are passed explicitly through constructors, never accessed globally:
```go
// ✅ Good
func NewManager(store Store, client *http.Client) *Manager

// ❌ Bad
func NewManager() *Manager {
    store := keychain.NewStore()  // Hard dependency
}
```

### 3. Minimal Mocking

Instead of mock frameworks, we provide real test implementations:
- `TestStore` - In-memory credential storage
- `FixtureCalendarClient` - Pre-loaded calendar data
- `httptest.Server` - OAuth flow testing

## Module Descriptions

### auth

Handles OAuth2 flow and credential storage.

**Files:**
- `manager.go` - OAuth flow coordinator, token refresh
- `store.go` - Keychain interface and production implementation
- `teststore.go` - In-memory implementation for tests

**Interfaces:**
```go
type Store interface {
    SaveToken(email string, token *oauth2.Token) error
    GetToken(email string) (*oauth2.Token, error
    DeleteToken(email string) error
    ListAccounts() ([]string, error)
}
```

**Key Design Decisions:**
- Automatic token refresh before expiry
- Two Keychain services: one for tokens, one for OAuth credentials
- State parameter for CSRF protection in OAuth flow
- Dynamic port allocation for local OAuth server

### calendar

Google Calendar API integration.

**Files:**
- `client.go` - Calendar API wrapper
- `events.go` - Event fetching, filtering, and transformation
- `types.go` - Event and calendar data structures
- `fixtureclient.go` - Test implementation with fixtures

**Interfaces:**
```go
type CalendarProvider interface {
    GetEvents(ctx context.Context, timeMin, timeMax time.Time) ([]*Event, error)
    GetCalendars(ctx context.Context) ([]*CalendarInfo, error)
}
```

**Key Design Decisions:**
- Events are fetched with color information
- Support for multiple calendars per account
- Context propagation for cancellation
- Fixtures loaded from JSON for consistent testing
- Calendar filtering by enabled IDs

### config

Configuration management following XDG Base Directory specification.

**Files:**
- `config.go` - Config data structures
- `manager.go` - Config persistence, atomic writes
- `testhelpers.go` - Test utilities

**Key Design Decisions:**
- XDG Base Directory support (`~/.config/urgent/config.yaml`)
- YAML format for human readability
- Atomic file writes (write to temp, rename)
- Flexible calendar selection format (string or object with name)
- Optional name field for human readability, only ID used for filtering
- No concurrent access management (CLI runs one command at a time)

**Config Structure:**
```yaml
accounts:
  user@example.com:
    email: user@example.com
    enabled_calendars:
      - id: primary
        name: user@example.com
      - id: work@group.calendar.google.com
        name: Work Calendar
```

### output

Output formatting for table and JSON.

**Files:**
- `formatter.go` - Formatter interface and factory
- `json.go` - JSON formatter implementation
- `table.go` - Table formatter using Bubbles
- `types.go` - Output data structures

**Interfaces:**
```go
type Formatter interface {
    FormatEvents(events []*Event, filter string) (string, error)
    FormatNextEvent(event *Event, minutesUntil int) (string, error)
}
```

**Key Design Decisions:**
- Consistent JSON schema across commands
- Table formatter uses adaptive colors
- Exit codes integrated with output (0 for success, 1 for no data)

### tui

Terminal UI components and styling.

**Files:**
- `styles.go` - Adaptive color palette and reusable styles
- `table.go` - Event table component
- `list.go` - Account selection list
- `spinner.go` - Loading spinner
- `confirm.go` - Confirmation dialog
- `status.go` - Success/error status screens
- `input.go` - Text input with validation

**Key Design Decisions:**
- Adaptive colors using `lipgloss.AdaptiveColor` and `termenv`
- State machine pattern for multi-screen flows
- Keyboard navigation follows standard conventions
- Icons + color for accessibility (not color alone)

### config

Application configuration (minimal, most data in Keychain).

**Files:**
- `config.go` - Runtime configuration
- `oauth.go` - OAuth credential retrieval

**Key Design Decisions:**
- No file-based configuration
- All sensitive data in Keychain
- Runtime configuration only for non-sensitive settings

## Data Flow

### OAuth Authentication Flow

```
1. User runs: urgent connect
2. Command retrieves OAuth credentials from Keychain
3. Auth manager starts local HTTP server on random port
4. Browser opens to Google OAuth consent page
5. User grants permission
6. Google redirects to localhost:{PORT}/callback
7. Manager exchanges auth code for tokens
8. Tokens saved to Keychain with user's email as key
9. Success screen displayed to user
```

### Event Fetching Flow

```
1. User runs: urgent today
2. Command retrieves all account tokens from Keychain
3. Calendar client created for each account
4. Events fetched in parallel from all accounts
5. Events merged and sorted by start time
6. Output formatter generates table or JSON
7. Result displayed to user
```

### Token Refresh Flow

```
1. Calendar client attempts API call
2. If token expired, refresh automatically
3. If refresh succeeds, save new token to Keychain
4. Retry original API call
5. If refresh fails, return auth error
```

## Error Handling Strategy

### Error Types

1. **User Errors** - Invalid input, missing setup
   - Clear error messages with guidance
   - Exit code 1
   - No stack traces

2. **System Errors** - Network failures, API errors
   - Retry with exponential backoff
   - Detailed error context
   - Wrapped errors for debugging

3. **Auth Errors** - Expired/invalid credentials
   - Prompt user to re-authenticate
   - Clear instructions
   - Never show raw OAuth errors

### Error Wrapping

```go
if err != nil {
    return fmt.Errorf("failed to fetch events for %s: %w", email, err)
}
```

## Testing Strategy

### Unit Tests

Test individual functions and methods with:
- Table-driven tests for multiple scenarios
- Test implementations of interfaces
- Edge cases and boundary conditions
- Target: 95%+ coverage

### Integration Tests

Test interactions between modules:
- Full command flows with test implementations
- Error propagation
- State transitions
- Located in `test/integration/`

### TUI Tests

Test terminal interactions:
- go-expect for simulating user input
- Screen content verification
- Keyboard navigation
- Located alongside TUI components

## Security Considerations

### Credential Storage

- **Never** store credentials in files
- **Always** use macOS Keychain via go-keyring
- **Two services**:
  - `com.urgent.cli` - User OAuth tokens
  - `com.urgent.cli.oauth` - OAuth client credentials

### OAuth Security

- State parameter prevents CSRF attacks
- Loopback redirect (localhost) only
- No credentials in URLs or logs
- Automatic token refresh

### Sensitive Data Handling

```go
// ✅ Good
log.Printf("Fetching events for %s", email)

// ❌ Bad
log.Printf("Token: %s", token.AccessToken)
```

## Performance Characteristics

### Parallel Operations

- Multiple account event fetching is parallelized
- Bounded concurrency (max 10 concurrent requests)
- Context cancellation propagates to all goroutines

### Resource Management

- HTTP servers shutdown gracefully
- Contexts timeout after reasonable duration
- Database connections (if added) pooled

### Caching

- Currently no caching (fetches are fast enough)
- Future: Consider short-term event cache
- Token refresh handled by oauth2 library

## Future Enhancements

### Planned

1. Event creation and modification
2. Event notifications (desktop/system)
3. Recurring event support
4. Calendar color customization
5. Config file for non-sensitive settings (theme preferences, etc.)

### Considered but Deferred

1. Linux Keychain support (needs keyring abstraction)
2. Windows Credential Manager support
3. Event search across all calendars
4. Multiple time zones support

## Design Decisions Log

### Why Keychain for OAuth Credentials?

**Decision:** Store OAuth client ID/secret in Keychain instead of config file.

**Rationale:**
- More secure than file storage
- Prevents accidental credential commits
- Consistent with token storage
- Better UX (one-time setup command)

**Alternatives Considered:**
- Config file: Less secure, risk of git commits
- Environment variables: Difficult for CLI users
- Hardcoded: Impossible (each user has own credentials)

### Why Dynamic Port for OAuth Server?

**Decision:** Use `localhost:0` for OAuth callback server.

**Rationale:**
- Prevents port conflicts
- Google supports loopback IP with any port
- More flexible for users
- No configuration needed

**Alternatives Considered:**
- Fixed port (8080): Port conflicts likely
- Port range (8080-8090): Partial solution, still conflicts
- Manual port selection: Poor UX

### Why Minimal Mocking?

**Decision:** Use test implementations instead of mock frameworks.

**Rationale:**
- Easier to understand and maintain
- Forces good interface design
- Faster test execution
- Fewer dependencies

**Alternatives Considered:**
- gomock: More boilerplate, generated code
- testify/mock: Runtime verification, complex setup
- Manual mocks per test: Too much duplication

### Why Cobra Over Other CLI Frameworks?

**Decision:** Use spf13/cobra for CLI framework.

**Rationale:**
- Industry standard (kubectl, gh, hugo use it)
- Excellent flag handling
- Subcommand support
- Auto-generated help/completion

**Alternatives Considered:**
- urfave/cli: Less feature-rich
- kong: Different paradigm, less common
- Standard flag: Too basic for subcommands

## Debugging Guide

### Enable Verbose Logging

```go
// Add to main.go for debugging
log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
```

### Test Keychain Integration

```go
// Create test account
store := auth.NewKeychainStore()
token := &oauth2.Token{AccessToken: "test"}
err := store.SaveToken("test@example.com", token)
```

### Simulate OAuth Flow

See `internal/auth/manager_test.go` for httptest examples.

### Debug TUI Rendering

```go
// Output View() to file for inspection
os.WriteFile("debug.txt", []byte(model.View()), 0644)
```
