# ADR: XDG Config for Preferences, Keychain for Credentials

**Date:** 2024-12-25
**Status:** Accepted

## Context

The application needs to store two types of data:
1. User preferences (enabled calendars per account)
2. Sensitive credentials (OAuth tokens, client ID/secret)

Options considered:
1. All in Keychain
2. All in config file
3. Split: config file for preferences, Keychain for credentials

## Decision

Use XDG Base Directory specification for non-sensitive configuration:
- **Location**: `~/.config/urgent/config.yaml`
- **Contents**: Enabled calendar IDs per account with optional names

Use macOS Keychain for all credentials (see [Keychain ADR](./20241225_keychain_storage.md)).

## Consequences

**Positive:**
- Human-readable config file for non-sensitive data
- Easy to backup/version calendar preferences
- Users can manually edit config if needed
- Credentials remain secure in Keychain
- Follows Unix conventions (XDG spec)

**Negative:**
- Two storage locations to manage
- Config file could be accidentally deleted
- Need to handle missing config gracefully
