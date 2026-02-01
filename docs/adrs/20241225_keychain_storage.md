# ADR: Use macOS Keychain for All Credential Storage

**Date:** 2024-12-25
**Status:** Accepted

## Context

The application needs to store sensitive OAuth credentials:
- OAuth client ID and secret (app credentials)
- User access tokens and refresh tokens (per-account)

Options considered:
1. Config file (plain text or encrypted)
2. Environment variables
3. macOS Keychain
4. Separate secrets manager

## Decision

Store all credentials in macOS Keychain using `go-keyring`:
- **OAuth client credentials**: Service `com.urgent.cli.oauth`
- **User tokens**: Service `com.urgent.cli`, account `{email}`

## Consequences

**Positive:**
- Credentials never written to disk in plain text
- Cannot accidentally commit credentials to git
- System-level security (Touch ID, password protection)
- Consistent with macOS security best practices
- Simple API via go-keyring

**Negative:**
- macOS only (no Linux/Windows support currently)
- Requires Keychain access permission grant
- Testing requires mock Store interface
- Users must re-enter credentials if Keychain is reset
