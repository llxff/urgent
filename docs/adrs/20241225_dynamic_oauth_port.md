# ADR: Use Dynamic Port for OAuth Callback Server

**Date:** 2024-12-25
**Status:** Accepted

## Context

OAuth2 flow requires a local HTTP server to receive the callback with the authorization code. The server needs to listen on a port.

Options considered:
1. Fixed port (e.g., 8080)
2. Configurable port
3. Dynamic port (`:0`)

## Decision

Use `localhost:0` which lets the OS assign any available port. The redirect URL is constructed dynamically as `http://localhost:{PORT}/callback`.

Google OAuth supports loopback addresses with any port for desktop apps.

## Consequences

**Positive:**
- No port conflicts with other applications
- No configuration needed
- Works on any machine without setup
- Google explicitly supports this pattern

**Negative:**
- Redirect URL changes each time (not an issue for loopback)
- Slightly more complex code to extract assigned port
- Cannot bookmark the callback URL (not relevant for this use case)
