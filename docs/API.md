# Internal API Documentation

This document describes the internal APIs and interfaces used in urgent.

## auth Package

### Store Interface

```go
type Store interface {
    SaveToken(email string, token *oauth2.Token) error
    GetToken(email string) (*oauth2.Token, error)
    DeleteToken(email string) error
    ListAccounts() ([]string, error)
}
```

**SaveToken** stores an OAuth2 token for a Google account.
- Parameters: email (account identifier), token (OAuth2 token)
- Returns: error if save fails
- Thread-safe: Yes

**GetToken** retrieves a stored token.
- Parameters: email (account identifier)
- Returns: token or error if not found
- Thread-safe: Yes

**DeleteToken** removes a stored token.
- Parameters: email (account identifier)
- Returns: error if deletion fails
- Thread-safe: Yes

**ListAccounts** returns all stored account emails.
- Returns: slice of email addresses or error
- Thread-safe: Yes

### Manager

```go
type Manager struct {
    store Store
    config *oauth2.Config
}

func NewManager(store Store, clientID, clientSecret string) *Manager
func (m *Manager) Authenticate(ctx context.Context) (*oauth2.Token, string, error)
func (m *Manager) RefreshToken(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error)
```

**Authenticate** performs OAuth2 flow.
- Returns: token, email, error
- Starts local HTTP server on random port
- Opens browser for user consent
- Waits for callback with timeout

**RefreshToken** refreshes an expired token.
- Returns: new token or error
- Automatically called by calendar client
- Saves refreshed token to store

## calendar Package

### CalendarProvider Interface

```go
type CalendarProvider interface {
    GetEvents(ctx context.Context, timeMin, timeMax time.Time) ([]*Event, error)
    GetCalendars(ctx context.Context) ([]*CalendarInfo, error)
}
```

**GetEvents** fetches events in time range.
- Parameters: context, start time, end time, calendar IDs (optional)
- Returns: events or error
- Automatically refreshes expired tokens
- If calendar IDs provided, filters events to those calendars only

**GetCalendars** fetches calendar list.
- Returns: calendar metadata including colors
- Used for event color coding

### Event Structure

```go
type Event struct {
    ID           string
    Summary      string
    Description  string
    Location     string
    Start        time.Time
    End          time.Time
    CalendarID   string
    CalendarName string
    ColorHex     string
    Account      string
    Status       string
}
```

### Client

```go
type Client struct {
    service *calendar.Service
    email   string
    store   auth.Store
}

func NewClient(ctx context.Context, email string, store auth.Store) (*Client, error)
```

**NewClient** creates an authenticated calendar client.
- Retrieves token from store
- Creates Google Calendar API service
- Configures automatic token refresh

## output Package

### Formatter Interface

```go
type Formatter interface {
    FormatEvents(events []*calendar.Event, filter string) (string, error)
    FormatNextEvent(event *calendar.Event, minutesUntil int) (string, error)
}
```

**FormatEvents** formats event list.
- Parameters: events, filter ("all" or "remaining")
- Returns: formatted string or error

**FormatNextEvent** formats single next event.
- Parameters: event, minutes until start
- Returns: formatted string or error

### JSON Structures

```go
type EventsOutput struct {
    Events []EventJSON `json:"events"`
    Count  int         `json:"count"`
    Filter string      `json:"filter"`
}

type NextEventOutput struct {
    HasEvent     bool      `json:"hasEvent"`
    Event        EventJSON `json:"event,omitempty"`
    MinutesUntil int       `json:"minutesUntil,omitempty"`
}
```

## tui Package

### Styles

```go
var (
    PrimaryColor   lipgloss.AdaptiveColor
    SuccessColor   lipgloss.AdaptiveColor
    WarningColor   lipgloss.AdaptiveColor
    ErrorColor     lipgloss.AdaptiveColor
    MutedColor     lipgloss.AdaptiveColor
)

func TitleStyle() lipgloss.Style
func BoxStyle() lipgloss.Style
func ErrorStyle() lipgloss.Style
func SuccessStyle() lipgloss.Style
```

All styles use adaptive colors that work with light and dark terminals.

### Components

Each TUI component implements the bubbletea Model interface:

```go
type Model interface {
    Init() tea.Cmd
    Update(tea.Msg) (tea.Model, tea.Cmd)
    View() string
}
```

Components are composable and can be nested within command-specific models.

## config Package

### Manager Functions

```go
func Load() (*Config, error)
func Save(cfg *Config) error
func GetEnabledCalendarIDs(email string) ([]string, error)
func SetEnabledCalendars(email string, selections []CalendarSelection) error
```

**Load** reads configuration from disk.
- Returns: config or error
- Location: `$XDG_CONFIG_HOME/urgent/config.yaml` or `~/.config/urgent/config.yaml`
- Creates empty config if file doesn't exist

**Save** writes configuration to disk with atomic write.
- Parameters: config
- Returns: error if save fails
- Atomic: writes to temp file, then renames

**GetEnabledCalendarIDs** retrieves enabled calendar IDs for an account.
- Parameters: email (account identifier)
- Returns: calendar IDs or error
- Only returns IDs, ignores name field

**SetEnabledCalendars** updates enabled calendars for an account.
- Parameters: email, calendar selections (ID + optional name)
- Returns: error if update fails
- Preserves existing accounts

### Config Structure

```go
type Config struct {
    Accounts map[string]AccountConfig `yaml:"accounts"`
}

type AccountConfig struct {
    Email            string             `yaml:"email"`
    EnabledCalendars []CalendarSelection `yaml:"enabled_calendars"`
}

type CalendarSelection struct {
    ID   string `yaml:"id"`
    Name string `yaml:"name,omitempty"` // Optional, for human readability
}
```

**CalendarSelection** supports two YAML formats:
- Simple: `- primary` (string)
- Detailed: `- id: primary` with optional `name: My Calendar`

Custom `UnmarshalYAML` handles both formats automatically.

## Usage Examples

### Authenticate and Fetch Events

```go
// Initialize store
store := auth.NewKeychainStore()

// Authenticate
manager := auth.NewManager(store, clientID, clientSecret)
result, err := manager.Authenticate(ctx)
if err != nil {
    return err
}

// Save token
err = store.SaveToken(result.Email, result.Token)
if err != nil {
    return err
}

// Create calendar client
client, err := calendar.NewClient(ctx, result.Email, store)
if err != nil {
    return err
}

// Fetch today's events
now := time.Now()
start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
end := start.Add(24 * time.Hour)

events, err := client.GetEvents(ctx, start, end)
if err != nil {
    return err
}

// Format output
formatter := output.NewJSONFormatter()
output, err := formatter.FormatEvents(events, "all")
if err != nil {
    return err
}

fmt.Println(output)
```

### List All Accounts

```go
store := auth.NewKeychainStore()
accounts, err := store.ListAccounts()
if err != nil {
    return err
}

for _, email := range accounts {
    fmt.Println(email)
}
```

### Delete Account

```go
store := auth.NewKeychainStore()
err := store.DeleteToken("user@example.com")
if err != nil {
    return err
}
```

## Error Handling

All errors are wrapped with context using `fmt.Errorf("...%w", err)`.

Common error types:
- `auth.ErrNotFound` - Account not in Keychain
- `auth.ErrInvalidToken` - Token invalid or expired (after refresh attempt)
- `calendar.ErrAPIQuotaExceeded` - Google API quota exceeded
- `context.DeadlineExceeded` - Operation timeout

Check specific errors with `errors.Is()`:

```go
if errors.Is(err, auth.ErrNotFound) {
    // Handle missing account
}
```
