# Coding Standards for urgent

These standards guide development for the urgent project.

## Core Principles

1. **Security First**: All credentials in Keychain, never in files or environment variables
2. **Minimal Mocking**: Use interface-based test implementations (TestStore, FixtureClient)
3. **Clean Architecture**: Dependencies flow inward, interfaces at boundaries
4. **Table-Driven Tests**: Comprehensive edge case coverage
5. **Godoc Everything**: Every exported symbol has documentation

## Code Style

### Go Standards

```go
// Good
func NewClient(ctx context.Context, email string, store auth.Store) (*Client, error) {
    // Clear parameter names, explicit dependencies
}

// Bad
func NewClient(ctx context.Context, e string, s interface{}) (*Client, error) {
    // Vague names, interface{} instead of specific type
}
```

### Error Handling

```go
// Good - Wrap with context
return nil, fmt.Errorf("failed to fetch events: %w", err)

// Bad - Lose error context
return nil, err
```

### Interface Design

```go
// Good - Small, focused interface
type Store interface {
    SaveToken(email string, token *oauth2.Token) error
    GetToken(email string) (*oauth2.Token, error)
}

// Bad - Large, unfocused interface
type Store interface {
    SaveToken(email string, token *oauth2.Token) error
    GetToken(email string) (*oauth2.Token, error)
    ValidateToken(token *oauth2.Token) bool
    RefreshToken(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error)
    // ... 10 more methods
}
```

## Testing Standards

### Test Implementations Over Mocks

```go
// Good - Real in-memory implementation
type TestStore struct {
    tokens map[string]*oauth2.Token
}

func (s *TestStore) SaveToken(email string, token *oauth2.Token) error {
    s.tokens[email] = token
    return nil
}

// Bad - Mock framework
mockStore := new(MockStore)
mockStore.On("SaveToken", "user@example.com", mock.Anything).Return(nil)
```

### Table-Driven Tests

```go
// Good
func TestGetColorHex(t *testing.T) {
    tests := []struct {
        name    string
        colorID string
        want    string
    }{
        {"primary color", "1", "a4bdfc"},
        {"invalid color", "999", "4285f4"},
        {"empty string", "", "4285f4"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := getColorHex(tt.colorID)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### Coverage Goals

- **Business Logic**: 95%+ (config, calendar conversion, output formatting)
- **TUI Components**: 90%+ (state management, keyboard handling)
- **Commands**: 80%+ (flags, validation, testable logic)
- **Integration Tests**: Critical flows (connect, calendars, filtering)

## Module-Specific Rules

### auth/

- **Never log tokens or credentials**
- Always use context.Context for OAuth flows
- Token refresh happens automatically before expiry
- Two Keychain services: `com.urgent.cli` (tokens), `com.urgent.cli.oauth` (client creds)

```go
// Good - Automatic refresh
client := oauth2.NewClient(ctx, oauth2.ReuseTokenSource(nil, tokenSource))

// Bad - Manual refresh logic scattered
if token.Expiry.Before(time.Now()) {
    // refresh token...
}
```

### calendar/

- Context propagation for cancellation
- Events include calendar color information
- Support optional calendar ID filtering
- Graceful degradation if calendar list fails

```go
// Good - Optional filtering
func (c *Client) GetEvents(ctx context.Context, timeMin, timeMax time.Time, calendarIDs []string) ([]*Event, error)

// Bad - Always fetches all calendars
func (c *Client) GetEvents(ctx context.Context, timeMin, timeMax time.Time) ([]*Event, error)
```

### config/

- XDG Base Directory spec compliance
- Atomic writes (temp file + rename)
- Support both string and object YAML formats
- No concurrent access protection needed (CLI is sequential)

```go
// Good - Atomic write
tmpFile := configPath + ".tmp"
ioutil.WriteFile(tmpFile, data, 0644)
os.Rename(tmpFile, configPath)

// Bad - Direct write (can corrupt on interrupt)
ioutil.WriteFile(configPath, data, 0644)
```

### output/

- Consistent JSON schema across commands
- Exit codes: 0 for success/data found, 1 for no data
- Table formatter uses adaptive colors
- JSON formatter escapes all strings properly

### tui/

- **Adaptive Colors**: Use `lipgloss.AdaptiveColor`, not hardcoded colors
- State machines for multi-screen flows
- Keyboard shortcuts follow conventions (q=quit, esc=back, enter=confirm)
- Loading states with spinner for async operations

```go
// Good - Adaptive color
var PrimaryColor = lipgloss.AdaptiveColor{
    Light: "63",  // ANSI blue
    Dark:  "63",  // ANSI blue
}

// Bad - Hardcoded color
style := lipgloss.NewStyle().Foreground(lipgloss.Color("#0000FF"))
```

## Command Development

### Adding a New Command

1. **Create `cmd/command.go`**
   - Define cobra command with Use, Short, Long
   - Add flags with clear names and descriptions
   - Implement RunE function

2. **Handle Output Formats**
   ```go
   if GetOutputFormat() == "json" {
       formatter := output.NewJSONFormatter()
   } else {
       formatter := output.NewTableFormatter()
   }
   ```

3. **TUI Flow Pattern**
   ```go
   type commandStage int

   const (
       stageInitializing commandStage = iota
       stageLoading
       stageSuccess
       stageError
   )

   type commandModel struct {
       stage commandStage
       err   error
   }
   ```

4. **Add Tests**
   - Command structure (flags, defaults)
   - Business logic (separate from TUI)
   - Error handling

## Documentation Standards

### Godoc

```go
// Good
// GetEvents fetches calendar events within the specified time range.
// If calendarIDs is nil or empty, events from all calendars are fetched.
// Returns an error if the API request fails or token refresh fails.
func (c *Client) GetEvents(ctx context.Context, timeMin, timeMax time.Time, calendarIDs []string) ([]*Event, error)

// Bad
// GetEvents gets events
func (c *Client) GetEvents(ctx context.Context, timeMin, timeMax time.Time, calendarIDs []string) ([]*Event, error)
```

### Architecture Decisions

Document in `docs/ARCHITECTURE.md`:
- Why this approach over alternatives
- Trade-offs considered
- Examples of the pattern

## Git Commit Messages

```
# Good
feat(calendar): add calendar ID filtering to GetEvents

Allows filtering events by specific calendar IDs instead of fetching
all calendars. Maintains backward compatibility by treating nil/empty
as "fetch all".

# Bad
update calendar
```

Format: `<type>(<scope>): <description>`

Types: feat, fix, docs, test, refactor, chore

## Common Pitfalls to Avoid

### 1. Hardcoded Colors
```go
// Bad
style := lipgloss.NewStyle().Foreground(lipgloss.Color("blue"))

// Good
style := lipgloss.NewStyle().Foreground(PrimaryColor)
```

### 2. Global State
```go
// Bad
var globalStore Store

func DoSomething() {
    globalStore.SaveToken(...)
}

// Good
func DoSomething(store Store) {
    store.SaveToken(...)
}
```

### 3. Ignoring Context
```go
// Bad
func FetchData() error {
    resp, err := http.Get(url)
}

// Good
func FetchData(ctx context.Context) error {
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    resp, err := client.Do(req)
}
```

### 4. Not Using Interfaces
```go
// Bad
func ProcessEvents(client *calendar.Client) error {
    // Hard to test, tightly coupled
}

// Good
func ProcessEvents(provider calendar.CalendarProvider) error {
    // Easy to test with FixtureClient
}
```

### 5. Verbose Error Messages
```go
// Bad
return fmt.Errorf("Error: Failed to save token for user %s to keychain: %v", email, err)

// Good
return fmt.Errorf("failed to save token: %w", err)
```

## Quick Reference

### Running Tests
```bash
task test               # All tests with coverage
go test -v ./cmd/...   # Specific package
go test -run TestName  # Specific test
```

### Formatting
```bash
go fmt ./...           # Format all files
goimports -w .         # Fix imports
```

### Building
```bash
task build             # Build to bin/urgent
task run -- today      # Run command directly
```

### Coverage
```bash
task test              # Generates coverage.html
open coverage.html     # View in browser
```

## Success Criteria

Code is ready to merge when:
- All tests pass
- Coverage meets targets (95%+ for business logic)
- All exported symbols have Godoc
- No hardcoded colors in TUI
- Dependencies injected via interfaces
- Error messages include context
- README and docs updated
- `task dev` runs clean (lint + test)
