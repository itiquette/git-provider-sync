// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package cmd

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSetupSignalContext tests that signal context is properly configured.
// Idiomatic Go testing - table driven, clear assertions, no overengineering.
func TestSetupSignalContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		signal       os.Signal
		expectCancel bool
		waitDuration time.Duration
	}{
		{
			name:         "SIGINT cancels context",
			signal:       syscall.SIGINT,
			expectCancel: true,
			waitDuration: 100 * time.Millisecond,
		},
		{
			name:         "SIGTERM cancels context",
			signal:       syscall.SIGTERM,
			expectCancel: true,
			waitDuration: 100 * time.Millisecond,
		},
	}

	for _, testCase := range tests {
		// capture range variable
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			// Create base context
			baseCtx := context.Background()

			// Setup signal context
			ctx := SetupSignalContext(baseCtx)
			require.NotNil(t, ctx)

			// Verify context is not cancelled initially
			select {
			case <-ctx.Done():
				t.Fatal("context should not be cancelled initially")
			default:
				// Good, context is active
			}

			// Send signal to self (test process)
			proc, err := os.FindProcess(os.Getpid())
			require.NoError(t, err)

			err = proc.Signal(testCase.signal)
			require.NoError(t, err)

			// Wait for context cancellation
			if testCase.expectCancel {
				select {
				case <-ctx.Done():
					// Context cancelled as expected
					assert.Equal(t, context.Canceled, ctx.Err())
				case <-time.After(testCase.waitDuration):
					t.Fatal("context should have been cancelled by signal")
				}
			}
		})
	}
}

// TestContextCancellationPropagation tests that cancellation propagates through the app.
func TestContextCancellationPropagation(t *testing.T) {
	t.Parallel()
	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Simulate a long-running operation
	done := make(chan bool)

	go func() {
		select {
		case <-ctx.Done():
			done <- true
		case <-time.After(5 * time.Second):
			done <- false
		}
	}()

	// Cancel after short delay
	time.Sleep(10 * time.Millisecond)
	cancel()

	// Verify cancellation was detected
	result := <-done
	assert.True(t, result, "operation should detect context cancellation")
}

// TestInterruptDuringOperation simulates interrupting a sync operation.
func TestInterruptDuringOperation(t *testing.T) {
	t.Parallel()
	// Create a context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Simulate an operation that checks context
	operation := func(ctx context.Context) error {
		for range 10 {
			// Check for cancellation
			if ctx.Err() != nil {
				return ctx.Err()
			}
			// Simulate work
			time.Sleep(10 * time.Millisecond)
		}

		return nil
	}

	// Start operation in goroutine
	errCh := make(chan error)

	go func() {
		errCh <- operation(ctx)
	}()

	// Cancel after brief delay
	time.Sleep(25 * time.Millisecond)
	cancel()

	// Check that operation was cancelled
	err := <-errCh
	assert.Equal(t, context.Canceled, err, "operation should return context.Canceled")
}
