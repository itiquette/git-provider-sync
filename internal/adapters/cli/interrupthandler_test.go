// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestInterruptHandler_ShowInterruptible tests showing interruptible operations.
func TestInterruptHandler_ShowInterruptible(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		operation      string
		contextDone    bool
		expectedOutput string
	}{
		{
			name:           "shows message for active context",
			operation:      "Cloning repository",
			contextDone:    false,
			expectedOutput: "→ Cloning repository (Press Ctrl-C to cancel)\n",
		},
		{
			name:           "no message for cancelled context",
			operation:      "Cloning repository",
			contextDone:    true,
			expectedOutput: "",
		},
	}

	for _, testCase := range tests {
		// Capture range variable
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			handler := NewInterruptHandler(&buf)

			ctx := context.Background()
			if testCase.contextDone {
				var cancel context.CancelFunc

				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			handler.ShowInterruptible(ctx, testCase.operation)
			assert.Equal(t, testCase.expectedOutput, buf.String())
		})
	}
}

// TestInterruptHandler_ShowProgress tests progress display.
func TestInterruptHandler_ShowProgress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		current        int
		total          int
		description    string
		expectedPrefix string
	}{
		{
			name:           "shows percentage progress",
			current:        5,
			total:          10,
			description:    "Processing files",
			expectedPrefix: "\r[ 50%] Processing files",
		},
		{
			name:           "shows indeterminate progress",
			current:        5,
			total:          0,
			description:    "Fetching data",
			expectedPrefix: "\r[...] Fetching data",
		},
	}

	for _, testCase := range tests {
		// Capture range variable
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			handler := NewInterruptHandler(&buf)

			handler.ShowProgress(context.Background(), testCase.current, testCase.total, testCase.description)

			output := buf.String()
			assert.True(t, strings.HasPrefix(output, testCase.expectedPrefix))
		})
	}
}

// TestInterruptHandler_ShowCancelled tests cancellation message.
func TestInterruptHandler_ShowCancelled(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	handler := NewInterruptHandler(&buf)

	handler.ShowCancelled(context.Background(), "Repository sync")

	output := buf.String()
	assert.Contains(t, output, "× Repository sync cancelled")
}

// TestInterruptHandler_RateLimiting tests that progress updates work without internal rate limiting.
// Rate limiting is now the caller's responsibility in the stateless design.
func TestInterruptHandler_RateLimiting(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	handler := NewInterruptHandler(&buf)
	ctx := context.Background()

	// Rapid fire progress updates - all should be written in stateless design
	for i := range 10 {
		handler.ShowProgress(ctx, i, 100, "Processing")
		time.Sleep(10 * time.Millisecond)
	}

	// In stateless design, all updates are written (no internal rate limiting)
	output := buf.String()
	lines := strings.Count(output, "\r")
	assert.Equal(t, 10, lines, "all progress updates should be written in stateless design")
}

// TestInterruptHandler_ThreadSafety tests concurrent access.
func TestInterruptHandler_ThreadSafety(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	handler := NewInterruptHandler(&buf)
	ctx := context.Background()

	// Run concurrent operations
	done := make(chan bool, 3)

	go func() {
		for range 10 {
			handler.ShowInterruptible(ctx, "Operation A")
		}

		done <- true
	}()

	go func() {
		for i := range 10 {
			handler.ShowProgress(ctx, i, 10, "Progress B")
		}

		done <- true
	}()

	go func() {
		for range 10 {
			handler.ShowCancelled(ctx, "Operation C")
		}

		done <- true
	}()

	// Wait for all goroutines
	for range 3 {
		<-done
	}

	// Should complete without panic or race conditions
	assert.NotEmpty(t, buf.String(), "should have some output")
}
