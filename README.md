# urgent - Google Calendar CLI for Automation

**Get your Google Calendar events from the command line.** Use them in scripts, status bars, notifications, or anywhere you need calendar data.

## Why urgent?

Your calendar data is locked in Google Calendar's web UI. **urgent** frees it:

- **Build custom notifications** - Get alerts 10 minutes before meetings, not Google's default
- **Status bar integration** - Show your next meeting in Polybar, i3status, or tmux
- **Script your workflow** - Block focus time when you have meetings, auto-set Slack status
- **Aggregate multiple accounts** - See work + personal calendars in one place
- **Pipe to anything** - JSON output works with jq, scripts, webhooks, whatever

## Quick Example

```bash
# What's my next meeting?
$ urgent next --within 60 -o json | jq '.event.summary'
"Team Standup"

# Am I free for the next hour?
$ urgent next --within 60 && echo "Meeting soon!" || echo "You're free"

# Today's schedule
$ urgent today -o json | jq -r '.events[] | "\(.start[11:16]) \(.summary)"'
09:00 Team Standup
11:00 1:1 with Manager
14:00 Sprint Planning
```

## Features

- **JSON output** for scripting and automation
- **Multiple Google accounts** in one place
- **Calendar filtering** - enable only the calendars you care about
- **Secure storage** - credentials in macOS Keychain, never in files
- **Fast** - parallel fetching, automatic token refresh

## Installation

```bash
# From source
git clone https://github.com/llxff/urgent.git
cd urgent && task build
sudo mv bin/urgent /usr/local/bin/

# Or with Go
go install github.com/llxff/urgent@latest
```

## Setup (One-time)

### 1. Create Google OAuth Credentials

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a project and enable Google Calendar API
3. Create OAuth 2.0 credentials (Desktop app)
4. Add redirect URI: `http://localhost`

### 2. Store Credentials

```bash
urgent setup
# Enter Client ID and Client Secret when prompted
```

### 3. Connect Your Account

```bash
urgent connect
# Browser opens for authorization, then select calendars
```

## Commands

| Command | Description |
|---------|-------------|
| `urgent today` | Today's events |
| `urgent today --remaining` | Only future events today |
| `urgent today -o json` | JSON output |
| `urgent next --within 30` | Next event within 30 minutes |
| `urgent calendars` | Manage which calendars to show |
| `urgent connect` | Add a Google account |
| `urgent disconnect` | Remove an account |

## Automation Examples

### Desktop Notifications (cron)

```bash
# Every 5 min, notify if meeting in 10 min
*/5 * * * * urgent next --within 10 -o json | jq -r 'select(.hasEvent) | "Meeting: \(.event.summary) in \(.event.minutesUntil)min"' | xargs -I {} terminal-notifier -message "{}"
```

### Status Bar (Polybar)

```ini
[module/calendar]
type = custom/script
exec = urgent next --within 60 -o json | jq -r 'if .hasEvent then "[\(.event.minutesUntil)m] \(.event.summary)" else "" end'
interval = 60
```

### Slack Status Script

```bash
#!/bin/bash
EVENT=$(urgent next --within 5 -o json)
if echo "$EVENT" | jq -e '.hasEvent' > /dev/null; then
  SUMMARY=$(echo "$EVENT" | jq -r '.event.summary')
  curl -X POST "https://slack.com/api/users.profile.set" \
    -H "Authorization: Bearer $SLACK_TOKEN" \
    -d "{\"profile\":{\"status_text\":\"In: $SUMMARY\",\"status_emoji\":\":calendar:\"}}"
fi
```

### tmux Status Line

```bash
# In .tmux.conf
set -g status-right '#(urgent next --within 30 -o json | jq -r "if .hasEvent then .event.summary else \"\" end")'
```

### Block Distractions During Meetings

```bash
#!/bin/bash
# Run every minute via cron
if urgent next --within 0 -o json | jq -e '.hasEvent' > /dev/null; then
  # Currently in a meeting - block distracting sites
  sudo cp /etc/hosts.blocked /etc/hosts
else
  sudo cp /etc/hosts.normal /etc/hosts
fi
```

## JSON Output Format

### `urgent today -o json`

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
      "colorHex": "4285f4"
    }
  ],
  "count": 1,
  "filter": "all"
}
```

### `urgent next --within 30 -o json`

```json
{
  "hasEvent": true,
  "minutesUntil": 15,
  "event": {
    "summary": "1:1 with Manager",
    "start": "2025-12-25T14:00:00-08:00",
    "end": "2025-12-25T14:30:00-08:00",
    "calendar": "Work"
  }
}
```

Exit code: `0` if event found, `1` if no event (useful for conditionals).

## Configuration

Calendar selections stored in `~/.config/urgent/config.yaml`:

```yaml
accounts:
  user@example.com:
    enabled_calendars:
      - id: primary
        name: Main Calendar
      - id: work@group.calendar.google.com
        name: Work
```

Credentials stored in macOS Keychain (never in files).

## Development

```bash
task dev      # Lint + test (before commit)
task build    # Build binary
task test     # Run tests with coverage
```

See [AGENTS.md](AGENTS.md) for architecture and development guide.

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Keychain access denied | Grant Terminal access in System Preferences > Privacy |
| OAuth client not found | Run `urgent setup` |
| No accounts connected | Run `urgent connect` |
| Token expired | Tokens auto-refresh; if failing, `urgent disconnect` then `urgent connect` |

## Links

- [AGENTS.md](AGENTS.md) - Development guide
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) - Technical architecture
- [GitHub Issues](https://github.com/llxff/urgent/issues) - Bug reports
