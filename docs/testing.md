<!--
SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors

SPDX-License-Identifier: CC0-1.0

-->

# Testing Guide

## Quick Start

### Running Tests

```bash
# Run all unit tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...

# Run specific test
go test -v ./path/to/package -run TestSpecificFunction

# Run integration tests
go test -tags=integration ./...

# Project-specific commands
just test        # Run all tests (recommended)
just verify      # Run full verification suite  
just quality     # Run quality checks
```

### Quick Reference

| Command | Purpose |
|---------|---------|
| `just test` | Run all tests |
| `just verify` | Full verification |
| `go test ./...` | Basic test run |
| `go test -race ./...` | Race detection |
| `go test -cover ./...` | Coverage report |

## Writing Tests

### Basic Test Pattern

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

### Test Requirements

- Use `testify/require` for critical assertions (stops test on failure)
- Use `testify/assert` for non-critical verifications
- Use table-driven tests for multiple scenarios
- Test both success and error cases
- Clean up resources in `defer` statements

### Using Temporary Directories

```go
func TestWithTmpDir(t *testing.T) {
    tmpDir := t.TempDir() // Automatically cleaned up

    // Use tmpDir for test operations
    testFile := filepath.Join(tmpDir, "test.txt")
    err := os.WriteFile(testFile, []byte("content"), 0644)
    require.NoError(t, err)
}
```

## Test Architecture

### 1. Unit Tests (`*_test.go`)
- **Purpose**: Fast, isolated testing of business logic
- **Dependencies**: Mocked using testify/mock
- **Isolation**: No external dependencies (filesystem, network, etc.)
- **Location**: Alongside source files
- **Run**: `go test ./...`

### 2. Integration Tests (`*_test.go` with `//go:build integration`)
- **Purpose**: Test component interactions with real implementations
- **Dependencies**: Real adapters with controlled environments
- **Isolation**: Temporary directories for filesystem operations
- **Location**: Alongside source files
- **Run**: `go test -tags=integration ./...`

### 3. Enhanced Tests (`*_enhanced_test.go`)
- **Purpose**: Comprehensive scenario testing with advanced edge cases
- **Dependencies**: Mix of real and mock implementations
- **Isolation**: Sophisticated temporary directory management
- **Location**: Domain and adapter layers
- **Run**: `go test -run Enhanced ./...`

### 4. End-to-End Tests (Future)
- **Purpose**: Full workflow testing
- **Dependencies**: External services (GitHub, GitLab) in test mode
- **Isolation**: Dedicated test organizations/namespaces

## Testing Principles

### 1. Testable Architecture
- **Dependency Injection**: All dependencies injected via interfaces
- **Pure Functions**: Domain logic is pure and easily testable
- **Minimal Side Effects**: Side effects isolated to adapters

### 2. Temporary Directory Management

```go
// Always use temporary directories for filesystem operations
tmpDir := createTestTmpDir(t, "test-prefix")
defer cleanupTestTmpDir(t, tmpDir)

// Use context with tmp directory
ctx := model.WithTmpDir(context.Background(), tmpDir)
```

### 3. Mock Strategy
- **Shared Mocks**: Common mocks in `mocks_test.go` to reduce duplication
- **Lenient Mocks**: Use `.Maybe()` for optional method calls
- **Focused Mocks**: Only mock what's needed for the specific test

### 4. Error Testing
- **Positive Cases**: Happy path scenarios
- **Negative Cases**: Error conditions and edge cases
- **Boundary Conditions**: Empty inputs, nil values, etc.

## Test Structure

### Unit Test Example

```go
func TestUseCase_Execute(t *testing.T) {
    tests := []struct {
        name           string
        setupMocks     func(*MockProvider, *MockLogger)
        input          Request
        expectedOutput Response
        expectedError  bool
    }{
        {
            name: "successful_operation",
            setupMocks: func(provider *MockProvider, logger *MockLogger) {
                provider.On("Method", mock.Anything).Return(result, nil)
                logger.On("Info", mock.Anything, mock.Anything, mock.Anything).Maybe()
            },
            input:          validRequest,
            expectedOutput: expectedResponse,
            expectedError:  false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup mocks
            mockProvider := &MockProvider{}
            mockLogger := &MockLogger{}
            tt.setupMocks(mockProvider, mockLogger)

            // Create use case
            useCase := NewUseCase(mockProvider, mockLogger)

            // Execute
            result, err := useCase.Execute(context.Background(), tt.input)

            // Verify
            if tt.expectedError {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
                assert.Equal(t, tt.expectedOutput, result)
            }

            // Verify mocks
            mockProvider.AssertExpectations(t)
            mockLogger.AssertExpectations(t)
        })
    }
}
```

### Integration Test Example

```go
//go:build integration

func TestAdapter_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    // Create temporary directory
    tmpDir := createAdapterTestTmpDir(t, "integration-test")
    defer cleanupAdapterTestTmpDir(t, tmpDir)

    // Test with real implementations
    adapter := NewRealAdapter()

    // Test actual operations
    result, err := adapter.Operation(ctx, tmpDir)

    // Verify results
    require.NoError(t, err)
    require.NotNil(t, result)
}
```

### Enhanced Test Example

```go
func TestEnhancedScenarios(t *testing.T) {
    tests := []struct {
        name          string
        setupTest     func(t *testing.T) (string, TestRequest)
        setupMocks    func(*MockProvider, *MockGitOps, *MockLogger)
        expectedError bool
        validateAfter func(t *testing.T, response TestResponse, tmpDir string)
    }{
        {
            name: "comprehensive_workflow_with_protection_cycle",
            setupTest: func(t *testing.T) (string, TestRequest) {
                tmpDir := createTestTmpDir(t, "comprehensive-test")
                // Complex test setup...
                return tmpDir, request
            },
            setupMocks: func(provider, gitOps, logger) {
                // Detailed mock expectations...
            },
            validateAfter: func(t *testing.T, response TestResponse, tmpDir string) {
                // Complex validation including filesystem state...
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tmpDir, request := tt.setupTest(t)
            defer cleanupTestTmpDir(t, tmpDir)

            // Execute enhanced test scenario...
        })
    }
}
```

## Test Categories

### 1. Domain Layer Tests
- **Location**: `internal/domain/*/`
- **Focus**: Business logic, use cases, validation
- **Dependencies**: Mocked ports
- **Examples**:
  - Repository filtering logic
  - Sync orchestration
  - Validation rules

### 2. Adapter Layer Tests
- **Location**: `internal/adapters/*/`
- **Focus**: External system integration
- **Dependencies**: Real implementations with controlled environments
- **Examples**:
  - Git operations with temporary repositories
  - Provider API calls with test endpoints
  - File system operations in temporary directories

### 3. Configuration Tests
- **Location**: `internal/configuration/`
- **Focus**: Configuration loading, validation, printing
- **Dependencies**: Temporary files for configuration
- **Examples**:
  - YAML/JSON parsing
  - Validation rules
  - Environment variable handling

## Test Data Management

### 1. Test Builders

```go
func createTestRepository(name string) entities.Repository {
    builder := entities.NewRepositoryBuilder()
    builder, _ = builder.WithName(name)
    builder, _ = builder.WithHTTPSURL("https://github.com/test/" + name + ".git")
    // ... configure other fields
    repo, _ := builder.Build()
    return repo
}
```

### 2. Test Fixtures
- Store common test data in `testdata/` directories
- Use golden files for expected outputs
- Version control test configurations

### 3. Environment Isolation
- Each test gets its own temporary directory
- Tests must not depend on external state
- Clean up all resources in defer statements

## Enhanced Testing Features

### 1. Temporary Directory Management

Always use `t.TempDir()` for automatic cleanup:

```go
func TestWithTmpDir(t *testing.T) {
    tmpDir := t.TempDir() // Automatically cleaned up
    // Use tmpDir for test operations
}
```

### 2. Mock Strategy
- Use testify/mock for dependencies
- Mock at interface boundaries (ports)
- Keep mocks simple and focused

### 3. Test Scenarios
- Protection cycles and force push operations
- Dry run validation and error recovery
- Concurrent operations and large repositories

## Performance Testing

### 1. Benchmarks

```go
func BenchmarkOperation(b *testing.B) {
    setup := createBenchmarkSetup()
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        _ = operation(setup.input)
    }
}
```

### 2. Memory Profiling
- Monitor memory usage in large operations
- Check for memory leaks in long-running processes
- Optimize based on profiling results

### 3. Concurrency Testing
- Test parallel execution safety
- Validate race condition prevention
- Ensure proper resource cleanup

## Test Execution

### Local Development

```bash
# Basic test commands
go test ./...                    # Unit tests
go test -tags=integration ./...  # Integration tests
go test -race ./...              # Race detection
go test -cover ./...             # Coverage report

# Project-specific
just test                        # Run all tests
just verify                      # Full verification
```

### Continuous Integration
- Unit tests run on every commit  
- Integration tests run on pull requests
- Coverage tracking and performance regression detection

## Best Practices

### Key Guidelines
- **Organization**: One test file per source file, descriptive names
- **Assertions**: Use `require` for critical checks, `assert` for verifications  
- **Isolation**: Independent tests, no shared state, proper cleanup
- **Mocking**: Mock at interface boundaries (ports), keep mocks simple
- **Coverage**: Test success/failure scenarios, edge cases, and error conditions

### Tools & Libraries
- **Framework**: Go `testing` package with `testify` for assertions/mocks
- **Coverage**: Built-in Go coverage tools
- **Commands**: `just test` for project testing, `go test -race` for race detection

## Troubleshooting

### Common Issues
- **Flaky Tests**: Usually timing issues or shared state
- **Slow Tests**: Use mocks instead of real implementations  
- **Cleanup Failures**: Always use `defer` and `t.TempDir()`

### Debugging Tips  
- Use `t.Logf()` for debug output and `go test -v` for verbose mode
- Run `go test -race` to detect race conditions
- Check temporary directory cleanup in CI environments

This testing strategy ensures reliable, maintainable test coverage for the hexagonal architecture implementation.
