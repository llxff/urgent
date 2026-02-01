# Architecture Decision Records

This folder contains ADRs documenting important technical and architectural decisions.

## Format

Files are named: `YYYYMMDD_short_name.md`

Each ADR follows this structure:

```markdown
# ADR: [Title]

**Date:** YYYY-MM-DD
**Status:** Accepted | Superseded | Deprecated

## Context

What is the issue or question that motivated this decision?

## Decision

What is the decision that was made?

## Consequences

What are the results of this decision? (positive and negative)
```

## Index

| Date | Decision | Status |
|------|----------|--------|
| 2024-12-25 | [Keychain for credential storage](./20241225_keychain_storage.md) | Accepted |
| 2024-12-25 | [Dynamic port for OAuth server](./20241225_dynamic_oauth_port.md) | Accepted |
| 2024-12-25 | [Test implementations over mocks](./20241225_test_implementations.md) | Accepted |
| 2024-12-25 | [XDG config with Keychain credentials](./20241225_config_strategy.md) | Accepted |

## AI Instructions

When making significant decisions during development:
1. Create a new ADR in this folder
2. Use today's date and a descriptive name
3. Document context, decision, and consequences
4. Update the index above
