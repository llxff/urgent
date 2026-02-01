# Contributing to urgent

Thank you for your interest in contributing! This guide will help you get started.

## Development Setup

### Prerequisites

- Go 1.25 or later
- golangci-lint
- Task (taskfile.dev)
- macOS (for Keychain integration)

### Initial Setup

```bash
# Clone the repository
git clone https://github.com/llxff/urgent.git
cd urgent

# Install dependencies
task install

# Install development tools
task setup

# Run tests
task test

# Build
task build
```

## Development Workflow

### Before Starting

1. Check existing issues or create a new one
2. Read `AGENTS.md` for project overview
3. Review `.cursor/rules.md` for coding standards
4. Look at similar code for patterns

### Making Changes

```bash
# Create a feature branch
git checkout -b feature/your-feature-name

# Make changes

# Run pre-commit checks
task dev

# Commit with conventional commits format
git commit -m "feat(module): add new feature"
```

### Commit Message Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `test`: Adding or updating tests
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `style`: Code style changes (formatting, missing semi-colons, etc.)
- `chore`: Changes to build process or auxiliary tools

**Examples:**
```
feat(auth): add automatic token refresh

fix(calendar): handle events without end time

docs(api): update calendar client documentation

test(tui): add go-expect tests for connect command
```

## Code Standards

### File Organization

- One primary type/interface per file
- Test files alongside implementation
- Test implementations in separate files (e.g., `teststore.go`)
- Keep files under 500 lines

### Naming Conventions

- Interfaces: `Store`, `Provider`, `Formatter`
- Implementations: `KeychainStore`, `GoogleCalendarClient`
- Test implementations: `TestStore`, `FixtureCalendarClient`

### Documentation

Every exported symbol must have a doc comment:

```go
// Store manages OAuth2 token persistence.
//
// All methods are safe for concurrent use.
type Store interface {
    // SaveToken stores an OAuth2 token for the given email address.
    SaveToken(email string, token *oauth2.Token) error
}
```

### Testing

- Write tests for all new code
- Target 95%+ coverage
- Use table-driven tests
- Provide test implementations, not mocks
- Use meaningful test names

```go
func TestStore_SaveToken(t *testing.T) {
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
```

## Pull Request Process

### Before Submitting

1. Run `task dev` and ensure all checks pass
2. Add tests for new functionality
3. Update documentation if needed
4. Verify test coverage remains above 95%

### PR Description

Include:
- What does this PR do?
- Why is this change needed?
- How has this been tested?
- Related issues (if any)

### Review Process

1. Automated checks must pass
2. Code review by maintainers
3. Address feedback
4. Squash commits if requested
5. Merge when approved

## Project Structure

```
cmd/              - Cobra commands
internal/
  auth/           - OAuth2 & Keychain
  calendar/       - Google Calendar API
  output/         - Formatters
  tui/            - Terminal UI components
  config/         - Configuration
test/
  integration/    - Integration tests
  fixtures/       - Test data
docs/             - Documentation
.cursor/          - AI assistant rules
```

## Common Tasks

### Adding a New Command

1. Create `cmd/newcommand.go`
2. Implement command logic
3. Add to root command in `init()`
4. Create tests in `cmd/newcommand_test.go`
5. Update documentation

### Adding a TUI Component

1. Create `internal/tui/component.go`
2. Implement bubbletea Model interface
3. Use adaptive colors from `styles.go`
4. Add unit tests
5. Add go-expect integration test

### Modifying an Interface

1. Update interface definition
2. Update all implementations
3. Update test implementations
4. Run all tests
5. Update API documentation

## Testing Guidelines

### Unit Tests

```bash
# Run all tests
task test

# Run specific package
go test ./internal/auth/...

# Run with coverage
go test -cover ./...

# Generate coverage report
task test
open coverage.html
```

### Integration Tests

```bash
# Run integration tests only
go test ./test/integration/...
```

### TUI Tests

TUI tests use go-expect to simulate terminal interaction:

```go
func TestConnectCommand(t *testing.T) {
    console, err := expect.NewConsole()
    if err != nil {
        t.Fatal(err)
    }
    defer console.Close()
    
    // Simulate user input
    console.SendLine("client-id")
    console.SendLine("client-secret")
    
    // Verify output
    _, err = console.ExpectString("Success!")
    if err != nil {
        t.Fatal(err)
    }
}
```

## Debugging

### Enable Verbose Logging

Add to command for debugging:
```go
log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
log.Printf("Debug: %v", value)
```

### Test Keychain Locally

```go
store := auth.NewKeychainStore()
token := &oauth2.Token{AccessToken: "test"}
err := store.SaveToken("test@example.com", token)
fmt.Printf("Error: %v\n", err)
```

### Debug TUI Rendering

Output View() to file:
```go
os.WriteFile("debug.txt", []byte(model.View()), 0644)
```

## Release Process

1. Update version in code
2. Update CHANGELOG.md
3. Create git tag
4. Push tag
5. CI builds and publishes release

## Getting Help

- Check documentation in `docs/`
- Look at existing code for patterns
- Ask in issues or discussions
- Review `.cursor/prompts.md` for common tasks

## Code of Conduct

- Be respectful and inclusive
- Provide constructive feedback
- Focus on what is best for the project
- Show empathy towards other contributors

## License

By contributing, you agree that your contributions will be licensed under the project's license.
