# urgent - Documentation Index

## Overview

**urgent** is a CLI tool that extracts Google Calendar events for use in scripts, status bars, notifications, and automation workflows.

| | |
|---|---|
| **Type** | CLI Application |
| **Language** | Go 1.25+ |
| **Platform** | macOS |
| **Storage** | Keychain (credentials), `~/.config/urgent/` (preferences) |

## Documentation

| Document | Purpose |
|----------|---------|
| [README.md](../README.md) | User guide, installation, automation examples |
| [AGENTS.md](../AGENTS.md) | Development entry point, project context, rules |
| [ARCHITECTURE.md](./ARCHITECTURE.md) | System design, data flows, design decisions |
| [STYLE.md](./STYLE.md) | Coding standards and best practices |
| [API.md](./API.md) | Internal interfaces and usage |

## Source Structure

```
urgent/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command, global flags
│   ├── setup.go           # OAuth credential setup
│   ├── connect.go         # Account connection
│   ├── disconnect.go      # Account removal
│   ├── calendars.go       # Calendar management
│   ├── today.go           # Today's events
│   └── next.go            # Next event
├── internal/
│   ├── auth/              # OAuth2 + Keychain
│   ├── calendar/          # Google Calendar API
│   ├── config/            # Configuration management
│   ├── output/            # JSON/table formatters
│   └── tui/               # Terminal UI components
├── docs/                  # This documentation
├── main.go                # Entry point
└── go.mod                 # Module definition
```

## Key Interfaces

```go
// Credential storage
type Store interface {
    SaveToken(email string, token *oauth2.Token) error
    GetToken(email string) (*oauth2.Token, error)
    DeleteToken(email string) error
    ListAccounts() ([]string, error)
}

// Calendar data
type CalendarProvider interface {
    GetEvents(ctx context.Context, timeMin, timeMax time.Time, calendarIDs []string) ([]*Event, error)
    GetCalendars(ctx context.Context) ([]*CalendarInfo, error)
}

// Output formatting
type Formatter interface {
    FormatEvents(events []*calendar.Event, filter string) (string, error)
    FormatNextEvent(event *calendar.Event, minutesUntil int) (string, error)
}
```

## Development

```bash
task dev      # Lint + test
task build    # Build binary
task test     # Run tests
```

Start with [AGENTS.md](../AGENTS.md) for development context.
