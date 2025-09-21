// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

//nolint:paralleltest // Signal tests cannot run in parallel as they send signals to the process
package cmd

import (
	"context"
	"errors"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignalExitCodes(t *testing.T) {
	// Cannot run in parallel as we're sending signals to the process
	tests := []struct {
		name     string
		signal   syscall.Signal
		expected int
	}{
		{"SIGHUP returns 129", syscall.SIGHUP, ExitSIGHUP},
		{"SIGINT returns 130", syscall.SIGINT, ExitSIGINT},
		{"SIGQUIT returns 131", syscall.SIGQUIT, ExitSIGQUIT},
		{"SIGTERM returns 143", syscall.SIGTERM, ExitSIGTERM},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Cannot run in parallel as we're sending signals to the process
			// Setup signal context
			ctx, getExitCode := SetupSignalContext(context.Background())
			require.NotNil(t, ctx)

			// Send signal to self
			proc, err := os.FindProcess(os.Getpid())
			require.NoError(t, err)
			err = proc.Signal(testCase.signal)
			require.NoError(t, err)

			// Wait for signal processing
			select {
			case <-ctx.Done():
				// Signal received and context cancelled
			case <-time.After(100 * time.Millisecond):
				t.Fatal("Signal not processed in time")
			}

			// Verify exit code
			exitCode := getExitCode()
			assert.Equal(t, testCase.expected, exitCode, "Exit code should be 128 + signal number")
		})
	}
}

func TestNoSignalReturnsZero(t *testing.T) {
	t.Parallel()

	// Setup signal context
	_, getExitCode := SetupSignalContext(context.Background())

	// Without sending any signal, exit code should be 0
	exitCode := getExitCode()
	assert.Equal(t, 0, exitCode, "No signal should return exit code 0")
}

func TestDetermineExitCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "nil error returns success",
			err:      nil,
			expected: ExitSuccess,
		},
		{
			name:     "configuration error returns 78",
			err:      errors.New("configuration file not found"),
			expected: ExitConfigError,
		},
		{
			name:     "invalid argument returns 2",
			err:      errors.New("invalid argument provided"),
			expected: ExitMisuse,
		},
		{
			name:     "401 unauthorized returns 77",
			err:      errors.New("401 Unauthorized"),
			expected: ExitNoPermission,
		},
		{
			name:     "403 forbidden returns 77",
			err:      errors.New("403 Forbidden"),
			expected: ExitNoPermission,
		},
		{
			name:     "command not found returns 127",
			err:      errors.New("executable file not found in PATH"),
			expected: ExitNotFound,
		},
		{
			name:     "git permission denied returns 126",
			err:      errors.New("permission denied: cannot execute git"),
			expected: ExitCantExecute,
		},
		{
			name:     "general permission denied returns 77",
			err:      errors.New("permission denied: cannot read file"),
			expected: ExitNoPermission,
		},
		{
			name:     "unknown error returns 1",
			err:      errors.New("something went wrong"),
			expected: ExitGeneralError,
		},
		{
			name:     "context canceled returns success",
			err:      context.Canceled,
			expected: ExitSuccess,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			exitCode := DetermineExitCode(testCase.err)
			assert.Equal(t, testCase.expected, exitCode)
		})
	}
}
