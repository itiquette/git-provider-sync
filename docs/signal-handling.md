<!--
SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors

SPDX-License-Identifier: CC0-1.0

-->

# Signal Handling and User Interruption

Git Provider Sync provides graceful handling of user interruption (Ctrl-C) following idiomatic Go patterns.

## User Escape Mechanisms

### Ctrl-C (SIGINT) Support

The application handles Ctrl-C gracefully at all times:

- **Immediate Response**: Ctrl-C is always responsive, even during network I/O
- **Clean Shutdown**: Operations are cancelled cleanly without corruption
- **Automatic Cleanup**: Temporary directories are removed on interruption
- **Clear Feedback**: Users see confirmation when operations are cancelled

### How It Works

1. **Signal Context**: The application creates a signal-aware context at startup that listens for SIGINT (Ctrl-C) and SIGTERM signals
2. **Context Propagation**: This context flows through all operations via the hexagonal architecture
3. **Cancellation Checks**: Long-running operations check for context cancellation before expensive steps
4. **Cleanup Handlers**: Deferred cleanup ensures temporary resources are freed

### Implementation Details

Following the codebase principles: **Be hexagonal, functional, idiomatic, and don't overengineer**

#### Entry Point Signal Handling

```go
// cmd/root.go - Simple, idiomatic signal handling
ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
```

#### Domain Layer Checks

```go
// Check for cancellation before expensive operations
if ctx.Err() != nil {
    return fmt.Errorf("operation cancelled: %w", ctx.Err())
}
```

#### Automatic Cleanup

```go
// Temporary directories cleaned up on cancellation
defer func() {
    if cleanupErr := entities.DeleteTmpDir(ctx); cleanupErr != nil {
        logger.Warn().Err(cleanupErr).Msg("Failed to cleanup")
    }
}()
```

## User Experience

### During Operations

When running long operations, users see:

```text
⚡ Cloning repository example-repo (Press Ctrl-C to cancel)
[45%] Processing files...
```

### On Cancellation

When pressing Ctrl-C:

```text
✗ Repository sync cancelled by user (Ctrl-C)
  Cleaning up temporary files...
```

## Architecture

The signal handling follows hexagonal architecture:

- **Domain Layer**: Defines `InterruptHandler` port interface
- **Adapter Layer**: Implements user feedback in `cli.InterruptHandler`
- **Use Cases**: Check context cancellation at strategic points
- **No Global State**: Context flows explicitly through function calls

## Testing

Signal handling is tested at multiple levels:

- **Unit Tests**: Test context cancellation propagation
- **Integration Tests**: Verify cleanup on interruption
- **No Race Conditions**: Thread-safe interrupt handling

## Comparison to Other Tools

Unlike tools that trap users (like vim's modal interface), git-provider-sync:

- **Always Responsive**: Ctrl-C works immediately
- **Clear Escape**: No special commands or modes to learn
- **Familiar Behavior**: Acts like standard Unix tools
- **No Hanging**: Network timeouts are bounded and interruptible

## Best Practices

1. **Check Context Early**: Before expensive operations
2. **Defer Cleanup**: Use defer for resource cleanup
3. **Provide Feedback**: Show users that Ctrl-C is available
4. **Graceful Degradation**: Complete current operation unit if safe

## Example Usage

```bash
# Start a sync operation
$ gitprovidersync sync --config config.yaml

⚡ Fetching repository list (Press Ctrl-C to cancel)
⚡ Cloning repository project-1 (Press Ctrl-C to cancel)
[25%] Cloning repository project-2...

# Press Ctrl-C at any time
^C
✗ Repository sync cancelled by user (Ctrl-C)
  Cleaning up temporary files...

# Exit cleanly with status code
$ echo $?
130
```

## Implementation Status

✅ Signal context at entry point
✅ Context propagation through commands  
✅ Cleanup handlers for temp directories
✅ Progress feedback for long operations
✅ Cancellation checks in domain layer
✅ Thread-safe interrupt handling
✅ Comprehensive test coverage
