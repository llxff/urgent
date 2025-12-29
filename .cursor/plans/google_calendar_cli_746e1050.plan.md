---
name: Google Calendar CLI
overview: Build a beautiful TUI-based Google Calendar CLI using Charm libraries that supports multiple Google accounts, OAuth authentication via local web server, and displays events in a color-coded table view.
todos:
  - id: setup-project
    content: Initialize project structure and install dependencies
    status: completed
  - id: create-ai-docs
    content: Create AI-friendly documentation (.cursor/rules.md, ARCHITECTURE.md, etc)
    status: completed
    dependencies:
      - setup-project
  - id: implement-setup-cmd
    content: Build setup command to store OAuth credentials in Keychain
    status: completed
    dependencies:
      - setup-project
      - create-ai-docs
  - id: implement-auth
    content: Build OAuth flow with local server and credential storage
    status: completed
    dependencies:
      - implement-setup-cmd
  - id: test-auth
    content: Write tests for auth module (Keychain integration, token management)
    status: completed
    dependencies:
      - implement-auth
  - id: implement-calendar-client
    content: Create Calendar API client wrapper and event fetching
    status: completed
    dependencies:
      - setup-project
  - id: test-calendar-client
    content: Write tests for calendar client (event parsing, filtering)
    status: completed
    dependencies:
      - implement-calendar-client
  - id: implement-output-formatter
    content: Build output formatter for table and JSON formats
    status: completed
    dependencies:
      - setup-project
  - id: test-output-formatter
    content: Write tests for output formatting (JSON schema, table)
    status: completed
    dependencies:
      - implement-output-formatter
  - id: build-tui-components
    content: Implement Bubbletea models, table view, and styling
    status: completed
    dependencies:
      - setup-project
  - id: test-tui-components
    content: Write tests for TUI components (model state, rendering)
    status: completed
    dependencies:
      - build-tui-components
  - id: implement-connect-cmd
    content: Build connect command with OAuth flow
    status: completed
    dependencies:
      - implement-auth
      - test-auth
  - id: test-connect-cmd
    content: Write integration tests for connect command
    status: completed
    dependencies:
      - implement-connect-cmd
  - id: implement-today-cmd
    content: Build today command with TUI table and remaining flag
    status: completed
    dependencies:
      - implement-calendar-client
      - build-tui-components
      - implement-output-formatter
      - test-calendar-client
      - test-tui-components
      - test-output-formatter
  - id: test-today-cmd
    content: Write tests for today command (both output formats)
    status: completed
    dependencies:
      - implement-today-cmd
  - id: implement-disconnect-cmd
    content: Build disconnect command with account selection TUI
    status: completed
    dependencies:
      - implement-auth
      - build-tui-components
      - test-auth
      - test-tui-components
  - id: test-disconnect-cmd
    content: Write tests for disconnect command
    status: completed
    dependencies:
      - implement-disconnect-cmd
  - id: implement-next-cmd
    content: Build next command with time filtering
    status: completed
    dependencies:
      - implement-calendar-client
      - implement-output-formatter
      - test-calendar-client
      - test-output-formatter
  - id: test-next-cmd
    content: Write tests for next command (both output formats)
    status: completed
    dependencies:
      - implement-next-cmd
  - id: add-color-coding
    content: Implement calendar color fetching and event styling
    status: completed
    dependencies:
      - implement-today-cmd
      - test-today-cmd
  - id: test-color-coding
    content: Write tests for color coding functionality
    status: completed
    dependencies:
      - add-color-coding
  - id: integration-testing
    content: Run full integration tests and E2E scenarios
    status: completed
    dependencies:
      - test-connect-cmd
      - test-today-cmd
      - test-disconnect-cmd
      - test-next-cmd
      - test-color-coding
---

# Google Calendar CLI Application

## Architecture Overview

```mermaid
graph TD
    CLI[CLI Entry Point] --> Auth[Auth Manager]
    CLI --> Events[Events Manager]
    CLI --> TUI[Charm TUI Components]

    Auth --> OAuth[Google OAuth2]
    Auth --> Store[Credential Store]

    Events --> CalAPI[Google Calendar API]
    Events --> Multi[Multi-Account Handler]

    TUI --> Table[Bubbles Table]
    TUI --> Style[Lipgloss Styling]

    Store --> Keychain[macOS Keychain]
```

## Testing Architecture

```mermaid
graph TD
    Tests[Test Suite] --> UnitTests[Unit Tests]
    Tests --> IntegrationTests[Integration Tests]

    UnitTests --> AuthTests[Auth Tests]
    UnitTests --> CalTests[Calendar Tests]
    UnitTests --> TUITests[TUI Tests]

    AuthTests --> TestStore[In-Memory Store]
    AuthTests --> MockOAuth[Mock OAuth Server]

    CalTests --> Fixtures[JSON Fixtures]
    CalTests --> TestClient[Fixture Client]

    TUITests --> ModelTests[State Tests]
    TUITests --> ExpectTests[go-expect TUI Tests]

    IntegrationTests --> RealFlow[Real Component Flow]
    RealFlow --> TestStore
    RealFlow --> Fixtures
```

## Technology Stack

- **CLI Framework**: `cobra` for command structure
- **TUI Components**: `bubbletea` (framework), `bubbles` (components), `lipgloss` (styling)
- **Google API**: `google.golang.org/api/calendar/v3`
- **OAuth2**: `golang.org/x/oauth2`
- **Config Management**: `viper` for configuration handling
- **Secure Storage**: `github.com/zalando/go-keyring` for macOS Keychain integration

## Project Structure

```javascript
cal/
├── cmd/
│   ├── root.go           # Root command setup + global flags
│   ├── root_test.go      # Root command tests
│   ├── connect.go        # Connect account command
│   ├── connect_test.go   # Connect command integration tests
│   ├── disconnect.go     # Disconnect account command
│   ├── disconnect_test.go
│   ├── today.go          # Show today's events
│   ├── today_test.go
│   └── next.go           # Show next event
│   └── next_test.go
├── internal/
│   ├── auth/
│   │   ├── manager.go    # OAuth flow & token management
│   │   ├── manager_test.go
│   │   ├── store.go      # Credential persistence
│   │   ├── store_test.go
│   │   └── teststore.go  # In-memory store for testing
│   ├── calendar/
│   │   ├── client.go     # Calendar API wrapper
│   │   ├── client_test.go
│   │   ├── events.go     # Event fetching & filtering
│   │   ├── events_test.go
│   │   └── testdata/     # JSON fixtures for calendar events
│   ├── output/
│   │   ├── formatter.go  # Output formatting (table, json)
│   │   ├── formatter_test.go
│   │   └── types.go      # Output data structures
│   ├── tui/
│   │   ├── model.go      # Bubbletea model
│   │   ├── model_test.go
│   │   ├── table.go      # Event table component
│   │   ├── table_test.go
│   │   ├── list.go       # Account list component
│   │   ├── list_test.go
│   │   ├── spinner.go    # Loading spinner component
│   │   ├── confirm.go    # Confirmation dialog component
│   │   ├── confirm_test.go
│   │   ├── status.go     # Status/success screens
│   │   ├── status_test.go
│   │   └── styles.go     # Lipgloss styling
│   └── config/
│       ├── config.go     # App configuration
│       └── config_test.go
├── test/
│   ├── integration/      # Integration tests
│   │   ├── auth_flow_test.go
│   │   ├── json_output_test.go
│   │   └── end_to_end_test.go
│   └── fixtures/         # Shared test fixtures
│       └── events.json
├── docs/
│   ├── ARCHITECTURE.md   # System architecture & design decisions
│   ├── CONTRIBUTING.md   # Development guidelines
│   └── API.md            # Internal API documentation
├── .cursor/
│   ├── rules.md          # Cursor AI coding rules
│   └── prompts.md        # Common prompts for AI assistance
├── .cursorignore         # Files to exclude from AI context
├── .golangci.yml         # golangci-lint configuration
├── Taskfile.yml          # Task runner for dev commands
├── AGENTS.md             # AI Agent guide (start here for any AI)
├── README.md             # Project overview & quick start
├── go.mod
└── main.go
```

## Implementation Plan

### 1. Core Setup

**Dependencies to add**:

- `github.com/spf13/cobra` - CLI framework
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/bubbles` - TUI components (table, list, spinner)
- `github.com/charmbracelet/lipgloss` - Styling
- `google.golang.org/api/calendar/v3` - Google Calendar API
- `golang.org/x/oauth2/google` - OAuth2 authentication
- `github.com/spf13/viper` - Configuration management
- `github.com/zalando/go-keyring` - macOS Keychain secure storage
- `github.com/stretchr/testify` - Testing assertions (minimal use)
- `github.com/netflix/go-expect` - TUI testing (bubbletea programs)
- `github.com/muesli/termenv` - Terminal color support detection

**Main entry point** (`main.go`):

- Initialize cobra root command
- Execute CLI

**Development Tools**:**golangci-lint** (`.golangci.yml`):

```yaml
linters:
  default: all

linters-settings:
  gocyclo:
    min-complexity: 15
  gocritic:
    enabled-tags:
      - diagnostic
      - style
      - performance

formatters:
  settings:
    gofmt:
      simplify: true

run:
  timeout: 5m
  tests: true
```

**Taskfile** (`Taskfile.yml`):

```yaml
version: '3'

tasks:
  default:
    desc: Show available tasks
    cmds:
      - task --list

  install:
    desc: Install dependencies
    cmds:
      - go mod download
      - go mod tidy

  build:
    desc: Build the application
    cmds:
      - go build -o bin/urgent .
    generates:
      - bin/urgent

  run:
    desc: Run the application
    cmds:
      - go run . {{.CLI_ARGS}}

  test:
    desc: Run all tests with coverage
    cmds:
      - go test -v -race -coverprofile=coverage.out ./...
      - go tool cover -html=coverage.out -o coverage.html
      - echo "Coverage: $(go tool cover -func=coverage.out | grep total | awk '{print $3}')"

  lint:
    desc: Run golangci-lint with auto-fix
    cmds:
      - golangci-lint run --fix ./...

  clean:
    desc: Clean build artifacts
    cmds:
      - rm -rf bin/
      - rm -f coverage.out coverage.html

  dev:
    desc: Run linter and tests (pre-commit check)
    cmds:
      - task: lint
      - task: test

  ci:
    desc: Run CI checks (lint + test with coverage)
    cmds:
      - golangci-lint run ./...
      - task: test

  setup:
    desc: Setup development environment
    cmds:
      - go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
      - task: install
```

**Usage**:

```bash
# Show all tasks
task

# Install dependencies
task install

# Build
task build

# Run tests with coverage
task test

# Lint (auto-fixes formatting)
task lint

# Pre-commit check (lint + test)
task dev

# Run the app
task run -- connect
task run -- today -o json
```

### 2. Authentication System (`internal/auth/`)

**OAuth Flow** (local web server):

- Start temporary HTTP server on `localhost` and random port. Make sure the port is free.
- Open browser to Google OAuth consent screen
- Receive callback with authorization code
- Exchange code for access/refresh tokens
- Store tokens securely

**Token Storage** (`store.go`):

- Store OAuth tokens in macOS Keychain using `go-keyring`
- Service name: `com.urgent.cli`
- Account identifier: user's Google email address
- Store JSON-serialized token data (access token, refresh token, expiry)
- Support multiple accounts with unique email identifiers
- Auto-refresh expired tokens
- Fallback gracefully if Keychain access denied

**Interface Design** (for testability):

```go
type Store interface {
    SaveToken(email string, token *oauth2.Token) error
    GetToken(email string) (*oauth2.Token, error)
    DeleteToken(email string) error
    ListAccounts() ([]string, error)
}

// Production implementation
type KeychainStore struct {}

// Test implementation (internal/auth/teststore.go)
type TestStore struct {
    tokens map[string]*oauth2.Token
    mu     sync.RWMutex
}
```

**Key Functions**:

```go
func Authenticate(ctx context.Context) (*oauth2.Token, error)
func SaveCredentials(email string, token *oauth2.Token) error
func LoadCredentials(email string) (*oauth2.Token, error)
func ListAccounts() ([]string, error)
func RemoveAccount(email string) error
```

**Tests** (`manager_test.go`, `store_test.go`):

- Use TestStore instead of Keychain
- Use httptest for OAuth server
- Test token refresh logic
- Test error scenarios

### 3. Calendar Integration (`internal/calendar/`)

**Client Manager** (`client.go`):

- Create authenticated Calendar API clients per account
- Pool management for multiple accounts
- Error handling & retry logic

**Interface Design** (for testability):

```go
type CalendarProvider interface {
    GetEvents(ctx context.Context, timeMin, timeMax time.Time) ([]*Event, error)
    GetCalendars(ctx context.Context) ([]*CalendarInfo, error)
}

// Production implementation
type GoogleCalendarClient struct {
    service *calendar.Service
    email   string
}

// Test implementation
type FixtureCalendarClient struct {
    events []*Event
}
```

**Event Operations** (`events.go`):

```go
func GetTodayEvents(clients []CalendarProvider) ([]*Event, error)
func GetRemainingEvents(clients []CalendarProvider) ([]*Event, error)
func GetNextEvent(clients []CalendarProvider, withinMinutes int) (*Event, error)
```

**Event Model**:

```go
type Event struct {
    Summary     string
    Start       time.Time
    End         time.Time
    Location    string
    Calendar    string
    Account     string
    Color       string  // Calendar color for styling
}
```

**Tests** (`client_test.go`, `events_test.go`):

- Use JSON fixtures from `testdata/`
- Test event parsing
- Test time-based filtering
- Test multi-account aggregation
- Test error handling

### 4. Output Formatting (`internal/output/`)

**Design Pattern**: Follow kubectl/gh CLI standard with `--output` flag**Formatter** (`formatter.go`):

```go
type OutputFormat string

const (
    FormatTable OutputFormat = "table"
    FormatJSON  OutputFormat = "json"
)

type Formatter interface {
    FormatEvents(events []*calendar.Event, filter string) (string, error)
    FormatNextEvent(event *calendar.Event, minutesUntil int) (string, error)
}

// JSON formatter (for data commands only)
type JSONFormatter struct{}

// Table/TUI formatter (for data commands only)
type TableFormatter struct{}
```

**Output Types** (`types.go`):

```go
// For today/remaining events
type EventsOutput struct {
    Events []EventJSON `json:"events"`
    Count  int         `json:"count"`
    Filter string      `json:"filter"`
}

// For next event
type NextEventOutput struct {
    HasEvent     bool       `json:"hasEvent"`
    Event        *EventJSON `json:"event,omitempty"`
    Message      string     `json:"message,omitempty"`
}

type EventJSON struct {
    Summary      string    `json:"summary"`
    Start        time.Time `json:"start"`
    End          time.Time `json:"end"`
    Location     string    `json:"location,omitempty"`
    Calendar     string    `json:"calendar"`
    Account      string    `json:"account"`
    ColorHEX     string    `json:"colorHex,omitempty"`
    Status       string    `json:"status,omitempty"` // upcoming, ongoing, past
    MinutesUntil *int      `json:"minutesUntil,omitempty"`
}
```

**Implementation Strategy**:

- Global `--output` flag on root command (only for `today` and `next` commands)
- Data commands (`today`, `next`) check output format
- If JSON: use JSONFormatter, print to stdout, exit
- If table: use TableFormatter or launch TUI
- Operational commands (`connect`, `disconnect`) use simple text output
- JSON is always valid, pretty-printed
- Consistent structure across data commands

**Tests** (`formatter_test.go`):

- Test JSON serialization
- Test JSON schema consistency
- Test table formatting
- Test edge cases (empty events, no accounts)

### 5. TUI Components (`internal/tui/`)

**Bubbletea Model** (`model.go`):

- Implement `tea.Model` interface
- Handle keyboard navigation (↑/↓, q to quit)
- Update event data periodically
- State management for multi-screen flows

**Table View** (`table.go`):

- Use `bubbles/table` component
- Columns: Time, Duration, Summary, Location, Calendar, Account
- Color-code rows by calendar color using lipgloss
- Highlight current/next event

**List View** (`list.go`):

- Use `bubbles/list` component for account selection
- Custom item rendering with lipgloss
- Keyboard navigation support
- Used by disconnect command

**Spinner Components** (`spinner.go`):

- Use `bubbles/spinner` for loading states
- OAuth flow progress indicator
- Keychain operations feedback
- Used by connect/disconnect commands

**Confirmation Dialog** (`confirm.go`):

- Reusable confirmation component
- Yes/No keyboard input (y/n, Enter/Esc)
- Warning/info styling variants
- Used by disconnect command

**Status Screens** (`status.go`):

- Success/error message screens
- Auto-dismiss or press any key
- Consistent styling for all status messages
- Used by connect/disconnect commands

**Styling** (`styles.go`):

- **Adaptive color palette** using `lipgloss.AdaptiveColor` and `termenv`
- Detect terminal capabilities (TrueColor, ANSI256, ANSI16, ASCII)
- Use ANSI color codes (0-255) instead of hex values
- Fallback gracefully for no-color terminals
- Title banners with lipgloss
- Status bar at bottom
- Calendar-specific colors (use Google's color IDs, adapt to terminal)
- Box styles (bordered, padded)
- Icon library (✓, ⚠, ✗, etc.)

**Color Strategy**:

```go
// Example adaptive color definition
var (
    // Use ANSI codes that work across themes
    PrimaryColor = lipgloss.AdaptiveColor{
        Light: "63",  // Blue for light terminals
        Dark: "63",   // Blue for dark terminals
    }

    SuccessColor = lipgloss.AdaptiveColor{
        Light: "42",  // Green
        Dark: "42",
    }

    // Let terminal theme define the actual RGB values
    // Works with Solarized, Dracula, Nord, etc.
)

// Detect terminal profile
profile := termenv.ColorProfile()
// Automatically adapts to terminal capabilities
```

**No Hardcoded RGB/Hex**:

- ❌ Don't use: `lipgloss.Color("#5865F2")`
- ✅ Use: `lipgloss.AdaptiveColor{Light: "63", Dark: "63"}`
- ❌ Don't use: `lipgloss.Color("rgb(88,101,242)")`
- ✅ Use: ANSI codes that adapt to user's terminal theme

**Tests** (`model_test.go`, `table_test.go`, `list_test.go`, `confirm_test.go`):

- Test model state transitions (no visual rendering)
- Test Init/Update/View methods
- Test keyboard input handling
- Test table/list row generation
- Test confirmation logic
- Use `go-expect` for full TUI testing

### 6. CLI Commands (`cmd/`)

**Root Command** (`root.go`):

```bash
urgent [command]

Global Flags:
  -o, --output string   Output format: table, json (default "table")
```

**Output Format Strategy**:

- Follow kubectl/gh CLI pattern with `--output` flag
- Data-returning commands (`today`, `next`) support JSON output
- **All commands use beautiful TUI with best practices**
- Operational commands (`setup`, `connect`, `disconnect`) have polished visual flow
- Default to beautiful TUI table view for data commands
- JSON output for scripting/automation
- Consistent JSON schema across data commands

**Setup Command** (`setup.go`):

```bash
urgent setup
```

**Purpose**: Securely store OAuth2 credentials in macOS Keychain (one-time setup)

**Beautiful TUI Flow**:

1. **Welcome Screen**:
            - Title: "Google Calendar CLI - Initial Setup" (styled with lipgloss)
            - Instructions box:
     ```
     ╭────────────────────────────────────────────╮
     │  You need OAuth2 credentials from          │
     │  Google Cloud Console                      │
     │                                            │
     │  Steps:                                    │
     │  1. Create project in Google Cloud        │
     │  2. Enable Calendar API                   │
     │  3. Create OAuth2 Desktop credentials     │
     │  4. Enter Client ID and Secret below      │
     ╰────────────────────────────────────────────╯
     ```

2. **Input Screen**:
            - Use `bubbles/textinput` for interactive prompts
            - Client ID input (visible):
     ```
     ╭────────────────────────────────────────────╮
     │  OAuth Client ID:                          │
     │  ┌──────────────────────────────────────┐  │
     │  │ 123456.apps.googleusercontent.com_   │  │
     │  └──────────────────────────────────────┘  │
     ╰────────────────────────────────────────────╯
     ```
            - Client Secret input (masked):
     ```
     ╭────────────────────────────────────────────╮
     │  OAuth Client Secret:                      │
     │  ┌──────────────────────────────────────┐  │
     │  │ •••••••••••••••••••••••••••••••••_   │  │
     │  └──────────────────────────────────────┘  │
     ╰────────────────────────────────────────────╯

     Press Enter to save  •  Esc to cancel
     ```

3. **Validation**:
            - Spinner: "Validating credentials format..."
            - Check Client ID format (*.apps.googleusercontent.com)
            - Check Client Secret not empty

4. **Saving to Keychain**:
            - Spinner: "Saving to macOS Keychain..."
            - Service: `com.urgent.cli.oauth`
            - Account: `oauth-credentials`
            - Password: JSON with `{client_id, client_secret}`

5. **Success Screen**:
            - Green checkmark: "✓ Setup Complete!"
            - Info box:
     ```
     ╭────────────────────────────────────────────╮
     │  ✓ Credentials saved to Keychain           │
     │  ✓ Service: com.urgent.cli.oauth              │
     │  ✓ Stored securely (encrypted)             │
     │                                            │
     │  Next step:                                │
     │  $ cal connect                             │
     ╰────────────────────────────────────────────╯
     ```
            - Auto-exit after 2 seconds

6. **Update Mode** (if credentials already exist):
            - Show warning that credentials exist
            - Confirmation: "Overwrite existing credentials? (y/n)"
            - If yes, proceed with update

**Security Features**:
- ✅ No files created on disk
- ✅ No environment variables
- ✅ Client secret masked during input
- ✅ Stored encrypted in Keychain
- ✅ Requires macOS user authentication to access
- ✅ Nothing in shell history

**Error Handling**:
- Keychain access denied → Clear error message with instructions
- Invalid format → Show expected format
- Network issues (if validation requires online check) → Retry option

**UX Features**:
- `bubbles/textinput` with cursor and editing
- Tab/Shift+Tab to move between fields
- Enter to submit, Esc to cancel
- Smooth transitions with lipgloss styling
- Help text always visible at bottom

**Connect Command** (`connect.go`):

```bash
urgent connect
``````

**Beautiful TUI Flow**:

1. **Initial Screen**:

                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Title: "Connect Google Calendar Account" (styled with lipgloss)
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Instructions: "This will open your browser for authentication"
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Spinner: "Preparing OAuth flow..."

2. **Authentication Phase**:

                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - "✓ Browser opened"
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Spinner: "Waiting for authentication..."
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Status: "Complete the authentication in your browser"

3. **Success Screen**:

                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Success icon: "✓" (green, styled)
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Message: "Successfully connected!"
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Account info box:
     ```javascript
          ╭───────────────────────────────╮
          │ Account: user@gmail.com       │
          │ Status:  Connected            │
          │ Time:    Dec 25, 2025 3:45 PM │
          ╰───────────────────────────────╯
     ```




                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Auto-exit after 2 seconds or press any key

4. **Error Handling**:

                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Red error message with helpful hints
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Option to retry or cancel

**UX Features**:

- Smooth spinners using `bubbles/spinner`
- Status transitions with lipgloss styling
- Graceful error messages
- Non-blocking - can be interrupted with Ctrl+C
- Progress indicators throughout flow

**Today Command** (`today.go`):

```bash
urgent today              # Show all events (TUI table)
urgent today --remaining  # Show only remaining events (TUI table)
urgent today -o json      # JSON output
urgent today --remaining -o json  # Remaining events as JSON
```

**Table output** (default):

- Launch TUI with table view
- Color-coded by calendar
- Real-time updates
- Interactive navigation

**JSON output** (`-o json`):

```json
{
  "events": [
    {
      "summary": "Team Standup",
      "start": "2025-12-25T09:00:00-08:00",
      "end": "2025-12-25T09:30:00-08:00",
      "location": "Zoom",
      "calendar": "Work",
      "account": "user@gmail.com",
      "colorHex": "ffff00",
      "status": "upcoming"
    }
  ],
  "count": 1,
  "filter": "all"
}
```

**Next Command** (`next.go`):

```bash
urgent next --within 30   # Show next event if within 30 minutes
urgent next --within 30 -o json
```

**Table output** (default):

- Simple text: "Meeting with Team in 15 minutes (3:00 PM - 4:00 PM)"
- Exit code 0 if event exists, 1 if not (useful for scripts)

**JSON output** (`-o json`):

```json
{
  "hasEvent": true,
  "event": {
    "summary": "Meeting with Team",
    "start": "2025-12-25T15:00:00-08:00",
    "end": "2025-12-25T16:00:00-08:00",
    "location": "Room 101",
    "minutesUntil": 15,
    "calendar": "Work",
    "account": "user@gmail.com",
    "colorHex": "ffff00"
  }
}
```

Or when no event:

```json
{
  "hasEvent": false,
  "message": "No events within 30 minutes"
}
```

**Disconnect Command** (`disconnect.go`):

```bash
urgent disconnect                              # Interactive TUI selection
urgent disconnect --account user@gmail.com     # Direct disconnect
```

**Interactive TUI Mode** (default):

1. **Account List Screen**:

                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Title: "Disconnect Google Calendar Account" (styled)
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Use `bubbles/list` component with custom styling
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - List format:
     ```javascript
          ╭────────────────────────────────────────╮
          │  Select account to disconnect:         │
          │                                        │
          │  › [color indication] user1@gmail.com  │
          │    Connected: Dec 20, 2025             │
          │                                        │
          │    [color indication] user2@gmail.com  │
          │    Connected: Dec 21, 2025             │
          ╰────────────────────────────────────────╯

          ↑/↓: Navigate  •  Enter: Select  •  q: Cancel
     ```




                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Highlight selected account
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Show connection date for each account

2. **Confirmation Screen**:

                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Warning icon and message
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Account details in styled box:
     ```javascript
          ╭─────────────────────────────────────────────╮
          │ ⚠ Disconnect Account?                       │
          │                                             │
          │ Account: [color indication] user1@gmail.com │
          │ Connected: Dec 20, 2025                     │
          │                                             │
          │ This will remove all credentials            │
          │ from your Keychain.                         │
          ╰─────────────────────────────────────────────╯

          y: Confirm  •  n: Cancel
     ```




3. **Processing**:

                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Spinner: "Removing credentials from Keychain..."
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Progress indicator

4. **Success Screen**:

                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Green checkmark: "✓ Successfully disconnected!"
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Account info:
     ```javascript
          ╭─────────────────────────────────────────────╮
          │ Account: [color indication] user1@gmail.com │
          │ Status:  Disconnected                       │
          │ Credentials removed from Keychain           │
          ╰─────────────────────────────────────────────╯
     ```

**Direct Mode** (with `--account` flag):

- Skip list selection, go directly to confirmation
- Same confirmation → processing → success flow
- Styled output throughout

**UX Features**:

- Keyboard navigation (↑/↓, Enter, y/n, q/Esc)
- Visual feedback for every action
- Clear status messages
- Graceful error handling
- Cancellable at any step (Ctrl+C or q)
- Consistent styling with rest of CLI

### 7. Configuration (`internal/config/`)

**OAuth Credentials Storage**: Keychain (via `urgent setup` command)

**Dynamic Port for OAuth**:
- Server starts on `localhost:0` (random available port)
- Google OAuth2 supports loopback IP redirect with any port
- No need to configure redirect URL in advance
- Redirect URL is generated dynamically: `http://localhost:{PORT}/callback`

**Setup Flow**:

Users will need to:

1. Create a Google Cloud Project
2. Enable Calendar API
3. Create OAuth2 credentials (Desktop app type)
   - Configure authorized redirect URIs: `http://localhost` (Google allows any port for loopback)
4. Run `urgent setup` to store client ID/secret in Keychain

**Note**: Both OAuth client credentials AND user tokens are stored securely in macOS Keychain, never in config files.

## Terminal UI/UX Best Practices

### Design Principles

1. **Progressive Disclosure**

- Show only relevant information at each step
- Multi-screen flows for complex operations (connect, disconnect)
- Clear progression through states

2. **Visual Hierarchy**

- Use lipgloss borders and boxes to group related info
- Title bars for context
- Status bars for help text
- Consistent spacing and padding

3. **Feedback & Responsiveness**

- Immediate visual feedback for every action
- Spinners for async operations (OAuth, API calls)
- Loading states with descriptive messages
- Success/error screens with clear outcomes

4. **Keyboard UX**

- Intuitive navigation (↑/↓ for lists, Enter to select)
- Universal quit (q, Esc, Ctrl+C)
- Confirmation prompts (y/n) for destructive actions
- Help text always visible at bottom

5. **Error Handling**

- Friendly error messages (not stack traces)
- Actionable suggestions ("Try: cal connect")
- Option to retry or cancel
- Clear error states with ✗ icons

6. **Accessibility**

- Graceful degradation for no-color terminals
- No reliance on color alone (use icons + color)
- Readable contrast ratios
- Works with terminal color schemes (via termenv)

7. **Performance**

- Responsive UI (no blocking)
- Async data fetching
- Smooth animations (60fps spinners)
- Instant keyboard response

8. **Consistency**

- Same key bindings across all screens
- Consistent box styles and borders
- Unified color palette
- Standard icons (✓, ✗, ⚠, →)

### Component Design Patterns

**Loading States**:

```javascript
⠋ Loading events from Google Calendar...
```

**List Selection**:

```javascript
╭─────────────────────────╮
│ › Item 1                │
│   Item 2                │
│   Item 3                │
╰─────────────────────────╯
↑/↓: Navigate • Enter: Select
```

**Confirmation**:

```javascript
╭────────────────────────╮
│ ⚠ Are you sure?        │
│                        │
│ [Details here]         │
╰────────────────────────╯
y: Yes • n: No
```

**Success**:

```javascript
╭────────────────────────╮
│ ✓ Success!             │
│                        │
│ [Action details]       │
╰────────────────────────╯
Press any key to continue
```

**Error**:

```
╭────────────────────────────╮
│ ✗ Error                    │
│                            │
│ [Error message]            │
│ [Suggested action]         │
╰────────────────────────────╯
Press any key to continue
```

### Specific TUI Implementations

**Connect Command Flow**:
- State machine: idle → preparing → authenticating → processing → success/error
- Bubbletea model with state transitions
- Non-blocking OAuth server in goroutine
- Update UI based on channel messages from OAuth server
- Timeout handling (2 minutes)

**Disconnect Command Flow**:
- State machine: list → confirmation → processing → success/error
- List model with arrow key navigation
- Confirmation model with y/n input
- Smooth transitions between states
- Cancel at any step returns to previous state

**Today/Next Command TUI**:
- Real-time event updates
- Smooth scrolling for long event lists
- Time-aware highlighting (current event in different color)
- Keyboard shortcuts: r (refresh), q (quit), ↑/↓ (scroll)
- Auto-refresh every 60 seconds for today view

### Animation & Polish

**Spinners**:
- Use bubbles/spinner with "dot" style for consistency
- Color: primary accent color from palette
- Descriptive loading messages

**Transitions**:
- Fade in/out for screen changes (optional, if performance allows)
- Instant feedback for keyboard input
- Smooth list scrolling

**Colors** (using termenv for terminal adaptation):
- Primary: `lipgloss.AdaptiveColor{Light: "63", Dark: "63"}` (blue)
- Success: `lipgloss.AdaptiveColor{Light: "42", Dark: "42"}` (green)
- Warning: `lipgloss.AdaptiveColor{Light: "214", Dark: "214"}` (yellow)
- Error: `lipgloss.AdaptiveColor{Light: "160", Dark: "160"}` (red)
- Muted: `lipgloss.AdaptiveColor{Light: "240", Dark: "240"}` (gray)
- Calendar colors: Use `lipgloss.Color()` with Google's color IDs
- Adapt to terminal color scheme via termenv profile detection



## Key Features Implementation

### Color-Coding Events

- Fetch calendar metadata to get color IDs
- Map Google Calendar colors to lipgloss styles
- Apply to table rows dynamically

### Multi-Account Support

- Store credentials per account in Keychain (keyed by email)
- Fetch events from all accounts in parallel
- Display account name in table column

### macOS Keychain Integration

- Use `go-keyring` library for secure token storage
- Each Google account stored as separate keychain entry
- Service: `com.urgent.cli`
- Account: Google email address
- Password field: JSON-serialized OAuth token
- Supports multiple accounts natively
- No plaintext credentials on disk

### OAuth Local Server Flow

1. Start ephemeral HTTP server on `localhost:0` (OS assigns random available port)
2. Get assigned port from server's listener
3. Generate OAuth URL with dynamic redirect: `http://localhost:{PORT}/callback`
4. Add state parameter for CSRF protection
5. Open browser automatically using `open` command
6. Wait for callback on `/callback` endpoint
7. Exchange authorization code for tokens
8. Shutdown server immediately

## Error Handling

- Graceful handling of network errors (retry with exponential backoff)
- Clear error messages for missing credentials
- Token refresh on 401/403 errors
- Rate limiting awareness (Google Calendar API quotas)

## Testing Approach

After implementation:

1. Test OAuth flow with one account
2. Add second account, verify both work
3. Test disconnect command
4. Verify table rendering with multiple events
5. Test edge cases (no events, single event, many events)
6. Test `next` command with various time windows

## Comprehensive Testing Strategy

### Philosophy: Minimal Mocking, Maximum Real Testing

We will use **real implementations** wherever possible, with test doubles only for external services (Google API, Keychain). Architecture will support dependency injection for testability.

### 1. Unit Tests

**Auth Module** (`internal/auth/*_test.go`):

- **Store Tests**: Use in-memory test store (`teststore.go`) instead of mocking
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Implements same interface as Keychain store
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Store/retrieve/delete tokens
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Multiple accounts
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Error handling
- **Manager Tests**: Use test store + httptest server for OAuth callbacks
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - OAuth flow state management
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Token exchange (mock Google OAuth endpoint)
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Token refresh logic
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Error scenarios (invalid code, network errors)

**Calendar Module** (`internal/calendar/*_test.go`):

- **Client Tests**: Use JSON fixtures from `testdata/` directory
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Parse Google Calendar API responses
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Handle various event formats (all-day, recurring, with/without location)
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Error responses from API
- **Events Tests**: Use in-memory event data
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Filter today's events
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Filter remaining events (time-based)
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Find next event within time window
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Multi-account event aggregation
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Time zone handling

**TUI Module** (`internal/tui/*_test.go`):

- **Model Tests**: Test bubbletea model state transitions
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Init state
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Update on key presses
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Event data updates
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - No visual rendering needed (test state only)
- **Table Tests**: Test table generation logic
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Row generation from events
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Column formatting
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Color assignment
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Sorting by time

**Config Module** (`internal/config/*_test.go`):

- Test OAuth credentials retrieval from Keychain
- Test default values
- Test validation
- Note: No file-based config, everything in Keychain

**Output Module** (`internal/output/*_test.go`):

- **Formatter Tests**: Test JSON and table output
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - JSON schema validation
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Pretty-printing
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Empty data handling
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Special characters in event names
- **Types Tests**: Test JSON marshaling/unmarshaling
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Time format consistency (RFC3339)
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Optional fields (omitempty)
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Null vs empty arrays

### 2. Integration Tests

**Command Tests** (`cmd/*_test.go`):

- Use real Cobra command execution
- Use test store (not Keychain)
- Use fixture data (not real API calls)
- Test full command flow end-to-end
- **Test both output formats for data commands** (`today`, `next`)

Example for `connect_test.go`:

```go
func TestConnectCommand(t *testing.T) {
    // Setup test store
    store := auth.NewTestStore()

    // Setup mock OAuth server
    server := httptest.NewServer(mockOAuthHandler())
    defer server.Close()

    // Execute command
    cmd := NewConnectCommand(store, server.URL)
    err := cmd.Execute()

    // Verify token stored
    accounts := store.ListAccounts()
    assert.Len(t, accounts, 1)
}
```

Example for `today_test.go`:

```go
func TestTodayCommandJSONOutput(t *testing.T) {
    // Setup with fixture data
    client := calendar.NewFixtureClient(testEvents)

    // Execute with JSON output
    cmd := NewTodayCommand(client)
    cmd.SetArgs([]string{"--output", "json"})

    output := captureStdout(func() {
        cmd.Execute()
    })

    // Parse and validate JSON
    var result output.EventsOutput
    err := json.Unmarshal([]byte(output), &result)
    require.NoError(t, err)
    assert.Equal(t, 3, result.Count)
    assert.Len(t, result.Events, 3)
}
```
````javascript

**Integration Test Suite** (`test/integration/*_test.go`):

- Full workflows with real components
- Use test fixtures for API responses
- Test account management flow:

    1. Connect account → verify stored
    2. List events → verify parsing
    3. Disconnect → verify removed

- Multi-account scenarios
- Error recovery
- **JSON output integration tests** (`json_output_test.go`):
    - `today` and `next` commands with `--output json`
    - JSON schema validation
    - Consistent time formats across commands
    - Edge cases (no events, empty responses)

### 3. Testability Architecture

**Interfaces for Dependency Injection**:

```go
// internal/auth/store.go
type Store interface {
    SaveToken(email string, token *oauth2.Token) error
    GetToken(email string) (*oauth2.Token, error)
    DeleteToken(email string) error
    ListAccounts() ([]string, error)
}

// Production: KeychainStore
type KeychainStore struct {}

// Testing: InMemoryStore
type TestStore struct {
    tokens map[string]*oauth2.Token
}
````
```go
// internal/calendar/client.go
type CalendarProvider interface {
    GetEvents(ctx context.Context, timeMin, timeMax time.Time) ([]*Event, error)
    GetCalendars(ctx context.Context) ([]*CalendarInfo, error)
}

// Production: GoogleCalendarClient
type GoogleCalendarClient struct {
    service *calendar.Service
    email   string
}

// Testing: FixtureCalendarClient
type FixtureCalendarClient struct {
    events []*Event
}
```



### 4. Test Data Strategy

**Fixtures** (`internal/calendar/testdata/`, `test/fixtures/`):

- Real Google Calendar API JSON responses
- Various scenarios:
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Empty calendar
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Single event
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Multiple events (past, current, future)
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - All-day events
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Recurring events
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Events with/without location
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Multiple calendars
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Multiple accounts

Example fixture structure:

```json
{
  "kind": "calendar#events",
  "items": [
    {
      "id": "event1",
      "summary": "Team Standup",
      "start": {"dateTime": "2025-12-25T09:00:00-08:00"},
      "end": {"dateTime": "2025-12-25T09:30:00-08:00"},
      "colorId": "1"
    }
  ]
}
```



### 5. TUI Testing

**Bubbletea Testing** with `go-expect`:

- Send keypresses to TUI
- Capture rendered output
- Verify table contents
- Test navigation
- **Test connect/disconnect flows**
- **Test confirmation dialogs**
- **Test spinner states**

Example for event table:

```go
func TestTodayCommandTUI(t *testing.T) {
    console, _, err := expect.NewConsole(expect.WithStdout(os.Stdout))
    require.NoError(t, err)
    defer console.Close()

    // Create TUI model with test data
    model := tui.NewModel(testEvents)

    // Run program
    go tea.NewProgram(model).Run()

    // Verify output contains event
    console.ExpectString("Team Standup")

    // Send quit key
    console.Send("q")
}
```

Example for disconnect flow:

```go
func TestDisconnectCommandTUI(t *testing.T) {
    console, _, err := expect.NewConsole()
    require.NoError(t, err)
    defer console.Close()

    // Setup with multiple accounts
    store := auth.NewTestStore()
    store.SaveToken("user1@gmail.com", testToken1)
    store.SaveToken("user2@gmail.com", testToken2)

    // Run disconnect command
    cmd := NewDisconnectCommand(store)
    go cmd.Execute()

    // Verify list appears
    console.ExpectString("Select account to disconnect")
    console.ExpectString("user1@gmail.com")
    console.ExpectString("user2@gmail.com")

    // Navigate and select
    console.Send("↓") // Move down
    console.Send("\n") // Select user2

    // Verify confirmation
    console.ExpectString("Disconnect Account?")
    console.ExpectString("user2@gmail.com")

    // Confirm
    console.Send("y")

    // Verify success
    console.ExpectString("✓ Successfully disconnected")

    // Verify account removed
    accounts := store.ListAccounts()
    assert.Len(t, accounts, 1)
    assert.Equal(t, "user1@gmail.com", accounts[0])
}
```

Example for connect flow:

```go
func TestConnectCommandTUI(t *testing.T) {
    console, _, err := expect.NewConsole()
    require.NoError(t, err)
    defer console.Close()

    // Setup mock OAuth server
    server := httptest.NewServer(mockOAuthHandler())
    defer server.Close()

    store := auth.NewTestStore()
    cmd := NewConnectCommand(store, server.URL)

    go cmd.Execute()

    // Verify initial screen
    console.ExpectString("Connect Google Calendar Account")
    console.ExpectString("Preparing OAuth flow")

    // Wait for OAuth completion (simulated)
    time.Sleep(100 * time.Millisecond)

    // Verify success
    console.ExpectString("✓ Successfully connected")
    console.ExpectString("@gmail.com")
}
```



### 6. Test Organization

**File naming**:

- `*_test.go` - unit tests (same package)
- `test/integration/*_test.go` - integration tests
- `testdata/` - fixtures
- `teststore.go`, `testclient.go` - test implementations

**Test build tags** (if needed):

```go
//go:build integration
```



### 7. Coverage Goals

- **Unit tests**: 95%+ coverage for business logic
- **Integration tests**: All user workflows covered
- **Edge cases**: Error handling, empty states, time boundaries

### 8. Running Tests

```bash
# All tests
go test ./...

# Unit tests only
go test ./internal/...

# Integration tests
go test ./test/integration/...

# With coverage
go test -cover ./...

# Verbose
go test -v ./...
```



### 9. CI/CD Considerations

**Tests should**:
- Not require Google Cloud credentials
- Not require actual Keychain access
- Run in CI environment (GitHub Actions, etc.)
- Complete quickly (< 30 seconds total)

**CI Pipeline** (using Taskfile):

```yaml
# .github/workflows/ci.yml
name: CI

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
                        - uses: actions/checkout@v3
                        - uses: go-task/setup-task@v1
                        - uses: actions/setup-go@v4
        with:
          go-version: '1.25'

      # Install golangci-lint
                        - name: golangci-lint
        uses: golangci/golangci-lint-action@v3

      # Run CI checks
                        - name: Run CI
        run: task ci
```

**Local Development Workflow**:

```bash
# Before committing
task dev

# This runs:
# 1. Lint with auto-fix (includes formatting via gofmt/goimports)
# 2. Tests with coverage
```

## Notes

- Google Calendar API requires OAuth2 scopes: `https://www.googleapis.com/auth/calendar.readonly`
- Credentials stored securely in macOS Keychain (not on disk)
- The TUI should handle terminal resize gracefully
- Keychain access requires user permission on first use
- JSON output follows kubectl/gh CLI patterns with `--output json` or `-o json`
- Only data commands (`today`, `next`) support JSON output
- Operational commands (`connect`, `disconnect`) use beautiful TUI with best practices
- JSON format uses RFC3339 for timestamps
- Exit codes: 0 for success, 1 for errors (useful for scripting)
- All TUI screens follow consistent design patterns
- Spinners for loading states, confirmations for destructive actions
- Keyboard navigation: ↑/↓ (navigate), Enter (select), y/n (confirm), q/Esc (quit)

## AI-Friendly Development Practices

This project follows best practices for AI-assisted development and future maintenance with Cursor.

### 1. Code Organization

**Clear Module Boundaries**:
- Each package has a single, well-defined responsibility
- Public APIs are minimal and well-documented
- Internal complexity is hidden behind clean interfaces
- File names clearly indicate their purpose

**Consistent Naming**:
- Interfaces end with `-er` suffix (e.g., `Store`, `Formatter`, `CalendarProvider`)
- Test implementations prefixed with `Test` or `Fixture` (e.g., `TestStore`, `FixtureClient`)
- TUI components named by function (e.g., `confirm.go`, `spinner.go`, `status.go`)

### 2. Documentation Standards

**File-Level Comments**:
Every source file starts with a package comment explaining its purpose:

```go
// Package auth provides OAuth2 authentication and credential management
// for Google Calendar API access. It supports multiple accounts with
// secure token storage in macOS Keychain.
package auth
```

**Function Documentation**:
All exported functions have clear godoc comments:

```go
// SaveToken stores an OAuth2 token for the given email address in the
// macOS Keychain. Returns an error if Keychain access is denied or the
// token cannot be serialized.
//
// Service: com.urgent.cli
// Account: email address
func SaveToken(email string, token *oauth2.Token) error
```

**Inline Comments for Complex Logic**:
- Explain WHY, not WHAT
- Document non-obvious behavior
- Reference external resources (RFCs, API docs)

```go
// OAuth state parameter prevents CSRF attacks per RFC 6749 Section 10.12
state := generateSecureState()
```

### 3. AI Agent Documentation

**AGENTS.md** - Universal guide for any AI assistant:

```markdown
# AI Agents Guide

> **Start here if you're an AI agent working with this codebase.**

## Project Overview

**cal** is a Google Calendar CLI with beautiful terminal UI. It supports:
- Multiple Google accounts
- OAuth2 authentication with macOS Keychain storage
- Terminal UI using Charm libraries (bubbletea, bubbles, lipgloss)
- JSON output for automation

**Language**: Go 1.25.5
**Architecture**: Clean architecture with interface-based design
**Testing**: 80%+ coverage, minimal mocking, real implementations

## Quick Navigation

| What you need | Where to look |
|---------------|---------------|
| System architecture | [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) |
| How to add features | [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) |
| Cursor-specific rules | [.cursor/rules.md](.cursor/rules.md) |
| Common AI prompts | [.cursor/prompts.md](.cursor/prompts.md) |
| Internal APIs | [docs/API.md](docs/API.md) |

## Project Structure

```
cmd/          # CLI commands (connect, disconnect, today, next)
internal/     # Business logic
  auth/       # OAuth + Keychain
  calendar/   # Google Calendar API
  tui/        # Terminal UI components
  output/     # JSON/table formatters
  config/     # Configuration
test/         # Integration tests
```

## Key Principles

### 1. Interfaces Over Concrete Types
Every external dependency is abstracted:
- `auth.Store` - Credential storage (Keychain in prod, memory in tests)
- `calendar.CalendarProvider` - Event fetching (API in prod, fixtures in tests)
- `output.Formatter` - Output formatting (table or JSON)

### 2. Dependency Injection
All dependencies passed explicitly (no globals):
```go
func NewCommand(store auth.Store, provider calendar.CalendarProvider) *cobra.Command
```

### 3. Test Patterns
- Use real implementations: `TestStore`, `FixtureClient` (NOT mocks)
- Table-driven tests for multiple scenarios
- Fixtures in `testdata/` directories
- Integration tests in `test/integration/`

### 4. Code Organization
- Files < 300 lines
- One primary responsibility per file
- Tests colocated with implementation
- Clear naming (e.g., `confirm.go` = confirmation dialog)

## Common Tasks

### Adding a New Command
1. Create `cmd/newcmd.go` following `cmd/today.go` pattern
2. Inject required dependencies (store, calendar provider, formatter)
3. Handle `--output` flag if returning data (or omit for operational commands)
4. Create TUI flow using components from `internal/tui/`
5. Add tests in `cmd/newcmd_test.go`
6. Register in `cmd/root.go`

### Adding a TUI Component
1. Create `internal/tui/component.go`
2. Implement bubbletea `tea.Model` interface (Init, Update, View)
3. Use lipgloss for styling (reference `internal/tui/styles.go`)
4. Add keyboard navigation
5. Write state transition tests in `component_test.go`

### Adding Tests
1. Follow existing patterns in same package
2. Use `TestStore` instead of real Keychain
3. Use fixture JSON instead of real API calls
4. Use table-driven tests for multiple cases
5. Aim for 80%+ coverage

### Modifying OAuth Flow
See: `internal/auth/manager.go`
Key: Local HTTP server receives callback, exchanges code for token

### Modifying Event Display
See: `internal/tui/table.go` for table layout
See: `internal/output/formatter.go` for JSON output

## Code Style

```go
// ✅ Good: Interface-based, injected dependencies
func NewService(store auth.Store) *Service {
    return &Service{store: store}
}

// ❌ Bad: Concrete type, global dependency
func NewService() *Service {
    return &Service{store: globalStore}
}

// ✅ Good: Wrapped errors with context
if err != nil {
    return fmt.Errorf("failed to save token: %w", err)
}

// ❌ Bad: Lost error context
if err != nil {
    return err
}

// ✅ Good: Table-driven tests
tests := []struct{name, input, want string}{ /* ... */ }
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) { /* ... */ })
}
```

## Architecture Patterns

### TUI State Machines
Commands with multiple screens use state machines:
```go
type state int
const (
    stateList state = iota
    stateConfirm
    stateProcessing
    stateSuccess
)
```
Example: `cmd/disconnect.go`

### Output Formatting
Data commands check `--output` flag:
```go
if outputFormat == "json" {
    return jsonFormatter.Format(events)
}
return tuiModel.Run() // Launch TUI
```

### Testing Without Mocks
Use test implementations of interfaces:
```go
// Production
type KeychainStore struct { /* uses real Keychain */ }

// Testing
type TestStore struct {
    tokens map[string]*oauth2.Token
}
```

## Important Constraints

### ❌ Don't Do This
- Import `cmd/` from `internal/` (breaks layering)
- Use mocks (use test implementations instead)
- Skip tests for exported functions
- Create files > 300 lines (split into smaller files)
- Use global state or singletons

### ✅ Do This
- Follow existing patterns
- Update docs when changing architecture
- Add godoc comments for exported symbols
- Use consistent error wrapping
- Write table-driven tests

## Getting Help

- **Architecture questions**: Read `docs/ARCHITECTURE.md`
- **Testing patterns**: Look at existing `*_test.go` files
- **TUI patterns**: Reference `internal/tui/` components
- **Cursor-specific**: Check `.cursor/rules.md`

## Debugging Tips

```bash
# Run with verbose output
go run . connect -v

# Run specific tests
go test -v ./internal/auth/...

# Run with race detector
go test -race ./...

# Check test coverage
go test -cover ./...
```

## For Cursor / AI Assistants

This codebase is designed for AI-assisted development:
- Consistent patterns → easy to learn one, apply everywhere
- Small files → fit in single context window
- Interface-based → easy to extend without modifying existing code
- Comprehensive tests → safe refactoring

When making changes:
1. Find similar existing code as reference
2. Follow the same pattern
3. Add tests following existing test patterns
4. Update docs if architecture changes
```

### 4. Cursor-Specific Files

**.cursor/rules.md** - Coding guidelines for AI:

```markdown
# Cursor AI Coding Rules for cal Project

## Code Style
- Use standard Go formatting (gofmt)
- Follow Effective Go guidelines
- Prefer table-driven tests
- Use interfaces for dependencies

## Architecture Rules
- Never import `cmd/` from `internal/`
- Keep TUI logic in `internal/tui/`
- All external API calls go through interface abstractions
- Test implementations live next to production code

## Testing Rules
- Every exported function must have tests
- Use real implementations, not mocks (see teststore.go pattern)
- Integration tests in `test/integration/`
- TUI tests use go-expect

## Documentation
- All exported symbols need godoc comments
- Complex functions need inline comments explaining WHY
- Update ARCHITECTURE.md when adding new modules

## Common Patterns
- OAuth flow: see internal/auth/manager.go
- TUI state machines: see cmd/disconnect.go
- JSON output: see internal/output/formatter.go
```

**.cursor/prompts.md** - Common AI prompts:

```markdown
# Common Prompts for cal Project

## Adding a New Command
"Add a new command `cal list-accounts` that shows all connected accounts in a TUI table. Follow the pattern from cmd/today.go and use the existing TUI components from internal/tui/."

## Adding Tests
"Add comprehensive tests for the [function/file] following the patterns in [existing_test_file]. Use TestStore instead of real Keychain."

## Refactoring
"Refactor [component] to follow the interface pattern used in internal/auth/store.go. Create a new interface and test implementation."

## TUI Enhancement
"Add a loading spinner to the connect command while waiting for OAuth. Use the spinner component from internal/tui/spinner.go and follow the state machine pattern."
```

**.cursorignore**:

```
# Exclude from AI context
*.log
.DS_Store
node_modules/
vendor/
.git/
dist/
*.test
coverage.txt

# Keep focused on source, not dependencies
go.sum

# Large generated files
internal/calendar/testdata/*.json
```

### 5. README Structure

**README.md** - AI-parseable structure:

```markdown
# cal - Google Calendar CLI

Beautiful terminal UI for Google Calendar with multi-account support.

## Quick Start
[Installation and setup instructions]

## Architecture
See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for system design.

## Project Structure
- `cmd/` - CLI commands (connect, disconnect, today, next)
- `internal/auth/` - OAuth2 and Keychain integration
- `internal/calendar/` - Google Calendar API wrapper
- `internal/tui/` - Terminal UI components (bubbletea)
- `internal/output/` - Output formatters (table, json)
- `test/` - Integration tests and fixtures

## Key Interfaces
- `auth.Store` - Credential storage (production: Keychain, test: memory)
- `calendar.CalendarProvider` - Event fetching (production: API, test: fixtures)
- `output.Formatter` - Output formatting (table or JSON)

## Development
See [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) for development setup.

## Testing
```bash
# All tests
go test ./...

# Integration tests
go test ./test/integration/...
```
```

### 6. Architecture Documentation

**docs/ARCHITECTURE.md**:

```markdown
# Architecture

## Overview
`urgent` is built with clean architecture principles, separating concerns into layers.

## Layers
1. **Commands** (cmd/) - User-facing CLI interface
2. **Business Logic** (internal/) - Core functionality
3. **External Services** - Google API, Keychain (abstracted via interfaces)

## Module Descriptions

### auth
Handles OAuth2 flow and credential storage.
- `manager.go`: OAuth flow coordinator
- `store.go`: Keychain interface and implementation
- `teststore.go`: In-memory implementation for tests

### calendar
Google Calendar API integration.
- `client.go`: API wrapper
- `events.go`: Event filtering and aggregation
- `testdata/`: Fixture JSON responses

### tui
Terminal UI components using Charm libraries.
- `model.go`: Bubbletea app model
- `table.go`: Event table display
- `list.go`: Account selection
- `confirm.go`: Y/N confirmation dialog
- `spinner.go`: Loading indicators
- `status.go`: Success/error screens
- `styles.go`: Lipgloss styling

### output
Output formatting for data commands.
- `formatter.go`: JSON and table formatters
- `types.go`: Output structure definitions

## Data Flow

[Mermaid diagram here]

## State Machines

### Connect Flow
idle → preparing → authenticating → processing → success/error

### Disconnect Flow
list → confirmation → processing → success/error

## Testing Strategy
See project plan for comprehensive testing approach.
```

### 7. Code Patterns for AI

**Consistent Error Handling**:

```go
// Pattern: Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to fetch events: %w", err)
}
```

**Dependency Injection**:

```go
// Pattern: Pass dependencies explicitly
func NewConnectCommand(store auth.Store, oauthConfig *oauth2.Config) *cobra.Command {
    return &cobra.Command{
        Use: "connect",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Use injected dependencies
        },
    }
}
```

**Table-Driven Tests**:

```go
// Pattern: Use test tables for multiple scenarios
func TestFormatEvents(t *testing.T) {
    tests := []struct {
        name     string
        events   []*Event
        want     string
        wantErr  bool
    }{
        {
            name:   "empty events",
            events: []*Event{},
            want:   `{"events":[],"count":0}`,
        },
        // More cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FormatEvents(tt.events)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
            }
            if got != tt.want {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

### 8. AI-Friendly Features

**Small, Focused Files**:
- Each file < 300 lines
- One primary responsibility per file
- Easy for AI to understand in single context

**Clear Separation of Concerns**:
- Business logic separate from presentation
- External services abstracted behind interfaces
- Tests separate from implementation (but colocated)

**Self-Documenting Code**:
- Descriptive variable and function names
- Minimal comments needed due to clarity
- Type names indicate their purpose

**Predictable Structure**:
- All commands follow same pattern
- All tests follow same pattern
- All TUI components follow same pattern
- AI can learn from one example and apply to others

### 9. Maintenance Prompts

Common maintenance tasks AI can help with:

```markdown
## Add New Calendar Source
"Add support for Microsoft Calendar following the same pattern as Google Calendar. Create a new provider implementation and update the multi-account logic."

## Add New Output Format
"Add YAML output format support following the JSON formatter pattern in internal/output/formatter.go."

## Enhance TUI Component
"Add a progress bar to the today command showing event timeline. Use bubbles/progress and follow the spinner component pattern."

## Add New Test Fixtures
"Create test fixtures for recurring events in internal/calendar/testdata/ following the existing event fixtures pattern."
```

### 10. Version Control for AI Context

**Meaningful Commit Messages**:
```
feat(tui): add confirmation dialog component

- Implement reusable Y/N confirmation dialog
- Add keyboard navigation (y/n, Enter/Esc)
- Include tests for state transitions
- Used by disconnect command for account removal

Follows pattern established in spinner.go
```

**PR Descriptions Include**:
- What changed
- Why it changed
- Which patterns/files to reference
- How to test

### 11. Benefits for Future AI Maintenance

✅ **Fast Context Loading**: Clear structure, well-documented
✅ **Pattern Recognition**: Consistent patterns across codebase
✅ **Safe Refactoring**: Comprehensive tests catch regressions
✅ **Easy Extension**: Interface-based design, clear examples
✅ **Reduced Ambiguity**: Explicit rules in .cursor/rules.md
✅ **Self-Contained**: Each module understandable independently

## Implementation Summary

This plan delivers a production-ready Google Calendar CLI with:

### Core Features
✅ Multi-account Google Calendar support
✅ Secure credential storage in macOS Keychain
✅ OAuth2 authentication with local server flow
✅ Beautiful TUI using Charm libraries (bubbletea, bubbles, lipgloss)
✅ JSON output for automation (`today`, `next` commands)
✅ Color-coded events by calendar

### Commands
1. **`urgent setup`** - Store OAuth client credentials securely in Keychain
2. **`urgent connect`** - Beautiful TUI flow for OAuth authentication
3. **`urgent disconnect`** - Interactive account selection with confirmation
4. **`urgent today`** - Show all/remaining events in TUI table or JSON
5. **`urgent next --within N`** - Show next event if within N minutes

### Technical Excellence
✅ Comprehensive testing (95%+ coverage, minimal mocking)
✅ Dependency injection for testability
✅ Interface-based architecture
✅ Test fixtures for Google Calendar API responses
✅ In-memory stores for testing (no real Keychain/API calls)
✅ TUI testing with go-expect

### UX Excellence
✅ State-machine driven TUI flows
✅ Progressive disclosure (multi-screen flows)
✅ Loading spinners for async operations
✅ Confirmation dialogs for destructive actions
✅ Graceful error handling with actionable messages
✅ Keyboard-first navigation
✅ Terminal color scheme adaptation
✅ Responsive, non-blocking UI

### AI-Friendly Development
✅ Comprehensive documentation (AGENTS.md, .cursor/rules.md, ARCHITECTURE.md)
✅ Universal AI agent guide (AGENTS.md) as entry point
✅ Cursor-specific rules and prompts
✅ Consistent code patterns across all modules
✅ Small, focused files (<300 lines each)
✅ Self-documenting code with clear naming
✅ Interface-based architecture for easy extension
✅ Table-driven tests for predictable patterns
✅ Inline comments explaining WHY, not WHAT
✅ Common prompts documented for future maintenance

### Project Statistics
- **22 TODO tasks** organized with dependencies (includes AI docs)
- **~15 source files** + tests (excluding fixtures)
- **6 major modules**: auth, calendar, output, tui, config, cmd
- **4 TUI components**: table, list, spinner, confirmation
- **3 test types**: unit, integration, TUI
- **6 doc files**: AGENTS.md, README, ARCHITECTURE, CONTRIBUTING, API, .cursor/rules


```````
