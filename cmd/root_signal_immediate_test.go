// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package cmd

import (
	"bytes"
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSignalImmediateFeedback tests that we get immediate feedback on Ctrl-C.
func TestSignalImmediateFeedback(t *testing.T) {
	t.Parallel()

	// Redirect stderr to capture output
	oldStderr := os.Stderr
	reader, writer, _ := os.Pipe()
	os.Stderr = writer

	defer func() {
		os.Stderr = oldStderr
	}()

	// Setup signal context
	ctx := SetupSignalContext(context.Background())
	require.NotNil(t, ctx)

	// Send signal to self
	proc, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)

	// Send SIGINT
	err = proc.Signal(syscall.SIGINT)
	require.NoError(t, err)

	// Give brief time for signal handler
	time.Sleep(50 * time.Millisecond)

	// Close writer and read output
	_ = writer.Close()

	var buf bytes.Buffer

	_, _ = buf.ReadFrom(reader)

	output := buf.String()

	// Should contain immediate feedback
	assert.Contains(t, output, "Interrupted",
		"Should show immediate feedback on interrupt")

	// Context should be cancelled
	select {
	case <-ctx.Done():
		// Good, context cancelled
	default:
		t.Fatal("Context should be cancelled after signal")
	}
}
