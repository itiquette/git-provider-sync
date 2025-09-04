<!--
SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors

SPDX-License-Identifier: CC0-1.0

-->

# Testing Guide

## Quick Start

```bash
just test        # Run all tests
just verify      # Full verification suite
just quality     # Quality checks

# Specific test scenarios
go test ./...                    # Unit tests
go test -tags=integration ./...  # Integration tests
go test -race ./...              # Race detection
go test -cover ./...             # Coverage report
go test -run TestName ./path     # Single test
```

## Writing Tests

### Test Pattern

```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    Input
        expected Output
        wantErr  bool
    }{
        {
            name:     "success_case",
            input:    validInput,
            expected: expectedOutput,
            wantErr:  false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Function(tt.input)

            if tt.wantErr {
                require.Error(t, err)
                return
            }

            require.NoError(t, err)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### Test Types

| Type | Tag | Purpose |
|------|-----|---------|
| Unit | none | Business logic, mocked dependencies |
| Integration | `integration` | Component interactions |
| Enhanced | none | Edge cases and scenarios |

### Key Patterns

## Temporary Directories

```go
tmpDir := t.TempDir() // Auto-cleaned
```

## Context with Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
```

## Mocking

```go
mockProvider := new(mocks.MockRepositoryProvider)
mockProvider.On("ListRepositories", mock.Anything).Return(repos, nil)
```

## Test Utilities

### Git Test Environment

```go
import "itiquette/git-provider-sync/internal/integrationtest/testutil"

env, err := testutil.SetupGitTestEnvironment(t, gitOps, opts)
// Creates source/target repos with test content
```

### Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Coverage requirements
# - Domain logic: >80%
# - Use cases: All paths covered
# - Adapters: Integration tested
```

## Debugging Tests

```bash
# Verbose output
go test -v ./path/to/package

# Debug specific test
dlv test ./path/to/package -- -test.run TestName

# Keep temp directories for inspection
KEEP_TEMP=1 go test ./...
```

## CI/CD Integration

```yaml
# GitHub Actions example
- name: Test
  run: |
    just test
    just verify
```

## Best Practices

1. **Use table-driven tests** for multiple scenarios
2. **Test error paths** not just success cases
3. **Mock external dependencies** in unit tests
4. **Use t.Parallel()** for independent tests
5. **Clean up resources** with defer or t.Cleanup()
6. **Name tests clearly**: `Test<Function>_<Scenario>`
7. **Keep tests focused** on single behaviors
8. **Use testify** for assertions (require/assert)

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Flaky tests | Check for race conditions, use `-race` |
| Slow tests | Use t.Parallel(), mock heavy operations |
| Can't find tests | Check build tags, use `-tags=integration` |
| Cleanup issues | Use t.TempDir() instead of manual cleanup |
