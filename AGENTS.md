# AGENTS.md - AI Agent Guide

**START HERE** - This is the universal entry point for all AI agents working on this project.

## Project Overview

`urgent` is a TUI-based Google Calendar CLI built in Go that supports multiple accounts, secure credential storage, and automation-friendly output.

## Quick Context

- **Language**: Go 1.25+
- **Architecture**: Clean architecture with interface-based design
- **CLI Framework**: Cobra
- **TUI**: Bubbletea + Bubbles + Lipgloss
- **Auth**: OAuth2 with macOS Keychain storage
- **Testing**: Minimal mocking, 95%+ coverage goal

## Key Files to Understand

1. **`docs/STYLE.md`** - Coding standards and patterns
2. **`docs/ARCHITECTURE.md`** - System architecture and design decisions
3. **`docs/API.md`** - Internal API documentation
4. **`docs/adrs/`** - Architecture Decision Records (important decisions)

## Module Structure

```
cmd/          - Cobra commands (user-facing interface)
internal/
  auth/       - OAuth2 + Keychain storage
  calendar/   - Google Calendar API wrapper
  output/     - JSON and table formatters
  tui/        - Bubbletea components and styling
  config/     - XDG-compliant configuration management
```

## Core Interfaces

### Auth Module
```go
type Store interface {
    SaveToken(email string, token *oauth2.Token) error
    GetToken(email string) (*oauth2.Token, error)
    DeleteToken(email string) error
    ListAccounts() ([]string, error)
}
```

### Calendar Module
```go
type CalendarProvider interface {
    GetEvents(ctx context.Context, timeMin, timeMax time.Time, calendarIDs []string) ([]*Event, error)
    GetCalendars(ctx context.Context) ([]*CalendarInfo, error)
}
```

### Config Module
```go
type Manager interface {
    Load() (*Config, error)
    Save(cfg *Config) error
    GetEnabledCalendarIDs(email string) ([]string, error)
    SetEnabledCalendars(email string, selections []CalendarSelection) error
}
```

### Output Module
```go
type Formatter interface {
    FormatEvents(events []*calendar.Event, filter string) (string, error)
    FormatNextEvent(event *calendar.Event, minutesUntil int) (string, error)
}
```

## Commands

1. **`urgent setup`** - Store OAuth credentials in Keychain (one-time)
2. **`urgent calendars`** - Unified account and calendar management (add/remove accounts, select calendars)
3. **`urgent today`** - Show today's events from enabled calendars (TUI table or JSON)
4. **`urgent next --within N`** - Show next event within N minutes (TUI or JSON)

## Testing Strategy

- **Unit Tests**: Test with in-memory implementations (TestStore, FixtureCalendarClient)
- **Integration Tests**: Test real component interactions with test implementations
- **TUI Tests**: Use go-expect for terminal interaction testing
- **No External Calls**: Tests never hit real Keychain or Google APIs

## Common Tasks

### Adding a New Command
1. Create `cmd/command_name.go` with cobra command
2. Implement TUI flow using components from `internal/tui/`
3. Handle `--output json` flag if returning data
4. Add tests in `cmd/command_name_test.go`
5. Add integration tests in `test/integration/`

### Adding a New TUI Component
1. Create component in `internal/tui/component.go`
2. Implement bubbletea Model interface (Init, Update, View)
3. Use adaptive colors from `styles.go`
4. Add unit tests for state transitions
5. Add go-expect tests for user interaction

### Modifying an Interface
1. Update interface definition
2. Update production implementation
3. Update test implementation
4. Run tests to ensure compatibility
5. Update documentation

## Best Practices

- **Interfaces for Testability**: All external dependencies behind interfaces
- **Dependency Injection**: Pass dependencies explicitly
- **Table-Driven Tests**: Use for comprehensive test coverage
- **Adaptive Colors**: Use `lipgloss.AdaptiveColor` for theme compatibility
- **Error Handling**: Return errors, let caller decide how to handle
- **Context Propagation**: Pass context.Context for cancellation support

## Development Workflow

```bash
task dev      # Format, lint, test (before commit)
task build    # Build binary
task run -- setup  # Run command
```

## Documentation Standards

- **Godoc**: Every exported symbol must have documentation
- **Examples**: Add examples for complex functions
- **ADRs**: Create ADRs for significant decisions (see above)
- **Architecture**: Update `docs/ARCHITECTURE.md` for system changes
- **API Changes**: Update `docs/API.md` for interface changes

## Getting Help

- Read `docs/adrs/` for past architectural decisions
- Read `docs/ARCHITECTURE.md` for system design
- Look at existing tests for patterns
- Follow patterns in similar modules

## Architecture Decision Records (ADRs)

**IMPORTANT**: AI agents must follow these ADR guidelines.

### When Starting Work

Always load existing ADRs to understand past decisions:
```bash
ls docs/adrs/
```
Read relevant ADRs before making changes to related areas.

### When to Create ADRs

Create a new ADR when making decisions about:
- Storage mechanisms or data persistence
- Authentication/security approaches
- Testing strategies or frameworks
- External service integrations
- Significant architectural patterns
- Technology or library choices

### ADR Format

Create files in `docs/adrs/` with name `YYYYMMDD_short_name.md`:

```markdown
# ADR: [Title]

**Date:** YYYY-MM-DD
**Status:** Accepted | Superseded | Deprecated

## Context
What motivated this decision?

## Decision
What was decided?

## Consequences
**Positive:** Benefits
**Negative:** Trade-offs
```

After creating an ADR, update `docs/adrs/README.md` index.

## Critical Principles

1. **Security First**: All credentials in Keychain, never on disk
2. **User Experience**: Beautiful TUI, clear error messages
3. **Testability**: Interfaces for all external dependencies
4. **Maintainability**: Clear code > clever code
5. **AI-Friendly**: Clear structure, good documentation, consistent patterns
6. **Document Decisions**: Create ADRs for significant architectural choices

## Configuration

- **Location**: `~/.config/urgent/config.yaml` (XDG Base Directory spec)
- **Format**: YAML with calendar IDs and optional names
- **Purpose**: Store enabled calendar selections per account
- **Editing**: Use `urgent calendars` command (manual editing supported)
