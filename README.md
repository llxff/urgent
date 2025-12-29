# urgent - Beautiful Google Calendar CLI

A powerful command-line interface for Google Calendar with a beautiful terminal UI, multiple account support, and automation-friendly JSON output.

## Features

✨ **Beautiful Terminal UI** - Adaptive colors, smooth animations, intuitive navigation  
🔐 **Secure** - All credentials stored in macOS Keychain  
👥 **Multiple Accounts** - Manage multiple Google accounts simultaneously  
🤖 **Automation-Friendly** - JSON output for scripts and integrations  
⚡ **Fast** - Parallel event fetching, automatic token refresh  
🎨 **Color-Coded** - Events styled by calendar colors  

## Installation

### From Source

```bash
git clone <repository-url>
cd urgent
task build
sudo mv bin/urgent /usr/local/bin/
```

### Using Go

```bash
go install github.com/yourusername/urgent@latest
```

## Quick Start

### 1. Setup OAuth Credentials

First, create OAuth credentials in Google Cloud Console:

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing
3. Enable Google Calendar API
4. Create OAuth 2.0 credentials (Desktop app type)
5. Configure authorized redirect URIs: `http://localhost`

Then run setup:

```bash
urgent setup
```

Enter your Client ID and Client Secret when prompted. They'll be stored securely in Keychain.

### 2. Connect Your Google Account

```bash
urgent connect
```

A browser window will open for you to authorize the app. Once authorized, you'll be prompted to select which calendars to enable. Your selections are stored securely.

### 3. Manage Calendar Selection

```bash
# Select calendars for an account
urgent calendars

# Select calendars for specific account
urgent calendars --account user@example.com

# List all calendars and their enabled/disabled status
urgent calendars --list
```

### 4. View Today's Events

```bash
# Show all events for today
urgent today

# Show only remaining events
urgent today --remaining

# Get JSON output for scripting
urgent today -o json
```

### 5. Check Next Event

```bash
# Show next event if it's within 30 minutes
urgent next --within 30

# JSON output
urgent next --within 30 -o json
```

### 6. Disconnect an Account

```bash
urgent disconnect
```

## Commands

### `urgent setup`

Store OAuth client credentials in Keychain (one-time setup).

```bash
urgent setup
```

### `urgent connect`

Connect a Google account with beautiful TUI flow. After OAuth authorization, you'll select which calendars to enable.

```bash
urgent connect
```

### `urgent calendars`

Manage calendar selection for your accounts.

```bash
# Select calendars (interactive)
urgent calendars

# Select calendars for specific account
urgent calendars --account user@example.com

# List all calendars and their enabled/disabled status
urgent calendars --list
```

**List Output:**
```
user@example.com:
  ✓ primary (user@example.com)
  ✓ work@group.calendar.google.com (Work Calendar)
    personal@example.com (Personal) [disabled]
```

You can change calendar selection at any time. Disabled calendars won't show events in `urgent today` or `urgent next`.

### `urgent today`

Show today's calendar events from enabled calendars.

```bash
urgent today                    # All events (TUI table)
urgent today --remaining        # Only remaining events
urgent today -o json            # JSON output
urgent today --remaining -o json
```

**JSON Output:**
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

### `urgent next`

Show the next event if it's within N minutes.

```bash
urgent next --within 30         # Check next event within 30 min
urgent next --within 10 -o json # JSON output
```

**JSON Output:**
```json
{
  "hasEvent": true,
  "event": {
    "summary": "Meeting",
    "start": "2025-12-25T15:00:00-08:00",
    "end": "2025-12-25T16:00:00-08:00",
    "minutesUntil": 15,
    "calendar": "Work",
    "account": "user@gmail.com"
  }
}
```

Exit code 0 if event found, 1 if no event.

### `urgent disconnect`

Disconnect a Google account.

```bash
urgent disconnect                         # Interactive selection
urgent disconnect --account user@gmail.com # Direct disconnect
```

## Automation Examples

### Cron Job for Notifications

Check for upcoming events every 5 minutes:

```bash
*/5 * * * * /usr/local/bin/urgent next --within 10 -o json | jq -r 'select(.hasEvent) | "Event: \(.event.summary) in \(.event.minutesUntil) minutes"' | terminal-notifier
```

### Status Bar Integration (Polybar)

```ini
[module/urgent]
type = custom/script
exec = urgent next --within 60 -o json | jq -r 'if .hasEvent then "\(.event.summary) (\(.event.minutesUntil)min)" else "" end'
interval = 60
```

### Alfred Workflow

```bash
#!/bin/bash
# Show today's events in Alfred
urgent today -o json | jq -r '.events[] | "\(.start) - \(.summary)"'
```

## Configuration

### Keychain Storage

All sensitive data is stored in macOS Keychain:
- **OAuth client credentials**: Service `com.urgent.cli.oauth`
- **User tokens**: Service `com.urgent.cli`, account `{email}`

### Calendar Preferences

Calendar selections are stored in:
- **Location**: `~/.config/urgent/config.yaml`
- **Format**: YAML with calendar IDs and names

**Example config:**
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

The config file is human-readable and can be edited manually, though using `urgent calendars` is recommended.

### XDG Base Directory Support

Config location follows XDG Base Directory specification:
1. `$XDG_CONFIG_HOME/urgent/config.yaml`
2. `~/.config/urgent/config.yaml` (default)

To use a custom location:
```bash
export XDG_CONFIG_HOME=/custom/path
```

## Development

### Prerequisites

- Go 1.25+
- Task (taskfile.dev)
- golangci-lint
- macOS

### Setup

```bash
task install    # Install dependencies
task setup      # Install development tools
```

### Common Tasks

```bash
task            # List all tasks
task build      # Build binary
task test       # Run tests with coverage
task lint       # Run linter with auto-fix
task dev        # Pre-commit check (lint + test)
task run -- today  # Run command
```

### Running Tests

```bash
task test               # All tests with coverage
go test ./internal/...  # Specific package
go test -v -run TestName # Specific test
```

### Project Structure

```
cmd/              - Cobra commands
internal/
  auth/           - OAuth2 & Keychain storage
  calendar/       - Google Calendar API client
  output/         - JSON & table formatters
  tui/            - Terminal UI components
  config/         - Configuration
test/
  integration/    - Integration tests
  fixtures/       - Test data
docs/             - Documentation
```

## Documentation

- **[AGENTS.md](AGENTS.md)** - AI agent guide (start here)
- **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)** - System architecture
- **[docs/API.md](docs/API.md)** - Internal API documentation
- **[docs/CONTRIBUTING.md](docs/CONTRIBUTING.md)** - Development guidelines
- **[.cursor/rules.md](.cursor/rules.md)** - Coding standards

## Troubleshooting

### "Keychain access denied"

Grant Terminal/iTerm2 access to Keychain in System Preferences → Privacy & Security.

### "OAuth client not found"

Run `urgent setup` to store OAuth credentials.

### "No accounts connected"

Run `urgent connect` to authorize a Google account.

### "Token expired"

Tokens are automatically refreshed. If this fails, disconnect and reconnect:

```bash
urgent disconnect --account user@gmail.com
urgent connect
```

## Security

- All credentials stored in macOS Keychain
- OAuth2 with automatic token refresh
- State parameter prevents CSRF attacks
- No credentials in logs or files
- Local OAuth server on random port

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](docs/CONTRIBUTING.md) for guidelines.

1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Run `task dev` to verify
5. Submit a pull request

## License

[Your License Here]

## Acknowledgments

- [Charm](https://charm.sh/) - Beautiful TUI libraries
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Google Calendar API](https://developers.google.com/calendar) - Calendar integration

## Support

- 🐛 **Bug Reports**: [GitHub Issues](your-repo-url/issues)
- 💡 **Feature Requests**: [GitHub Discussions](your-repo-url/discussions)
- 📖 **Documentation**: [docs/](docs/)

---

Made with ❤️ and ☕ by [Your Name]
