# Common AI Prompts for urgent

This file contains frequently used prompts for AI assistance when working on the urgent codebase.

## Adding Features

### New Command
```
Add a new command called [command-name] that [description].
The command should:
- Use TUI with [specific components]
- Support --output json flag [if applicable]
- Follow the patterns in existing commands
- Include comprehensive tests
```

### New TUI Component
```
Create a new TUI component for [purpose] in internal/tui/[name].go.
Requirements:
- Implement bubbletea Model interface
- Use adaptive colors from styles.go
- Handle keyboard navigation
- Include state transitions tests
- Add go-expect integration test
```

## Debugging

### Test Failures
```
Tests in [package] are failing with error: [error message]
Please:
1. Analyze the failure
2. Identify the root cause
3. Fix the issue
4. Ensure all tests pass
5. Verify coverage remains above 95%
```

### TUI Issues
```
The TUI in [command] has [problem description].
Expected behavior: [description]
Actual behavior: [description]
Please debug and fix while maintaining the adaptive color system.
```

## Refactoring

### Extract Interface
```
Extract an interface from [concrete type] in [package].
Requirements:
- Follow naming conventions (Store, Provider, etc.)
- Create test implementation
- Update all usage sites
- Ensure all tests still pass
```

### Improve Test Coverage
```
Improve test coverage in [package].
Current coverage: [X]%
Target coverage: 95%+
Focus on:
- Edge cases
- Error scenarios
- Boundary conditions
```

## Code Review

### Pre-commit Check
```
Review my changes before commit:
1. Check Godoc comments on all exported symbols
2. Verify test coverage
3. Check for hardcoded values (colors, paths, credentials)
4. Ensure errors are properly wrapped
5. Verify golangci-lint would pass
```

### Performance Review
```
Review [package/function] for performance issues.
Consider:
- Context cancellation
- Resource leaks
- Memory allocations
- Concurrent access patterns
```

## Documentation

### Update Architecture
```
I've made changes to [component].
Please update docs/ARCHITECTURE.md to reflect:
- New design decisions
- Interface changes
- Data flow changes
```

### Generate API Docs
```
Generate/update API documentation in docs/API.md for:
- [Package name]
Include all exported functions, types, and interfaces with examples.
```

## Testing

### Add Integration Test
```
Create an integration test for the workflow:
[describe user workflow]
The test should:
- Use test implementations (TestStore, FixtureCalendarClient)
- Cover happy path and error cases
- Verify all interactions
```

### TUI Testing
```
Add go-expect test for [command] TUI flow.
Test scenarios:
1. [scenario 1]
2. [scenario 2]
3. [error scenario]
Verify all screen transitions and user interactions.
```

## Troubleshooting

### Keychain Issues
```
The Keychain integration is failing with: [error]
Context: [MacOS version, command being run, etc.]
Please help diagnose and fix.
```

### OAuth Flow Problems
```
OAuth flow is failing at [stage] with error: [error]
Details:
- Browser opens: [yes/no]
- Redirect received: [yes/no]
- Error message: [message]
Please debug and fix.
```

## Best Practices

### Code Style Audit
```
Audit [file/package] for adherence to project standards:
- Naming conventions
- Error handling patterns
- Interface usage
- Comment quality
- Import grouping
Suggest improvements.
```

### Security Review
```
Review [component] for security issues:
- Credential handling
- Sensitive data logging
- CSRF protection
- Input validation
Suggest fixes for any issues found.
```

## Examples

### Usage Example Request
```
Create a usage example for [function/command] showing:
- Basic usage
- Common options
- Error handling
- Expected output
Add to relevant documentation file.
```

### README Update
```
Update README.md to include:
- [new feature]
- Installation instructions for [platform]
- Updated examples
Keep existing content and formatting style.
```
