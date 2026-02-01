# ADR: Use Test Implementations Instead of Mock Frameworks

**Date:** 2024-12-25
**Status:** Accepted

## Context

Testing code that depends on external services (Keychain, Google API) requires some form of test doubles.

Options considered:
1. Mock frameworks (gomock, testify/mock)
2. Manual mocks per test
3. Real in-memory test implementations

## Decision

Create real test implementations that satisfy interfaces:
- `TestStore` - In-memory credential storage
- `FixtureCalendarClient` - Returns pre-loaded calendar data

These live alongside production code (e.g., `teststore.go`).

## Consequences

**Positive:**
- Test implementations are real, working code
- No mock framework dependencies
- Easier to understand and debug
- Forces good interface design
- Can be reused across many tests
- Faster than mock setup/verification

**Negative:**
- Must maintain test implementations
- More upfront code to write
- Test implementations could diverge from production behavior
- Less granular verification (can't assert "method called with X")
